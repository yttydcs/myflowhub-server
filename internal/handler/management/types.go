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

type listNodesReq struct{}

type listNodesResp struct {
	Code  int        `json:"code"`
	Msg   string     `json:"msg,omitempty"`
	Nodes []nodeInfo `json:"nodes,omitempty"`
}

type configGetReq struct {
	Key string `json:"key"`
}

type configSetReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type configResp struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type configListReq struct{}

type configListResp struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg,omitempty"`
	Keys []string `json:"keys,omitempty"`
}

type nodeInfo struct {
	NodeID      uint32 `json:"node_id"`
	HasChildren bool   `json:"has_children,omitempty"`
}

type listSubtreeReq struct{}

type listSubtreeResp struct {
	Code  int        `json:"code"`
	Msg   string     `json:"msg,omitempty"`
	Nodes []nodeInfo `json:"nodes,omitempty"`
}
