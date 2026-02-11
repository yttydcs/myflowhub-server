package varstore

import protocol "github.com/yttydcs/myflowhub-proto/protocol/varstore"

// 本包为兼容壳：保留原 import path `github.com/yttydcs/myflowhub-server/protocol/varstore`，
// 其类型/常量全部委托到独立协议仓库 MyFlowHub-Proto（wire 不变）。

const (
	SubProtoVarStore uint8 = protocol.SubProtoVarStore

	ActionSet           = protocol.ActionSet
	ActionAssistSet     = protocol.ActionAssistSet
	ActionSetResp       = protocol.ActionSetResp
	ActionAssistSetResp = protocol.ActionAssistSetResp
	ActionUpSet         = protocol.ActionUpSet
	ActionNotifySet     = protocol.ActionNotifySet

	ActionGet           = protocol.ActionGet
	ActionAssistGet     = protocol.ActionAssistGet
	ActionGetResp       = protocol.ActionGetResp
	ActionAssistGetResp = protocol.ActionAssistGetResp

	ActionList           = protocol.ActionList
	ActionAssistList     = protocol.ActionAssistList
	ActionListResp       = protocol.ActionListResp
	ActionAssistListResp = protocol.ActionAssistListResp

	ActionRevoke           = protocol.ActionRevoke
	ActionAssistRevoke     = protocol.ActionAssistRevoke
	ActionRevokeResp       = protocol.ActionRevokeResp
	ActionAssistRevokeResp = protocol.ActionAssistRevokeResp
	ActionUpRevoke         = protocol.ActionUpRevoke
	ActionNotifyRevoke     = protocol.ActionNotifyRevoke

	ActionSubscribe           = protocol.ActionSubscribe
	ActionAssistSubscribe     = protocol.ActionAssistSubscribe
	ActionSubscribeResp       = protocol.ActionSubscribeResp
	ActionAssistSubscribeResp = protocol.ActionAssistSubscribeResp
	ActionUnsubscribe         = protocol.ActionUnsubscribe
	ActionAssistUnsubscribe   = protocol.ActionAssistUnsubscribe

	ActionVarChanged = protocol.ActionVarChanged
	ActionVarDeleted = protocol.ActionVarDeleted

	VisibilityPublic  = protocol.VisibilityPublic
	VisibilityPrivate = protocol.VisibilityPrivate
)

type Message = protocol.Message

type SetReq = protocol.SetReq
type GetReq = protocol.GetReq
type ListReq = protocol.ListReq
type SubscribeReq = protocol.SubscribeReq
type VarResp = protocol.VarResp
