package exec

import "encoding/json"

// 子协议：exec（网络特殊能力调用）。
const SubProtoExec uint8 = 7

const (
	actionCall     = "call"
	actionCallResp = "call_resp"
)

const permExecCall = "exec.call"

type message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type CallReq struct {
	ReqID        string          `json:"req_id"`
	ExecutorNode uint32          `json:"executor_node"`
	TargetNode   uint32          `json:"target_node"`
	Method       string          `json:"method"`
	Args         json.RawMessage `json:"args,omitempty"`
	TimeoutMs    int             `json:"timeout_ms,omitempty"`
}

type CallResp struct {
	ReqID        string          `json:"req_id"`
	Code         int             `json:"code"`
	Msg          string          `json:"msg,omitempty"`
	ExecutorNode uint32          `json:"executor_node,omitempty"`
	TargetNode   uint32          `json:"target_node,omitempty"`
	Method       string          `json:"method,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
}
