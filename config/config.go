package config

import (
	v1 "github.com/ligao-cloud-native/xwc-controller/pkg/componentconfig/controller/v1"
	"io/ioutil"
	"sync"
)

var Config Configure
var once sync.Once

type Configure struct {
	ControllerConfig *v1.ControllerConfig
	SSHPrivateKey    []byte
	SSHPublicKey     []byte
	SecretId         string
	SecretKey        string
}

func InitConfigure(cfg *v1.ControllerConfig) {
	once.Do(func() {
		Config = Configure{
			ControllerConfig: cfg,
		}

		//TODO: 如果没ssh private.key和public.key文件,则自动生成
		privateKey, err := ioutil.ReadFile("/etc/xwc-controller/private.key")
		if err == nil {
			Config.SSHPrivateKey = privateKey
		} else {
		}

		publicKey, err := ioutil.ReadFile("/etc/xwc-controller/public.key")
		if err == nil {
			Config.SSHPublicKey = publicKey
		} else {
		}
	})
}
