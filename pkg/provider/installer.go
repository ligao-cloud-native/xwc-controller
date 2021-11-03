package provider

import v1 "github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"

type Installer interface {
	Install(wc *v1.WorkloadCluster) (jobPath string, err error)
	Remove(wc *v1.WorkloadCluster) (jobPath string, err error)
	Cleanup(wc *v1.WorkloadCluster)
	Scale(wc *v1.WorkloadCluster) (jobPath string, err error)
	Reduce(wc *v1.WorkloadCluster) (jobPath string, err error)
}
