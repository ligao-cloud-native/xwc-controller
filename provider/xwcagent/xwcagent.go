package xwcagent

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/provider"
	"k8s.io/client-go/kubernetes"
	"strings"
	//ctlconfig"github.com/ligao-cloud-native/xwc-controller/config"
	//"github.com/ligao-cloud-native/xwc-controller/provider/xwcagent/agentclient"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type XwcAgentProvider struct {
	name       string
	prechecker provider.Prechecker
	installer  provider.Installer
	//agentClient *agentclient.Client
}

func NewXwcAgentProvider(name string, kubeClient kubernetes.Interface, timeout int64) *XwcAgentProvider {

	return &XwcAgentProvider{
		name: name,
		//agentClient: initAgentClient(),
		prechecker: NewAgentPreChecker(),
		installer:  NewInstaller(name, kubeClient, timeout),
	}
}

func (a *XwcAgentProvider) Precheck(nodes []v1.Node, resultCh chan<- provider.PrecheckResultInterface, finished chan<- interface{}) {
	var wg sync.WaitGroup
	for _, node := range nodes {
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

func (a *XwcAgentProvider) Install(wc *v1.WorkloadCluster) (jobPath string) {
	jobPath, err := a.installer.Install(wc)
	if err != nil {
		klog.Error(err.Error())
		return
	}

	return jobPath

}

func (a *XwcAgentProvider) Scale(wc *v1.WorkloadCluster) (jobPath string) {
	jobPath, err := a.installer.Scale(wc)
	if err != nil {
		klog.Error(err.Error())
		return
	}

	return jobPath
}

func (a *XwcAgentProvider) Reduce(wc *v1.WorkloadCluster) (jobPath string) {
	jobPath, err := a.installer.Reduce(wc)
	if err != nil {
		klog.Error(err.Error())
		return
	}

	return jobPath
}

func (a *XwcAgentProvider) Remove(wc *v1.WorkloadCluster) (jobPath string) {
	jobPath, err := a.installer.Remove(wc)
	if err != nil {
		klog.Error(err.Error())
		return
	}

	return jobPath
}

func (a *XwcAgentProvider) Cleanup(wc *v1.WorkloadCluster) {
	a.installer.Cleanup(wc)
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
