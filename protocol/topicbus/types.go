package topicbus

import protocol "github.com/yttydcs/myflowhub-proto/protocol/topicbus"

// 本包为兼容壳：保留原 import path `github.com/yttydcs/myflowhub-server/protocol/topicbus`，
// 其类型/常量全部委托到独立协议仓库 MyFlowHub-Proto（wire 不变）。

const SubProtoTopicBus uint8 = protocol.SubProtoTopicBus

const (
	ActionSubscribe          = protocol.ActionSubscribe
	ActionSubscribeResp      = protocol.ActionSubscribeResp
	ActionSubscribeBatch     = protocol.ActionSubscribeBatch
	ActionSubscribeBatchResp = protocol.ActionSubscribeBatchResp

	ActionUnsubscribe          = protocol.ActionUnsubscribe
	ActionUnsubscribeResp      = protocol.ActionUnsubscribeResp
	ActionUnsubscribeBatch     = protocol.ActionUnsubscribeBatch
	ActionUnsubscribeBatchResp = protocol.ActionUnsubscribeBatchResp

	ActionListSubs     = protocol.ActionListSubs
	ActionListSubsResp = protocol.ActionListSubsResp

	ActionPublish = protocol.ActionPublish
)

type Message = protocol.Message

type SubscribeReq = protocol.SubscribeReq
type SubscribeBatchReq = protocol.SubscribeBatchReq
type PublishReq = protocol.PublishReq
type Resp = protocol.Resp
type ListResp = protocol.ListResp
