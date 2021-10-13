package callback

type JobCallback struct {
	ClusterName  string
	TaskType     string
	BearerToken  string
	KubeConfig   string
	OperatorTime float64
	Message      string
}
