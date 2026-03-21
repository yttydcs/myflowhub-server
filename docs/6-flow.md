flow 协议（SubProto=6）规范（草案）
=============================

范围
----
- `flow` 子协议用于在任意节点上保存/触发/调度一个“有向无环图（DAG）工作流”。
- 本协议只定义“工作流编排与调度”：
  - 工作流的**实际执行者**固定为接收 `flow.set` 的节点（该节点需要实现 `flow` 协议）。
  - DAG 节点的执行方式：对 DAG 进行拓扑排序并逐个执行（可失败节点可继续，不可失败节点失败则立即终止）。
- “跨节点的特殊能力调用”不由 `flow` 直接承载，使用 `exec` 子协议（SubProto=7）。

总览
----
- 控制帧编码：UTF-8 JSON，envelope 固定为 `{"action":"...","data":{...}}`
- 典型动作：
  - `set`：设置/更新工作流（需要权限 `flow.set`）
  - `delete`：删除工作流（需要权限 `flow.delete`）
  - `run`：手动触发一次运行（第一版可选）
  - `status`：查询运行状态（第一版可选）
- 触发器（当前）：支持 `interval` / `event` / `var_changed`。
  - `event`：由 `topicbus.publish` / `topicbus.received` 事件驱动，按 `event_mode` + `event_name`/`event_topic` 匹配。
  - `var_changed`：由 `varstore.changed` 与 `varstore.deleted` 事件驱动，按 `var_owner`/`var_name` 过滤。

权限
----
- 权限节点格式：`协议.action`
- 第一版最小权限：
  - `flow.set`：允许写入/更新工作流定义（落盘并生效）
  - `flow.delete`：允许删除工作流定义（删除时立即中断该 flow 的运行中 run）

HeaderTcp 与路由约定
--------------------
- SubProto 固定为 `6`（预留给 `flow`）。
- `TargetID` 仍由核心路由自动转发到目标节点；本协议的“逐级授权”不依赖 `TargetID=0` 等特殊语义。
- Major 约定（统一框架规则）：
  - 请求帧（`set/delete/run/status/list/get`）：`MajorCmd`（逐跳可见，需要进入 handler 参与裁决/执行）。
  - 响应帧（`*_resp`）：`MajorOKResp`（按 `TargetID` 由 Core 快速转发；中间节点不需要进 handler 转发）。
  - 失败响应仍使用 `MajorOKResp`，错误通过 payload 的 `code/msg` 表达。

逐级授权模型（统一规则）
----------------------
前提：
- 子节点无条件信任父节点；父节点对子节点有绝对控制权。
- 因此“父→子”的控制请求属于“向下控制”，可视为天然可控（具体是否免检由各协议定义）。

本协议 `flow.set` / `flow.delete` 的逐级授权规则：
- 请求方将 `flow.set` 或 `flow.delete` 发给任意实现了 `flow` 的节点（称为“执行者/Executor”，它负责落盘与调度状态变更）。
- 执行者**不直接裁决权限**，而是将请求逐级上送，直到某一级节点能够在其子树内完成裁决并返回结果。
- 裁决节点（通常为最近公共祖先或其上级）对动作做权限校验：
  - 校验主体建议为“请求方节点”（`origin_node`）。
  - `set` 权限名固定为 `flow.set`；`delete` 权限名固定为 `flow.delete`。
  - 通过：返回对应 `*_resp(code=1)`；拒绝：返回对应 `*_resp(code=403)`。
- 执行者仅在收到 `code=1` 后才执行落盘变更：
  - `set`：写入/更新定义并生效。
  - `delete`：删除定义并立即中断该 flow 的运行中 run。

> 注：上送转发的具体实现可参考 `varstore` 的 `assist_*` 思路；但 `flow.set` 需要一个明确的最终响应（允许/拒绝），以满足“是否生效”的语义。

控制帧格式
----------
- 载荷编码：JSON(UTF-8)
- JSON envelope：`{"action":"...","data":{...}}`
- 统一字段建议：
  - `req_id`：UUID 字符串（用于幂等/关联响应）
  - `origin_node`：最初发起请求的节点 ID（用于权限主体）
  - `executor_node`：工作流执行者节点 ID（通常为接收方；可用于一致性校验）

### action=set（设置/更新工作流，权限：flow.set）

请求 `data`：
- `req_id`：UUID（必填）
- `origin_node`：uint32（可选；默认取 `hdr.SourceID`）
- `executor_node`：uint32（可选；默认取“接收此请求的节点 ID”）
- `flow_id`：UUID（必填）
- `name`：string（可选）
- `trigger`：object（必填）
  - `type`：`"interval"` | `"event"` | `"var_changed"`
  - `interval` 触发：
    - `every_ms`：uint64（必填，>0）
  - `event` 触发（匹配 TopicBus 发布事件）：
    - `event_mode`：`publish` | `received` | `any`（可选，默认 `publish`）
    - `event_name`：string（可选）
    - `event_topic`：string（可选）
    - 约束：`event_name` 与 `event_topic` 不能同时为空
  - `var_changed` 触发（匹配变量变化/删除事件）：
    - `var_owner`：uint32（可选，0 表示不过滤 owner）
    - `var_name`：string（可选，空表示不过滤 name）
- `graph`：object（必填）
  - `nodes`：array
  - `edges`：array

响应 `action=set_resp`，`data`：
- `req_id`：回显
- `code`：`1` 成功；`400/403/404/500` 等失败
- `msg`：可选错误说明
- `flow_id`：回显

### action=delete（删除工作流，权限：flow.delete）

请求 `data`：
- `req_id`：UUID（必填）
- `origin_node`：uint32（可选；默认取 `hdr.SourceID`）
- `executor_node`：uint32（可选；默认取“接收此请求的节点 ID”）
- `flow_id`：UUID（必填）

删除语义（第一版）：
- 删除成功后，执行者必须立即删除对应工作流定义（例如 `./flows/<flow_id>.json`）。
- 若该 `flow_id` 存在运行中 run，执行者必须立即中断/取消这些 run；不得等待当前节点执行完成。
- 删除是终态变更；删除后该 `flow_id` 不再参与后续触发与调度，直到再次 `set`。

响应 `action=delete_resp`，`data`：
- `req_id`：回显
- `code`：`1` 成功；`400/403/404/500` 等失败
- `msg`：可选错误说明
- `flow_id`：回显

#### DAG 结构

`graph.nodes[]`（每个 DAG 节点）：
- `id`：string（必填，图内唯一）
- `kind`：string（必填）
  - `"local"`：由执行者节点本地执行（例如调用既有协议、或执行者内置逻辑）
  - `"exec"`：通过 `exec.call` 调用网络中的“特殊能力”（SubProto=7）
- `allow_fail`：bool（可选，默认 false）
  - false：不可接受失败（失败/超时/重试耗尽 => 立即终止整个 flow）
  - true：可接受失败（记录失败并继续）
- `retry`：int（可选，默认 1；表示失败后最多额外重试次数）
- `timeout_ms`：int（可选，默认 3000）
- `spec`：object（必填，按 kind 区分）
  - kind=local：由执行者自定义（建议用已注册的 `namespace::method` 或直接声明“调用某协议动作”的参数）
  - kind=exec：
    - `target`：uint32（必填，目标节点）
    - `method`：string（必填，形如 `namespace::method`）
    - `args`：object（可选）

`graph.edges[]`：
- `from`：string（必填，节点 id）
- `to`：string（必填，节点 id）

执行语义（第一版）：
- 对 `graph` 做拓扑排序；按顺序逐个执行 `nodes`。
- 每个节点执行采用 `timeout_ms` 控制单次尝试；失败时按 `retry` 进行重试。
- 失败处理：
  - `allow_fail=false`：立刻结束，flow=failed
  - `allow_fail=true`：记录节点失败，继续执行后续节点

持久化与目录
------------
- 默认目录：运行目录下 `./flows`
  - 工作流定义：`./flows/<flow_id>.json`
  - （可选）运行记录：`./flows/runs/<run_id>.json`
- 建议配置项：
  - `flow.base_dir`：默认 `./flows`
  - `flow.max_flows` / `flow.max_concurrent_runs`：可选

错误码建议
----------
- `1`：ok
- `400`：invalid request / invalid graph
- `403`：permission denied（`flow.set` / `flow.delete`）
- `404`：not found（例如无法找到可裁决节点/无父节点/删除目标 flow 不存在）
- `409`：conflict（可选，例如已有运行中的同名 flow 不允许覆盖）
- `500`：internal error

