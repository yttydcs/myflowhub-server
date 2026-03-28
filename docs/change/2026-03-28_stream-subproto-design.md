# 2026-03-28_stream-subproto-design

## 变更背景 / 目标

- 用户希望在现有协议体系中设计一个更通用的“流”子协议，用于传输语音、视频和自定义流内容。
- 用户随后明确：目标模型不是一次性一对一建链，而是“每个设备声明流源，consumer 订阅 producer，整体单生产者多消费者，且 producer 可标记流种类”。
- 用户继续追加：内建流种类在 `music`、`video` 之外，再增加一个预定义的 `text`。
- 用户继续追加：consumer 在创建时就要声明自己消费的种类；consumer 也需要支持被列出；控制侧还要能把同种类的 source / consumer 连接或断开。
- 当前 `file` 子协议已经具备成熟的 `control + data + ack` 双平面模型，而现有 `flow` 子协议已经稳定表示 DAG workflow orchestration。
- 本轮目标是在 `MyFlowHub-Server/docs` 中补齐长期 requirements/specs 真相，明确：
  - 为什么不能复用现有 `flow`
  - 为什么新的通用流协议应沿用 `file` 的骨架思路
  - 后续跨仓实现时应遵循怎样的稳定合同

## 具体变更内容

### 新增

- `docs/requirements/stream.md`
  - 新增 Stream 子协议的长期需求文档，覆盖目标、范围、场景、功能 / 非功能需求、边界异常与验收标准。
- `docs/specs/stream.md`
  - 新增 Stream 子协议技术规范草案，覆盖：
    - naming / compatibility 边界
    - source / consumer / subscribe / connect / delivery 模型
    - `music|video|text|custom` 的内建 kind
    - DATA / ACK 二进制小头
    - `connect` 的逐级协调时序与 coordinator 规则
    - owner / coordinator / transit hub 的状态归属
    - `disconnect` / `withdraw_*` / 节点掉线的清理规则
    - 路由 / 权限 / profile 设计

### 修改

- `plan.md`
  - 新增本轮 workflow 的初始化、Stage 1、Stage 2、Stage 3.1、Stage 3.2、Stage 3.3、Stage 4 记录。
- `docs/requirements/README.md`
  - 新增 `stream.md` 的入口。
- `docs/specs/README.md`
  - 新增 `stream.md` 的入口。
- `docs/change/README.md`
  - 新增本次 change 归档入口。

### 删除

- 无。

## Requirements impact

- `updated`

## Specs impact

- `updated`

## Lessons impact

- `none`

## Related requirements

- `docs/requirements/flow_data_dag.md`
- `docs/requirements/stream.md`

## Related specs

- `docs/specs/file.md`
- `docs/specs/flow.md`
- `docs/specs/topicbus.md`
- `docs/specs/stream.md`

## Related lessons

- `none`

## 对应 `plan.md` 任务映射

- `STRM-DOC-1`
  - `docs/requirements/stream.md`
  - `docs/requirements/README.md`
- `STRM-DOC-2`
  - `docs/specs/stream.md`
  - `docs/specs/README.md`
- `STRM-DOC-3`
  - `plan.md`
- `STRM-DOC-4`
  - `docs/change/2026-03-28_stream-subproto-design.md`
  - `docs/change/README.md`

## 经验 / 教训摘要

- 在 MyFlowHub 当前语义里，`flow` 已经稳定表示 workflow DAG；任何“媒体 / 字节流”设计若继续沿用 `flow` 这个名字，都会造成协议边界塌陷。
- `topicbus` 适合做“谁订了什么”的类比，但它没有 producer 声明、订阅确认、per-consumer ACK，因此不能直接作为流协议。
- `file` 当前虽然只服务文件场景，但它的 control/data/ack 三层结构已经是设计更通用流协议的最佳直接参考。
- 当需求从“source 可被订阅”继续扩展为“consumer 也要被声明和控制连接”时，最小安全演进不是推翻现有 source 模型，而是在其旁边补一个对称的 consumer endpoint 描述层。
- 当控制侧开始管理外部 consumer 时，必须把“谁是 connect 的协调节点、谁创建 delivery_id、谁负责断线清理”提前写清，否则后续实现很容易出现半开会话和多点抢状态。
- 对于音视频这类 live 内容，当前可靠字节流 transport 更适合统一 source / delivery / routing / framing 语义，而不是承诺 RTP 级实时性。

## 可复用排查线索

- 症状
  - “想做通用流协议，但现有 `flow` 已经有别的含义”
  - “想做单生产者多消费者订阅，不知道应该参考 `topicbus` 还是 `file`”
- 触发条件
  - 需要引入连续数据面，而现有协议多数仍是 JSON 控制面
- 关键词
  - `stream`
  - `file control data ack`
  - `flow workflow`
  - `topic subscribe`
  - `single producer multi consumers`
  - `music video text custom stream`
  - `consumer endpoint connect disconnect`
- 快速检查
  - 先看 `docs/specs/file.md`
  - 再看 `docs/specs/flow.md`
  - 最后看 `docs/specs/stream.md`

## 关键设计决策与权衡

- 选择“新增 `stream` 子协议”，而不是复用现有 `flow`
  - 好处：避免 breaking change 与历史语义污染。
  - 代价：需要新的 SubProto 编号与后续多仓实现工作。
- 选择“逻辑 source + 每次订阅或连接独立 delivery”，而不是把所有 consumer 混成一个共享 ACK 状态
  - 好处：单 producer 多 consumer 可以成立，同时反压与 timeout 仍可独立处理。
  - 代价：实现层需要维护 source 表和 delivery 表两层状态。
- 选择把 `text` 设为内建一级 kind，而不是让所有文本类内容都落入 `custom`
  - 好处：consumer 可以直接识别“这是文本流”，便于字幕、转写、日志等常见场景快速分流。
  - 代价：`kind` 枚举增加一个固定值，但仍通过 `content_type/tags` 控制更细粒度扩展。
- 选择补充 `consumer endpoint` 声明层，而不是把 consumer 能力临时塞进 `subscribe` 请求
  - 好处：consumer 可被 list / get，且 control plane 能稳定引用 `consumer_id` 执行连接或断开。
  - 代价：协议需要多一组 consumer 目录动作和状态表。
- 选择新增 `connect/disconnect`，而不是只保留 consumer 自主 `subscribe/unsubscribe`
  - 好处：既满足控制侧编排，也保留 consumer 自主订阅能力。
  - 代价：需要额外权限点与 kind 匹配校验。
- 选择由协调节点分配 `delivery_id` 并仅保存最小 route index，而不是让控制侧或所有中间节点长期持有完整 delivery 状态
  - 好处：沿用 `file` 的逐级判权思路，同时避免控制侧进入数据面，状态边界也更清楚。
  - 代价：实现层需要明确 coordinator 的建立与清理时序。
- 选择复用 `file` 的 control/data/ack 骨架，而不是直接扩展 `topicbus`
  - 好处：和现有逐跳判权、端到端数据面模型一致，迁移成本低。
  - 代价：live 媒体能力仍受到底层可靠字节流 transport 的自然限制。

## 测试与验证方式 / 结果

- 方式
  - 人工对照以下事实源进行审阅：
    - `docs/specs/file.md`
    - `docs/specs/flow.md`
    - `docs/specs/topicbus.md`
    - `repo/MyFlowHub-Proto/protocol/file/types.go`
    - `repo/MyFlowHub-Proto/protocol/flow/types.go`
    - `repo/MyFlowHub-Proto/protocol/topicbus/types.go`
    - `repo/MyFlowHub-SubProto/file/handler.go`
    - `repo/MyFlowHub-SubProto/flow/handler.go`
  - 索引检查：
    - `docs/requirements/README.md`
    - `docs/specs/README.md`
    - `docs/change/README.md`
- 结果
  - 通过。新的 requirements/specs 已能独立说明设计边界和后续实现方向。

## 3.3 Code Review 结论

- 需求覆盖：通过。requirements 已覆盖目标、范围、场景、验收标准与边界。
- 架构合理性：通过。`stream` 与现有 `file` / `flow` / `exec` / `topicbus` 的职责边界清晰。
- 性能风险：通过。spec 明确要求固定二进制小头、有界缓存和 ACK / credit。
- 可读性与一致性：通过。术语统一为 `source / consumer / subscribe / connect / delivery / producer`。
- 可扩展性与配置化：通过。`kind` / `mode` / `unit_mode` / `tags` / consumer endpoint 描述字段 与 `signal.op` 都可扩展。
- 稳定性与安全：通过。source、consumer、delivery、direction、position 校验以及断线清理都已明确。
- 测试覆盖情况：通过。本轮是文档 workflow，人工审阅与索引检查已完成。
- 子Agent治理与审计：通过。本轮未使用子Agent。

## 潜在影响

- `docs/specs/stream.md` 现在成为后续 `Proto` / `SubProto` / `SDK` / `Win` 讨论通用流协议时的正式入口。
- 若未来实现选择偏离本设计，需要先回到 requirements/specs 层更新，而不是只改代码。

## 回滚方案

- 回退以下文件即可撤销本轮设计归档：
  - `plan.md`
  - `docs/requirements/stream.md`
  - `docs/requirements/README.md`
  - `docs/specs/stream.md`
  - `docs/specs/README.md`
  - `docs/change/2026-03-28_stream-subproto-design.md`
  - `docs/change/README.md`

## 子Agent执行轨迹

- 本轮未使用子Agent。
- Task ID → Agent → Worktree → 文件 → 验收结果
  - `STRM-DOC-1` → 主Agent → `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design` → `docs/requirements/stream.md`, `docs/requirements/README.md` → 通过
  - `STRM-DOC-2` → 主Agent → `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design` → `docs/specs/stream.md`, `docs/specs/README.md` → 通过
  - `STRM-DOC-3` → 主Agent → `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design` → `plan.md` → 通过
  - `STRM-DOC-4` → 主Agent → `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design` → `docs/change/2026-03-28_stream-subproto-design.md`, `docs/change/README.md` → 通过
