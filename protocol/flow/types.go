package flow

import protocol "github.com/yttydcs/myflowhub-proto/protocol/flow"

// 本包为兼容壳：保留原 import path `github.com/yttydcs/myflowhub-server/protocol/flow`，
// 其类型/常量全部委托到独立协议仓库 MyFlowHub-Proto（wire 不变）。

const SubProtoFlow uint8 = protocol.SubProtoFlow

const (
	ActionSet        = protocol.ActionSet
	ActionSetResp    = protocol.ActionSetResp
	ActionRun        = protocol.ActionRun
	ActionRunResp    = protocol.ActionRunResp
	ActionStatus     = protocol.ActionStatus
	ActionStatusResp = protocol.ActionStatusResp
	ActionList       = protocol.ActionList
	ActionListResp   = protocol.ActionListResp
	ActionGet        = protocol.ActionGet
	ActionGetResp    = protocol.ActionGetResp
)

const PermFlowSet = protocol.PermFlowSet

type Message = protocol.Message

type Trigger = protocol.Trigger
type Graph = protocol.Graph
type Node = protocol.Node
type Edge = protocol.Edge

type SetReq = protocol.SetReq
type SetResp = protocol.SetResp
type RunReq = protocol.RunReq
type RunResp = protocol.RunResp
type StatusReq = protocol.StatusReq
type NodeStatus = protocol.NodeStatus
type StatusResp = protocol.StatusResp
type ListReq = protocol.ListReq
type FlowSummary = protocol.FlowSummary
type ListResp = protocol.ListResp
type GetReq = protocol.GetReq
type GetResp = protocol.GetResp
