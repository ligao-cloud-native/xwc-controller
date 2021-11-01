package v1

type ControllerConfig struct {
	InstallerServer string
	All             *All
	Addon           *Addon
	Env             *Env
}

type All struct {
	LoadbalanceApiserver      LoadbalanceApiserver `json:"loadbalance_apiserver"`
	ServiceSubnet             string               `json:"kube_service_addresses"`
	PodSubnet                 string               `json:"kube_pods_subnet"`
	WorkloadClusterName       string               `json:"workload_cluster_name"`
	TaskCompletionCallbackUrl string               `json:"task_completion_callback_url"`
	PksControlPlane           string               `json:"pks_control_plane"`
	DockerInsecureRegistries  string               `json:"docker_insecure_registries"`
	AnsibleSSHPrivateKeyFile  string               `json:"ansible_ssh_private_key_file"`
	AnsibleSSHCommonArgs      string               `json:"ansible_ssh_common_args"`
	AnsibleSSHPass            string               `json:"ansible_ssh_pass"`
}

type Addon struct {
	IngressControllerEnabled bool   `json:"ingress-controller-enabled"`
	MonitoringEnabled        bool   `json:"monitoring_enabled"`
	LoggingEnabled           bool   `json:"monitoring_enabled"`
	MetricsServerEnabled     bool   `json:"metrics_server_enabled"`
	KubeNetworkPlugin        string `json:"kube_netrork_plugin"`
}

type Env struct {
	// installer image
	RegistryUrl        string `josn:"registry_url"`
	InstallerImageName string `josn:"installer_image_name"`

	// callback url, controller api server
	CallbackUrl string `josn:"callback_url"`

	// a k8s service vmagent, defaukt http://vmserver.kmc-nats:8000
	VmserverUrl string `josn:"vmserver_url"`
	// vmagent server url, default /vmas/VmAgentService
	VmserverPath string `josn:"vmserver_path"`
	// vmagent server secretID
	VmserverSecretid string `josn:"vmserver_secretid"`
	// vmagent server secretKey
	VmserverSecretkey string `josn:"vmserver_secretkey"`

	// file server
	PkgserverUrl string `josn:"pkgserver_url"`

	InstallerLogLevel string `josn:"installer_log_level"`
}

type LoadbalanceApiserver struct {
	Address string `json:"Address"`
	Port    string `json:"Port"`
}
