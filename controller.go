package main

import (
	"context"
	"encoding/json"
	ctlv1 "github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"
	xwcclient "github.com/ligao-cloud-native/kubemc/pkg/client/clientset/versioned"
	controllercfg "github.com/ligao-cloud-native/xwc-controller/config"
	config "github.com/ligao-cloud-native/xwc-controller/pkg/componentconfig/controller/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/metrics"
	"github.com/ligao-cloud-native/xwc-controller/provider"
	"github.com/ligao-cloud-native/xwc-controller/provider/xwcagent"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"time"
)

const (
	xwcFinalizer                string = "finalizer.xwc.kubemc.io"
	defaultK8sVersion                  = "1.13.11"
	defaultDockerRuntimeVersion        = "18.09"
	defaultPodCird                     = "192.168.0.0/16"
	defaultServiceCird                 = "10.233.0.0/18"
	defaultServiceDomain               = "cluster.local"

	defaultPrecheckTimeout = 60 * time.Second
)

// Controller is the controller implementation for Foo resources
type XWCController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeClientSet kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	xwcClientSet xwcclient.Interface
	// install k8s provider
	xwcProvider provider.Interface
	// xwc cache
	xwcCacheStore cache.Store
	// k8s an cni version compatible
	compatibleVersion map[string][]string
	// runtime type and version
	// metrics
	metrics *metrics.Metrics
	// install timeout
	timeout int64
}

func NewController(cfg *config.ControllerConfig, metric *metrics.Metrics, timeout int64) *XWCController {
	controllercfg.InitConfigure(cfg)

	kubeConfig, err := buildConfig()
	if err != nil {
		klog.Errorf("Failed to build config, err: %v", err)
		os.Exit(1)
	}
	kubeClient := kubernetes.NewForConfigOrDie(kubeConfig)
	xwcClient := xwcclient.NewForConfigOrDie(kubeConfig)

	var installer provider.Interface
	switch provider.ProviderType(installProvider) {
	case provider.XWCAgentProvider:
		installer = xwcagent.NewXwcAgentProvider(installProvider, kubeClient, timeout)
	case provider.RPCProvider:
	}

	return &XWCController{
		kubeClientSet:     kubeClient,
		xwcClientSet:      xwcClient,
		xwcProvider:       installer,
		compatibleVersion: initCompatibleVersion(),
		metrics:           metric,
		timeout:           timeout,
	}

}

func (c *XWCController) Run() {
	// update controller configmap
	go c.updateConfigMap()
	c.startController()
}

func (c *XWCController) updateConfigMap() {
	envJson, err := json.Marshal(controllercfg.Config.ControllerConfig.Env)
	if err != nil {
		klog.Errorf("Marshal env error, %v", err)
		return
	}

	cms, err := c.kubeClientSet.CoreV1().ConfigMaps("xwc-system").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("List system configmap error, %v", err)
		return
	}

	for _, cm := range cms.Items {
		if _, ok := cm.Data["env.json"]; !ok {
			cm.Data["env.json"] = string(envJson)
			_, err := c.kubeClientSet.CoreV1().ConfigMaps("xwc-system").Update(context.TODO(), &cm, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("update configmap %s error, %v", cm.Name, err)
			} else {
				klog.Infof("update configmap %s success", cm.Name)
			}
		}
	}

}

func (c *XWCController) startController() {
	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (result runtime.Object, err error) {
				return c.xwcClientSet.WorkloadClustersV1().Foos("").List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return c.xwcClientSet.SamplecontrollerV1alpha1().Foos("").Watch(context.TODO(), options)
			},
		},
		&ctlv1.WorkloadCluster{},
		1*time.Minute,
		c)
	go controller.Run(wait.NeverStop)
	c.xwcCacheStore = store
}

// build Config from flags
func buildConfig() (conf *rest.Config, err error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		return nil, err
	}
	kubeConfig.ContentType = "application/json"

	return kubeConfig, nil
}

// initCompatibleVersion init k8s and calico  cni version compatible
func initCompatibleVersion() map[string][]string {
	return map[string][]string{
		"v1.13.11": {"v3.5.8"},
		"v1.15.12": {"v3.13.4", "v3.14.1"},
		"v1.16.12": {"v3.13.4", "v3.14.1", "v3.15.1"},
		"v1.17.9":  {"v3.13.4", "v3.14.1", "v3.15.1"},
		"v1.18.6":  {"v3.14.1", "v3.15.1"},
		"v1.19.2":  {"v3.14.1", "v3.15.1"},
	}

}
