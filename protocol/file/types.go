package file

import protocol "github.com/yttydcs/myflowhub-proto/protocol/file"

// 本包为兼容壳：保留原 import path `github.com/yttydcs/myflowhub-server/protocol/file`，
// 其类型/常量全部委托到独立协议仓库 MyFlowHub-Proto（wire 不变）。

const SubProtoFile uint8 = protocol.SubProtoFile

const (
	KindCtrl byte = protocol.KindCtrl
	KindData byte = protocol.KindData
	KindAck  byte = protocol.KindAck
)

const (
	ActionRead      = protocol.ActionRead
	ActionWrite     = protocol.ActionWrite
	ActionReadResp  = protocol.ActionReadResp
	ActionWriteResp = protocol.ActionWriteResp
)

const (
	OpPull     = protocol.OpPull
	OpOffer    = protocol.OpOffer
	OpList     = protocol.OpList
	OpReadText = protocol.OpReadText
)

type Message = protocol.Message

type ReadReq = protocol.ReadReq
type ReadResp = protocol.ReadResp
type WriteReq = protocol.WriteReq
type WriteResp = protocol.WriteResp
