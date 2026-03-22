flow 协议（SubProto=6）规范
===========================

范围
----

- `flow` 子协议用于在任意节点上保存、触发和调度一个 DAG 工作流。
- 本协议同时定义：
  - 工作流定义与调度动作（`set/delete/run/status/list/get`）
  - 节点执行语义
  - 节点间数据传递模型
- 第一版数据流增强后的正式节点类型为：
  - `call`
  - `compose`
- `exec`（SubProto=7）继续负责远程方法调用与权限裁决；`flow` 不复制 `exec.call` 的路由与权限模型。

总览
----

- 控制帧编码：UTF-8 JSON，envelope 固定为 `{"action":"...","data":{...}}`
- 典型动作：
  - `set`：设置/更新工作流（需要权限 `flow.set`）
  - `delete`：删除工作流（需要权限 `flow.delete`）
  - `run`：手动触发一次运行
  - `status`：查询运行状态摘要
  - `list`：列出执行者当前已知的工作流摘要
  - `get`：读取指定工作流定义
- 触发器（当前）：支持 `interval` / `event` / `var_changed`
  - `event`：由 `topicbus.publish` / `topicbus.received` 事件驱动
  - `var_changed`：由 `varstore.changed` / `varstore.deleted` 事件驱动

权限
----

- 权限节点格式：`协议.action`
- 第一版最小权限：
  - `flow.set`
  - `flow.delete`
- 当前版本未为 `run/status/list/get` 单独定义额外权限；它们按 `executor_node` 路由到执行者处理。

HeaderTcp 与路由约定
--------------------

- SubProto 固定为 `6`
- Major 约定：
  - 请求帧（`set/delete/run/status/list/get`）：`MajorCmd`
  - 响应帧（`*_resp`）：`MajorOKResp`
  - 失败响应也使用 `MajorOKResp`，错误通过 payload 的 `code/msg` 表达

逐级授权模型
------------

本协议的逐级授权仅适用于 `flow.set` / `flow.delete`：

- 请求方将请求发给任意实现了 `flow` 的节点（称为执行者）
- 执行者不直接裁决权限，而是逐级上送，直到某一级可以在其子树内完成裁决
- 裁决节点以 `origin_node` 为权限主体：
  - `set` 对应 `flow.set`
  - `delete` 对应 `flow.delete`
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

响应 `action=run_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/404/500`
- `msg`：可选错误说明
- `flow_id`：回显
- `run_id`：成功时必填

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

响应 `action=status_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/404/500`
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

### action=list

请求 `data`：

- `req_id`：UUID（必填）
- `origin_node`：uint32（可选）
- `executor_node`：uint32（可选）

响应 `action=list_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/500`
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

响应 `action=get_resp`，`data`：

- `req_id`：回显
- `code`：`1/400/404/500`
- `msg`：可选错误说明
- `executor_node`：实际执行者节点 ID
- `flow_id`：回显
- `name`：工作流名称（可选）
- `trigger`：触发器定义
- `graph`：完整 DAG 定义

触发器定义
----------

`trigger`：

- `type`：`interval` | `event` | `var_changed`
- `interval`
  - `every_ms`：uint64，必填，`>0`
- `event`
  - `event_mode`：`publish` | `received` | `any`
  - `event_name`：string（可选）
  - `event_topic`：string（可选）
  - 约束：`event_name` 与 `event_topic` 不能同时为空
- `var_changed`
  - `var_owner`：uint32（可选，`0` 表示不过滤 owner）
  - `var_name`：string（可选，空表示不过滤 name）

图与节点模型
------------

`graph.nodes[]`：

- `id`：string（必填，图内唯一）
- `kind`：string（必填）
  - 新写入契约：`"call"` | `"compose"`
- `allow_fail`：bool（可选，默认 `false`）
- `retry`：int（可选，默认 `1`）
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

引用规则：

- 只允许引用祖先节点
- 不允许引用当前节点或未来节点
- `to` 必须是合法 JSON Pointer
- `required=true` 且来源不存在时，当前节点必须失败

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

执行语义
--------

- 对 `graph` 做拓扑排序；按顺序逐个执行 `nodes`
- 每个节点执行前，先完成输入物化
- 每个节点执行后，把状态和结果写入 `RunContext`
- `allow_fail=false`
  - 当前节点失败后立即结束整个 run
- `allow_fail=true`
  - 记录当前节点失败并继续后续节点

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
- `call` / `compose` 的 spec 字段完整
- `inputs` 中引用的 `node_id` 存在
- `inputs` 中引用的 `node_id` 是当前节点祖先
- `to` / `source.path` 是合法 JSON Pointer

持久化与配置
------------

- 默认目录：`./flows`
  - 工作流定义：`./flows/<flow_id>.json`
- 建议配置项：
  - `flow.base_dir`
  - `flow.max_retained_runs`

结果保留策略：

- 运行时可在内存中保留有限数量的历史 run 摘要
- 完整节点结果不承诺长期持久化，除非后续新增专门的 run detail / archive 能力

错误码建议
----------

- `1`：ok
- `400`：invalid request / invalid graph / invalid flow_id / invalid binding
- `403`：permission denied
- `404`：not found
- `408`：timeout
- `500`：internal error

Related Requirements
--------------------

- [../requirements/flow_data_dag.md](../requirements/flow_data_dag.md)

Related Changes
---------------

- 待本次 workflow 完成后补充。
