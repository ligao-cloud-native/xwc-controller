package provider

type Prechecker interface {
	PreCheck(ip, uuid string) (int, string)
}

type PrecheckResultInterface interface {
	IsSuccess() bool
	HostInfo() string
	ResultMessage() string
}
