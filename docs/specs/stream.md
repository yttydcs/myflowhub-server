stream 协议（建议 SubProto=8）规范（草案）
=========================================

范围
----

- `stream` 子协议用于在节点间声明、发现和连接“逻辑流源”和“逻辑消费端”，并为每次订阅或控制连接建立独立交付状态。
- 该协议主要承载：
  - 音乐
  - 视频
  - 文本流
  - 自定义流内容
- 本协议目标是统一“流源目录 + 订阅 + 二进制交付”的 wire 和会话模型，而不是替换：
  - `file` 的既有 wire
  - `flow` 的 DAG workflow 语义
  - `topicbus` 的事件总线语义

命名与兼容边界
--------------

- **现有 `flow`（SubProto=6）继续表示 DAG workflow orchestration**，不得复用为流媒体协议。
- 本轮建议新增独立 `stream` 子协议，并为其分配新的 SubProto 编号（当前建议 `8`）。
- `file` 继续保留现有 `SubProto=5` 和 wire，不做破坏性替换。
- `topicbus` 继续表示“无状态事件广播”：
  - 可以类比“订阅主题”
  - 但不能替代 `stream` 的 producer 声明、订阅确认、per-consumer ACK

核心模型
--------

本协议区分三个层级：

### 1. 逻辑流源（Source）

- 由 producer 声明
- 在协议期内保持稳定标识
- 可以被多个 consumer 订阅
- 至少包含：
  - `source_id`
  - `producer`
  - `kind`
  - `name`
  - `tags`
  - `metadata`

### 2. 逻辑消费端（Consumer Endpoint）

- 由 consumer 节点声明
- 在协议期内保持稳定标识
- 用于表达“这个端点准备消费哪一类流”
- 至少包含：
  - `consumer_id`
  - `consumer`
  - `kind`
  - `name`
  - `tags`
  - `metadata`

说明：

- 为保持首版匹配语义简单明确，一个 `ConsumerDescriptor` 只声明一个 `kind`。
- 若同一设备需要消费多种类型，应声明多个 consumer endpoint。

### 3. 交付会话（Delivery）

- 每个 consumer 对某个 source 的自主订阅，或控制侧对某个 `source + consumer endpoint` 的连接，都会生成一个独立 `delivery_id`
- `delivery_id` 绑定：
  - producer
  - consumer
  - `consumer_id`
  - source
  - 进度 / ACK / timeout / window
- 这样同一 source 下多个 consumer 的消费速度不同，也不会互相污染

说明：

- 这是“单 producer，多 consumer”的逻辑模型。
- 但 wire / ACK 层仍以“每次订阅或连接独立 delivery”建模，避免把多 consumer 反压混成一份状态。

与现有子协议边界
----------------

- `file`
  - 适合文件、目录、文本预览等 bounded 内容。
  - `stream` 借鉴其 control/data/ack 三层结构，但新增了 source catalog 与 subscribe 模型。
- `flow`
  - 适合 workflow 的定义、调度和运行状态，不承载原始媒体帧或长时字节流。
- `exec`
  - 负责 RPC / capability 调用与权限裁决，不承载长期数据面。
- `topicbus`
  - 适合事件 pub/sub。
  - `stream` 与它的最大区别是：
    - 有 producer 声明的 source 描述
    - 有 consumer 声明的消费端描述
    - 有订阅确认
    - 有 per-consumer delivery state
    - 有二进制数据面与 ACK

权限
----

- 权限节点格式继续采用 `协议.action`。
- 建议最小权限：
  - `stream.publish`
    - producer 声明 / 撤销 source
  - `stream.consume`
    - consumer 声明 / 撤销消费端
  - `stream.subscribe`
    - consumer 查询 source 并自主订阅 / 取消订阅
  - `stream.connect`
    - 控制侧查询 consumer 并发起 connect / disconnect
- `signal`
  - 默认只允许该 delivery 的 producer 或 consumer 本方发起。
  - 若未来需要第三方控制，再单独扩展权限模型。

帧分类（由 payload[0] 决定）
--------------------------

本协议延续 `file` 的 kind 分流：

- `0x01`：CTRL（控制帧，后续为 JSON）
- `0x02`：DATA（二进制数据帧）
- `0x03`：ACK（二进制确认帧）

建议 Major：

- CTRL 请求：`MajorCmd`
- CTRL 响应：`MajorOKResp`
- DATA / ACK：`MajorMsg`
- 失败响应仍使用 `MajorOKResp`，错误通过 payload `code/msg` 表达

HeaderTcp 与路由约定
--------------------

- SubProto：建议 `8`
- `TargetID=0` 的广播语义保持不变，**不得**用于表达“上送父节点”。
- CTRL 请求使用 `MajorCmd` 逐跳进入 handler：
  - 便于复用 `file` 风格的 LCA 判权 / 转交
  - 避免子节点直接填写最终目标从而绕过判权链
- CTRL 响应使用 `MajorOKResp` 端到端返回请求方。
- DATA / ACK 使用 `MajorMsg` 端到端快速转发。

端到端 SourceID / TargetID 约定：

- `announce/withdraw/list_sources/get_source/announce_consumer/withdraw_consumer/list_consumers/get_consumer/subscribe/unsubscribe/connect/disconnect/signal` 请求阶段：
  - `SourceID=请求方`
  - 首发 `TargetID=请求方的直接父 Hub`
- 授权后下发：
  - `SourceID=请求方`
  - `TargetID=最终对端`
- `*_resp`：
  - `SourceID=响应方`
  - `TargetID=请求方`
- DATA：
  - `SourceID=producer`
  - `TargetID=consumer`
- ACK：
  - `SourceID=consumer`
  - `TargetID=producer`

来源校验
--------

沿用当前系统的统一校验前提：

- 来自父连接：放行
- 来自子连接：仅当 `SourceID` 为该连接自身或其后代时放行

控制帧（CTRL）格式
-----------------

- 载荷编码：`payload = [0x01] + JSON(UTF-8)`
- JSON envelope：`{"action":"...","data":{...}}`

建议 action：

- `announce`
- `announce_resp`
- `withdraw`
- `withdraw_resp`
- `list_sources`
- `list_sources_resp`
- `get_source`
- `get_source_resp`
- `announce_consumer`
- `announce_consumer_resp`
- `withdraw_consumer`
- `withdraw_consumer_resp`
- `list_consumers`
- `list_consumers_resp`
- `get_consumer`
- `get_consumer_resp`
- `subscribe`
- `subscribe_resp`
- `unsubscribe`
- `unsubscribe_resp`
- `connect`
- `connect_resp`
- `disconnect`
- `disconnect_resp`
- `signal`
- `signal_resp`

通用错误码建议：

- `1`：ok
- `400`：invalid request / invalid source / invalid delivery
- `403`：permission denied
- `404`：source not found / target not found
- `406`：unsupported kind or format
- `408`：timeout
- `409`：duplicate source / duplicate subscribe / conflict
- `413`：payload too large
- `429`：too many deliveries / buffers exhausted
- `500`：internal error

流源描述模型（SourceDescriptor）
------------------------------

建议字段：

- `source_id`：UUID 或 producer 内稳定唯一字符串
- `producer`：producer nodeID
- `name`：可选展示名
- `kind`：首批建议内建值：
  - `music`
  - `video`
  - `text`
  - `custom`
- `content_type`：可选，更精确的格式提示
  - 示例：`audio/opus`
  - 示例：`video/h264`
  - 示例：`text/plain`
  - 示例：`application/octet-stream`
- `mode`：`live` | `bounded`
- `unit_mode`：`frame` | `chunk`
- `tags`：可选字符串数组，承载 producer 定义的扩展标签
- `metadata`：可选 JSON 对象

说明：

- `kind` 是 consumer 最先看到的一级分类。
- `tags` 用于扩展，例如：
  - `camera`
  - `screen`
  - `live`
  - `lossless`
- `content_type` 用于更精确的格式提示，不参与第一层订阅语义。

消费端描述模型（ConsumerDescriptor）
----------------------------------

建议字段：

- `consumer_id`：UUID 或 consumer 节点内稳定唯一字符串
- `consumer`：consumer nodeID
- `name`：可选展示名
- `kind`：必填，表示该 consumer endpoint 消费的一级种类
  - `music`
  - `video`
  - `text`
  - `custom`
- `content_type`：可选，希望接收的更精确格式提示
- `tags`：可选字符串数组
- `metadata`：可选 JSON 对象

说明：

- 首版匹配规则首先看 `kind`。
- 若 `kind=custom`，可以再结合 `content_type/tags` 做更细粒度约束。
- 若一个设备要消费多种 `kind`，建议注册多个 `consumer_id`。

### action=announce

用于 producer 声明一个可被订阅的逻辑流源。

请求 `data` 建议字段：

- `req_id`：UUID（必填）
- `source`：`SourceDescriptor`（必填）

响应 `announce_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `source`

约束建议：

- 同一 producer 下 `source_id` 必须唯一。
- 重复 `announce` 相同 `source_id`：
  - 若描述完全相同，可视为幂等更新
  - 若描述冲突，返回 `409`

### action=withdraw

用于 producer 撤销一个流源，并终止其下全部 delivery。

请求 `data` 建议字段：

- `req_id`
- `source_id`

响应 `withdraw_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `source_id`

说明：

- producer 撤销 source 后，应结束其下全部 delivery。

### action=list_sources

用于 consumer 或控制端查询某个 producer 当前公开的 source 列表。

请求 `data` 建议字段：

- `req_id`
- `producer`：目标 producer nodeID
- `kind`：可选过滤
- `tag`：可选过滤

响应 `list_sources_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `producer`
- `sources`：`SourceDescriptor[]`

### action=get_source

用于按 `source_id` 精确读取单个 source 描述。

请求 `data` 建议字段：

- `req_id`
- `producer`
- `source_id`

响应 `get_source_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `source`

### action=announce_consumer

用于 consumer 节点声明一个可被连接或自主订阅引用的逻辑消费端。

请求 `data` 建议字段：

- `req_id`
- `consumer_endpoint`：`ConsumerDescriptor`

响应 `announce_consumer_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `consumer_endpoint`

约束建议：

- 同一 consumer 节点下 `consumer_id` 必须唯一。
- 每个 `consumer_id` 首版只绑定一个 `kind`。

### action=withdraw_consumer

用于 consumer 节点撤销一个逻辑消费端，并终止其下全部 delivery。

请求 `data` 建议字段：

- `req_id`
- `consumer_id`

响应 `withdraw_consumer_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `consumer_id`

### action=list_consumers

用于控制侧或其他具备权限的节点查询某个 consumer 当前公开的消费端列表。

请求 `data` 建议字段：

- `req_id`
- `consumer`：目标 consumer nodeID
- `kind`：可选过滤
- `tag`：可选过滤

响应 `list_consumers_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `consumer`
- `consumer_endpoints`：`ConsumerDescriptor[]`

### action=get_consumer

用于按 `consumer_id` 精确读取单个 consumer endpoint 描述。

请求 `data` 建议字段：

- `req_id`
- `consumer`
- `consumer_id`

响应 `get_consumer_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `consumer_endpoint`

### action=subscribe

用于 consumer 使用自己已声明的 `consumer_id` 订阅某个 source，并创建独立 delivery。

请求 `data` 建议字段：

- `req_id`
- `producer`：目标 producer nodeID
- `source_id`
- `consumer_id`：发起订阅的 consumer endpoint 标识
- `resume_from`：可选，仅 `bounded + chunk` 常用
- `window_bytes`：可选
- `ack_interval_ms`：可选

响应 `subscribe_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `accept`
- `source`
  - 回传 producer 声明的 source 描述，让 consumer 在订阅成功时再次拿到 `kind/tags`
- `consumer_endpoint`
  - 回传本次绑定的消费端描述
- `delivery_id`
- `producer`
- `consumer`
- `consumer_id`
- `start_position`
- `window_bytes`
- `ack_interval_ms`

校验建议：

- `consumer_id` 必须已由请求方所属 consumer 节点声明。
- `source.kind` 必须等于 `consumer_endpoint.kind`，否则返回 `406`。

### action=unsubscribe

用于关闭某个 delivery。

请求 `data` 建议字段：

- `req_id`
- `delivery_id`
- `reason`：可选

响应 `unsubscribe_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `delivery_id`
- `reason`

说明：

- consumer 主动退订，只关闭自己的 delivery。
- producer 也可以通过 `delivery_id` 主动结束某个 consumer 的交付，不影响其他 consumer。

### action=connect

用于控制侧把一个 `source` 和一个 `consumer endpoint` 连接起来，并创建独立 delivery。

请求 `data` 建议字段：

- `req_id`
- `producer`
- `source_id`
- `consumer`
- `consumer_id`
- `resume_from`：可选，仅 `bounded + chunk` 常用
- `window_bytes`：可选
- `ack_interval_ms`：可选

响应 `connect_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `accept`
- `source`
- `consumer_endpoint`
- `delivery_id`
- `producer`
- `consumer`
- `consumer_id`
- `start_position`
- `window_bytes`
- `ack_interval_ms`

说明：

- `connect` 是控制侧触发的 delivery 建立动作。
- 成功后数据面仍然是 `producer -> consumer`，并不经过控制侧转发。
- 首版强制校验 `source.kind == consumer_endpoint.kind`。

逐级处理建议：

- `connect` 首发时仍按 `file` 风格设置：
  - `SourceID=控制侧请求方`
  - `TargetID=请求方的直接父 Hub`
- 沿途节点不要直接假定自己是最终处理点。
- **第一个同时满足以下条件的节点应成为本次 connect 的协调节点（coordinator）**：
  - 能证明 `producer` 在自己的本地或子树路由范围内
  - 能证明 `consumer` 在自己的本地或子树路由范围内
  - 能对 `hdr.SourceID` 对应的请求方执行 `stream.connect` 权限判断
- 若当前节点无法同时覆盖 `producer` 与 `consumer`，则继续上送父节点。
- 协调节点必须在返回 `connect_resp(code=1)` 之前完成：
  - 校验 source 存在
  - 校验 consumer endpoint 存在
  - 校验 `source.kind == consumer_endpoint.kind`
  - 分配新的 `delivery_id`
  - 让 producer 侧安装发送状态
  - 让 consumer 侧安装接收状态
  - 写入最小 delivery 路由索引
- 若上述任一步失败，返回失败 `connect_resp`，且不得留下半开 delivery。

### action=disconnect

用于控制侧按 `delivery_id` 断开一个现存连接。

请求 `data` 建议字段：

- `req_id`
- `delivery_id`
- `reason`：可选

响应 `disconnect_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `delivery_id`
- `reason`

说明：

- `disconnect` 由控制侧发起，但关闭后仍只影响目标 delivery。
- 推荐由持有该 `delivery_id` 路由索引的协调节点负责拆除。

逐级处理建议：

- 请求路径与 `connect` 相同，仍先上送直接父 Hub。
- 若当前节点持有该 `delivery_id` 的协调索引，则：
  - 标记 delivery 为 `closing`
  - 向 producer / consumer 两侧发出拆除指令
  - 等待两侧确认或超时
  - 删除协调索引
  - 返回 `disconnect_resp`
- 若当前节点不持有该 `delivery_id`，但能根据本地索引判定上游协调节点，则继续向上转交。
- 若无法找到该 `delivery_id`，返回 `404`。

### action=signal

用于 delivery 运行中的轻量控制。

请求 `data` 建议字段：

- `req_id`
- `delivery_id`
- `op`：建议首批支持：
  - `pause`
  - `resume`
  - `metadata_update`
  - `keyframe_request`
  - `custom`
- `data`：可选 JSON 对象

响应 `signal_resp.data` 建议字段：

- `code` / `msg`
- `req_id`
- `delivery_id`
- `op`

二进制数据帧（DATA）
-------------------

payload 结构：

- `DATA = [0x02] + StreamDataHeaderV1 + <bytes...>`

`StreamDataHeaderV1` 建议字段（网络序，大端）：

- `ver`：`uint8`（固定 1）
- `flags`：`uint8`
  - bit0：`EOS`
  - bit1：`KEYFRAME`
  - bit2：`CONFIG`
  - bit3：`DISCONTINUITY`
- `delivery_id`：`[16]byte`
- `position`：`uint64`
  - `chunk`：本块起始 byte offset
  - `frame`：单调递增 frame sequence
- `pts_ms`：`uint64`
  - `frame` 推荐填写
  - `chunk` 可为 `0`

说明：

- DATA 绑定的是 `delivery_id`，不是逻辑 `source_id`。
- 这样同一 source 下多个 consumer 的 ACK 可以独立收敛。

确认帧（ACK）
------------

payload 结构：

- `ACK = [0x03] + StreamAckHeaderV1`

`StreamAckHeaderV1` 建议字段（网络序，大端）：

- `ver`：`uint8`（固定 1）
- `flags`：`uint8`
- `delivery_id`：`[16]byte`
- `position`：`uint64`
  - `chunk`：consumer 已持久化的下一期望 byte offset
  - `frame`：consumer 已消费的下一期望 sequence
- `credit_units`：`uint32`
  - 允许 producer 继续发送的窗口提示
- `reserved`：`uint32`

ACK 语义：

- `bounded + chunk`
  - ACK 主要用于 resume / progress / bounded buffer
- `live + frame`
  - ACK 主要用于消费进度与本地反压
  - 不表示 RTP 式网络层重传

source / delivery 状态与校验
---------------------------

source 至少应绑定：

- `source_id`
- `producer`
- `kind`
- `mode`
- `unit_mode`
- `tags`

consumer endpoint 至少应绑定：

- `consumer_id`
- `consumer`
- `kind`
- `tags`

delivery 至少应绑定：

- `delivery_id`
- `source_id`
- `producer`
- `consumer`
- `consumer_id`
- 当前 `position`
- 协商后的 `window_bytes` / `ack_interval_ms`

节点本地状态建议
----------------

### 1. source owner（producer 所在节点）

至少维护本节点拥有的：

- `sources[source_id] -> SourceDescriptor`
- `producer_deliveries[delivery_id]`
  - `source_id`
  - `consumer`
  - `consumer_id`
  - `position`
  - `acked_position`
  - `window_bytes` / `credit_units`
  - `state`
  - `last_active`

职责：

- 响应 `announce/withdraw/list_sources/get_source`
- 在 connect / subscribe 成功后安装发送侧 delivery
- 发送 DATA 并消费 ACK
- source 撤销或节点掉线时清理本节点相关 delivery

### 2. consumer owner（consumer 所在节点）

至少维护本节点拥有的：

- `consumers[consumer_id] -> ConsumerDescriptor`
- `consumer_deliveries[delivery_id]`
  - `source_id`
  - `producer`
  - `consumer_id`
  - `expected_position`
  - `last_ack_position`
  - `window_bytes` / `credit_units`
  - `state`
  - `last_active`

职责：

- 响应 `announce_consumer/withdraw_consumer/list_consumers/get_consumer`
- 在 connect / subscribe 成功后安装接收侧 delivery
- 接收 DATA 并发送 ACK
- consumer endpoint 撤销或节点掉线时清理本节点相关 delivery

### 3. coordinator（connect / subscribe 被确认的协调节点）

至少维护最小 delivery 路由索引：

- `delivery_routes[delivery_id]`
  - `requester`
  - `producer`
  - `source_id`
  - `consumer`
  - `consumer_id`
  - `kind`
  - `created_at`
  - `state`
  - `last_control_at`

职责：

- 负责最终接受或拒绝 `connect`
- 推荐也作为跨子树 `subscribe` 的协调节点
- 分配 `delivery_id`
- 协调 producer / consumer 两侧安装或拆除 delivery
- 持有 `disconnect` 所需的最小路由索引
- 本身不进入正常 DATA / ACK 路径

内部协调口径（本轮锁定）：

- 下列动作只作为 `stream` handler 内部 helper action 使用，不属于对外公开 wire：
  - `delivery_prepare` / `delivery_prepare_resp`
  - `delivery_activate` / `delivery_activate_resp`
  - `delivery_abort` / `delivery_abort_resp`
  - `delivery_close` / `delivery_close_resp`
- 协调节点在 `connect` / `subscribe` 成功前，必须先：
  - 向 producer owner 与 consumer owner 发 `delivery_prepare`
  - 等待两侧返回存在性 / `kind` / 初始窗口等确认
  - 本地写入 `delivery_routes[delivery_id]`
  - 再向两侧发 `delivery_activate`
- 在 `delivery_activate` 两侧都成功之前：
  - producer 不得开始发送 DATA
  - consumer 不得把该 delivery 视为可确认 ACK 的 active delivery
- 任一侧 `prepare` / `activate` 失败时，协调节点必须回滚为：
  - 对已安装一侧发送 `delivery_abort`
  - 删除本地 `delivery_routes` 暂存索引
- `withdraw` / `withdraw_consumer` / `disconnect` / `unsubscribe` 触发关闭时：
  - owner 可以先清理本地 state
  - 但仍应通过协调节点补发 `delivery_close`，确保另一侧与 coordinator 路由索引一起收敛

### 4. transit hub（普通中转节点）

- 不建议维护长期 delivery 状态。
- 依赖 Core 现有路由表与父子连接关系完成 CTRL 转交、DATA/ACK 快速转发。
- 除调试 / 观测外，不应复制 producer / consumer 的业务状态。

DATA / ACK 必须校验：

- `delivery_id` 是否存在
- `source.kind` 与 `consumer_endpoint.kind` 是否匹配
- 方向是否匹配：
  - DATA 只能 `producer -> consumer`
  - ACK 只能 `consumer -> producer`
- `position` 是否单调有效

无效情形处理建议：

- 非法 delivery / direction：直接丢弃
- 非法 `position`：
  - `bounded + chunk` 可拒绝并要求重新订阅或重新协商
  - `live + frame` 可丢弃并等待关键帧请求或重新订阅

connect / subscribe 推荐时序
---------------------------

### A 控制 B 和 C 建链

假设：

- A：控制侧请求方
- B：producer 所在节点，拥有 `source_id=s1`
- C：consumer 所在节点，拥有 `consumer_id=c1`

推荐时序：

1. A 发起 `connect(producer=B, source_id=s1, consumer=C, consumer_id=c1, ...)`，首发给自己的直接父 Hub。
2. 请求沿父链上送，直到某个节点同时可达 B 与 C；该节点成为 coordinator。
3. coordinator 校验 A 是否具备 `stream.connect`。
4. coordinator 校验 B 侧 `source_id=s1` 是否存在。
5. coordinator 校验 C 侧 `consumer_id=c1` 是否存在。
6. coordinator 校验 `source.kind == consumer.kind`。
7. coordinator 分配新的 `delivery_id=d1`。
8. coordinator 让 B 安装 `producer_deliveries[d1]`。
9. coordinator 让 C 安装 `consumer_deliveries[d1]`。
10. producer / consumer 两侧都确认后，coordinator 写入 `delivery_routes[d1]`。
11. coordinator 返回 `connect_resp(delivery_id=d1, ...)` 给 A。
12. 后续数据面：
    - B -> C：`DATA(delivery_id=d1)`
    - C -> B：`ACK(delivery_id=d1)`
    - A 不参与正常 DATA / ACK 路径

### consumer 自主订阅

- `subscribe` 仍然走相同的逐级协调思路。
- 区别只在于请求方本身就是 consumer owner。
- 若 producer 与 consumer 不在同一节点，推荐仍由能同时覆盖两端的 coordinator 分配 `delivery_id` 并建立两侧状态。

delivery 生命周期与清理
----------------------

### 正常关闭

- `unsubscribe`
  - 由 consumer 发起
  - 推荐送达 coordinator 或 consumer owner，再由 coordinator 协调两端拆除
- `disconnect`
  - 由控制侧发起
  - 推荐由 coordinator 处理并拆除两端
- `withdraw`
  - 由 source owner 发起
  - 必须遍历该 source 下全部 delivery 并逐个拆除
- `withdraw_consumer`
  - 由 consumer owner 发起
  - 必须遍历该 consumer endpoint 下全部 delivery 并逐个拆除

### 掉线 / 路由失效清理

- producer 所在连接断开或节点不可达：
  - producer owner 必须停止发送 DATA
  - 与该 producer 相关的 delivery 进入 `closing`
  - coordinator 删除相关路由索引
- consumer 所在连接断开或节点不可达：
  - consumer owner 必须停止发送 ACK
  - 与该 consumer endpoint 相关的 delivery 进入 `closing`
  - coordinator 删除相关路由索引
- coordinator 自身失去到任一端的可达路由时：
  - 不得再接受该 delivery 的后续控制操作
  - 应尽快清理 `delivery_routes`
  - 允许实现层通过 TTL / janitor 兜底回收残留索引

### 清理顺序建议

1. 将 delivery 标记为 `closing`
2. 停止接收新的 DATA / ACK
3. 最佳努力通知对端停止发送 / 接收
4. 删除 producer / consumer 本地 delivery 状态
5. 删除 coordinator 的 `delivery_routes`

### TTL / janitor 建议

- producer / consumer owner 应为长期无活动 delivery 设置超时清理
- coordinator 对 `closing` 或孤儿 `delivery_routes` 应设置更短 TTL
- janitor 只能做兜底回收，不应替代显式 `disconnect/withdraw`

多消费者 fanout 语义
-------------------

- source 级别是“一个 producer，多个 consumer”。
- wire 级别是“每个 consumer endpoint 一次连接 / 订阅对应一个独立 delivery”。
- 实现可以在内部做 branch 级复用或 relay 优化，但对协议合同来说：
  - ACK 仍按 delivery 独立统计
  - timeout / unsubscribe / disconnect / pause 仍按 delivery 独立生效

反压与缓存策略
--------------

- `window_bytes` 与 `credit_units` 共同表达 receiver 的承载能力。
- sender 侧必须使用有界缓存，避免 live delivery 无限堆积。
- 对 `live + frame`，允许未来在实现层引入 `drop_policy=latest` 之类的局部策略，但不改变 wire 结构。
- 对 `bounded + chunk`，应优先保证顺序、完整和可恢复。

典型 profile
------------

### music profile

- `kind=music`
- `mode=live`
- `unit_mode=frame`
- `content_type` 示例：`audio/opus`

### video profile

- `kind=video`
- `mode=live`
- `unit_mode=frame`
- `KEYFRAME` / `CONFIG` flag 有意义

### text profile

- `kind=text`
- `mode=live` 或 `bounded`
- `unit_mode=frame` 或 `chunk`
- `content_type` 示例：
  - `text/plain`
  - `text/markdown`
  - `application/json`
- 适合字幕、转写结果、日志文本、聊天消息等“文本为主”的流内容

### custom profile

- `kind=custom`
- `mode=live` 或 `bounded`
- `unit_mode=frame` 或 `chunk`
- `tags` / `content_type` 给出更精细语义

### file bridge profile（未来）

- `kind=custom`
- `mode=bounded`
- `unit_mode=chunk`
- `metadata` 中可携带：
  - `dir`
  - `name`
  - `sha256`
  - `overwrite`

说明：

- 这是“语义映射方向”，不是说首版 `stream` 立刻替换现有 `file` wire。

实现建议
--------

- `MyFlowHub-Proto`
  - 新增 `protocol/stream/types.go`
  - 同步 `docs/protocol_map.md`
- `MyFlowHub-SubProto`
  - 新增 `stream/` module
  - 复用 `file` 风格的：
    - control routing
    - source / consumer / delivery table
    - DATA/ACK 小头解析
  - 可以借鉴 `topicbus` 的“查询/订阅”导航思路，但不要直接复用其无状态事件模型
- `MyFlowHub-Server`
  - 保持 docs 作为长期真相入口

测试策略
--------

后续实现最小测试矩阵建议：

- `announce` 成功 / 重复冲突 / producer 下线清理
- `announce_consumer` 成功 / 重复冲突 / consumer 下线清理
- `list_sources` / `get_source` 返回正确 `kind/tags`
- `list_consumers` / `get_consumer` 返回正确 `kind/tags`
- `text` 流源能被正确声明、查询和订阅
- `subscribe` 成功 / 权限不足 / source 不存在
- `subscribe` 的 `consumer_id` 未声明
- `connect` 成功 / 权限不足 / consumer 不存在
- `connect` 的 source.kind 与 consumer.kind 不匹配
- 同一 source 多 consumer 订阅
- 一个 consumer `unsubscribe` 不影响其他 consumer
- 一个控制侧 `disconnect` 不影响其他 delivery
- `withdraw` 关闭全部 delivery
- `withdraw_consumer` 关闭对应 consumer endpoint 下全部 delivery
- `signal(keyframe_request/pause/resume)` 正常回程
- `DATA` 非法 `delivery_id`
- `DATA` 非法方向
- `ACK` 非法方向
- `bounded + chunk` resume
- `live + frame` 反压和超时

已知风险
--------

- 当前底层 transport 主要是可靠字节流，live 音视频会天然受到 head-of-line blocking 和可靠传输延迟影响。
- `kind=music|video|text|custom` 满足当前用户模型。首版 consumer endpoint 采用“一个 endpoint 对应一个 kind”的最小模型；若未来需要复杂能力协商，再考虑引入 capability set，而不是过早把首版合同复杂化。

Related Requirements
--------------------

- [../requirements/stream.md](../requirements/stream.md)
- [../requirements/flow_data_dag.md](../requirements/flow_data_dag.md)

Related Changes
---------------

- 待本次 workflow 完成后补充。
