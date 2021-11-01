package main

import (
	"context"
	"fmt"
	"github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"time"
)

// OnAdd handle "new" or "" status
func (c *XWCController) OnAdd(obj interface{}) {
	wc, ok := obj.(*v1.WorkloadCluster)
	if !ok {
		klog.Error("not a WorkloadCluster object")
		return
	}

	if wc.Status.Phase == v1.WorkloadClusterNew || wc.Status.Phase == "" {
		c.onAdd(wc)
	}
	klog.V(2).Infof("OnAdd(%s) called: status is %s", wc.Name, wc.Status.Phase)

	//TODO: metric
	c.metrics.CurrentMetrics(c.xwcCacheStore)

}

func (c *XWCController) onAdd(wc *v1.WorkloadCluster) {
	if wc.GetFinalizers() == nil {
		wc.SetFinalizers([]string{xwcFinalizer})
	}

	// check k8s version
	if _, ok := c.compatibleVersion[wc.Spec.Cluster.Version]; !ok {
		klog.Warningf("no supported k8s version, use default version %s", defaultK8sVersion)
		wc.Spec.Cluster.Version = defaultK8sVersion
	}

	// check cni type and version
	switch wc.Spec.Cluster.Network.Type {
	case v1.NetworkTypeCalico, v1.NetworkTypeCalicoTypha:
		ifPairCalicoVersion := false
		for _, calicoV := range c.compatibleVersion[wc.Spec.Cluster.Version] {
			if wc.Spec.Cluster.Network.Version == calicoV {
				ifPairCalicoVersion = true
				break
			}
		}
		if !ifPairCalicoVersion {
			klog.Warningf("no supported calico version, use default version")
			wc.Spec.Cluster.Network.Version = c.compatibleVersion[wc.Spec.Cluster.Version][0]
		}
	case v1.NetworkTypeFlannel:
		klog.Warningf("only support flannel cni version 0.12.0")
		wc.Spec.Cluster.Network.Version = "v0.12.0"

	default:
		klog.Warningf("no supported cni type, use default cni %v", v1.NetworkTypeCalico)
		wc.Spec.Cluster.Network.Type = v1.NetworkTypeCalico
		wc.Spec.Cluster.Network.Version = c.compatibleVersion[wc.Spec.Cluster.Version][0]
	}

	if wc.Spec.Cluster.Network.PodCIDR == "" {

		wc.Spec.Cluster.Network.PodCIDR = defaultPodCird
	}
	if wc.Spec.Cluster.Network.ServiceCIDR == "" {
		wc.Spec.Cluster.Network.PodCIDR = defaultServiceCird
	}
	if wc.Spec.Cluster.Network.ServiceDomain == "" {
		wc.Spec.Cluster.Network.PodCIDR = defaultServiceDomain
	}

	// check kube-proxy mode
	if wc.Spec.Cluster.Network.KubeProxyMode != v1.KubeProxyModeIptables &&
		wc.Spec.Cluster.Network.KubeProxyMode != v1.KubeProxyModeIpvs {
		klog.Warningf("not supported kube-proxy mode, use default %v", v1.KubeProxyModeIptables)
		wc.Spec.Cluster.Network.KubeProxyMode = v1.KubeProxyModeIptables
	}

	// check docker runtime
	wc.Spec.Cluster.Runtime = v1.Runtime{
		Type:    v1.RuntimeTypeDocker,
		Version: defaultDockerRuntimeVersion,
	}

	// update workerloadcluster object
	var err error
	wc, err = c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().Update(context.TODO(), wc, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		return
	}

	wc.Status.Phase = v1.WorkloadClusterPrechecking
	wc.Status.Action = v1.WorkloadClusterActionInstall
	wc, err = c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().UpdateStatus(context.TODO(), wc, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("pwc %s status update error.", wc.Name)
	}
	// install
	c.startInstaller(wc)

	klog.Infof("onAdd[%s] install success. cluster status: %s", wc.Name, wc.Status.Phase)

}

func (c *XWCController) startInstaller(wc *v1.WorkloadCluster) {
	precheckResultCh := make(chan provider.PrecheckResultInterface)
	precheckFinishedCh := make(chan interface{})

	// precheck
	go c.xwcProvider.Precheck(wc, precheckResultCh, precheckFinishedCh)

	// precheck通过后安装,
	go func() {
		checkResult := ""
		select {
		case <-precheckFinishedCh:
			for res := range precheckResultCh {
				if !res.IsSuccess() {
					checkResult += fmt.Sprintf("[%s]%s", res.HostInfo(), res.ResultMessage())
				}
			}
		case <-time.After(defaultPrecheckTimeout):
		}

		//TODO: update xwc object check status
		if err := c.updatePrecheckStatus(wc, checkResult); err != nil {
			klog.Error(err)
		}
	}()

}

func (c *XWCController) updatePrecheckStatus(wc *v1.WorkloadCluster, checkResult string) error {
	c.xwcCacheStore.List()

	wc.Status.Reason = checkResult

	return nil
}

func (c *XWCController) OnUpdate(oldObj, newObj interface{}) {
	oldWC, isOldWC := oldObj.(*v1.WorkloadCluster)
	wc, isNewWC := newObj.(*v1.WorkloadCluster)
	if !isOldWC || isNewWC {
		klog.Error("not a WorkloadCluster object")
		return
	}

	if oldWC.Status.Phase != wc.Status.Phase {
		klog.Infof("OnUpdate[%s]: old state: %v, new state %v",
			oldWC.Name, oldWC.Status.Phase, wc.Status.Phase)
	}

	c.onUpdate(wc)

	c.metrics.CurrentMetrics(c.xwcCacheStore)

}

func (c *XWCController) onUpdate(wc *v1.WorkloadCluster) {
	// to update wc object
	switch wc.Status.Phase {
	case v1.WorkloadClusterPrechecking:
		klog.Infof("OnUpdate[%s] state: %v", wc.Name, wc.Status.Phase)
		return
	case v1.WorkloadClusterInstalling, v1.WorkloadClusterReducing, v1.WorkloadClusterScaling, v1.WorkloadClusterRemoving:
		// check timeout
		if metav1.Now().Unix()-wc.Status.LastTransitionTime.Unix() > c.timeout {
			wc.Status.Phase = v1.WorkloadClusterTimeout
			wc.Status.LastTransitionTime = metav1.Time{}
			// TODO: update wc, if error return
		}
	case v1.WorkloadClusterGone:
		// wc is remove
		wc.Finalizers = nil
		if len(wc.Spec.Cluster.Workers) == 0 {
			wc.Spec.Cluster.Workers = []v1.Node{}
		}
		// TODO: update wc, if error return
	//case v1.WorkloadClusterPrecheckFail, v1.WorkloadClusterFailed, v1.WorkloadClusterTimeout:
	//	return
	case v1.WorkloadClusterPrecheckFail:

		return
	case v1.WorkloadClusterSuccess:

		return
	}

	// to done k8s
	c.xwcProvider.Install(wc)
	// TODO: update wc

}

func (c *XWCController) OnDelete(obj interface{}) {
	//TODO: delete

	c.metrics.Trigger.WithLabelValues("pwc_controller", "delete").Inc()
	c.metrics.CurrentMetrics(c.xwcCacheStore)
}
