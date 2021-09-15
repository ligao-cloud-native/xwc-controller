package provider

type PrecheckResultInterface interface {
	IsSuccess() bool
	HostInfo() string
	ResultMessage() string
}
