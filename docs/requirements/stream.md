# Stream Subprotocol Requirements

## Background

- 当前 `file` 子协议已经证明：在 MyFlowHub 内，`control + binary data + ack` 的双平面模型适合“需要逐跳判权、端到端传输”的能力。
- 当前 `flow` 子协议已经稳定表示“DAG workflow orchestration”，其长期 requirements/specs 和实现都围绕 `set/delete/run/status/list/get` 展开，不适合承载媒体流或自定义字节流。
- 当前 `topicbus` 具备“主题发布/订阅”的多消费者传播模型，但它是无状态事件总线：
  - 不维护 producer 声明的流描述
  - 不提供订阅后的独立交付状态
  - 不提供 ACK / 反压 / resume
- 用户当前明确希望的模型更接近：
  - 每个设备都能声明自己生产的流
  - 每个设备也能声明自己消费的端点
  - consumer 可以自己订阅 producer 的流，也可以由控制侧把两端连起来
  - 整体是“单生产者，多消费者”
  - producer 可以标记流的种类，consumer 创建时也要声明自己消费的种类
  - producer / consumer 都需要可被查询与列出

## Goal

- 新增一个长期稳定的 `stream` 子协议需求，用于统一表达：
  - `music`
  - `video`
  - `text`
  - 自定义流内容
- 该协议应采用“逻辑流源 + 逻辑消费端 + 自主订阅/控制连接 + 每交付独立状态”的模型：
  - 一个逻辑流源只有一个 producer
  - 每个逻辑消费端创建时必须声明自己消费的 `kind`
  - 同一个逻辑流源可以连接多个 consumer
  - 每次订阅或控制连接都拥有独立的 delivery / ACK / timeout 状态

## Scope

### Must

- 新协议必须与现有 `flow`（workflow）分离，不能复用 `SubProto=6` 与 `flow` 语义。
- 新协议必须支持 producer 先声明可供消费的“逻辑流源”。
- 新协议必须支持 consumer 先声明可消费的“逻辑消费端”。
- consumer 必须能查询或直接订阅某个 producer 声明的流源。
- 控制侧必须能查询 consumer 端点并发起连接 / 断开操作。
- 流源模型必须是：
  - 单生产者
  - 多消费者
- producer 声明流源时必须携带可被 consumer 读取的类型信息。
- consumer 声明消费端时必须携带自己消费的 `kind`。
- 首批默认流种类必须覆盖：
  - `music`
  - `video`
  - `text`
- 同时必须保留扩展能力，用于承载其他内容类型。
- producer 与 consumer 都必须支持被列出与按标识读取。
- consumer 必须支持两种建立 delivery 的方式：
  - 自己订阅 producer 的 source
  - 由控制侧把 source 与 consumer 端点连接起来
- 控制侧连接时必须校验 source.kind 与 consumer.kind 匹配。
- 每次订阅或控制连接成功后都必须形成独立的交付会话，使不同 consumer 的 ACK / 反压互不污染。
- 数据面必须为二进制帧，不能把真实流内容塞回 JSON 控制帧。
- hub 必须只承担最小状态和路由职责，不缓存整段流内容。

### Optional

- 支持流源列表变化通知，例如 producer 上线/下线或流源撤销。
- 支持消费端列表变化通知，例如 consumer 上线/下线或消费端撤销。
- 支持 producer 为流源声明更多标签，例如 `camera`、`screen`、`live`。
- 支持 bounded 内容的 resume / start position。

### Not In Scope

- 不替换当前 `file` 的现有 wire。
- 不替换当前 `flow` 的 DAG workflow 语义。
- 不定义 codec 转码、混流、录制、渲染等媒体处理逻辑。
- 不在首版中承诺真正的 RTP 级低时延 / 抗抖动 / 丢包恢复能力。
- 不在首版中引入一对多广播的持久历史回放。

## Scenarios

1. 节点 A 声明一个 `music` 流源，多个 consumer 同时订阅播放。
2. 节点 A 声明一个 `video` 流源，多个 consumer 订阅观看。
3. 节点 B 创建一个 `text` consumer 端点，用于接收字幕、转写结果、日志文本或聊天内容。
4. 控制侧列出某个 producer 的 `music` 流源和某个 consumer 的 `music` 端点后，发起连接建立 delivery。
5. consumer 自己使用已声明的 `consumer_id` 去订阅某个 source。
6. 节点 A 声明一个自定义流源，例如屏幕增量、传感器帧或日志分片。
7. producer 或 consumer 撤销端点后，所有相关 delivery 都会收到结束语义并关闭。

## Functional Requirements

1. 每个设备都必须能够声明自己生产的流源集合。
2. 每个设备都必须能够声明自己消费的 consumer 端点集合。
3. 每个流源至少必须包含：
   - `source_id`
   - `producer_node`
   - `kind`
   - 可选 `name`
   - 可选 `tags`
   - 可选 `metadata`
4. 每个 consumer 端点至少必须包含：
   - `consumer_id`
   - `consumer_node`
   - `kind`
   - 可选 `name`
   - 可选 `tags`
   - 可选 `metadata`
5. consumer 端点在创建时必须声明自己消费的 `kind`，首版一个 consumer 端点只对应一个 `kind`。
6. source 与 consumer 都必须能在连接前被查询或列出。
7. consumer 对某个流源的每一次自主订阅，或控制侧对 source / consumer 的每一次连接，都必须产生独立的 delivery 标识。
8. delivery 必须绑定：
   - producer
   - consumer
   - `consumer_id`
   - source
   - 进度 / ACK 状态
9. receiver 必须能通过 ACK / credit 告知发送方“已消费到哪里”和“还能继续多少”。
10. 控制侧发起 `connect` 时，数据面仍必须保持 `producer -> consumer` 直连；控制侧不进入正常 DATA / ACK 路径。
11. 协议必须明确 connect 的协调节点与状态归属，避免多跳下出现谁创建 / 谁清理 delivery 的歧义。
10. producer 撤销流源时，系统必须终止该流源下的全部 delivery。
11. consumer 撤销消费端时，系统必须终止该消费端下的全部 delivery。
12. consumer 主动取消订阅时，必须只影响自己的 delivery，不影响同一流源下其他 consumer。
13. 控制侧主动断开连接时，必须只影响目标 delivery，不影响同一 source 下其他 delivery。
14. 协议必须拒绝 `source.kind != consumer.kind` 的连接或订阅请求。
15. producer / consumer 所在节点掉线或连接失效时，系统必须清理相关 delivery，并阻止后续 DATA / ACK 被误接收。
15. 协议必须允许 future profile 将 `file` 的 bounded transfer 映射为 `stream` 的特例。
16. 协议必须允许 producer 在流源级声明默认种类为 `music` / `video` / `text`，并允许扩展更多业务标签。

## Non-functional Requirements

- 兼容性：
  - 新协议不能破坏现有 `file` / `flow` / `exec` / `topicbus` 合同。
- 可读性与可审计性：
  - `source / consumer / subscribe / connect / delivery / ack` 语义必须能直接从 requirements/specs 读出。
- 性能：
  - 数据热路径应避免 JSON 编码和无意义拷贝。
  - 必须防止无界队列和无界缓存增长。
- 可扩展性：
  - 新增流种类、标签和交付模式不应破坏已有字段。
- 安全性：
  - 非法来源、非法 delivery、非法进度值必须显式拒绝或丢弃。

## Edge Cases

- producer 声明了流源，但在 consumer 订阅前已经撤销。
- consumer 声明了消费端，但在控制侧连接前已经撤销。
- 同一个 consumer 重复订阅同一个流源。
- 控制侧尝试把 `video` source 连接到 `text` consumer。
- 多个 consumer 的消费速度不同，ACK / 反压差异很大。
- producer 下线，已经建立的 delivery 需要统一结束。
- controller 发起了 `connect`，但 producer 或 consumer 在提交前掉线。
- consumer 只能按一级 `kind` 做匹配，但看不懂更细粒度标签。
- 某个流源是 bounded 内容，而另一个流源是 live 内容，两者的进度语义不同。

## Acceptance Criteria

1. requirements 明确说明：现有 `flow` 继续表示 workflow，新流协议必须独立命名与编号。
2. requirements 明确说明：整体模型是“单生产者，多消费者”，不是一次性一对一建链模型。
3. requirements 明确说明：producer 声明流源、consumer 声明消费端、每次订阅或控制连接都形成独立 delivery 是首版核心模型。
4. requirements 明确说明：consumer 在创建时必须声明自己消费的 `kind`，且 producer / consumer 都支持被列出。
5. requirements 明确说明：控制侧可以连接或断开同种类的 source / consumer。
6. requirements 明确说明：默认流种类至少包含 `music`、`video`、`text`，同时保留扩展能力。
7. requirements 明确说明：consumer 在查询或订阅 / 连接时都能拿到 producer 声明的流种类和标签。
8. requirements 明确说明：控制侧负责建链编排，但正常 DATA / ACK 不经过控制侧。

## Related Specs

- [../specs/file.md](../specs/file.md)
- [../specs/flow.md](../specs/flow.md)
- [../specs/topicbus.md](../specs/topicbus.md)
- [../specs/stream.md](../specs/stream.md)

## Related Changes

- 待本次 workflow 完成后补充。
