package xwcagent

import (
	"encoding/json"
	"fmt"
	ctlconfig "github.com/ligao-cloud-native/xwc-controller/config"
	"github.com/ligao-cloud-native/xwc-controller/provider/xwcagent/agentclient"
	"k8s.io/klog/v2"
	"time"
)

const (
	PrecheckCommand = "time /usr/sbin/ip addr |grep "
	PrecheckTimeout = 60 * time.Second
)

type PreChecker struct {
	agentClient   *agentclient.Client
	command       string
	timeout       time.Duration
	retryTimes    int
	retryWaitTime time.Duration
}

func NewAgentPreChecker() *PreChecker {
	return &PreChecker{
		agentClient:   initAgentClient(),
		timeout:       PrecheckTimeout,
		retryTimes:    3,
		retryWaitTime: 2 * time.Second,
	}
}

func (c *PreChecker) PreCheck(ip, uuid string) (int, string) {
	cmd := PrecheckCommand + ip
	resp, err := c.agentClient.SendCmd(uuid, cmd)
	if err != nil {
		return -1, err.Error()
	}
	if resp.Data.ReqId == "" {
		return -1, "no request id"
	}

	for i := 1; i <= c.retryTimes; {
		rec, err := c.agentClient.ReceiveCmd(resp.Data.ReqId)
		if err == nil && (rec.Data.State == 1 || rec.Data.State == -1) {
			klog.Infof("node command response: [%v %v][%v][%v]", ip, uuid, cmd, rec.Data.Result)
			exitValue := rec.Data.Result.ExitValue
			exitString := rec.Data.Result.Error
			if exitString == "" {
				if rec.Data.Result.OutPut == "" {
					exitString = fmt.Sprintf("vmserver return empty output/error, but exitValue=%v", exitValue)
				} else {
					exitString = rec.Data.Result.OutPut
				}
			}
			return exitValue, exitString
		}

		time.Sleep(c.retryWaitTime)
		i++
	}

	return -1, fmt.Sprintf("vmagent no response for command %v, try %v times", cmd, c.retryTimes)

}

func initAgentClient() *agentclient.Client {
	vmserver := ctlconfig.Config.ControllerConfig.Env.XwcServer + ctlconfig.Config.ControllerConfig.Env.XwcServerUrl
	cre := agentclient.Credential{
		SecretId:  ctlconfig.Config.SecretId,
		SecretKey: ctlconfig.Config.SecretKey,
	}

	client, err := agentclient.NewClient(vmserver, cre)
	if err != nil {
		klog.Error(err)
		return nil
	}

	return client

}
