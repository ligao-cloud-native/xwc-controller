package metrics

import (
	"github.com/ligao-cloud-native/kubemc/pkg/apis/xwc/v1"
	"github.com/ligao-cloud-native/xwc-controller/pkg/callback"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"
)

type Metrics struct {
	Trigger          *prometheus.CounterVec
	CurrentStatus    *prometheus.GaugeVec
	CurrentOperator  *prometheus.CounterVec
	SuccessDate      *prometheus.CounterVec
	SuccessDateTotal *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	// kubemc_controller_operator_triggered_total{controller="activation",  action="install"} 22
	trigger := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "kubemc_controller",
		Subsystem: "operator",
		Name:      "triggered_total",
		Help: "Total number of a pwc install successed/failed, " +
			"operator triggered the pwc controller to reconcile an object",
	}, []string{"controller", "action"})

	// kubemc_controller_current_status{controller="",status="", name=""} 1
	currentStatus := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kubemc_controller_current_status",
		Help: "number of object a pwc current status",
	}, []string{"controller", "status", "name"})

	currentOperator := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kubemc_controller_operator",
		Help: "number of object a pwc operator scale/reduce/upgrade",
	}, []string{"controller", "status", "name"})

	successDate := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kubemc_controller_operator_success_date_millisecond",
		Help: "pwc operates successfully date in millisecond",
	}, []string{"controller", "status", "name"})

	successDateTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kubemc_controller_operator_success_date_millisecond_total",
		Help: "pwc operates success, remove date total in millisecond",
	}, []string{"controller", "status"})

	prometheus.MustRegister(
		trigger,
		currentStatus,
		currentOperator,
		successDate,
		successDateTotal,
	)

	return &Metrics{
		Trigger:          trigger,
		CurrentStatus:    currentStatus,
		CurrentOperator:  currentOperator,
		SuccessDate:      successDate,
		SuccessDateTotal: successDateTotal,
	}

}

func (metrics *Metrics) HandleMetrics(jcb *callback.JobCallback) {
	switch jcb.TaskType {
	case "install-complete":
		metrics.Trigger.WithLabelValues("pwc_controller", "install_success").Inc()
		metrics.SuccessDate.WithLabelValues("pwc_controller", "success", jcb.ClusterName).Add(jcb.OperatorTime)
		metrics.SuccessDateTotal.WithLabelValues("pwc_controller", "install_success").Add(jcb.OperatorTime)
	case "install_failed":
		metrics.Trigger.WithLabelValues("pwc_controller", "install_failed").Inc()
	case "scale-complete":
		metrics.Trigger.WithLabelValues("pwc_controller", "scale_success").Inc()
		metrics.SuccessDateTotal.WithLabelValues("pwc_controller", "scale_success").Add(jcb.OperatorTime)
	case "scale_failed":
		metrics.Trigger.WithLabelValues("pwc_controller", "scale_failed").Inc()
	case "reduce-complete":
		metrics.Trigger.WithLabelValues("pwc_controller", "reduce_success").Inc()
	case "reduce_failed":
		metrics.Trigger.WithLabelValues("pwc_controller", "reduce_failed").Inc()
	case "upgrade-complete":
		metrics.Trigger.WithLabelValues("pwc_controller", "upgrade_success").Inc()
	case "upgrade_failed":
		metrics.Trigger.WithLabelValues("pwc_controller", "upgrade_failed").Inc()
	case "remove-complete":
		metrics.SuccessDate.WithLabelValues("pwc_controller", "success", jcb.ClusterName).Add(jcb.OperatorTime)
	}

}

func (metrics *Metrics) CurrentMetrics(store cache.Store) {
	metrics.CurrentStatus.Reset()
	for _, v := range store.List() {
		if pwc, ok := v.(*v1.WorkloadCluster); ok {
			metrics.CurrentStatus.WithLabelValues("pwc_controller", string(pwc.Status.Phase), pwc.Name).Inc()
		}
	}
}
