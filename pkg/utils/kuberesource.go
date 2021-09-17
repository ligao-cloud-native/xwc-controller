package utils

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const pksInstallerNamespace = "xwc-installer"

func GetConfigMap(clientSet kubernetes.Interface, cmName string) (*corev1.ConfigMap, error) {
	return clientSet.CoreV1().ConfigMaps(pksInstallerNamespace).Get(context.TODO(), cmName, metav1.GetOptions{})
}

func CreateConfigMap(clientSet kubernetes.Interface, cmObject *corev1.ConfigMap) error {
	_, err := clientSet.CoreV1().ConfigMaps(pksInstallerNamespace).Create(context.TODO(), cmObject, metav1.CreateOptions{})
	return err
}

func GetSecret(clientSet kubernetes.Interface, secretName string) (*corev1.Secret, error) {
	return clientSet.CoreV1().Secrets(pksInstallerNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
}

func CreateSecret(clientSet kubernetes.Interface, secretObject *corev1.Secret) error {
	_, err := clientSet.CoreV1().Secrets(pksInstallerNamespace).Create(context.TODO(), secretObject, metav1.CreateOptions{})
	return err
}

func CreateJob(clientSet kubernetes.Interface, jobObject *batchv1.Job) error {
	_, err := clientSet.BatchV1().Jobs(pksInstallerNamespace).Create(context.TODO(), jobObject, metav1.CreateOptions{})
	return err
}
