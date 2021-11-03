package provider

import (
	"github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/provider"
)

type Interface interface {
	Precheck(nodes []v1.Node, resultCh chan<- provider.PrecheckResultInterface, finished chan<- interface{})
	// install cluster
	Install(wc *v1.WorkloadCluster) (jobPath string)
	// remove cluster， and clean pwc resource
	Remove(wc *v1.WorkloadCluster) (jobPath string)
	Cleanup(wc *v1.WorkloadCluster)
	// scale node
	Scale(wc *v1.WorkloadCluster) (jobPath string)
	// reduce node
	Reduce(wc *v1.WorkloadCluster) (jobPath string)
}
