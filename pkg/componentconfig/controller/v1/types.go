package v1

type ControllerConfig struct {
	InstallerServer string
	All             *All
	Addon           *Addon
	Env             *Env
}

type All struct {
	LoadbalanceApiserver     LoadbalanceApiserver `json:"loadbalancer_apiserver"`
	ServiceSubnet            string               `json:"kube_service_addresses"`
	PodSubnet                string               `json:"kube_pods_subnet"`
	WorkloadClusterName      string               `json:"workload_cluster_name"`
	AnsibleSSHPrivateKeyFile string               `json:"ansible_ssh_private_key_file"`
	AnsibleSSHCommonArgs     string               `json:"ansible_ssh_common_args"`
	AnsibleSSHPass           string               `json:"ansible_ssh_pass"`
}

type Addon struct {
	IngressControllerEnable bool
	MonitoringEnable        bool
	LoggingEnable           bool
}

type Env struct {
	RegistryUrl string

	// a k8s service vmagent, defaukt http://vmserver.kmc-nats:8000
	XwcServer string
	// vmagent server url, default /vmas/VmAgentService
	XwcServerUrl string
	// vmagent server secretID
	XwcServerSecretId string
	// vmagent server secretKey
	XwcServerSecretKey string

	FileServer         string
	InstallerImageName string `josn:"installer_image_name"`
	InstallerLogLevel  string `josn:"installer_log_level"`
}

type LoadbalanceApiserver struct {
	Address string `json:"Address"`
	Port    string `json:"Port"`
}
