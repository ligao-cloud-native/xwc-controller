package v1

type ControllerConfig struct {
	InstallerServer string
	Base            *Base
	Addon           *Addon
	Env             *Env
}

type Base struct {
	ServiceSubnet string
	PodSubnet     string
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
	InstallerImageName string
}
