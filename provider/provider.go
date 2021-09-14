package provider

import "github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"

type Interface interface {
	Precheck(wc *v1.WorkloadCluster)
	Install()
}
