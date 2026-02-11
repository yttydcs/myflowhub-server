package exec

import protocol "github.com/yttydcs/myflowhub-proto/protocol/exec"

// 本包为兼容壳：保留原 import path `github.com/yttydcs/myflowhub-server/protocol/exec`，
// 其类型/常量全部委托到独立协议仓库 MyFlowHub-Proto（wire 不变）。

const SubProtoExec uint8 = protocol.SubProtoExec

const (
	ActionCall     = protocol.ActionCall
	ActionCallResp = protocol.ActionCallResp
)

const PermExecCall = protocol.PermExecCall

type Message = protocol.Message
type CallReq = protocol.CallReq
type CallResp = protocol.CallResp
