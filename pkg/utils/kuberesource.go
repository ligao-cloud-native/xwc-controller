package utils

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const pksInstallerNamespace = "pks-installer"

func GetConfigMap(clientSet kubernetes.Interface, cmName string) (*corev1.ConfigMap, error) {
	return clientSet.CoreV1().ConfigMaps(pksInstallerNamespace).Get(context.TODO(), cmName, metav1.GetOptions{})
}

func CreateConfigMap(clientSet kubernetes.Interface, cmObject *corev1.ConfigMap) error {
	_, err := clientSet.CoreV1().ConfigMaps(pksInstallerNamespace).Create(context.TODO(), cmObject, metav1.CreateOptions{})
	return err
}

func DeleteConfigMap(clientSet kubernetes.Interface, cmName string) error {
	return clientSet.CoreV1().ConfigMaps(pksInstallerNamespace).Delete(context.TODO(), cmName, metav1.DeleteOptions{})
}

func GetSecret(clientSet kubernetes.Interface, secretName string) (*corev1.Secret, error) {
	return clientSet.CoreV1().Secrets(pksInstallerNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
}

func CreateSecret(clientSet kubernetes.Interface, secretObject *corev1.Secret) error {
	_, err := clientSet.CoreV1().Secrets(pksInstallerNamespace).Create(context.TODO(), secretObject, metav1.CreateOptions{})
	return err
}

func DeleteSecret(clientSet kubernetes.Interface, secretName string) error {
	return clientSet.CoreV1().Secrets(pksInstallerNamespace).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
}

func CreateJob(clientSet kubernetes.Interface, jobObject *batchv1.Job) (*batchv1.Job, error) {
	return clientSet.BatchV1().Jobs(pksInstallerNamespace).Create(context.TODO(), jobObject, metav1.CreateOptions{})
}

func ListJob(clientSet kubernetes.Interface, opts metav1.ListOptions) (*batchv1.JobList, error) {
	return clientSet.BatchV1().Jobs(pksInstallerNamespace).List(context.TODO(), opts)
}

func DeleteJob(clientSet kubernetes.Interface, name string) error {
	return clientSet.BatchV1().Jobs(pksInstallerNamespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func ListPod(clientSet kubernetes.Interface, opts metav1.ListOptions) (*corev1.PodList, error) {
	return clientSet.CoreV1().Pods(pksInstallerNamespace).List(context.TODO(), opts)
}

func DeletePod(clientSet kubernetes.Interface, name string) error {
	return clientSet.CoreV1().Pods(pksInstallerNamespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
