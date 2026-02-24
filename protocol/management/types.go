package management

import protocol "github.com/yttydcs/myflowhub-proto/protocol/management"

// 本包为兼容壳：保留原 import path `github.com/yttydcs/myflowhub-server/protocol/management`，
// 其类型/常量全部委托到独立协议仓库 MyFlowHub-Proto（wire 不变）。

const SubProtoManagement uint8 = protocol.SubProtoManagement

const (
	ActionNodeEcho        = protocol.ActionNodeEcho
	ActionNodeEchoResp    = protocol.ActionNodeEchoResp
	ActionNodeInfo        = protocol.ActionNodeInfo
	ActionNodeInfoResp    = protocol.ActionNodeInfoResp
	ActionListNodes       = protocol.ActionListNodes
	ActionListNodesResp   = protocol.ActionListNodesResp
	ActionListSubtree     = protocol.ActionListSubtree
	ActionListSubtreeResp = protocol.ActionListSubtreeResp
	ActionConfigGet       = protocol.ActionConfigGet
	ActionConfigGetResp   = protocol.ActionConfigGetResp
	ActionConfigSet       = protocol.ActionConfigSet
	ActionConfigSetResp   = protocol.ActionConfigSetResp
	ActionConfigList      = protocol.ActionConfigList
	ActionConfigListResp  = protocol.ActionConfigListResp
)

type Message = protocol.Message

type NodeEchoReq = protocol.NodeEchoReq
type NodeEchoResp = protocol.NodeEchoResp
type NodeInfoReq = protocol.NodeInfoReq
type NodeInfoResp = protocol.NodeInfoResp
type ListNodesReq = protocol.ListNodesReq
type ListNodesResp = protocol.ListNodesResp
type ConfigGetReq = protocol.ConfigGetReq
type ConfigSetReq = protocol.ConfigSetReq
type ConfigResp = protocol.ConfigResp
type ConfigListReq = protocol.ConfigListReq
type ConfigListResp = protocol.ConfigListResp
type NodeInfo = protocol.NodeInfo
type ListSubtreeReq = protocol.ListSubtreeReq
type ListSubtreeResp = protocol.ListSubtreeResp
