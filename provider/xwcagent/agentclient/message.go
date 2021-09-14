package agentclient

// call vmagent
type DataBean struct {
	Cmd    string      `json:"cdm,omitempty"`
	Params CommandBean `json:"paras,omitempty"`
}

type CommandBean struct {
	Command string `json:"command,omitempty"`
}

type SendCommandArgs struct {
	Uuid string   `json:"uuid,omitempty"`
	Data DataBean `json:"data,omitempty"`
}

type ReceiveCommandArgs struct {
	Success int
	Data    ReceiveData
}

type ReceiveData struct {
	ReqId string
}

// query result
type QueryReqResult struct {
	ReqId string `json:"reqId,omitempty"`
}

type QueryResultRes struct {
	Success int
	Data    QueryReqResultData
}

type QueryReqResultData struct {
	ReqId  string
	State  int
	Result QueryRequest
}

type QueryRequest struct {
	OutPut    string
	ExitValue int
	Error     string
}
