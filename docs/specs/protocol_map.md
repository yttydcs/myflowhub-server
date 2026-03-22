# Protocol Mapping（Client <-> Server）

> 本文档为“半自动文档”：
> - single source-of-truth：`protocol/*/types.go`
> - **不要手工修改** `<!-- BEGIN GENERATED -->` 与 `<!-- END GENERATED -->` 中间的内容
> - canonical 生成入口位于 `MyFlowHub-Proto/docs/specs/protocol_map.md`
> - 本仓库保留同步副本：`docs/specs/protocol_map.md`

<!-- BEGIN GENERATED -->
## SubProto Overview

| SubProto | Name | Package |
|---:|---|---|
| 1 | Management | `protocol/management` |
| 2 | Auth | `protocol/auth` |
| 3 | VarStore | `protocol/varstore` |
| 4 | TopicBus | `protocol/topicbus` |
| 5 | File | `protocol/file` |
| 6 | Flow | `protocol/flow` |
| 7 | Exec | `protocol/exec` |

## Management (SubProto=1)

**Actions**
- `ActionConfigGet = "config_get"`
- `ActionConfigGetResp = "config_get_resp"`
- `ActionConfigList = "config_list"`
- `ActionConfigListResp = "config_list_resp"`
- `ActionConfigSet = "config_set"`
- `ActionConfigSetResp = "config_set_resp"`
- `ActionListNodes = "list_nodes"`
- `ActionListNodesResp = "list_nodes_resp"`
- `ActionListSubtree = "list_subtree"`
- `ActionListSubtreeResp = "list_subtree_resp"`
- `ActionNodeEcho = "node_echo"`
- `ActionNodeEchoResp = "node_echo_resp"`
- `ActionNodeInfo = "node_info"`
- `ActionNodeInfoResp = "node_info_resp"`

**Payload types**
- `ConfigGetReq`
- `ConfigListReq`
- `ConfigListResp`
- `ConfigResp`
- `ConfigSetReq`
- `ListNodesReq`
- `ListNodesResp`
- `ListSubtreeReq`
- `ListSubtreeResp`
- `Message`
- `NodeEchoReq`
- `NodeEchoResp`
- `NodeInfo`
- `NodeInfoReq`
- `NodeInfoResp`

## Auth (SubProto=2)

**Actions**
- `ActionAssistLogin = "assist_login"`
- `ActionAssistLoginResp = "assist_login_resp"`
- `ActionAssistOffline = "assist_offline"`
- `ActionAssistQueryCred = "assist_query_credential"`
- `ActionAssistQueryCredResp = "assist_query_credential_resp"`
- `ActionAssistRegister = "assist_register"`
- `ActionAssistRegisterResp = "assist_register_resp"`
- `ActionGetPerms = "get_perms"`
- `ActionGetPermsResp = "get_perms_resp"`
- `ActionListRoles = "list_roles"`
- `ActionListRolesResp = "list_roles_resp"`
- `ActionLogin = "login"`
- `ActionLoginResp = "login_resp"`
- `ActionOffline = "offline"`
- `ActionPermsInvalidate = "perms_invalidate"`
- `ActionPermsSnapshot = "perms_snapshot"`
- `ActionRegister = "register"`
- `ActionRegisterResp = "register_resp"`
- `ActionRevoke = "revoke"`
- `ActionRevokeResp = "revoke_resp"`
- `ActionUpLogin = "up_login"`
- `ActionUpLoginResp = "up_login_resp"`

**Payload types**
- `InvalidateData`
- `ListRolesReq`
- `LoginData`
- `Message`
- `OfflineData`
- `PermsQueryData`
- `QueryCredData`
- `RegisterData`
- `RespData`
- `RevokeData`
- `RolePermEntry`
- `UpLoginData`

## VarStore (SubProto=3)

**Actions**
- `ActionAssistGet = "assist_get"`
- `ActionAssistGetResp = "assist_get_resp"`
- `ActionAssistList = "assist_list"`
- `ActionAssistListResp = "assist_list_resp"`
- `ActionAssistRevoke = "assist_revoke"`
- `ActionAssistRevokeResp = "assist_revoke_resp"`
- `ActionAssistSet = "assist_set"`
- `ActionAssistSetResp = "assist_set_resp"`
- `ActionAssistSubscribe = "assist_subscribe"`
- `ActionAssistSubscribeResp = "assist_subscribe_resp"`
- `ActionAssistUnsubscribe = "assist_unsubscribe"`
- `ActionGet = "get"`
- `ActionGetResp = "get_resp"`
- `ActionList = "list"`
- `ActionListResp = "list_resp"`
- `ActionNotifyRevoke = "notify_revoke"`
- `ActionNotifySet = "notify_set"`
- `ActionRevoke = "revoke"`
- `ActionRevokeResp = "revoke_resp"`
- `ActionSet = "set"`
- `ActionSetResp = "set_resp"`
- `ActionSubscribe = "subscribe"`
- `ActionSubscribeResp = "subscribe_resp"`
- `ActionUnsubscribe = "unsubscribe"`
- `ActionUpRevoke = "up_revoke"`
- `ActionUpSet = "up_set"`
- `ActionVarChanged = "var_changed"`
- `ActionVarDeleted = "var_deleted"`

**Payload types**
- `GetReq`
- `ListReq`
- `Message`
- `SetReq`
- `SubscribeReq`
- `VarResp`

**Other constants**
- `VisibilityPrivate = "private"`
- `VisibilityPublic = "public"`

## TopicBus (SubProto=4)

**Actions**
- `ActionListSubs = "list_subs"`
- `ActionListSubsResp = "list_subs_resp"`
- `ActionPublish = "publish"`
- `ActionSubscribe = "subscribe"`
- `ActionSubscribeBatch = "subscribe_batch"`
- `ActionSubscribeBatchResp = "subscribe_batch_resp"`
- `ActionSubscribeResp = "subscribe_resp"`
- `ActionUnsubscribe = "unsubscribe"`
- `ActionUnsubscribeBatch = "unsubscribe_batch"`
- `ActionUnsubscribeBatchResp = "unsubscribe_batch_resp"`
- `ActionUnsubscribeResp = "unsubscribe_resp"`

**Payload types**
- `ListResp`
- `Message`
- `PublishReq`
- `Resp`
- `SubscribeBatchReq`
- `SubscribeReq`

## File (SubProto=5)

**Actions**
- `ActionRead = "read"`
- `ActionReadResp = "read_resp"`
- `ActionWrite = "write"`
- `ActionWriteResp = "write_resp"`

**Payload types**
- `Message`
- `ReadReq`
- `ReadResp`
- `WriteReq`
- `WriteResp`

**Other constants**
- `KindAck = 0x03`
- `KindCtrl = 0x01`
- `KindData = 0x02`
- `OpList = "list"`
- `OpOffer = "offer"`
- `OpPull = "pull"`
- `OpReadText = "read_text"`

## Flow (SubProto=6)

**Actions**
- `ActionDelete = "delete"`
- `ActionDeleteResp = "delete_resp"`
- `ActionGet = "get"`
- `ActionGetResp = "get_resp"`
- `ActionList = "list"`
- `ActionListResp = "list_resp"`
- `ActionRun = "run"`
- `ActionRunResp = "run_resp"`
- `ActionSet = "set"`
- `ActionSetResp = "set_resp"`
- `ActionStatus = "status"`
- `ActionStatusResp = "status_resp"`

**Payload types**
- `DeleteReq`
- `DeleteResp`
- `Edge`
- `FlowSummary`
- `GetReq`
- `GetResp`
- `Graph`
- `ListReq`
- `ListResp`
- `Message`
- `Node`
- `NodeStatus`
- `RunReq`
- `RunResp`
- `SetReq`
- `SetResp`
- `StatusReq`
- `StatusResp`
- `Trigger`

**Other constants**
- `PermFlowDelete = "flow.delete"`
- `PermFlowSet = "flow.set"`

## Exec (SubProto=7)

**Actions**
- `ActionCall = "call"`
- `ActionCallResp = "call_resp"`

**Payload types**
- `CallReq`
- `CallResp`
- `Message`

**Other constants**
- `PermExecCall = "exec.call"`

<!-- END GENERATED -->

## Notes（Manual）
- Auth：login/register 使用签名（ES256）+ nonce + timestamp（具体语义以实现侧为准；此处仅做提示）。


