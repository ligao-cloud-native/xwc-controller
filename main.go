package main

import (
	"context"
	"flag"
	config "github.com/ligao-cloud-native/xwc-controller/pkg/componentconfig/controller/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/leaderelection"
	"k8s.io/klog/v2"
)

var (
	masterURL        string
	kubeConfig       string
	installProvider  string
	controllerConfig string
)

func main() {
	cfg := config.NewDefaultControllerConfig()
	err := cfg.LoadConfigFromFileOrEnv(controllerConfig, true)
	if err != nil {
		klog.Fatal(err)
	}

	// set up signals so we handle the first shutdown signal gracefully
	//stopCh := signals.SetupSignalHandler()

	startController := func(ctx context.Context) {
		c := NewController(cfg)
		go func() {
			c.Run()
		}()
	}
	leaderelection.Run(startController)
}

func init() {
	klog.InitFlags(nil)

	flag.StringVar(&kubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	flag.StringVar(&installProvider, "provider", "xwcagent", "agent to install k8s cluster")
	flag.StringVar(&controllerConfig, "config", "/etc/xwc-controller/config.json", "controller config file")

	flag.Parse()
}
