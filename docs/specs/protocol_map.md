# Protocol Mapping（Client <-> Server）

> 本文档为“半自动文档”：
> - single source-of-truth：`protocol/*/types.go`
> - **不要手工修改** `<!-- BEGIN GENERATED -->` 与 `<!-- END GENERATED -->` 中间的内容
> - 更新方式：`go run ./cmd/protocolmapgen -write -out docs/protocol_map.md`

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
| 8 | Stream | `protocol/stream` |

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
- `ActionApproveRegister = "approve_register"`
- `ActionApproveRegisterResp = "approve_register_resp"`
- `ActionAssistLogin = "assist_login"`
- `ActionAssistLoginResp = "assist_login_resp"`
- `ActionAssistOffline = "assist_offline"`
- `ActionAssistQueryCred = "assist_query_credential"`
- `ActionAssistQueryCredResp = "assist_query_credential_resp"`
- `ActionAssistRegister = "assist_register"`
- `ActionAssistRegisterResp = "assist_register_resp"`
- `ActionAuthorityPolicySync = "authority_policy_sync"`
- `ActionGetPerms = "get_perms"`
- `ActionGetPermsResp = "get_perms_resp"`
- `ActionIssueRegisterPermit = "issue_register_permit"`
- `ActionIssueRegisterPermitResp = "issue_register_permit_resp"`
- `ActionListPendingRegisters = "list_pending_registers"`
- `ActionListPendingRegistersResp = "list_pending_registers_resp"`
- `ActionListRoles = "list_roles"`
- `ActionListRolesResp = "list_roles_resp"`
- `ActionLogin = "login"`
- `ActionLoginResp = "login_resp"`
- `ActionOffline = "offline"`
- `ActionPermsInvalidate = "perms_invalidate"`
- `ActionPermsSnapshot = "perms_snapshot"`
- `ActionRegister = "register"`
- `ActionRegisterResp = "register_resp"`
- `ActionRejectRegister = "reject_register"`
- `ActionRejectRegisterResp = "reject_register_resp"`
- `ActionRevoke = "revoke"`
- `ActionRevokeRegisterPermit = "revoke_register_permit"`
- `ActionRevokeRegisterPermitResp = "revoke_register_permit_resp"`
- `ActionRevokeResp = "revoke_resp"`
- `ActionUpLogin = "up_login"`
- `ActionUpLoginResp = "up_login_resp"`

**Payload types**
- `ApproveRegisterReq`
- `ApproveRegisterResp`
- `AuthorityPolicySyncData`
- `InvalidateData`
- `IssueRegisterPermitReq`
- `IssueRegisterPermitResp`
- `ListPendingRegistersReq`
- `ListPendingRegistersResp`
- `ListRolesReq`
- `LoginData`
- `Message`
- `OfflineData`
- `PendingRegisterInfo`
- `PermsQueryData`
- `QueryCredData`
- `RegisterData`
- `RejectRegisterReq`
- `RejectRegisterResp`
- `RespData`
- `RevokeData`
- `RevokeRegisterPermitReq`
- `RevokeRegisterPermitResp`
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
- `ActionCapHeartbeat = "cap_heartbeat"`
- `ActionCapQuery = "cap_query"`
- `ActionCapQueryResp = "cap_query_resp"`
- `ActionCapSnapshot = "cap_snapshot"`
- `ActionCapSyncResp = "cap_sync_resp"`
- `ActionCapUpsert = "cap_upsert"`
- `ActionCapWithdraw = "cap_withdraw"`

**Payload types**
- `CallReq`
- `CallResp`
- `CapHeartbeatReq`
- `CapQueryReq`
- `CapQueryResp`
- `CapSnapshotReq`
- `CapSyncResp`
- `CapUpsertReq`
- `CapWithdrawReq`
- `CapabilityDescriptor`
- `CapabilityKey`
- `CapabilityRoute`
- `Message`

**Other constants**
- `PermExecCall = "exec.call"`
- `PermExecCapQuery = "exec.cap.query"`
- `PermExecCapSync = "exec.cap.sync"`

## Stream (SubProto=8)

**Actions**
- `ActionAnnounce = "announce"`
- `ActionAnnounceConsumer = "announce_consumer"`
- `ActionAnnounceConsumerResp = "announce_consumer_resp"`
- `ActionAnnounceResp = "announce_resp"`
- `ActionConnect = "connect"`
- `ActionConnectResp = "connect_resp"`
- `ActionDisconnect = "disconnect"`
- `ActionDisconnectResp = "disconnect_resp"`
- `ActionGetConsumer = "get_consumer"`
- `ActionGetConsumerResp = "get_consumer_resp"`
- `ActionGetSource = "get_source"`
- `ActionGetSourceResp = "get_source_resp"`
- `ActionListConsumers = "list_consumers"`
- `ActionListConsumersResp = "list_consumers_resp"`
- `ActionListSources = "list_sources"`
- `ActionListSourcesResp = "list_sources_resp"`
- `ActionSignal = "signal"`
- `ActionSignalResp = "signal_resp"`
- `ActionSubscribe = "subscribe"`
- `ActionSubscribeResp = "subscribe_resp"`
- `ActionUnsubscribe = "unsubscribe"`
- `ActionUnsubscribeResp = "unsubscribe_resp"`
- `ActionWithdraw = "withdraw"`
- `ActionWithdrawConsumer = "withdraw_consumer"`
- `ActionWithdrawConsumerResp = "withdraw_consumer_resp"`
- `ActionWithdrawResp = "withdraw_resp"`

**Payload types**
- `AnnounceConsumerReq`
- `AnnounceConsumerResp`
- `AnnounceReq`
- `AnnounceResp`
- `ConnectReq`
- `ConnectResp`
- `ConsumerDescriptor`
- `DisconnectReq`
- `DisconnectResp`
- `GetConsumerReq`
- `GetConsumerResp`
- `GetSourceReq`
- `GetSourceResp`
- `ListConsumersReq`
- `ListConsumersResp`
- `ListSourcesReq`
- `ListSourcesResp`
- `Message`
- `SignalReq`
- `SignalResp`
- `SourceDescriptor`
- `StreamAckHeaderV1`
- `StreamDataHeaderV1`
- `SubscribeReq`
- `SubscribeResp`
- `UnsubscribeReq`
- `UnsubscribeResp`
- `WithdrawConsumerReq`
- `WithdrawConsumerResp`
- `WithdrawReq`
- `WithdrawResp`

**Other constants**
- `DataFlagConfig = 4`
- `DataFlagDiscontinuity = 8`
- `DataFlagEOS = 1`
- `DataFlagKeyframe = 2`
- `HeaderVersionV1 = 1`
- `KindAck = 0x03`
- `KindCtrl = 0x01`
- `KindData = 0x02`
- `ModeBounded = "bounded"`
- `ModeLive = "live"`
- `PermStreamConnect = "stream.connect"`
- `PermStreamConsume = "stream.consume"`
- `PermStreamPublish = "stream.publish"`
- `PermStreamSubscribe = "stream.subscribe"`
- `SignalOpCustom = "custom"`
- `SignalOpKeyframeRequest = "keyframe_request"`
- `SignalOpMetadataUpdate = "metadata_update"`
- `SignalOpPause = "pause"`
- `SignalOpResume = "resume"`
- `StreamKindCustom = "custom"`
- `StreamKindMusic = "music"`
- `StreamKindText = "text"`
- `StreamKindVideo = "video"`
- `UnitModeChunk = "chunk"`
- `UnitModeFrame = "frame"`

<!-- END GENERATED -->

## Notes（Manual）
- Management（Nodes）：
  - `list_nodes`：仅返回 downstream children（直连子节点）；不包含 upstream parent link。
  - `list_subtree`：返回 `list_nodes` 的结果 + self（不递归；更接近 “direct + self”）。
  - `nodes[].has_children`：best-effort hint（可能缺失/为 false），客户端应以实际 `list_nodes` 结果为准。
- Auth：login/register 使用签名（ES256）+ nonce + timestamp（具体语义以实现侧为准；此处仅做提示）。

