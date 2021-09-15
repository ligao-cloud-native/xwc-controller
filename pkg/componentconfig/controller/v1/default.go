package v1

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
)

func NewDefaultControllerConfig() *ControllerConfig {
	return &ControllerConfig{
		InstallerServer: "",
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

	//TODO: set other env
}
