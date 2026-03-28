package stream

import protocol "github.com/yttydcs/myflowhub-proto/protocol/stream"

// 本包为兼容壳：保留原 import path `github.com/yttydcs/myflowhub-server/protocol/stream`，
// 其类型/常量全部委托到独立协议仓库 MyFlowHub-Proto（wire 不变）。

const SubProtoStream uint8 = protocol.SubProtoStream

const (
	KindCtrl byte = protocol.KindCtrl
	KindData byte = protocol.KindData
	KindAck  byte = protocol.KindAck
)

const (
	ActionAnnounce             = protocol.ActionAnnounce
	ActionAnnounceResp         = protocol.ActionAnnounceResp
	ActionWithdraw             = protocol.ActionWithdraw
	ActionWithdrawResp         = protocol.ActionWithdrawResp
	ActionListSources          = protocol.ActionListSources
	ActionListSourcesResp      = protocol.ActionListSourcesResp
	ActionGetSource            = protocol.ActionGetSource
	ActionGetSourceResp        = protocol.ActionGetSourceResp
	ActionAnnounceConsumer     = protocol.ActionAnnounceConsumer
	ActionAnnounceConsumerResp = protocol.ActionAnnounceConsumerResp
	ActionWithdrawConsumer     = protocol.ActionWithdrawConsumer
	ActionWithdrawConsumerResp = protocol.ActionWithdrawConsumerResp
	ActionListConsumers        = protocol.ActionListConsumers
	ActionListConsumersResp    = protocol.ActionListConsumersResp
	ActionGetConsumer          = protocol.ActionGetConsumer
	ActionGetConsumerResp      = protocol.ActionGetConsumerResp
	ActionSubscribe            = protocol.ActionSubscribe
	ActionSubscribeResp        = protocol.ActionSubscribeResp
	ActionUnsubscribe          = protocol.ActionUnsubscribe
	ActionUnsubscribeResp      = protocol.ActionUnsubscribeResp
	ActionConnect              = protocol.ActionConnect
	ActionConnectResp          = protocol.ActionConnectResp
	ActionDisconnect           = protocol.ActionDisconnect
	ActionDisconnectResp       = protocol.ActionDisconnectResp
	ActionSignal               = protocol.ActionSignal
	ActionSignalResp           = protocol.ActionSignalResp
)

const (
	PermStreamPublish   = protocol.PermStreamPublish
	PermStreamConsume   = protocol.PermStreamConsume
	PermStreamSubscribe = protocol.PermStreamSubscribe
	PermStreamConnect   = protocol.PermStreamConnect
)

const (
	StreamKindMusic  = protocol.StreamKindMusic
	StreamKindVideo  = protocol.StreamKindVideo
	StreamKindText   = protocol.StreamKindText
	StreamKindCustom = protocol.StreamKindCustom
)

const (
	ModeLive    = protocol.ModeLive
	ModeBounded = protocol.ModeBounded
)

const (
	UnitModeFrame = protocol.UnitModeFrame
	UnitModeChunk = protocol.UnitModeChunk
)

const (
	SignalOpPause           = protocol.SignalOpPause
	SignalOpResume          = protocol.SignalOpResume
	SignalOpMetadataUpdate  = protocol.SignalOpMetadataUpdate
	SignalOpKeyframeRequest = protocol.SignalOpKeyframeRequest
	SignalOpCustom          = protocol.SignalOpCustom
)

const (
	HeaderVersionV1 uint8 = protocol.HeaderVersionV1

	DataFlagEOS           uint8 = protocol.DataFlagEOS
	DataFlagKeyframe      uint8 = protocol.DataFlagKeyframe
	DataFlagConfig        uint8 = protocol.DataFlagConfig
	DataFlagDiscontinuity uint8 = protocol.DataFlagDiscontinuity
)

type Message = protocol.Message

type SourceDescriptor = protocol.SourceDescriptor
type ConsumerDescriptor = protocol.ConsumerDescriptor

type AnnounceReq = protocol.AnnounceReq
type AnnounceResp = protocol.AnnounceResp
type WithdrawReq = protocol.WithdrawReq
type WithdrawResp = protocol.WithdrawResp
type ListSourcesReq = protocol.ListSourcesReq
type ListSourcesResp = protocol.ListSourcesResp
type GetSourceReq = protocol.GetSourceReq
type GetSourceResp = protocol.GetSourceResp
type AnnounceConsumerReq = protocol.AnnounceConsumerReq
type AnnounceConsumerResp = protocol.AnnounceConsumerResp
type WithdrawConsumerReq = protocol.WithdrawConsumerReq
type WithdrawConsumerResp = protocol.WithdrawConsumerResp
type ListConsumersReq = protocol.ListConsumersReq
type ListConsumersResp = protocol.ListConsumersResp
type GetConsumerReq = protocol.GetConsumerReq
type GetConsumerResp = protocol.GetConsumerResp
type SubscribeReq = protocol.SubscribeReq
type SubscribeResp = protocol.SubscribeResp
type UnsubscribeReq = protocol.UnsubscribeReq
type UnsubscribeResp = protocol.UnsubscribeResp
type ConnectReq = protocol.ConnectReq
type ConnectResp = protocol.ConnectResp
type DisconnectReq = protocol.DisconnectReq
type DisconnectResp = protocol.DisconnectResp
type SignalReq = protocol.SignalReq
type SignalResp = protocol.SignalResp

type StreamDataHeaderV1 = protocol.StreamDataHeaderV1
type StreamAckHeaderV1 = protocol.StreamAckHeaderV1
