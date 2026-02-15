# Plan - Server 模块装配层（modules）地基（PR2-1a）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-modules`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr2-server-modules\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`

## 1) 需求分析
### 目标
1) 为 `cmd/hub_server` 引入统一的“模块装配层（modules）”，集中管理子协议 handler 的启用集合。
2) 为后续“子协议可裁切/可组装（编译期裁切）”铺路：让启用集合从 `main.go` 中抽离，后续可用 build tags / 多套 set 选择。
3) 保持行为不变：不改 wire、不改 SubProto、不改权限语义、不改变现有 handler 的业务逻辑。

### 范围
#### 必须（本 PR）
- 新增 `modules` 包：提供 hub_server 默认启用集合（auth/varstore/topicbus/file/flow/management/exec + default forward）。
- `cmd/hub_server/main.go` 改为使用 `modules` 装配（替代手写 RegisterHandler 列表）。
- 兼容 `flow` 的 `BindServer` 需求（启动后绑定发送能力），不得回退功能。
- 回归：`go test ./...` 通过（建议统一使用 `-p 1` 规避环境 OOM）。
- 为装配层补齐最小单测（例如：SubProto 唯一性、模块集合非空）。

#### 可选（后续 PR）
- 将 `internal/handler/*` 去 internal 化（迁移到 `subproto/*` 可复用包）。
- build tags / 多套 set（例如 minimal/full）选择。

#### 不做（本 PR）
- 改 wire（action 名称/消息结构/SubProto/权限名）。
- assist/up/notify 的 wire 收敛（策略 A：仅做代码层抽象，wire 不变）。
- LoginServer 的装配改造（本 PR 聚焦 hub_server）。

### 使用场景
- 未来新增/裁切子协议时，不需要改动 `cmd/hub_server/main.go` 的大量重复注册代码。
- 为后续拆库做准备：modules 成为“统一装配入口”，子协议实现迁移时 main 不需要大改。

### 功能需求
- 输出一个 hub_server 的默认启用集合（handlers + fallback）。
- 提供装配工具（返回 slice 或一键注册函数），并对错误（nil handler/重复 SubProto）给出明确提示。
- 支持启动后 hook：对实现了 `BindServer(core.IServer)` 的 handler 执行绑定（用于 flow scheduler/发送能力）。

### 非功能需求
- 性能：装配层仅发生在启动期；不在运行期引入额外热路径开销。
- 可读性：main.go 聚焦“配置/启动/停止”，不承载模块清单。
- 可扩展性：新增子协议只需在 modules 中添加一处注册。

### 边界异常
- handler 构造返回 nil：启动失败并打印明确错误。
- handler SubProto 冲突：启动失败（dispatcher 已防护），测试也应覆盖。

### 验收标准
- `cmd/hub_server` 启动路径仍能注册所有既有子协议 handler。
- flow 的 `BindServer` 仍被调用（或通过统一 hook 机制保证）。
- `go test ./... -count=1 -p 1` 通过。

### 风险
- 若装配层做成“初始化即 Init”，可能改变 `dispatcher.RegisterHandler` 的生命周期假设；本 PR 必须保持由 dispatcher 负责 Init。

## 2) 架构设计（分析）
### 总体方案（含选型理由 / 备选对比）
- 方案 A（本 PR 采用）：modules 仅负责“生成集合/注册集合”，不触碰 handler 实现位置。
  - 优点：最小风险、最小 diff、可快速落地。
  - 缺点：handler 仍在 `internal/handler/*`，可复用性改善有限（留待后续 PR）。
- 方案 B（后续 PR）：同步将 handler 迁移到非 internal 包，并由 modules 引用新包。

### 模块职责
- `modules`：定义 hub_server 的启用集合（默认 set），并提供注册与 post-start 绑定 hook。
- `cmd/hub_server`：解析参数/构建 cfg/创建 dispatcher+server/启动停止；调用 modules 完成装配。

### 数据 / 调用流
1) main 构建 `cfg/log/dispatcher`
2) modules 返回 `handlers + fallback`
3) main 将 handlers 注册到 dispatcher，并设置 fallback
4) main 创建并启动 server
5) main 对实现 `BindServer(core.IServer)` 的 handler 调用绑定（post-start hook）

### 接口草案
- `modules.DefaultHub(cfg, log) (modules.Set, error)`：返回默认集合。
- `modules.RegisterAll(dispatcher, set) error`：将集合注册到 dispatcher（可选封装）。
- `modules.BindServerHooks(srv, set)`：对 handler 执行可选 hook（通过 interface 判定）。

### 错误与安全
- 启动期错误直接返回并退出（保持 main 原有策略）。
- 不改变现有权限判定/路由语义。

### 性能与测试策略
- 单测：验证默认集合 SubProto 唯一性；验证 `BindServerHooks` 仅对实现接口者调用。
- 回归：`go test ./... -count=1 -p 1`。

### 可扩展性设计点
- 预留多个 set（例如 `DefaultHub`、`MinimalHub`）的扩展位置；本 PR 只实现 DefaultHub。

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（本 PR 不涉及 wire 与跨仓依赖调整）。

### M1 - 新增 modules 包（默认启用集合）
- 目标：将 hub_server 需要的 handler 清单与装配逻辑从 main.go 抽离。
- 涉及模块/文件（预期）：
  - `modules/hub.go`（或同级文件）
  - `modules/hub_test.go`
- 验收条件：
  - 能构造并返回完整 handler 集合（不调用 Init）。
  - 单测覆盖：SubProto 唯一性、集合非空。
- 测试点：
  - `go test ./... -count=1 -p 1`
- 回滚点：
  - revert `modules/**`。

### M2 - hub_server main 改为通过 modules 装配
- 目标：main.go 只负责启动/停止与 cfg 构建，不手写 handler 清单。
- 涉及模块/文件：
  - `cmd/hub_server/main.go`
- 验收条件：
  - 所有 handler 仍被注册；默认 handler 仍生效；flow 的 `BindServer` 仍执行。
- 测试点：
  - `go test ./... -count=1 -p 1`
- 回滚点：
  - revert `cmd/hub_server/main.go`。

### M3 - 全量回归
- 目标：确认装配抽离后编译/测试稳定。
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过。

## 依赖关系
- 无（本 PR 仅 Server 内部重构）。

## 注意事项
- 禁止计划外改动：若发现必须同时迁移 handler 包位置或修改其它仓库，必须回到 3.1 更新计划并重新确认。
