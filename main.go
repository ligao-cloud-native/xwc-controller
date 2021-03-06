package main

import (
	"context"
	"flag"
	config "github.com/ligao-cloud-native/xwc-controller/pkg/componentconfig/controller/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/leaderelection"
	"github.com/ligao-cloud-native/xwc-controller/pkg/metrics"
	"k8s.io/klog/v2"
)

var (
	masterURL        string
	kubeConfig       string
	installProvider  string
	controllerConfig string
	timeout          int64
	version          = "unknow"
)

func main() {
	cfg := config.NewDefaultControllerConfig()
	err := cfg.LoadConfigFromFileOrEnv(controllerConfig, true)
	if err != nil {
		klog.Fatal(err)
	}

	metric := metrics.NewMetrics()

	// start http server
	go StartHttpServer(&Options{Metric: metric})

	// set up signals so we handle the first shutdown signal gracefully
	//stopCh := signals.SetupSignalHandler()

	startController := func(ctx context.Context) {
		c := NewController(cfg, metric, timeout)
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

	flag.StringVar(&controllerConfig, "config", "/etc/xwc-controller/config.json", "controller config file")
	flag.StringVar(&installProvider, "provider", "xwcagent", "agent to install k8s cluster")
	flag.Int64Var(&timeout, "timeout", 480, "xwc install timeout")

	flag.Parse()

	klog.V(1).Infoln("start pwc-controller version: ", version)
	klog.V(1).Infoln("precheck provider: ", installProvider)
	klog.V(1).Infoln("install timeout: ", timeout)

}
