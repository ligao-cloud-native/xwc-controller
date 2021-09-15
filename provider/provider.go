package provider

import (
	"github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/provider"
)

type Interface interface {
	Precheck(wc *v1.WorkloadCluster, resultCh chan<- provider.PrecheckResultInterface, finished chan<- interface{})
	Install()
}
