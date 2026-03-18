package exec

import protocol "github.com/yttydcs/myflowhub-proto/protocol/exec"

// 本包为兼容壳：保留原 import path `github.com/yttydcs/myflowhub-server/protocol/exec`，
// 其类型/常量全部委托到独立协议仓库 MyFlowHub-Proto（wire 不变）。

const SubProtoExec uint8 = protocol.SubProtoExec

const (
	ActionCall     = protocol.ActionCall
	ActionCallResp = protocol.ActionCallResp

	ActionCapSnapshot  = protocol.ActionCapSnapshot
	ActionCapUpsert    = protocol.ActionCapUpsert
	ActionCapWithdraw  = protocol.ActionCapWithdraw
	ActionCapHeartbeat = protocol.ActionCapHeartbeat
	ActionCapSyncResp  = protocol.ActionCapSyncResp
	ActionCapQuery     = protocol.ActionCapQuery
	ActionCapQueryResp = protocol.ActionCapQueryResp
)

const (
	PermExecCall     = protocol.PermExecCall
	PermExecCapSync  = protocol.PermExecCapSync
	PermExecCapQuery = protocol.PermExecCapQuery
)

type Message = protocol.Message
type CallReq = protocol.CallReq
type CallResp = protocol.CallResp
type CapabilityDescriptor = protocol.CapabilityDescriptor
type CapabilityKey = protocol.CapabilityKey
type CapSnapshotReq = protocol.CapSnapshotReq
type CapUpsertReq = protocol.CapUpsertReq
type CapWithdrawReq = protocol.CapWithdrawReq
type CapHeartbeatReq = protocol.CapHeartbeatReq
type CapSyncResp = protocol.CapSyncResp
type CapQueryReq = protocol.CapQueryReq
type CapabilityRoute = protocol.CapabilityRoute
type CapQueryResp = protocol.CapQueryResp
