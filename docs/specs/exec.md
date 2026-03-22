exec 协议（SubProto=7）规范（草案）
=============================

范围
----
- `exec` 子协议用于在网络中调用“特殊能力”（将来可用于插件系统，例如调用第三方服务商 Web API）。
- **注意**：`exec.call` 不是把 `flow` 交给别的节点执行；`flow` 的执行者永远是接收 `flow.set` 的节点。
- 权限：`exec.call`、`exec.cap.sync`、`exec.cap.query` 均可独立校验。

总览
----
- 控制帧编码：UTF-8 JSON，envelope 固定为 `{"action":"...","data":{...}}`
- 典型动作：
  - `call`：请求目标节点执行一个已注册的 `namespace::method`
  - `call_resp`：执行结果响应
  - `cap_snapshot/cap_upsert/cap_withdraw/cap_heartbeat`：能力注册中心逐级同步（草案）
  - `cap_query/cap_query_resp`：能力发现查询（草案）

权限
----
- 权限节点格式：`协议.action`
- 当前固定：
  - `exec.call`：允许“使用网络中的特殊能力”
  - `exec.cap.sync`：允许能力同步（snapshot/upsert/withdraw/heartbeat）
  - `exec.cap.query`：允许查询能力聚合索引
- 默认仍受 auth 默认权限策略影响（若默认 `*` 则等价放开）。

HeaderTcp 与路由约定
--------------------
- SubProto 固定为 `7`（预留给 `exec`）。
- `TargetID` 由核心路由自动转发到目标节点。
- 本协议依赖“逐级上送直到可向下转发”的一致性路由语义（见下文）。
- Major 约定（统一框架规则）：
  - 请求帧（`call`）：`MajorCmd`（逐跳可见，需要进入 handler 参与裁决/执行/转发）。
  - 响应帧（`call_resp`）：`MajorOKResp`（按 `TargetID` 由 Core 快速转发；中间节点不需要进 handler 转发）。
  - 能力同步请求帧（`cap_*`）：`MajorCmd`（逐级上送/逐级聚合）。
  - 能力同步响应（`cap_sync_resp`）与查询响应（`cap_query_resp`）：`MajorOKResp`。
  - 失败响应仍使用 `MajorOKResp`，错误通过 payload 的 `code/msg` 表达。

逐级上送与裁决（downstream 判定）
------------------------------
前提：
- 子节点无条件信任父节点；父节点对子节点有绝对控制权。

对 `exec.call` 的规则（你已确认）：
- `exec.call` 本质是“调用网络能力”，请求方（通常为 `flow` 执行者）应先发送给自己的**直接父节点**，再逐级向上。
- 任一节点收到 `call/assist_call` 后：
  1) 判断目标 `target_node` 是否在自己的子树内：
     - 若到 `target_node` 的下一跳连接属于 **downstream（子连接）**，则视为目标在本子树内，当前节点具备对该子树的控制权。
     - 否则（下一跳为 upstream 或不可达），继续向上转发给父节点。
  2) 当“可向下转发”成立时，当前节点即为“裁决/转发点”：
     - 若目标位于当前节点的 **downstream**，则可以直接向下转发（父控子语义）。
     - 权限校验：校验主体为 `executor_node` 是否具备 `exec.call`。
       - 通过：**直接转发**请求到目标（转发即同意，不需要先回 allow）。
       - 拒绝：直接回 `call_resp(code=403)` 给 `executor_node`。

特殊情况（无需权限判断）
----------------------
- 若目标节点位于“请求节点的 downstream”（即 `flow` 执行者对目标是父→子方向可控），则可由 `flow` 执行者直接向目标发送 `exec.call`，不经过逐级授权，也不做 `exec.call` 权限校验。
  - 该规则是“父控子”语义的直接体现：子节点无条件信任父节点。

> 注：上述“免检”仅描述 `exec.call` 这一能力的权限模型；目标节点仍可对 `namespace::method` 做入参校验/限流/失败返回。

控制帧格式
----------
- 载荷编码：JSON(UTF-8)
- JSON envelope：`{"action":"call","data":{...}}` / `{"action":"call_resp","data":{...}}`
- 统一字段建议：
  - `req_id`：UUID（用于匹配一次调用；同时可用于幂等/去重）
  - `executor_node`：发起这次调用的 `flow` 执行者节点（权限主体）

### action=call（执行特殊方法，权限：exec.call）

请求 `data`：
- `req_id`：UUID（必填）
- `executor_node`：uint32（必填）
- `target_node`：uint32（必填）
- `method`：string（必填，形如 `namespace::method`）
- `args`：object（可选）
- `timeout_ms`：int（可选，默认 3000；建议由 `flow` 节点的 timeout 传入）

响应 `action=call_resp`，`data`：
- `req_id`：回显
- `code`：`1` 成功；`400/403/404/408/500` 等失败
- `msg`：可选错误说明
- `executor_node`：回显（可选）
- `target_node`：回显（可选）
- `method`：回显（可选）
- `result`：object（可选，成功时返回）

响应投递建议：
- `TargetID` 设置为 `executor_node`，依赖核心路由将 `call_resp` 直接回到执行者。
- `SourceID` 建议为“执行方法的节点”（最终目标节点）。

方法注册（插件系统接口建议）
--------------------------
- 每个节点维护一个方法注册表：`method(string) -> handler(ctx, args) -> result`
- `method` 使用 `namespace::method` 作为唯一标识。
- 第一版不规定注册机制（静态内置或动态插件均可），但约束：
  - 未注册方法返回 `code=404`（method not found）
  - 入参不合法返回 `code=400`
  - 超时返回 `code=408`

错误码建议
----------
- `1`：ok
- `400`：invalid request / invalid args
- `403`：permission denied（`exec.call`）
- `404`：target or method not found / not reachable
- `408`：timeout
- `429`：too many requests（可选，限流）
- `500`：internal error

能力注册中心（挂载在 exec，逐级注册草案）
---------------------------------------
目标：
- 在不新增 subproto 的前提下，复用 `exec` 承载能力注册中心。
- 采用“子到父逐级同步 + 父到子树逐级聚合”模型，保证断父后子树自治可用。

核心模型：
- 每个节点维护：
  - `local_caps`：本节点自有能力
  - `child_caps[child]`：每个直连子节点上报的能力快照
  - `subtree_index`：本节点聚合索引（供本节点/子树查询）
- 连接断开（`conn.closed`）时清理对应 child 快照，并触发上行增量撤销同步。
- 子节点重连父节点后，发送 `cap_snapshot`（新 `epoch`）覆盖旧状态。

能力描述（CapabilityDescriptor）：
- `provider_node`：真正执行 method 的节点
- `method`：`namespace::method`
- `version`：能力版本（可选）
- `input_schema/output_schema`：建议 JSON Schema 子集（可选）
- `default_timeout_ms`：默认超时（可选）
- `permissions`：调用该能力的权限集合（可选）
- `tags`：扩展标签（可选）

同步动作（逐级上送）：
- `cap_snapshot`：全量上报（用于初次上线/重连）
  - `req_id? from_node epoch lease_ms caps[]`
- `cap_upsert`：增量新增/更新
  - `req_id? from_node epoch lease_ms caps[]`
- `cap_withdraw`：增量撤销
  - `req_id? from_node epoch keys[]`
- `cap_heartbeat`：续租
  - `req_id? from_node epoch lease_ms`
- 统一响应 `cap_sync_resp`
  - `req_id? code msg from_node epoch applied responder_node`

上行发送策略（当前实现）：
- 首次上行或父节点切换：发送 `cap_snapshot`（全量）
- 稳态变更：发送 `cap_upsert`/`cap_withdraw`（增量）
- 无差异：按租约窗口发送 `cap_heartbeat` 续租
- 若收到父节点 `cap_sync_resp` 且 `code in {404,409}` 或 `>=500`，本节点会清空上行缓存并触发一次全量 `cap_snapshot` 重同步

查询动作（本地优先，必要时上送）：
- `cap_query`：
  - `req_id requester_node? method? prefix? provider_node? limit? include_schema?`
- `cap_query_resp`：
  - `req_id code msg responder_node total routes[]`
  - `routes[]` 含 `provider_node`、`via_node`、`method`、`version`、`lease_expire_at` 等字段

与 `call` 的关系：
- `call` 保持现状（不破坏既有路由与权限语义）。
- 后续可在 executor 侧先 `cap_query` 再做 provider 选择与 `call` 下发。

