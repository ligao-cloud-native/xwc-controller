package agentclient

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	*http.Client
	host       string
	credential Credential
}

type Credential struct {
	SecretId  string
	SecretKey string
}

func NewClient(host string, cre Credential) (*Client, error) {
	return &Client{
		&http.Client{},
		host,
		cre,
	}, nil
}

func (c *Client) SendCmd(uuid string, command string) (*ReceiveCommandArgs, error) {
	cmd := SendCommandArgs{
		Uuid: uuid,
		Data: DataBean{
			Cmd: "CustomCmd",
			Params: CommandBean{
				Command: command,
			},
		},
	}
	res := &ReceiveCommandArgs{}

	err := c.requestAgent("callVmagent", cmd.Uuid, cmd.Data, res)
	if err != nil {
		return nil, err
	}

	return res, nil

}

func (c *Client) ReceiveCmd(reqId string) (*QueryResultRes, error) {
	args := QueryReqResult{
		ReqId: reqId,
	}
	res := &QueryResultRes{}

	err := c.requestAgent("queryReqResult", "", args, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) requestAgent(command string, uuid string, args interface{}, response interface{}) error {
	// Values used for query parameters and form values
	resValues := url.Values{}
	resValues.Set("uuid", uuid)
	resValues.Set("command", command)
	resValues.Set("apikey", c.credential.SecretId)

	signature := ""
	resValues.Set("signature", signature)

	if command == "callVmagent" {
		argsJson, err := json.Marshal(args)
		if err != nil {
			return err
		}
		base64Args := base64.StdEncoding.EncodeToString(argsJson)
		resValues.Set("data", base64Args)
	} else {
		//TODO:
	}

	query := resValues.Encode()
	url := fmt.Sprintf("%s?%s", c.host, query)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return err
	}

	return nil

}
