package leaderelection

import (
	"context"
	"fmt"
	"github.com/ligao-cloud-native/xwc-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"os"
	"time"
)

const (
	resourceName      = "xwc-controller-election-lock"
	resourceNamespace = "kube-system"
)

func Run(startAPP func(ctx context.Context)) {
	kubeClient, err := utils.KubeClient("")
	if err != nil {
		klog.Errorf("Create kubeClient for leaderElection error: %v", err)
		return
	}

	rl, err := newResourceLock(kubeClient)
	if err != nil {
		klog.Errorf("Create leaderElection resourceLock error: %v", err)
		return
	}

	leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
		Lock:          rl,
		Name:          resourceName,
		LeaseDuration: 30 * time.Second,
		RenewDeadline: 15 * time.Second,
		RetryPeriod:   10 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: startAPP,
			OnStoppedLeading: func() {
				klog.Errorf("leaderelection lost %s, gracefully terminate program", rl.Identity())
			},
			// called when the client observes a leader
			OnNewLeader: func(identity string) {
				if identity == rl.Identity() {
					return
				}
				klog.Infof("new leader election: %s", identity)
			},
		},
	})
}

func newResourceLock(client *kubernetes.Clientset) (resourcelock.Interface, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("unable to get hostname: %v", err)
	}
	id := hostname + "_" + string(uuid.NewUUID())

	// a event broadcaster
	eventBroadcaster := record.NewBroadcaster()
	// an EventRecorder that can be used to send events to this EventBroadcaster
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: resourceName})
	// sending events received from this EventBroadcaster to the given logging function.
	eventBroadcaster.StartLogging(klog.Infof)
	// starts sending events received from this EventBroadcaster to the given sink
	eventBroadcaster.StartRecordingToSink(&clientcorev1.EventSinkImpl{Interface: client.CoreV1().Events("")})

	rl, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		resourceNamespace,
		resourceName,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: recorder,
		})
	if err != nil {
		return nil, fmt.Errorf("couldn't create resource lock: %v", err)
	}

	return rl, nil

}
