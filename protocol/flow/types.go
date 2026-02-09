package flow

import "encoding/json"

const SubProtoFlow uint8 = 6

const (
	ActionSet        = "set"
	ActionSetResp    = "set_resp"
	ActionRun        = "run"
	ActionRunResp    = "run_resp"
	ActionStatus     = "status"
	ActionStatusResp = "status_resp"
	ActionList       = "list"
	ActionListResp   = "list_resp"
	ActionGet        = "get"
	ActionGetResp    = "get_resp"
)

const PermFlowSet = "flow.set"

type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type Trigger struct {
	Type    string `json:"type"`
	EveryMs uint64 `json:"every_ms,omitempty"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Node struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	AllowFail bool            `json:"allow_fail,omitempty"`
	Retry     *int            `json:"retry,omitempty"`
	TimeoutMs *int            `json:"timeout_ms,omitempty"`
	Spec      json.RawMessage `json:"spec"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type SetReq struct {
	ReqID        string  `json:"req_id"`
	OriginNode   uint32  `json:"origin_node,omitempty"`
	ExecutorNode uint32  `json:"executor_node,omitempty"`
	FlowID       string  `json:"flow_id"`
	Name         string  `json:"name,omitempty"`
	Trigger      Trigger `json:"trigger"`
	Graph        Graph   `json:"graph"`
}

type SetResp struct {
	ReqID  string `json:"req_id"`
	Code   int    `json:"code"`
	Msg    string `json:"msg,omitempty"`
	FlowID string `json:"flow_id,omitempty"`
}

type RunReq struct {
	ReqID        string `json:"req_id"`
	OriginNode   uint32 `json:"origin_node,omitempty"`
	ExecutorNode uint32 `json:"executor_node,omitempty"`
	FlowID       string `json:"flow_id"`
}

type RunResp struct {
	ReqID  string `json:"req_id"`
	Code   int    `json:"code"`
	Msg    string `json:"msg,omitempty"`
	FlowID string `json:"flow_id,omitempty"`
	RunID  string `json:"run_id,omitempty"`
}

type StatusReq struct {
	ReqID        string `json:"req_id"`
	OriginNode   uint32 `json:"origin_node,omitempty"`
	ExecutorNode uint32 `json:"executor_node,omitempty"`
	FlowID       string `json:"flow_id"`
	RunID        string `json:"run_id,omitempty"`
}

type NodeStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Code   int    `json:"code,omitempty"`
	Msg    string `json:"msg,omitempty"`
}

type StatusResp struct {
	ReqID        string       `json:"req_id"`
	Code         int          `json:"code"`
	Msg          string       `json:"msg,omitempty"`
	ExecutorNode uint32       `json:"executor_node,omitempty"`
	FlowID       string       `json:"flow_id,omitempty"`
	RunID        string       `json:"run_id,omitempty"`
	Status       string       `json:"status,omitempty"`
	Nodes        []NodeStatus `json:"nodes,omitempty"`
}

type ListReq struct {
	ReqID        string `json:"req_id"`
	OriginNode   uint32 `json:"origin_node,omitempty"`
	ExecutorNode uint32 `json:"executor_node,omitempty"`
}

type FlowSummary struct {
	FlowID     string `json:"flow_id"`
	Name       string `json:"name,omitempty"`
	EveryMs    uint64 `json:"every_ms,omitempty"`
	LastRunID  string `json:"last_run_id,omitempty"`
	LastStatus string `json:"last_status,omitempty"`
}

type ListResp struct {
	ReqID        string        `json:"req_id"`
	Code         int           `json:"code"`
	Msg          string        `json:"msg,omitempty"`
	ExecutorNode uint32        `json:"executor_node,omitempty"`
	Flows        []FlowSummary `json:"flows,omitempty"`
}

type GetReq struct {
	ReqID        string `json:"req_id"`
	OriginNode   uint32 `json:"origin_node,omitempty"`
	ExecutorNode uint32 `json:"executor_node,omitempty"`
	FlowID       string `json:"flow_id"`
}

type GetResp struct {
	ReqID        string  `json:"req_id"`
	Code         int     `json:"code"`
	Msg          string  `json:"msg,omitempty"`
	ExecutorNode uint32  `json:"executor_node,omitempty"`
	FlowID       string  `json:"flow_id,omitempty"`
	Name         string  `json:"name,omitempty"`
	Trigger      Trigger `json:"trigger,omitempty"`
	Graph        Graph   `json:"graph,omitempty"`
}
