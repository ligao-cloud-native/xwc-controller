package xwcagent

import (
	"github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"
	//ctlconfig"github.com/ligao-cloud-native/xwc-controller/config"
	//"github.com/ligao-cloud-native/xwc-controller/provider/xwcagent/agentclient"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type XwcAgentProvider struct {
	prechecker *PreChecker
	Installer
	//agentClient *agentclient.Client
}

func NewXwcAgentProvider() *XwcAgentProvider {

	return &XwcAgentProvider{
		//agentClient: initAgentClient(),
		prechecker: NewAgentPreChecker(),
	}
}

func (a *XwcAgentProvider) Precheck(wc *v1.WorkloadCluster) {
	masters := wc.Spec.Cluster.Masters
	workers := wc.Spec.Cluster.Workers

	//TODO: complement master and worker node
	masters = append(masters, workers...)
	//并发preheck每个节点

	var wg sync.WaitGroup
	for _, node := range masters {
		wg.Add(1)
		go func() {
			defer wg.Done()
			exitValue, output := a.prechecker.PreCheck(node.IP, node.NodeID)

		}()
	}
	wg.Wait()

}

func (a *XwcAgentProvider) Install() {}

func (a *XwcAgentProvider) preCheck(uuid, ip, command string) {

}

//func initAgentClient() *agentclient.Client {
//	vmserver := ctlconfig.Config.ControllerConfig.Env.XwcServer + ctlconfig.Config.ControllerConfig.Env.XwcServerUrl
//	cre := agentclient.Credential{
//		SecretId: ctlconfig.Config.SecretId,
//		SecretKey: ctlconfig.Config.SecretKey,
//	}
//
//	client, err := agentclient.NewClient(vmserver, cre)
//	if err != nil {
//		klog.Error(err)
//		return nil
//	}
//
//	return client
//
//
//}
