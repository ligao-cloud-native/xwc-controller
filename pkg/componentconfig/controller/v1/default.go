package v1

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
)

func NewDefaultControllerConfig() *ControllerConfig {
	return &ControllerConfig{
		Env: &Env{
			PkgserverUrl: "pcr-pub.paic.com.cn",
		},
		All: &All{
			ServiceSubnet:            "10.223.0.0/18",
			PodSubnet:                "192.168.0.0/24",
			AnsibleSSHPrivateKeyFile: "/opt/mycluster/private.key",
			AnsibleSSHCommonArgs:     "-o StrictHostKeyChecking=no",
			LoadbalanceApiserver: LoadbalanceApiserver{
				Port: "6443",
			},
		},
		Addon: &Addon{
			IngressControllerEnabled: true,
			MonitoringEnabled:        true,
			LoggingEnabled:           true,
			MetricsServerEnabled:     false,
			KubeNetworkPlugin:        "calico",
		},
	}

}

func (c *ControllerConfig) LoadConfigFromFileOrEnv(filename string, env bool) error {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		klog.Errorf("Failed to read configfile %s: %v", filename, err)
		return err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		klog.Errorf("Failed to unmarshal configfile %s: %v", filename, err)
		return err
	}

	if env {
		c.updateConfigFromEnv()
	}

	return nil

}

func (c *ControllerConfig) updateConfigFromEnv() {
	registryUrl := os.Getenv("REGISTRY_URL")
	if registryUrl != "" {
		c.Env.RegistryUrl = registryUrl
	}

	callbackUrl := os.Getenv("CALLBACK_URL")
	if registryUrl != "" {
		c.Env.CallbackUrl = callbackUrl
	}

	//TODO: set other env
}
