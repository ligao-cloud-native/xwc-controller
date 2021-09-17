package provider

type Prechecker interface {
	PreCheck(args ...string) (int, string)
}

type PrecheckResultInterface interface {
	IsSuccess() bool
	HostInfo() string
	ResultMessage() string
}
