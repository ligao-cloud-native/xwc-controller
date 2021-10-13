package utils

import (
	xwcclient "github.com/ligao-cloud-native/kubemc/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// KubeClient from config
func KubeClient(kubeConfig string) (*kubernetes.Clientset, error) {
	config, err := KubeConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func XwcClient(kubeConfig string) (*xwcclient.Clientset, error) {
	config, err := KubeConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return xwcclient.NewForConfigOrDie(config), nil
}

func KubeConfig(kubeConfig string) (*rest.Config, error) {
	if kubeConfig == "" {
		klog.Info("using in-cluster configuration")
		return rest.InClusterConfig()
	} else {
		klog.Infof("using configuration from %s", kubeConfig)
		conf, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return nil, err
		}
		conf.QPS = 100
		conf.Burst = 200
		conf.ContentType = "application/vnd.kubernetes.protobuf"
		conf.TLSClientConfig = rest.TLSClientConfig{Insecure: true}

		return conf, nil
	}

}
