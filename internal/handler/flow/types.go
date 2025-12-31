package flow

import "encoding/json"

// 子协议：flow（DAG 工作流调度）。
const SubProtoFlow uint8 = 6

const (
	actionSet        = "set"
	actionSetResp    = "set_resp"
	actionRun        = "run"
	actionRunResp    = "run_resp"
	actionStatus     = "status"
	actionStatusResp = "status_resp"
	actionList       = "list"
	actionListResp   = "list_resp"
	actionGet        = "get"
	actionGetResp    = "get_resp"
)

const permFlowSet = "flow.set"

type message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type trigger struct {
	Type    string `json:"type"`
	EveryMs uint64 `json:"every_ms,omitempty"`
}

type graph struct {
	Nodes []node `json:"nodes"`
	Edges []edge `json:"edges"`
}

type node struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	AllowFail bool            `json:"allow_fail,omitempty"`
	Retry     *int            `json:"retry,omitempty"`
	TimeoutMs *int            `json:"timeout_ms,omitempty"`
	Spec      json.RawMessage `json:"spec"`
}

type edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type setReq struct {
	ReqID        string  `json:"req_id"`
	OriginNode   uint32  `json:"origin_node,omitempty"`
	ExecutorNode uint32  `json:"executor_node,omitempty"`
	FlowID       string  `json:"flow_id"`
	Name         string  `json:"name,omitempty"`
	Trigger      trigger `json:"trigger"`
	Graph        graph   `json:"graph"`
}

type setResp struct {
	ReqID  string `json:"req_id"`
	Code   int    `json:"code"`
	Msg    string `json:"msg,omitempty"`
	FlowID string `json:"flow_id,omitempty"`
}

type runReq struct {
	ReqID        string `json:"req_id"`
	OriginNode   uint32 `json:"origin_node,omitempty"`
	ExecutorNode uint32 `json:"executor_node,omitempty"`
	FlowID       string `json:"flow_id"`
}

type runResp struct {
	ReqID  string `json:"req_id"`
	Code   int    `json:"code"`
	Msg    string `json:"msg,omitempty"`
	FlowID string `json:"flow_id,omitempty"`
	RunID  string `json:"run_id,omitempty"`
}

type statusReq struct {
	ReqID        string `json:"req_id"`
	OriginNode   uint32 `json:"origin_node,omitempty"`
	ExecutorNode uint32 `json:"executor_node,omitempty"`
	FlowID       string `json:"flow_id"`
	RunID        string `json:"run_id,omitempty"`
}

type nodeStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Code   int    `json:"code,omitempty"`
	Msg    string `json:"msg,omitempty"`
}

type statusResp struct {
	ReqID        string       `json:"req_id"`
	Code         int          `json:"code"`
	Msg          string       `json:"msg,omitempty"`
	ExecutorNode uint32       `json:"executor_node,omitempty"`
	FlowID       string       `json:"flow_id,omitempty"`
	RunID        string       `json:"run_id,omitempty"`
	Status       string       `json:"status,omitempty"`
	Nodes        []nodeStatus `json:"nodes,omitempty"`
}

type listReq struct {
	ReqID        string `json:"req_id"`
	OriginNode   uint32 `json:"origin_node,omitempty"`
	ExecutorNode uint32 `json:"executor_node,omitempty"`
}

type flowSummary struct {
	FlowID     string `json:"flow_id"`
	Name       string `json:"name,omitempty"`
	EveryMs    uint64 `json:"every_ms,omitempty"`
	LastRunID  string `json:"last_run_id,omitempty"`
	LastStatus string `json:"last_status,omitempty"`
}

type listResp struct {
	ReqID        string        `json:"req_id"`
	Code         int           `json:"code"`
	Msg          string        `json:"msg,omitempty"`
	ExecutorNode uint32        `json:"executor_node,omitempty"`
	Flows        []flowSummary `json:"flows,omitempty"`
}

type getReq struct {
	ReqID        string `json:"req_id"`
	OriginNode   uint32 `json:"origin_node,omitempty"`
	ExecutorNode uint32 `json:"executor_node,omitempty"`
	FlowID       string `json:"flow_id"`
}

type getResp struct {
	ReqID        string  `json:"req_id"`
	Code         int     `json:"code"`
	Msg          string  `json:"msg,omitempty"`
	ExecutorNode uint32  `json:"executor_node,omitempty"`
	FlowID       string  `json:"flow_id,omitempty"`
	Name         string  `json:"name,omitempty"`
	Trigger      trigger `json:"trigger,omitempty"`
	Graph        graph   `json:"graph,omitempty"`
}
