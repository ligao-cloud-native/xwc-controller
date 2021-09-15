package xwcagent

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/provider"
	"strings"

	//ctlconfig"github.com/ligao-cloud-native/xwc-controller/config"
	//"github.com/ligao-cloud-native/xwc-controller/provider/xwcagent/agentclient"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type XwcAgentProvider struct {
	prechecker *PreChecker
	//Installer
	//agentClient *agentclient.Client
}

func NewXwcAgentProvider() *XwcAgentProvider {

	return &XwcAgentProvider{
		//agentClient: initAgentClient(),
		prechecker: NewAgentPreChecker(),
	}
}

func (a *XwcAgentProvider) Precheck(wc *v1.WorkloadCluster, resultCh chan<- provider.PrecheckResultInterface, finished chan<- interface{}) {
	masters := wc.Spec.Cluster.Masters
	workers := wc.Spec.Cluster.Workers

	//TODO: complement master and worker node
	masters = append(masters, workers...)
	//并发preheck每个节点

	var wg sync.WaitGroup
	for _, node := range masters {
		wg.Add(1)
		go func(host v1.Node, resultCh chan<- provider.PrecheckResultInterface) {
			defer wg.Done()
			res := &PrecheckResult{Host: host.IP, NodeID: host.NodeID}

			exitValue, output := a.prechecker.PreCheck(host.IP, host.NodeID)
			res.Success = exitValue == 0
			res.Result = output
			if exitValue == 0 {
				isCostTime, err := isPrecheckTimeout(output, host.IP)
				if isCostTime {

					res.Success = false
				}
				if err != nil {
					res.Result = err.Error()
				}
			}

			resultCh <- res

		}(node, resultCh)
	}
	wg.Wait()
	close(resultCh)
	finished <- true
}

func (a *XwcAgentProvider) Install() {}

func (a *XwcAgentProvider) preCheck(uuid, ip, command string) {

}

func isPrecheckTimeout(cmdOutput, ip string) (bool, error) {
	scanner := bufio.NewScanner(bytes.NewReader([]byte(cmdOutput)))
	for scanner.Scan() {
		t := scanner.Text()
		if strings.HasPrefix(t, "real") {
			l := strings.Split(t, "\t")
			if len(l) == 2 {
				timeCost, err := time.ParseDuration(l[1])
				if err != nil {
					klog.Error(err)
					return false, err
				}
				klog.Infof("precheck command running time exceeded %v on remote host %v", timeCost, ip)
				if timeCost > time.Second {
					return true, fmt.Errorf("precheck command running time exceeded %v", timeCost)
				}
				return false, nil
			}
			break
		}
	}
	klog.Warning("found noting from output %v", cmdOutput)
	return false, nil
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
