TopicBus 协议（SubProto=4）规范
==============================

范围
----
- 仅描述当前 `MyFlowHub-Server/subproto/topicbus` 的实现与约定。
- 不涉及 login_server 或其他旧流程。

总览
----
- 主题（topic）订阅/退订 + 事件发布（publish）与逐级转发。
- topic 字符串**无约束**：不做 trim/校验，按原样存储与匹配（大小写敏感）。
- 发布/订阅均**无权限控制**。
- 订阅关系仅保存在**内存**；连接断开会自动清理；不做离线堆积与历史回放。
- 逐级订阅：父节点只记录“**直连子连接**订阅了哪些 topic”；当本节点某 topic 从 0→1/1→0 时，才向父节点汇总 `subscribe_batch`/`unsubscribe_batch`。
- 逐级转发：publish 在本层先向下转发给本节点的订阅者（不回显），再上送父节点；收到父节点的 publish 时只向下转发，不再向上回传。

消息格式（JSON）
---------------
- 统一 envelope：`{"action":"<name>","data":{...}}`
- 响应 action = `<req>_resp`，状态码位于 `data.code`。

头部与路由约定
--------------
- Major：
  - 请求与 publish：建议 `MajorCmd`
  - 响应：`MajorOKResp`
- SubProto 固定为 `4`。
- `SourceID`：必须为已认证连接的 nodeID（`SourceID=0` 的非登录子协议会被 PreRouting 丢弃）。
- `TargetID`：
  - 向父节点发送：填父节点 NodeID
  - **注意**：在核心里 `TargetID=0` 表示“广播给子节点（不回父）”，不要用 0 表示“上送父节点”。

动作与载荷
----------
以下示例均只展示 payload（不含 header）。

### subscribe
- 请求：`{"action":"subscribe","data":{"topic":"<string>"}}`
- 响应：`{"action":"subscribe_resp","data":{"code":1,"msg":"ok","topic":"..."}}`
- 说明：
  - 重复订阅同一 topic 为幂等 no-op（不会重复上送）。
  - 若请求体无法解析，返回 `code=400`。

### subscribe_batch
- 请求：`{"action":"subscribe_batch","data":{"topics":["t1","t2"]}}`
- 响应：`{"action":"subscribe_batch_resp","data":{"code":1,"msg":"ok","topics":["t1","t2"]}}`
- 说明：
  - `topics` 会去重（保留首次出现的顺序）。

### unsubscribe
- 请求：`{"action":"unsubscribe","data":{"topic":"<string>"}}`
- 响应：`{"action":"unsubscribe_resp","data":{"code":1,"msg":"ok","topic":"..."}}`
- 说明：幂等；无论该 topic 是否存在订阅，都返回 `code=1`（包括请求体无法解析时）。

### unsubscribe_batch
- 请求：`{"action":"unsubscribe_batch","data":{"topics":["t1","t2"]}}`
- 响应：`{"action":"unsubscribe_batch_resp","data":{"code":1,"msg":"ok","topics":["t1","t2"]}}`
- 说明：幂等；`topics` 会去重（保留首次出现的顺序）。

### list_subs
- 请求：`{"action":"list_subs","data":{}}`（data 内容忽略）
- 响应：`{"action":"list_subs_resp","data":{"code":1,"topics":["..."]}}`
- 说明：
  - 返回的是“**本节点记录的该连接**”订阅列表（用于子节点向父节点查询“我在你这里订了什么”）。
  - `topics` 始终存在：无订阅时返回 `[]`；有订阅时按字典序排序。

### publish
- 请求：`{"action":"publish","data":{"topic":"...","name":"...","ts":1730000000000,"payload":{...}}}`
- 响应：无（不回显/不 ack）。
- 字段约定：
  - `topic`：主题字符串（无约束，原样匹配）
  - `name`：事件名，**不能为空**（全空白会被丢弃）
  - `ts`：Unix 毫秒时间戳
  - `payload`：事件数据，可为任意 JSON 值（对象/数组/字符串/数值/布尔/null），可省略
- 不回显规则：发布连接不会收到自己的 publish（即便该连接订阅了该 topic）。

逐级订阅汇总与逐级转发
----------------------

订阅汇总（向上）
---------------
- 子节点对父节点发 `subscribe`/`subscribe_batch` 后，父节点只在本地记录该“直连连接”的订阅关系。
- 当某 topic 在本节点的订阅者数量从 0→1 时，本节点会向父节点发送一次 `subscribe_batch` 做汇总（best-effort）。
- 当某 topic 在本节点的订阅者数量从 1→0 时，本节点会向父节点发送一次 `unsubscribe_batch` 取消汇总（best-effort）。

事件转发（向下 + 向上）
----------------------
- 本节点收到 `publish` 后：
  1) 向下转发：将消息转发给本节点所有订阅了该 topic 的直连连接（**排除来源连接，不回显**）。
  2) 向上转发：若存在父连接且来源不是父连接，则将该 publish 上送父节点（用于跨子树传播）。
- 父节点收到 publish 后同理处理：向其订阅者下发，并继续向上上送，直到最上层节点。
- 若某一分支没有订阅者，该分支不会收到任何 publish；系统不做离线缓存。

连接断开与重订阅建议
--------------------
- 连接断开（`conn.closed`）会触发本节点清理该连接的所有 topic 订阅；必要时自动向父节点退订汇总。
- 因为不做持久化，节点在“登录成功/重连成功/更换父节点接入位置”后，应主动重新订阅所需 topic。

来源与追踪（可选）
----------------
- `publish.data` 默认只包含 `topic/name/ts/payload`，不包含“原始发布者 nodeID”。
- 接收方看到的 header `SourceID` 为**逐跳发送方**（当前转发节点），不保证等于原始发布者。
- 若业务需要追踪来源，建议在 `payload` 内自行携带 `source`（或在 `payload` 顶层附加 `publisher_node_id` 等字段）。

示例流程
--------
- 订阅：子节点（NodeID=10）向父节点（NodeID=1）发送：`{"action":"subscribe","data":{"topic":"t/a"}}`，收到 `subscribe_resp`。
- 查询：子节点向父节点发送 `list_subs`，父节点返回 `{"action":"list_subs_resp","data":{"code":1,"topics":["t/a"]}}`。
- 发布：任意节点发送：`{"action":"publish","data":{"topic":"t/a","name":"temp","ts":1730000000000,"payload":{"value":22.5}}}`。
  - 本层：转发给本节点所有订阅 `t/a` 的直连连接（排除来源，不回显）。
  - 向上：继续上送父节点，以便跨子树传播；父节点收到后同理向下+向上处理。

错误码约定
----------
- `1`：成功（subscribe/subscribe_batch/unsubscribe/unsubscribe_batch/list_subs 的正常返回）
- `400`：订阅请求体无法解析（仅 `subscribe`）

集成提示
--------
- Hub 注册：`dispatcher.RegisterHandler(topicbus.NewTopicBusHandlerWithConfig(cfg, log))`（建议 import `github.com/yttydcs/myflowhub-server/subproto/topicbus`）
- 子协议编号：`SubProto=4`。***
