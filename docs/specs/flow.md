flow 协议（SubProto=6）规范
===========================

范围
----

- `flow` 子协议用于在任意节点上保存、触发和调度一个 DAG 工作流。
- 本协议同时定义：
  - 工作流定义与调度动作（`set/delete/run/cancel_run/status/detail/list_runs/list/get`）
  - 节点执行语义
  - 节点间数据传递模型
  - 单次 run 局部变量模型
- 第一版数据流增强后的正式节点类型为：
  - `call`
  - `compose`
  - `set_var`
- `exec`（SubProto=7）继续负责远程方法调用与权限裁决；`flow` 不复制 `exec.call` 的路由与权限模型。

总览
----

- 控制帧编码：UTF-8 JSON，envelope 固定为 `{"action":"...","data":{...}}`
- 典型动作：
  - `set`：设置/更新工作流（需要权限 `flow.set`）
  - `delete`：删除工作流（需要权限 `flow.delete`）
  - `run`：手动触发一次运行（需要权限 `flow.run`）
  - `cancel_run`：取消指定运行（需要权限 `flow.run`）
  - `status`：查询运行状态摘要（需要权限 `flow.read`）
  - `detail`：查询指定 run / node 的结果详情（需要权限 `flow.read`）
  - `list_runs`：列出指定 flow 当前保留窗口内的运行摘要（需要权限 `flow.read`）
  - `list`：列出执行者当前已知的工作流摘要（需要权限 `flow.read`）
  - `get`：读取指定工作流定义（需要权限 `flow.read`）
- 触发器（当前）：支持 `interval` / `event` / `var_changed`
  - `event`：由 `topicbus.publish` / `topicbus.received` 事件驱动
  - `var_changed`：由 `varstore.changed` / `varstore.deleted` 事件驱动

权限
----

- 权限节点格式：`协议.action`
- 当前版本稳定权限：
  - `flow.set`
  - `flow.delete`
  - `flow.run`
  - `flow.read`
- 动作映射：
  - `set` -> `flow.set`
  - `delete` -> `flow.delete`
  - `run` / `cancel_run` -> `flow.run`
  - `status` / `detail` / `list_runs` / `list` / `get` -> `flow.read`
- `flow::run` capability descriptor 也必须声明 `flow.run`，以便 `exec.call`、`cap_query` 与 `cap_snapshot` 看到一致权限要求。

HeaderTcp 与路由约定
--------------------

- SubProto 固定为 `6`
- Major 约定：
  - 请求帧（`set/delete/run/cancel_run/status/detail/list_runs/list/get`）：`MajorCmd`
  - 响应帧（`*_resp`）：`MajorOKResp`
  - 失败响应也使用 `MajorOKResp`，错误通过 payload 的 `code/msg` 表达

逐级授权模型
------------

本协议的逐级授权适用于 `flow.set` / `flow.delete` / `flow.run` / `flow.read`：

- 请求方将请求发给任意实现了 `flow` 的节点（称为执行者）
- 执行者不直接裁决权限，而是逐级上送，直到某一级可以在其子树内完成裁决
- 裁决节点以 `origin_node` 为权限主体：
  - `set` 对应 `flow.set`
  - `delete` 对应 `flow.delete`
  - `run` / `cancel_run` 对应 `flow.run`
  - `status` / `detail` / `list_runs` / `list` / `get` 对应 `flow.read`
- 执行者仅在收到允许结果后执行本地持久化和调度状态变更

控制帧格式
----------

- 统一 envelope：`{"action":"...","data":{...}}`
- 统一字段：
  - `req_id`：请求关联 ID
  - `origin_node`：最初发起请求的节点 ID
  - `executor_node`：工作流执行者节点 ID
  - `flow_id`：UUID 字符串，必须通过 UUID 校验

动作契约
--------

### action=set

请求 `data`：

- `req_id`：UUID（必填）
- `origin_node`：uint32（可选；默认取 `hdr.SourceID`）
- `executor_node`：uint32（可选；默认取“接收此请求的节点 ID”）
- `flow_id`：UUID（必填）
- `name`：string（可选）
- `max_active_runs`：int（可选）
  - 省略表示保持 legacy 兼容行为
  - `0` 表示不限制活动 run 数
  - `>0` 表示统一的活动 run 上限
- `trigger`：object（必填）
- `graph`：object（必填）

响应 `action=set_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/403/404/500`
- `msg`：可选错误说明
- `flow_id`：回显

### action=delete

请求 `data`：

- `req_id`：UUID（必填）
- `origin_node`：uint32（可选）
- `executor_node`：uint32（可选）
- `flow_id`：UUID（必填）

删除语义：

- 删除成功后，执行者必须立即删除工作流定义
- 若该 `flow_id` 存在运行中 run，执行者必须立即中断/取消这些 run
- 删除后该 `flow_id` 不再参与后续触发和调度，直到再次 `set`
- 若启用 run archive，当前 retained window 内的已结束 run 仍可继续通过 `status/detail/list_runs` 查询，直到超出 retained window 被回收

响应 `action=delete_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/403/404/500`
- `msg`：可选错误说明
- `flow_id`：回显

### action=run

请求 `data`：

- `req_id`：UUID（必填）
- `origin_node`：uint32（可选）
- `executor_node`：uint32（可选）
- `flow_id`：UUID（必填）

运行语义：

- `run` 仅触发一次即时执行，不修改定义和触发器
- 成功时返回新的 `run_id`
- 若该 flow 的 effective active-run 上限已满，则返回 `409`
- 权限：`flow.run`

响应 `action=run_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/403/404/409/500`
- `msg`：可选错误说明
- `flow_id`：回显
- `run_id`：成功时必填

### action=cancel_run

请求 `data`：

- `req_id`：UUID（必填）
- `origin_node`：uint32（可选）
- `executor_node`：uint32（可选）
- `flow_id`：UUID（必填）
- `run_id`：UUID（必填）

取消语义：

- `cancel_run` 只中断指定 run，不删除 flow 定义
- 仅允许命中该 `flow_id` 的活动 run
- `run_id` 不存在或不属于该 `flow_id` 时，返回 `404`
- run 已处于 `succeeded` / `failed` / `cancelled` 时，返回 `409`
- 成功取消后，`status` 必须返回 `cancelled`；`detail` 查询相关节点时应能体现 run 已被取消
- 权限：`flow.run`

响应 `action=cancel_run_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/403/404/409/500`
- `msg`：可选错误说明；成功时可携带取消原因
- `executor_node`：实际执行者节点 ID
- `flow_id`：回显
- `run_id`：回显
- `status`：成功时为 `cancelled`；`409` 时回显已有终态

### action=status

请求 `data`：

- `req_id`：UUID（必填）
- `origin_node`：uint32（可选）
- `executor_node`：uint32（可选）
- `flow_id`：UUID（必填）
- `run_id`：UUID（可选）

状态语义：

- 若提供 `run_id`，查询该次运行摘要
- 若未提供 `run_id`，查询该 `flow_id` 的最近一次运行
- `status` 返回摘要，不默认包含完整节点结果
- 权限：`flow.read`

响应 `action=status_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/403/404/500`
- `msg`：可选错误说明；若运行被取消，可携带取消原因
- `executor_node`：实际执行者节点 ID
- `flow_id`：命中的工作流 ID
- `run_id`：命中的运行 ID
- `status`：`queued` | `running` | `succeeded` | `failed` | `cancelled`
- `nodes`：array
  - `id`：节点 ID
  - `status`：节点状态摘要
  - `code`：节点执行结果码
  - `msg`：可选错误说明

### action=detail

请求 `data`：

- `req_id`：UUID（必填）
- `origin_node`：uint32（可选）
- `executor_node`：uint32（可选）
- `flow_id`：UUID（必填）
- `run_id`：UUID（可选）
  - 为空时，默认命中该 `flow_id` 的最近一次运行
- `node_id`：string（必填）
- `path`：JSON Pointer（可选）
  - 为空时读取节点根结果

详情语义：

- `detail` 是重数据查询接口，与 `status` 分离
- 第一版只查询单个节点的结果详情，不返回完整局部变量视图
- 若 `path` 非空，则在节点根结果上应用 JSON Pointer
- 若 run / node / path 不存在，返回 `404`
- 若 `path` 非法，返回 `400`
- 节点自身状态失败时，仍允许查询其节点状态摘要；仅在请求结果路径不存在时返回 `404`
- 若该 run 已被取消，响应 `msg` 可携带取消原因；若命中的节点在取消时处于活动态，其节点状态应反映 `cancelled`
- 权限：`flow.read`

响应 `action=detail_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/403/404/500`
- `msg`：可选错误说明
- `executor_node`：实际执行者节点 ID
- `flow_id`：命中的工作流 ID
- `run_id`：命中的运行 ID
- `path`：回显命中的结果路径；空表示根结果
- `node`：object
  - `id`：节点 ID
  - `status`：节点状态摘要
  - `code`：节点执行结果码
  - `msg`：可选节点错误说明
- `result`：命中的 JSON 值

### action=list_runs

请求 `data`：

- `req_id`：UUID（必填）
- `origin_node`：uint32（可选）
- `executor_node`：uint32（可选）
- `flow_id`：UUID（必填）
- `limit`：uint32（可选）
  - `0` / 省略表示返回当前保留窗口内全部 run

历史语义：

- `list_runs` 返回指定 `flow_id` 当前保留窗口内的 run 摘要
- 返回顺序必须为最新到最旧
- 若 flow 已删除但保留窗口内仍存在该 `flow_id` 的 run，执行者仍可返回这些 retained run
- 若既没有活动定义也没有 retained run，返回 `404`
- `list_runs` 只返回 run 摘要，不返回完整节点结果
- 权限：`flow.read`

响应 `action=list_runs_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/403/404/500`
- `msg`：可选错误说明
- `executor_node`：实际执行者节点 ID
- `flow_id`：回显
- `runs`：array，按最新到最旧排序
  - `run_id`：运行 ID
  - `status`：`queued` | `running` | `succeeded` | `failed` | `cancelled`
  - `started_at_ms`：开始时间，Unix 毫秒
  - `ended_at_ms`：结束时间，Unix 毫秒；未结束可省略
  - `msg`：可选说明；取消原因等

### action=list

请求 `data`：

- `req_id`：UUID（必填）
- `origin_node`：uint32（可选）
- `executor_node`：uint32（可选）

读取语义：

- `list` 只返回当前执行者已知的 flow 摘要
- 权限：`flow.read`

响应 `action=list_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/403/500`
- `msg`：可选错误说明
- `executor_node`：实际执行者节点 ID
- `flows`：array
  - `flow_id`
  - `name`
  - `every_ms`
  - `last_run_id`
  - `last_status`

### action=get

请求 `data`：

- `req_id`：UUID（必填）
- `origin_node`：uint32（可选）
- `executor_node`：uint32（可选）
- `flow_id`：UUID（必填）

读取语义：

- `get` 返回指定 flow 的完整定义
- 权限：`flow.read`

响应 `action=get_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/403/404/500`
- `msg`：可选错误说明
- `executor_node`：实际执行者节点 ID
- `flow_id`：回显
- `name`：工作流名称（可选）
- `max_active_runs`：int（可选）
- `trigger`：触发器定义
- `graph`：完整 DAG 定义

触发器定义
----------

`trigger`：

- `type`：`interval` | `event` | `var_changed`
- 共享字段：
  - `dedup_window_ms`：int（可选，默认 `0`）
    - `0` / 省略表示关闭 trigger dedup
    - 仅 `event` / `var_changed` 支持 `>0`
- `interval`
  - `every_ms`：uint64，必填，`>0`
  - `dedup_window_ms` 仅允许省略或 `0`
- `event`
  - `event_mode`：`publish` | `received` | `any`
  - `event_name`：string（可选）
  - `event_topic`：string（可选）
  - 约束：`event_name` 与 `event_topic` 不能同时为空
  - 若 `dedup_window_ms>0`，则同一 flow 在窗口内遇到相同规范化事件上下文（`mode/topic/name/payload/ts`）时跳过本次启动
- `var_changed`
  - `var_owner`：uint32（可选，`0` 表示不过滤 owner）
  - `var_name`：string（可选，空表示不过滤 name）
  - 若 `dedup_window_ms>0`，则同一 flow 在窗口内遇到相同规范化变量变更上下文（`owner/name/op`）时跳过本次启动

图与节点模型
------------

`graph.nodes[]`：

- `id`：string（必填，图内唯一）
- `kind`：string（必填）
  - 新写入契约：`"call"` | `"compose"` | `"set_var"`
- `allow_fail`：bool（可选，默认 `false`）
- `retry`：int（可选，默认 `1`）
- `retry_backoff_ms`：int（可选，默认 `0`）
- `timeout_ms`：int（可选，默认 `3000`）
- `spec`：object（必填）

`graph.edges[]`：

- `from`：string（必填）
- `to`：string（必填）

写入契约与兼容边界：

- 新的 `set` 请求不得再写入 `local` / `exec`
- 运行期仍允许解释历史存量 `local` / `exec` 数据，以便旧落盘 flow 继续运行

数据流执行模型
--------------

单次 run 在执行期维护一个内部 `RunContext`，至少包含：

- `flow_id`
- `run_id`
- `executor_node`
- `trigger`
- `nodes.<node_id>`
  - `status`
  - `code`
  - `msg`
  - `result`
- `vars`
  - `<name>`
    - `value`
    - `writer_node_id`

`RunContext` 仅在运行时和调试链路中使用，不默认通过 `status_resp` 全量返回。

触发上下文规范化
----------------

执行器应在 run 开始时把触发源统一映射为可引用的 trigger 上下文：

- `interval`
  - 至少包含触发时间
- `event`
  - 至少包含 `mode`、`topic`、`name`、`payload`
- `var_changed`
  - 至少包含 `owner`、`name`、`op`

节点可以通过输入绑定显式消费这些 trigger 字段。

输入绑定模型
------------

输入绑定采用结构化对象，而不是自由模板字符串。

`InputBinding`：

- `to`：JSON Pointer，表示写入目标
- `source`：object，表示读取来源
- `required`：bool（可选，默认 `false`）

`source.kind` 首批支持：

- `node_result`
  - `node_id`：被引用节点 ID
  - `path`：可选 JSON Pointer，默认根结果
- `trigger`
  - `path`：可选 JSON Pointer，默认根 trigger 上下文
- `flow_meta`
  - `field`：当前仅允许 `flow_id`
- `run_meta`
  - `field`：当前仅允许 `run_id`
- `flow_var`
  - `name`：局部变量名（必填）
  - `path`：可选 JSON Pointer，默认根变量值

局部变量读取约束：

- 第一版通过 `source.kind=flow_var` 读取局部变量，不新增独立 `get_var` 节点
- 这样可以复用现有 `args_template/template + inputs` 物化路径，避免为“纯读取”再引入额外节点类型和结果包装层

引用规则：

- 只允许引用祖先节点
- 不允许引用当前节点或未来节点
- `to` 必须是合法 JSON Pointer
- `required=true` 且来源不存在时，当前节点必须失败
- `flow_var` 仅允许引用可唯一解析到祖先 `set_var` 写入者的变量名
- 同一路径上后续 `set_var` 可覆盖更早的同名值；若当前节点的祖先子图中存在多个不可唯一判定先后的同名写入者，则 `set` 阶段必须判定为歧义并拒绝保存

节点类型：call
--------------

`kind=call` 的正式写入 spec：

- `method`：string（必填，形如 `namespace::method`）
- `target`：uint32（可选）
  - `0` / 省略 / 本节点 ID：本地调用
  - 其他节点 ID：通过 `exec.call` 发起远程调用
- `args_template`：object（可选，默认 `{}`）
- `inputs`：`InputBinding[]`（可选）
- `_ui`：编辑器布局元数据（可选，不参与执行语义）

执行语义：

1. 以 `args_template` 为基础物化输入 JSON
2. 按声明顺序应用 `inputs`
3. 得到最终调用参数
4. `target` 为空、`0` 或等于本节点时：
   - 先查本地方法
   - 再查 capability registry
5. `target` 为其他节点时：
   - 由 `flow` 发起 `exec.call`
   - `exec` 负责路由与权限裁决
6. 成功时，将 `exec.call_resp.result` 或本地方法返回值写入 `RunContext.nodes[<id>].result`

兼容说明：

- 历史存量 `call` 节点若仍使用 `args` 直写且未声明 `inputs`，运行期可继续解释
- 新写入契约应使用 `args_template + inputs`

节点类型：compose
-----------------

`kind=compose` 的正式写入 spec：

- `template`：object（必填）
- `inputs`：`InputBinding[]`（可选）
- `_ui`：编辑器布局元数据（可选）

执行语义：

1. 以 `template` 为基础物化结果 JSON
2. 按声明顺序应用 `inputs`
3. 得到最终结果并写入 `RunContext.nodes[<id>].result`
4. `compose` 节点不发起远程调用，也不依赖 `exec.call`

节点类型：set_var
-----------------

`kind=set_var` 的正式写入 spec：

- `name`：string（必填，大小写敏感，仅允许 `[A-Za-z_][A-Za-z0-9_]*`）
- `template`：任意合法 JSON 值（可选，默认 `null`）
- `inputs`：`InputBinding[]`（可选）
- `_ui`：编辑器布局元数据（可选，不参与执行语义）

执行语义：

1. 以 `template` 为基础物化一个 JSON 值
2. 按声明顺序应用 `inputs`
3. 得到最终值后写入 `RunContext.vars[<name>]`
4. 同时把该值写入 `RunContext.nodes[<id>].result`
5. 当前 run 内，下游节点可通过 `source.kind=flow_var` 读取该值

执行语义
--------

- 对 `graph` 做拓扑排序；按顺序逐个执行 `nodes`
- 每个节点执行前，先完成输入物化
- 每个节点执行后，把状态和结果写入 `RunContext`
- `set_var` 额外更新 `RunContext.vars` 中对应变量的当前值和写入者
- `allow_fail=false`
  - 当前节点失败后立即结束整个 run
- `allow_fail=true`
  - 记录当前节点失败并继续后续节点
- `retry`
  - 表示失败后的额外重试次数；默认 `1` 表示“首次尝试失败后，最多再试一次”
- `retry_backoff_ms`
  - 第一版固定间隔策略；仅在“本次失败且后面仍有剩余 retry”时生效
  - `0` / 省略表示立即重试
  - `>0` 表示下一次尝试前等待指定毫秒数
  - 等待期间若 run 被 `cancel_run` 或 `delete` 中断，必须立即停止等待和后续重试
- `max_active_runs`
  - flow 级活动 run 上限，作用于同一 `flow_id`
  - 省略时保持 legacy 兼容行为：
    - 手动 `run` 继续允许并发重入
    - `interval/event/var_changed` trigger 继续在已有活动 run 时跳过本次启动
  - `0` 表示所有启动来源都不限制活动 run 数
  - `>0` 表示手动 `run` 与 trigger 启动都必须遵守统一 active-run 上限
  - 手动 `run` 超限时返回 `409`
  - trigger 超限时跳过本次启动，不生成新 run
- `trigger.dedup_window_ms`
  - `0` / 省略表示关闭 trigger dedup
  - 仅 `event` / `var_changed` 支持 `>0`
  - `>0` 表示执行器按“同一 flow + 同一规范化 trigger 上下文”在内存中做窗口去重
  - dedup 命中时跳过本次 trigger 启动，不生成新 run
  - dedup 状态不要求持久化；执行器重启后可清空

失败类型建议：

- 绑定配置非法：`400`
- 绑定运行期缺失必填值：`400`
- 目标方法不存在：`404`
- 调用超时：`408`
- 执行内部错误：`500`

图校验要求
----------

`set` 阶段至少校验：

- 图非空
- 节点 ID 唯一
- 边引用的节点存在
- 图无环
- `call` / `compose` / `set_var` 的 spec 字段完整
- `max_active_runs >= 0`
- `retry_backoff_ms >= 0`
- `trigger.dedup_window_ms >= 0`
- `interval` trigger 不支持 `dedup_window_ms > 0`
- `inputs` 中引用的 `node_id` 存在
- `inputs` 中引用的 `node_id` 是当前节点祖先
- `flow_var.name` 合法，且可唯一解析到祖先 `set_var` 写入者
- `to` / `source.path` 是合法 JSON Pointer

持久化与配置
------------

- 工作流定义持久化是可插拔的，运行期 `flow` handler 只依赖 `LoadAll/Save/Delete` 接口。
- 默认 backend：`json`
  - `flow.backend` 未配置或为空时，仍使用本地 JSON 文件。
  - 默认目录：`./flows`
  - 文件命名：`./flows/<flow_id>.json`
- `pg` backend：
  - `flow.backend=pg` 时，由 `Server` 注入 PG persistence。
  - PG 中直接存储完整 flow 定义本体，而不是本地 JSON 路径引用。
  - 启动时通过 persistence `LoadAll()` 预热到内存 `flows map`。
- 建议配置项：
  - `flow.backend`
  - `flow.base_dir`
  - `flow.max_retained_runs`
  - `flow.run_archive_enabled`
  - `state.pg.dsn`
  - `state.pg.flow_table`
- 不在本轮持久化范围：
  - 活动 run 状态
  - scheduler
  - 运行中的 runtime context
- `flow.run_archive_enabled=true` 时：
  - retained window 内的终态 run 会以本地 JSON sidecar 形式归档到 `flow.base_dir/_runs/<flow_id>/<run_id>.json`
  - 启动时执行器会从 archive 预热 retained run，供 `status/detail/list_runs` 继续查询
  - archive 仅覆盖 retained window，不承诺窗口外长期历史
- backend 已显式配置但不可用时，不静默降级到其他 backend。
- backend 切换时不自动迁移已有 JSON / PG 数据。

结果保留策略：

- 运行时可在内存中保留有限数量的历史 run 摘要
- `list_runs` 查询的正是这部分 retained window，不承诺返回窗口外历史
- 完整节点结果可在 retained window 内通过 `detail` 查询
- `flow.run_archive_enabled=true` 时，这部分 retained window 会被持久化并在启动时重新加载
- `flow.max_retained_runs` 同时约束 retained window 的内存与 archive 上限
- 未开启 run archive 时，retained window 继续保持仅内存语义
- flow 局部变量属于 `RunContext` 运行期状态，不参与定义持久化，也不承诺长期保留

错误码建议
----------

- `1`：ok
- `400`：invalid request / invalid graph / invalid flow_id / invalid binding
- `403`：permission denied
- `404`：not found
- `409`：conflict / run already terminal / active run limit reached
- `408`：timeout
- `500`：internal error

Related Requirements
--------------------

- [../requirements/flow_data_dag.md](../requirements/flow_data_dag.md)

Related Changes
---------------

- 待本次 workflow 完成后补充。
