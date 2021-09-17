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

var operateStage = map[string]string{
	fmt.Sprintf("%v%v", v1.WorkloadClusterInstalling, v1.WorkloadClusterActionInstall): "install",
	fmt.Sprintf("%v%v", v1.WorkloadClusterRemoving, v1.WorkloadClusterActionRemove):    "remove",
	fmt.Sprintf("%v%v", v1.WorkloadClusterScaling, v1.WorkloadClusterActionScale):      "scale",
	fmt.Sprintf("%v%v", v1.WorkloadClusterReducing, v1.WorkloadClusterActionReduce):    "reduce",
}

type XwcAgentProvider struct {
	name       string
	prechecker provider.Prechecker
	installer  provider.Installer
	//agentClient *agentclient.Client
}

func NewXwcAgentProvider(name string, kubeClient kubernetes.Interface) *XwcAgentProvider {

	return &XwcAgentProvider{
		name: name,
		//agentClient: initAgentClient(),
		prechecker: NewAgentPreChecker(),
		installer:  NewInstaller(name, kubeClient),
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

func (a *XwcAgentProvider) Install(wc *v1.WorkloadCluster) {
	switch operateStage[fmt.Sprintf("%v%v", wc.Status.Phase, wc.Status.Action)] {
	case "install":
		a.installer.Install(wc)
	case "remove":
		a.installer.Reset()
	case "scale":
		a.installer.Scale()
	case "reduce":
		a.installer.Reduce()
	default:
		klog.Warningf("do nothing.")
	}
}

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
