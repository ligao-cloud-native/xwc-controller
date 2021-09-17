package xwcagent

import (
	"encoding/json"
	"fmt"
	v1 "github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"
	ctlcfg "github.com/ligao-cloud-native/xwc-controller/config"
	config "github.com/ligao-cloud-native/xwc-controller/pkg/componentconfig/controller/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var (
	pksInstallerNamespace string = "xwc-installer"
	defaultBackoffLimit   int32  = 0
	defaultCompletions    int32  = 1
	defaultParallelism    int32  = 1
)

type JobType string

const (
	JobTypeInstall JobType = "install"
	JobTypeReset   JobType = "reset"
	JobTypeScale   JobType = "scale"
	JobTypeReduce  JobType = "reduce"
)

type Installer struct {
	clientSet    kubernetes.Interface
	config       *ctlcfg.Configure
	providerName string
}

func NewInstaller(name string, clientSet kubernetes.Interface) *Installer {
	return &Installer{
		clientSet:    clientSet,
		config:       ctlcfg.Config,
		providerName: name,
	}

}

func (i *Installer) Install(wc *v1.WorkloadCluster) {
	if err := i.createJobConfigMap(wc); err != nil {
		klog.Errorf("creater configmap for xwc %s error: %s", wc.Name, err.Error())
		return
	}

	if err := i.createJobSecret(wc); err != nil {
		klog.Errorf("creater secret for xwc %s error: %s", wc.Name, err.Error())
		return
	}

	i.createJob(wc)
}

func (i *Installer) Reset() {}

func (i *Installer) Scale() {}

func (i *Installer) Reduce() {}

func (i *Installer) createJobConfigMap(wc *v1.WorkloadCluster) error {
	if cm, err := utils.GetConfigMap(i.clientSet, wc.Name); cm != nil && err == nil {
		klog.Infof("ConfigMap for xwc %s is existed.", wc.Name)
		return nil
	}

	cm := i.buildConfigMap(wc)

	if cm == nil {
		return fmt.Errorf("build configMap from wc object %s error", wc.Name)
	}

	if err := utils.CreateConfigMap(i.clientSet, cm); err != nil {
		return fmt.Errorf("create configMap for wc %s error: %s", wc.Name, err.Error())
	}

	return nil
}

func (i *Installer) createJobSecret(wc *v1.WorkloadCluster) error {
	if secret, err := utils.GetSecret(i.clientSet, wc.Name); secret != nil && err == nil {
		klog.Infof("Secret for xwc %s is existed.", wc.Name)
		return nil
	}

	secret := i.buildSecret(wc)
	if secret == nil {
		return fmt.Errorf("build secret from wc object %s error", wc.Name)
	}

	if err := utils.CreateSecret(i.clientSet, secret); err != nil {
		return fmt.Errorf("create secret for wc %s error: %s", wc.Name, err.Error())
	}

	return nil

}

func (i *Installer) createJob(wc *v1.WorkloadCluster) error {
	job := i.buildJob(wc.Name, JobTypeInstall, []string{"k8s", "install"})
	if job == nil {
		return fmt.Errorf("build job from wc object %s error", wc.Name)
	}

	if err := utils.CreateJob(i.clientSet, job); err != nil {
		return fmt.Errorf("create job for wc %s error: %s", wc.Name, err.Error())
	}

	return nil
}

func (i *Installer) buildConfigMap(wc *v1.WorkloadCluster) *corev1.ConfigMap {
	hostData := getHostsYaml()
	if len(hostData) == 0 {
		return nil
	}
	nodeData := getNodesJson()
	if len(nodeData) == 0 {
		return nil
	}

	cm := corev1.ConfigMap{}
	cm.Name = wc.Name
	cm.Labels = map[string]string{
		"app": wc.Name,
	}
	cm.Data = map[string]string{
		"hosts.yaml": string(hostData),
		"nodes.json": string(nodeData),
	}

	for k, v := range i.getOtherData(wc) {
		cm.Data[k] = string(v)
	}

	return &cm

}

func (i *Installer) buildSecret(wc *v1.WorkloadCluster) *corev1.Secret {
	data := make(map[string][]byte)

	if len(i.config.SSHPublicKey) > 0 {
		data["private.key"] = i.config.SSHPublicKey
	}
	if len(i.config.SSHPrivateKey) > 0 {
		data["public.key"] = i.config.SSHPublicKey
	}

	secret := corev1.Secret{}
	secret.Name = wc.Name
	secret.Type = corev1.SecretTypeOpaque
	secret.Labels = map[string]string{
		"app": wc.Name,
	}
	secret.Data = data

	return &secret

}

func (i *Installer) buildJob(wcName string, jobType JobType, jobCmd []string) *batchv1.Job {
	job := batchv1.Job{}

	labels := map[string]string{
		"app": wcName,
	}

	// job meta
	job.Name = fmt.Sprintf("%s-%s-%s", wcName, jobType, utils.RandomStr(4))
	job.Namespace = pksInstallerNamespace
	job.Labels = labels

	// job spec
	job.Spec.BackoffLimit = &defaultBackoffLimit
	job.Spec.Completions = &defaultCompletions
	job.Spec.Parallelism = &defaultParallelism

	// job pod meta
	job.Spec.Template.Name = wcName
	job.Spec.Template.Labels = labels

	// job pod spec
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	job.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:            wcName,
			Image:           i.config.ControllerConfig.Env.RegistryUrl + "/" + i.config.ControllerConfig.Env.InstallerImageName,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Args:            jobCmd,
			Env: []corev1.EnvVar{
				{
					Name:  "PWC_NAME",
					Value: wcName,
				},
				{
					Name:  "PROVIDER",
					Value: i.providerName,
				},
				{
					Name:  "INSTALLER_LOG_LEVEL",
					Value: i.config.ControllerConfig.Env.InstallerLogLevel,
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "config",
					MountPath: "/opt/mycluster/hosts.yaml",
					SubPath:   "hosts.yaml",
					ReadOnly:  true,
				},
				{
					Name:      "config",
					MountPath: "/opt/mycluster/group_vars/all/all.yaml",
					SubPath:   "all.yaml",
					ReadOnly:  true,
				},
				{
					Name:      "config",
					MountPath: "/opt/mycluster/group_vars/k8s-cluster/addons.yaml",
					SubPath:   "addons.yaml",
					ReadOnly:  true,
				},
				{
					Name:      "config",
					MountPath: "/opt/nodes.josn",
					SubPath:   "nodes.josn",
					ReadOnly:  true,
				},
				{
					Name:      "config",
					MountPath: "/opt/env.json",
					SubPath:   "env.josn",
					ReadOnly:  true,
				},
				{
					Name:      "sshkey",
					MountPath: "/opt/mycluster/private.key",
					SubPath:   "private.key",
					ReadOnly:  true,
				},
				{
					Name:      "sshkey",
					MountPath: "/opt/mycluster/public.key",
					SubPath:   "public.key",
					ReadOnly:  true,
				},
			},
		},
	}
	job.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: wcName,
					},
				},
			},
		},
		{
			Name: "sshkey",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: wcName,
				},
			},
		},
	}

	return &job

}

func getHostsYaml() []byte {}
func getNodesJson() []byte {}

func (i *Installer) getOtherData(wc *v1.WorkloadCluster) map[string][]byte {
	cfg := *i.config.ControllerConfig

	cfg.All.ServiceSubnet = wc.Spec.Cluster.Network.ServiceCIDR
	cfg.All.PodSubnet = wc.Spec.Cluster.Network.PodCIDR
	cfg.All.WorkloadClusterName = wc.Spec.Cluster.LoadBalance
	cfg.All.LoadbalanceApiserver = config.LoadbalanceApiserver{
		Address: wc.Spec.Cluster.LoadBalance,
		Port:    "6443",
	}

	data := make(map[string][]byte)

	if allData, err := json.Marshal(cfg.All); err == nil {
		data["all.yaml"] = allData
	}
	if envData, err := json.Marshal(cfg.Env); err == nil {
		data["env.json"] = envData
	}
	if addonData, err := json.Marshal(cfg.Addon); err == nil {
		data["addons.yaml"] = addonData
	}

	return data

}
