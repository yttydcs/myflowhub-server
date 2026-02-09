package management

import "encoding/json"

const SubProtoManagement uint8 = 1

const (
	ActionNodeEcho     = "node_echo"
	ActionNodeEchoResp = "node_echo_resp"
	ActionListNodes    = "list_nodes"
	ActionListNodesResp = "list_nodes_resp"
	ActionListSubtree    = "list_subtree"
	ActionListSubtreeResp = "list_subtree_resp"
	ActionConfigGet      = "config_get"
	ActionConfigGetResp  = "config_get_resp"
	ActionConfigSet      = "config_set"
	ActionConfigSetResp  = "config_set_resp"
	ActionConfigList     = "config_list"
	ActionConfigListResp = "config_list_resp"
)

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type NodeEchoReq struct {
	Message string `json:"message"`
}

type NodeEchoResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Echo string `json:"echo,omitempty"`
}

type ListNodesReq struct{}

type ListNodesResp struct {
	Code  int        `json:"code"`
	Msg   string     `json:"msg,omitempty"`
	Nodes []NodeInfo `json:"nodes,omitempty"`
}

type ConfigGetReq struct {
	Key string `json:"key"`
}

type ConfigSetReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ConfigResp struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type ConfigListReq struct{}

type ConfigListResp struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg,omitempty"`
	Keys []string `json:"keys,omitempty"`
}

type NodeInfo struct {
	NodeID      uint32 `json:"node_id"`
	HasChildren bool   `json:"has_children,omitempty"`
}

type ListSubtreeReq struct{}

type ListSubtreeResp struct {
	Code  int        `json:"code"`
	Msg   string     `json:"msg,omitempty"`
	Nodes []NodeInfo `json:"nodes,omitempty"`
}
