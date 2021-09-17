package provider

import v1 "github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"

type Installer interface {
	Install(wc *v1.WorkloadCluster)
	Reset()
	Scale()
	Reduce()
}
