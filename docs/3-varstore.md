VarStore 协议（SubProto=3）新规范
=================================

总览
----
- 面向“节点变量”缓存与查询，键 = `owner_node_id:name`。
- 角色：**请求方**（发起 set/get/list/revoke）、**owner**（变量所属节点）、**祖先链**（请求所在节点→父→…→根）。
- 目标：所有读写/撤销均沿祖先链逐跳上送命中，逐跳回传；不再向所有子节点广播，避免性能开销。

数据模型
--------
- name：大小写敏感，仅字母/数字/下划线。
- owner：uint32，默认为请求方 SourceID。
- visibility：`public`（他人可读/改）或 `private`（仅 owner 可读/改）。
- type：字符串标签，空默认为 `"string"`。

动作与载荷（JSON：`{"action": "...", "data": {...}}`）
-----------------------------
- 请求方 -> 直接父节点：`set/get/list/revoke`（原始指令）。
- 父节点向上协助：`assist_set/assist_get/assist_list/assist_revoke`。
- 仅 set、revoke 有 `_up`（祖先链缓存同步）：`up_set/up_revoke`。
- 订阅/退订：`subscribe/assist_subscribe/subscribe_resp/assist_subscribe_resp`；退订 `unsubscribe/assist_unsubscribe`（可选返回 subscribe_resp）。
- 变更广播：`var_changed`（包含 name/owner/value/visibility/type）、`var_deleted`（仅 name/owner）。
- 响应：`set_resp/get_resp/list_resp/revoke_resp`（以及 assist_*_resp）。  
- 通知 owner：`notify_set`、`notify_revoke`（Cmd 帧，途经链路尽量缓存）。
- 示例：
  - set：`{"name":"temp","value":"22.5","visibility":"public","type":"string","owner":5}` → `{"code":1,"name":"temp","owner":5,"visibility":"public","type":"string"}`
  - get：`{"name":"temp","owner":5}` → `{"code":1,"name":"temp","value":"22.5","owner":5,"visibility":"public","type":"string"}`
  - list：`{"owner":5}` → `{"code":1,"owner":5,"names":["a","b"]}`
  - revoke：`{"name":"temp","owner":5}` → `{"code":1,"name":"temp","owner":5}`
  - notify_set/revoke：携带 name/owner/可选 value/type/visibility（更新时才携带），沿链路缓存。
  - 旧文档中的概要已从 core.md 移至此处。

路由与头部
----------
- Major：命令/状态类建议用 `MajorCmd`；响应可用 `MajorOKResp` 或 `MajorCmd`，需逐跳可见的用 `MajorCmd`。
- SubProto 固定为 3。转发时保留原始 `SourceID`，根据场景调整 `TargetID`。
- `TargetID=0` 在核心中意味着“广播子节点，不上行父链”，不要用 0 表示“上送父节点”。上送请显式填父节点/目标 Hub 的 NodeID。

处理与链路规则
--------------
- 判定子树：用路由表/连接索引 + localID 判断 owner 是否在当前子树（包含自己）。
- 权限：关键执行节点（通常是请求方与 owner 的最近公共祖先，或更上层拥有完整信息的节点）做 `var.private_set` / `var.revoke` / `var.subscribe` 判定，不提前拒绝。
- set/revoke（修改类）：
  1) 请求方 -> 直接父：发送 `set`/`revoke`（Target=父）。
  2) 父节点查子树是否含 owner。若否且有父：向上发 `assist_*`；若无父仍未找到，回 `*_resp` not found 给请求方。
  3) 命中的关键节点：执行业务并缓存，向上发 `up_*` 让祖先缓存；向请求方发 `*_resp`（若 requester≠owner）；向 owner 发 `notify_set`/`notify_revoke`（requester=owner 则只通知一次）。`up_*`/`notify_*` 经途节点尽可能缓存。
  4) 上层收到 `assist_*` 复用同逻辑；收到 `up_*` 先缓存，再继续向上转发（默认透传，可挂接校验/签名）。
- get/list（查询类）：
  1) 请求方 -> 父：发送 `get`/`list`。
  2) 父若子树含 owner 且本地有缓存：直接回 `*_resp`（Target=requester）。若未缓存（即便在子树）或不在子树且有父：向上 `assist_*`；无父则回 not found。
  3) 上层收到 `assist_*` 重复步骤 2。无 `_up`。
- 缓存策略：请求方是否缓存自定；`up_*` 与 `notify_*` 路径上的节点尽量缓存；`*_resp` 不强制缓存。

订阅/变更推送
--------------
- 订阅载荷：`{"name":"...","owner":N}`，owner 必填；订阅者节点为请求帧的 SourceID（assist_* 场景沿用原订阅者）。  
- 权限：公共变量无限制；私有变量需订阅者=owner 或具备 `var.subscribe` 权限。
- 存在性：订阅必须命中真实变量。若本地未命中，沿父链 `assist_subscribe` 逐级向上，根节点最终返回 `*_subscribe_resp`（code=1/4 等）；收到 code=1 后逐级记录订阅。
- 退订：`unsubscribe/assist_unsubscribe`，连接断开会自动退订；当本地及子节点均无订阅者时自动向上 `assist_unsubscribe`。
- 触发源：`up_set`、关键节点发送的 `notify_set`、`up_revoke`、`notify_revoke` 会触发变更通知；owner 的专有 notify 仍保留，订阅推送不重复发送给 owner。
- 变更广播：关键节点向已订阅的子节点发送 `var_changed`；收到后同样向下游订阅者转发（不回原来源链路）。可见性从 public 变 private 视为删除，发送 `var_deleted` 并清理订阅。撤销时发送 `var_deleted` 并自动退订。

错误码约定
----------
- `1` 成功
- `2` 参数非法
- `3` 权限不足
- `4` 未找到

示例 payload
------------
- set：`{"action":"set","data":{"name":"sensor_a","value":"22.5","visibility":"public","type":"string"}}`
- get：`{"action":"get","data":{"name":"sensor_a"}}`
- list：`{"action":"list","data":{"owner":5}}`
- revoke：`{"action":"revoke","data":{"name":"sensor_a"}}`
- 成功 get_resp：`{"action":"get_resp","data":{"code":1,"name":"sensor_a","value":"22.5","owner":5,"visibility":"public","type":"string"}}`
- 未命中：`{"action":"get_resp","data":{"code":404,"msg":"not found"}}`

集成提示
--------
- Dispatcher 注册：`dp.RegisterHandler(handler.NewVarStoreHandler(logger))`。
- handler 需在 `TargetID` 与 `shouldForwardUp` 逻辑上遵循“祖先链上送、逐跳回传”的规则，避免使用 0 表示上送。

后续演进
--------
- 若需持久化，可在 `assist_set` 命中节点挂接存储。
- 若需多活 owner/分区路由，可扩展 owner->Hub 映射，在预路由阶段直接路由到 owner 所在 Hub。***
