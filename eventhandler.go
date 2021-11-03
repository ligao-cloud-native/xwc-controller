package main

import "C"
import (
	"context"
	"fmt"
	"github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"strings"
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
		return
	}

	// precheck
	c.RunPrecheck(wc)

	klog.Infof("onAdd[%s] install success. cluster status: %s", wc.Name, wc.Status.Phase)

}

func (c *XWCController) OnUpdate(oldObj, newObj interface{}) {
	oldWC, isOldWC := oldObj.(*v1.WorkloadCluster)
	wc, isNewWC := newObj.(*v1.WorkloadCluster)
	if !isOldWC || !isNewWC {
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
	case v1.WorkloadClusterInstalling, v1.WorkloadClusterReducing, v1.WorkloadClusterScaling, v1.WorkloadClusterRemoving:
		// check timeout
		if metav1.Now().Unix()-wc.Status.LastTransitionTime.Unix() > c.timeout {
			klog.Warningf("cluster %s %s timeout", wc.Name, wc.Status.Action)
			wc.Status.Phase = v1.WorkloadClusterTimeout
			wc.Status.LastTransitionTime = metav1.Time{}
			_, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().UpdateStatus(context.TODO(), wc, metav1.UpdateOptions{})
			if err != nil {
				klog.Error(err)
			}
		}
	case v1.WorkloadClusterGone:
		// wc is remove
		wc.Finalizers = nil
		if len(wc.Spec.Cluster.Workers) == 0 {
			wc.Spec.Cluster.Workers = []v1.Node{}
		}
		_, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().Update(context.TODO(), wc, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err)
		}
	case v1.WorkloadClusterPrecheckFail, v1.WorkloadClusterFailed, v1.WorkloadClusterTimeout:
		// 处理集群PrecheckFail状态
		if wc.Status.Phase == v1.WorkloadClusterPrecheckFail {
			switch wc.Spec.NextAction {
			case v1.WorkloadClusterNextActionRetry:
				wc.Status.Phase = v1.WorkloadClusterPrechecking
				wc.Status.LastTransitionTime = metav1.Now()
				wc.Spec.NextAction = ""
				w, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().Update(context.TODO(), wc, metav1.UpdateOptions{})
				if err != nil {
					klog.Error(err)
				} else {
					c.RunPrecheck(w)
				}

			case v1.WorkloadClusterNextActionAbort:
				wc.Status.Phase = v1.WorkloadClusterSuccess
				wc.Status.LastTransitionTime = metav1.Now()
				wc.Spec.NextAction = ""
				if _, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().Update(context.TODO(), wc, metav1.UpdateOptions{}); err != nil {
					klog.Error(err)
				}
			}
		}

		// 集群异常状态处理
		if wc.Spec.Cluster.ForceOperation {

			if wc.DeletionTimestamp != nil {
				wc.Finalizers = nil
				if _, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().Update(context.TODO(), wc, metav1.UpdateOptions{}); err != nil {
					klog.Error(err)
				}
				return
			}

			if wc.Spec.Cluster.Workers != nil {
				wc.Status.Cluster.Workers = wc.Spec.Cluster.Workers
			} else {
				wc.Status.Cluster.Workers = nil
			}
			wc.Status.Phase = v1.WorkloadClusterSuccess
			wc.Status.Action = ""
			wc.Status.Reason = "force operation"
			w, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().UpdateStatus(context.TODO(), wc, metav1.UpdateOptions{})
			if err != nil {
				klog.Error(err)
			}

			if len(w.Spec.Cluster.Workers) == 0 {
				w.Spec.Cluster.Workers = []v1.Node{}
			}
			w.Spec.Cluster.ForceOperation = false
			if _, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().Update(context.TODO(), w, metav1.UpdateOptions{}); err != nil {
				klog.Error(err)
			}
		}
	case v1.WorkloadClusterSuccess:
		//after cluster install success，next to do scale/reduce/remove
		action := c.ensureAction(wc)
		if action == "" {
			if wc.Spec.Cluster.ForceOperation {
				wc.Spec.Cluster.ForceOperation = false
				if _, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().Update(context.TODO(), wc, metav1.UpdateOptions{}); err != nil {
					klog.Error(err)
				}
			}
			return
		}

		wc.Status.Action = action
		wc.Status.Phase = v1.WorkloadClusterPrechecking
		wc.Status.LastTransitionTime = metav1.Now()
		w, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().Update(context.TODO(), wc, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err)
		} else {
			c.RunPrecheck(w)
		}
	}

	// to done k8s
	var jobPath string
	if wc.Status.Action == v1.WorkloadClusterActionInstall && wc.Status.Phase == v1.WorkloadClusterInstalling {
		jobPath = c.xwcProvider.Install(wc)
	} else if wc.Status.Action == v1.WorkloadClusterActionScale && wc.Status.Phase == v1.WorkloadClusterScaling {
		jobPath = c.xwcProvider.Scale(wc)
	} else if wc.Status.Action == v1.WorkloadClusterActionReduce && wc.Status.Phase == v1.WorkloadClusterReducing {
		jobPath = c.xwcProvider.Reduce(wc)
	} else if wc.Status.Action == v1.WorkloadClusterActionRemove && wc.Status.Phase == v1.WorkloadClusterRemoving {
		jobPath = c.xwcProvider.Remove(wc)
	} else {
		jobPath = ""
	}

	if jobPath != "" {
		wc.Status.Action = ""
		wc.Status.LastTransitionTime = metav1.Now()
		_, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().UpdateStatus(context.TODO(), wc, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err)
		}
	}

	return
}

func (c *XWCController) ensureAction(wc *v1.WorkloadCluster) (action v1.WorkloadClusterAction) {
	if wc.DeletionTimestamp != nil {
		action = v1.WorkloadClusterActionRemove
		return
	}

	if len(wc.Spec.Cluster.Workers) > len(wc.Status.Cluster.Workers) {
		less, more := wc.Status.Cluster.Workers, wc.Spec.Cluster.Workers
		if isSubset(less, more) {
			action = v1.WorkloadClusterActionScale
		}
		return
	}

	if len(wc.Spec.Cluster.Workers) < len(wc.Status.Cluster.Workers) {
		less, more := wc.Spec.Cluster.Workers, wc.Status.Cluster.Workers
		if isSubset(less, more) {
			action = v1.WorkloadClusterActionReduce
		}
		return
	}

	return

}

func (c *XWCController) OnDelete(obj interface{}) {
	wc, ok := obj.(*v1.WorkloadCluster)
	if !ok {
		klog.Error("not a WorkloadCluster object")
		return
	}
	klog.Infof("onDelete(%v) called", wc.Name)

	c.onDelete(wc)

	c.metrics.Trigger.WithLabelValues("pwc_controller", "delete").Inc()
	c.metrics.CurrentMetrics(c.xwcCacheStore)
}

func (c *XWCController) onDelete(wc *v1.WorkloadCluster) {
	c.xwcProvider.Cleanup(wc)
	_ = c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().Delete(context.TODO(), wc.Name, metav1.DeleteOptions{})
}

func (c *XWCController) RunPrecheck(wc *v1.WorkloadCluster) {
	klog.Infof("start prechecking for %s@%v", wc.Name, wc.Status.Action)

	// get prechecked nodes
	precheckedNodes := getPrecheckNodes(wc)

	precheckResultCh := make(chan provider.PrecheckResultInterface, len(precheckedNodes))
	precheckFinishedCh := make(chan interface{})

	// precheck
	go c.xwcProvider.Precheck(precheckedNodes, precheckResultCh, precheckFinishedCh)

	go func() {
		checkMsg := ""
		checkSuccess := true
		select {
		case <-precheckFinishedCh:
			//install过程检查所有的master和worker节点，必须都precheck通过
			//scale/reduce/remve过程， 检查所有的master和操作的worker节点，至少有一个master节点和操作的worker节点precheck通过.
			isInstallAction := wc.Status.Action == v1.WorkloadClusterActionInstall
			checkedMasterErrNum := 0
			for res := range precheckResultCh {
				if !res.IsSuccess() {
					if isInstallAction {
						checkSuccess = false
						checkMsg += fmt.Sprintf("[%s]%s", res.HostInfo(), res.ResultMessage())
					} else {
						if isMasterNode(res.HostInfo(), wc.Spec.Cluster.Masters) {
							checkedMasterErrNum += 1
							checkMsg += fmt.Sprintf("[%s]%s", res.HostInfo(), res.ResultMessage())
							if checkedMasterErrNum == len(wc.Spec.Cluster.Masters) {
								checkSuccess = false
							}
						} else {
							checkSuccess = false
							checkMsg += fmt.Sprintf("[%s]%s", res.HostInfo(), res.ResultMessage())

						}
					}
				}
			}
			klog.Info(checkMsg)
		case <-time.After(defaultPrecheckTimeout):
			checkSuccess = false
			checkMsg = fmt.Sprintf("timeout on prechecking pwc %s", wc.Name)
			klog.Info(checkMsg)
		}

		//TODO: update xwc object check status
		if err := c.updatePrecheckStatus(wc, checkSuccess, checkMsg); err != nil {
			klog.Error(err)
		}
	}()

}

func (c *XWCController) updatePrecheckStatus(wc *v1.WorkloadCluster, checkSuccess bool, checkMsg string) error {
	wcName := wc.Name
	wc = c.getCachedWC(wcName)
	if wc == nil {
		return fmt.Errorf("updatePrecheckStatus error: not find pwc %s", wcName)
	}

	if !checkSuccess {
		wc.Status.Phase = v1.WorkloadClusterPrecheckFail
		wc.Status.Reason = checkMsg
	} else {
		switch wc.Status.Action {
		case v1.WorkloadClusterActionInstall:
			wc.Status.Phase = v1.WorkloadClusterInstalling
		case v1.WorkloadClusterActionScale:
			wc.Status.Phase = v1.WorkloadClusterScaling
		case v1.WorkloadClusterActionReduce:
			wc.Status.Phase = v1.WorkloadClusterReducing
		case v1.WorkloadClusterActionRemove:
			wc.Status.Phase = v1.WorkloadClusterRemoving
		}
		wc.Status.Reason = ""
	}
	wc.Status.LastTransitionTime = metav1.Now()
	_, err := c.xwcClientSet.WorkloadClustersV1().WorkloadClusters().UpdateStatus(context.TODO(), wc, metav1.UpdateOptions{})

	return err
}

func (c *XWCController) getCachedWC(name string) *v1.WorkloadCluster {
	_ = c.xwcCacheStore.Resync()
	for _, v := range c.xwcCacheStore.List() {
		if wc, ok := v.(*v1.WorkloadCluster); ok {
			if wc.Name == name {
				return wc
			}
		}
	}

	return nil
}

func getPrecheckNodes(wc *v1.WorkloadCluster) []v1.Node {
	var less, more []v1.Node

	if len(wc.Spec.Cluster.Workers) > len(wc.Status.Cluster.Workers) {
		// install cluster or scale node
		more = wc.Spec.Cluster.Workers
		less = wc.Status.Cluster.Workers
	} else if len(wc.Spec.Cluster.Workers) < len(wc.Status.Cluster.Workers) {
		// reduce node
		less = wc.Spec.Cluster.Workers
		more = wc.Status.Cluster.Workers
	} else {
		// remove cluster
		return append(wc.Spec.Cluster.Masters, wc.Spec.Cluster.Workers...)
	}

	// 获取操作的节点
	var subNodes []v1.Node
	for _, m := range more {
		found := false
		for _, l := range less {
			if l.IP == m.IP {
				found = true
				break
			}
		}
		if !found {
			subNodes = append(subNodes, m)
		}
	}

	return append(wc.Spec.Cluster.Masters, subNodes...)
}

func isMasterNode(node string, masters []v1.Node) bool {
	for _, m := range masters {
		if strings.Contains(node, m.IP) {
			return true
		}
	}

	return false
}

func isSubset(less, more []v1.Node) bool {
	if len(less) > len(more) {
		return false
	}

	for _, l := range less {
		found := false
		for _, m := range more {
			if m.IP == l.IP {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
