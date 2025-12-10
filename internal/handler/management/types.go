package management

import "encoding/json"

type mgmtMessage struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type nodeEchoReq struct {
	Message string `json:"message"`
}

type nodeEchoResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Echo string `json:"echo,omitempty"`
}
