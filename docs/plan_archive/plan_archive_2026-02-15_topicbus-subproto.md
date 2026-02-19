# Plan - TopicBus 迁移到 subproto/topicbus（PR2-TopicBus）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-subproto-topicbus`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr2-server-topicbus\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`
- 约束：commit 信息使用中文（前缀如 `refactor:` 可英文）

## 当前状态
- TopicBus 子协议实现位于 `internal/handler/topicbus/`，只能在本仓库内部引用。
- `modules/hub.go` 与 `tests/topicbus_handler_test.go` 直接 import `internal/handler/topicbus`。
- 文档 `docs/4-topicbus.md` 明确引用当前实现路径；迁移后需要同步更新。

> 环境备注（不进 git）：本仓库 `go.mod` 使用 `replace ../MyFlowHub-Core`、`../MyFlowHub-Proto`。  
> 在本 worktree 布局下，需要在 `d:\project\MyFlowHub3\worktrees\pr2-server-topicbus\` 下存在同名目录。  
> 当前通过 Junction 指向 `d:\project\MyFlowHub3\repo\MyFlowHub-Core` / `repo\MyFlowHub-Proto` 满足。

## 1) 需求分析
### 目标
1) 将 TopicBus handler 从 `internal/handler/topicbus` 迁移到 `subproto/topicbus`（公开可装配的子协议模块），为后续拆库/裁切做准备。
2) 更新 `modules`、测试与文档引用新路径，移除对 `internal/handler/topicbus` 的直接依赖。
3) 保持行为与 wire 不变：不调整 action 名称、不引入权限/持久化/离线堆积等新语义。

### 范围
#### 必须（本 PR）
- 新增 `subproto/topicbus/`，承载 `TopicBusHandler` 全部实现（由 `internal/handler/topicbus` 迁移）。
- `subproto/topicbus/types.go` 直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/topicbus`（减少对 Server 兼容壳的耦合；wire 不变）。
- `modules/hub.go` 默认集合改用 `subproto/topicbus.NewTopicBusHandlerWithConfig`。
- 更新测试 import：
  - `tests/topicbus_handler_test.go`
- 更新文档：
  - `docs/4-topicbus.md`（实现路径与集成提示）
- 清理：删除 `internal/handler/topicbus` 目录，确保仓库内不再引用该路径。
- 回归：`go test ./... -count=1 -p 1` 通过（Windows）。

#### 不做（本 PR）
- assist_* / up_* / notify_* 命名收敛（wire 变更）。
- 订阅持久化、历史回放、权限控制、限流/背压等行为改造。
- 将 TopicBus 的发送/响应构造统一迁移到 `subproto/kit`（避免引入行为差异）。
- Linux 构建验收。

### 使用场景
- hub_server 启动时由 `modules.DefaultHub` 装配并注册 SubProto=4 的 handler。
- 运行期处理 topic 的订阅/退订/列表与 publish 的逐级转发（本 PR 不改行为，仅迁移位置）。

### 功能需求（保持既有约定）
- subscribe/unsubscribe/list_subs 的响应保持既有 code/msg/字段语义。
- publish：
  - 本层先向下扇出（不回显给发布连接），再向上转发父节点；
  - publish 来自父节点时不再向上回传，避免回环。
- 逐级订阅汇总：topic 在本节点订阅者从 0→1/1→0 时，best-effort 向父节点发送 subscribe_batch/unsubscribe_batch。
- 连接断开清理：订阅关系仅在内存中维护，收到 `conn.closed` 事件后清理并在必要时向上退订汇总。
- 父连接变化：当父连接变更且已具备 nodeID 时触发全量重订阅（保持当前实现策略）。

### 非功能需求
- 性能：仅包路径迁移与 import 调整，不引入热路径额外开销；避免额外 I/O、重复计算。
- 可维护性：变更最小化、可回滚、文档与代码一致。

### 输入输出
- 输入：`OnReceive(ctx, conn, hdr, payload)`（payload 为 topicbus JSON envelope）。
- 输出：
  - OK 响应：`MajorOKResp`，`SubProto=4`；
  - publish 转发：`MajorCmd`，`SubProto=4`，向订阅者/父节点发送。

### 边界异常
- 非法 JSON / unknown action：仅告警/调试日志并丢弃（保持）。
- publish.name 为空：丢弃（保持）。
- 无 server context（`core.ServerFromContext(ctx)==nil`）：仅能走直连 `conn.SendWithHeader` 的分支（保持）。

### 验收标准
- `modules`、测试与文档不再引用 `internal/handler/topicbus`。
- `rg "internal/handler/topicbus" ./` 在仓库内无命中（允许 `plan_archive_*` 不参与验收）。
- `go test ./... -count=1 -p 1` 通过（Windows）。

### 风险
- 迁移漏改 import 导致编译失败（`go test` 可覆盖）。
- 文档更新遗漏导致“实现路径/集成提示”不一致（需同步更新）。

## 2) 架构设计（分析）
### 总体方案（含选型理由 / 备选对比）
- 方案 A（采用）：`git mv internal/handler/topicbus -> subproto/topicbus`，并切换装配层/测试/文档到新 import path。
  - 优点：最小 diff、行为稳定、符合“子协议模块可装配/可裁切”的方向。
  - 缺点：仍在 Server 仓库内；后续若要独立为单独库，再做下一轮拆分。
- 方案 B（不选）：保留 internal 实现，在 `subproto/topicbus` 做 wrapper 转发。
  - 缺点：引入额外间接层与重复维护点，不利于后续拆库。

### 模块职责
- `subproto/topicbus`：TopicBus 子协议处理（SubProto=4），包含订阅关系管理、publish 扇出与逐级转发、上游订阅汇总与重订阅逻辑。
- `modules`：装配入口，负责创建 handler 并注册到 dispatcher。
- `protocol/topicbus`：兼容壳（保留旧 import path），对外仍可用，但 `subproto/topicbus` 直接依赖 Proto 协议包以降低耦合。

### 数据 / 调用流
1) `modules.DefaultHub` 构造 `topicbus.NewTopicBusHandlerWithConfig(cfg, log)` 放入 `Set.Handlers`
2) `modules.RegisterAll` -> `dispatcher.RegisterHandler(handler)`
3) dispatcher 按 `SubProto()==4` 分发到 handler，handler 内部按 `msg.Action` 查 action entry 并处理

### 接口草案
- 对外构造：
  - `NewTopicBusHandler(log *slog.Logger) *TopicBusHandler`
  - `NewTopicBusHandlerWithConfig(cfg core.IConfig, log *slog.Logger) *TopicBusHandler`
- 常量：
  - `SubProtoTopicBus uint8`

### 错误与安全
- 不引入新的权限/鉴权逻辑；继续依赖 Core 的预路由与连接元数据（如 nodeID/role）保证基本约束。
- 维持“不回显 publish”“父链防回环”的安全默认行为。

### 性能与测试策略
- 性能：保持当前“快照订阅者后发送”的策略，避免持锁网络发送；本 PR 不改热路径算法。
- 测试：
  - 现有 `tests/topicbus_handler_test.go` 回归覆盖订阅/列表/退订、上游汇总、publish 扇出与不回显、向上转发。
  - 执行：`$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`

### 可扩展性设计点
- `subproto/<name>` 作为子协议模块统一落点，后续可与 `varstore/management/forward` 形成一致结构，并逐步抽成独立 module/library。

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（范围明确，且本 PR 坚持最小迁移，wire/行为不变）。

### TB0 - 归档旧 plan.md 并准备本 workflow 文档
- 目标：保留历史 plan 的可审计性，不影响当前 workflow。
- 涉及文件：
  - `plan_archive_2026-02-15_varstore-subproto.md`
  - `plan.md`
- 验收条件：新 `plan.md` 仅描述本次 TopicBus 迁移。
- 回滚点：revert 文档提交。

### TB1 - 迁移 TopicBus 到 subproto/topicbus
- 目标：`TopicBusHandler` 从 `internal/handler/topicbus` 迁移到 `subproto/topicbus`。
- 涉及模块/文件（预期）：
  - `internal/handler/topicbus/*` → `subproto/topicbus/*`
- 验收条件：
  - 包名保持 `topicbus`，对外构造函数签名不变。
  - `SubProto()==4` 等关键声明保持不变。
- 测试点：`go test ./...`。
- 回滚点：revert 本迁移提交。

### TB2 - subproto 直连 MyFlowHub-Proto 协议包
- 目标：`subproto/topicbus` 直接 import `github.com/yttydcs/myflowhub-proto/protocol/topicbus`。
- 涉及文件：
  - `subproto/topicbus/types.go`
- 验收条件：仅 import 路径变化，常量/类型一致，wire 不变。
- 回滚点：revert。

### TB3 - modules 装配切换到新路径
- 目标：`modules/hub.go` 使用 `subproto/topicbus`。
- 验收条件：默认装配集合仍启用 topicbus。
- 回滚点：revert。

### TB4 - 测试切换到新路径
- 目标：`tests/topicbus_handler_test.go` import 使用 `subproto/topicbus`。
- 验收条件：测试编译通过并运行通过。
- 回滚点：revert。

### TB5 - 文档同步更新
- 目标：`docs/4-topicbus.md` 更新实现路径与集成提示为 `subproto/topicbus`。
- 验收条件：文档不再提及 `internal/handler/topicbus`。
- 回滚点：revert。

### TB6 - 清理旧目录与引用
- 目标：移除 `internal/handler/topicbus`，并确保仓库内无引用。
- 验收条件：`rg "internal/handler/topicbus"` 无命中（排除 `plan_archive_*`）。
- 回滚点：revert。

### TB7 - 全量回归
- 目标：确保迁移不破坏编译/测试。
- 验收条件：`go test ./... -count=1 -p 1` 通过（Windows）。

### TB8 - Code Review + 归档变更
- 目标：按模板完成审查与归档。
- 涉及文件：
  - `docs/change/YYYY-MM-DD_topicbus-subproto.md`
- 验收条件：归档包含任务映射、关键决策、测试结果与回滚方案。

## 注意事项
- 禁止计划外改动：若迁移过程中发现必须同步修改其它子协议或引入 wire 变更，必须回到 3.1 更新计划并重新确认。

## 执行记录
- 2026-02-15：完成 TB1-TB7；回归 `go test ./... -count=1 -p 1` 通过。
