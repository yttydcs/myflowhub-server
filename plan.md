# Plan - modules/defaultset（默认装配集合解耦）（PR5-DefaultSet）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-modules-defaultset`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr5-server-defaultset\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`
- 约束：commit 信息使用中文（前缀如 `refactor:` 可英文）

## 当前状态
- `internal/handler/*` 已全部迁移到 `subproto/*`（management/varstore/topicbus/exec/flow/file/auth）。
- `modules/` 已成为 hub_server 的装配入口，但当前 `modules/hub.go` 仍直接 import 具体子协议包并在 `DefaultHub()` 内硬编码默认启用集合。
- 目标架构（见 `target.md`）建议引入 `modules/defaultset` 承载“默认启用模块集合”，为后续 build tags/裁切与更复杂的装配策略预留落点。

> 环境备注（不进 git）：本仓库 `go.mod` 使用 `replace ../MyFlowHub-Core`、`../MyFlowHub-Proto`。  
> 在本 worktree 布局下，需要在 `d:\project\MyFlowHub3\worktrees\pr5-server-defaultset\` 下存在同名目录。  
> 当前通过 Junction 指向 `d:\project\MyFlowHub3\repo\MyFlowHub-Core` / `d:\project\MyFlowHub3\repo\MyFlowHub-Proto` 满足。

## 1) 需求分析
### 目标
1) 将 hub_server 的“默认装配集合”从 `modules` 包内的硬编码构造逻辑中解耦出来，落到 `modules/defaultset`。
2) 保持行为不变：默认启用的子协议集合、构造方式（`New*WithConfig`）、启动期 `BindServerHooks` 机制不变。
3) 为后续“可裁切/可组装（build tags / 多 main 变体）”提供稳定落点与更清晰的依赖方向。

### 范围
#### 必须（本 PR）
- 新增 `modules/defaultset` 包，提供“hub_server 默认启用模块集合”的构造函数（handlers + default）。
- `modules.DefaultHub(cfg, log)` 改为委托 `modules/defaultset`，避免在 `modules` 包内直接 import 各 `subproto/*`。
- 回归：`go test ./... -count=1 -p 1` 通过（Windows）。

#### 可选（本 PR，如不增加风险）
- 在 `modules/defaultset` 中预留最小扩展点（例如明确的构造函数/类型命名），但不引入 build tags 与复杂裁切逻辑（避免一次 PR 过大）。

#### 不做（本 PR）
- 修改任何子协议 handler 的业务语义、wire 协议、权限点与错误码。
- 引入新的模块依赖解析（Deps）、模块注册表（Module interface）或 build tags 裁切（这些可作为下一轮目标）。
- Linux 构建验收。

### 使用场景
- `cmd/hub_server` 启动时调用 `modules.DefaultHub(cfg, log)` 获取默认集合，并注册到 dispatcher。

### 功能需求
- 默认集合仍包含（与当前一致）：`management/auth/varstore/topicbus/exec/flow/file` + default forward。
- `modules.RegisterAll`、`modules.BindServerHooks` 行为不变。

### 非功能需求
- 性能：仅装配构造逻辑的包边界调整，不引入运行期热路径额外开销。
- 可维护性：变更最小化、可回滚、文档与代码一致。

### 输入输出
- 输入：`DefaultHub(cfg, log)`（cfg 可为 nil；log 可为 nil）。
- 输出：`modules.Set{Handlers, Default}` 与错误（保持现有约定）。

### 边界异常
- `cfg == nil`：各 handler 仍应按既有实现处理（保持）。
- `log == nil`：保持现有默认 logger 处理方式（保持）。

### 验收标准
- `modules/hub.go` 不再直接 import `subproto/*`（只允许通过 `modules/defaultset` 间接依赖）。
- 默认启用集合不变（通过对比构造点 + `go test` 回归确保）。
- `go test ./... -count=1 -p 1` 通过（Windows）。

### 风险
- 漏迁移/漏包含某个 handler，导致 hub_server 默认能力缺失（通过构造点对比与回归测试降低风险）。
- 产生新的 import cycle（通过“defaultset 不反向依赖 modules”或清晰的依赖方向设计避免）。

## 2) 架构设计（分析）
### 总体方案（含选型理由 / 备选对比）
- 方案 A（采用）：新增 `modules/defaultset` 包承载默认集合构造；`modules.DefaultHub` 委托该包构造并继续负责校验。
  - 优点：实现小、风险低；让 `modules` 包更“稳定/抽象”，默认集合成为可替换的策略包；为后续 build tags/裁切预留明确落点。
  - 缺点：仍未引入 module registry/Deps；仅完成“默认集合构造”的分层调整。
- 方案 B（不选）：继续在 `modules` 包内硬编码默认集合，仅通过文件拆分/注释约束。
  - 缺点：`modules` 长期直接依赖所有子协议实现包，不利于裁切与后续拆库。

### 模块职责
- `modules`：定义装配所需抽象（Set/Dispatcher/RegisterAll/BindServerHooks/validateSet），并提供 `DefaultHub` 作为“对外稳定入口”（本 PR 改为委托 defaultset）。
- `modules/defaultset`：提供“默认启用集合”的具体策略实现（集中 import 各子协议实现包）。

### 数据 / 调用流
1) `cmd/hub_server` 调用 `modules.DefaultHub(cfg, log)`
2) `modules.DefaultHub` 委托 `defaultset.Hub(cfg, log)`（或同等命名）获取 handlers/default
3) `modules.DefaultHub` 继续执行 `validateSet` 并返回 `modules.Set`
4) `modules.RegisterAll` 注册 handlers + default
5) `modules.BindServerHooks` 启动期绑定（保持）

### 接口草案
- `modules/defaultset`（新增）：
  - `func Hub(cfg core.IConfig, log *slog.Logger) (handlers []core.ISubProcess, def core.ISubProcess)`
  - 或：`func DefaultHub(cfg core.IConfig, log *slog.Logger) (handlers []core.ISubProcess, def core.ISubProcess)`
  - 以实现清晰、避免 import cycle 为第一优先。

### 错误与安全
- 不引入新的权限/路由语义；仅装配边界调整。

### 性能与测试策略
- 性能：装配期一次性构造；不引入运行期额外开销。
- 测试：
  - 全量回归：`$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`

### 可扩展性设计点
- 后续可在 `modules/defaultset` 基础上引入 build tags / 多 main 变体，实现编译期裁切。
- 后续可引入 module registry（Module interface/Deps）而不破坏 `modules` 现有 API（通过新增 API 并逐步迁移）。

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（目标明确、wire/行为不变，且本 PR 仅做装配层解耦）。

### DS0 - 归档旧 plan.md 并准备本 workflow 文档
- 目标：保留历史 plan 的可审计性，不影响当前 workflow。
- 涉及文件：
  - `plan_archive_2026-02-16_auth-subproto.md`
  - `plan.md`
- 验收条件：新 `plan.md` 仅描述本次 defaultset 解耦。
- 回滚点：revert 文档提交。

### DS1 - 新增 modules/defaultset
- 目标：提供 hub_server 默认启用集合的构造函数（handlers + default）。
- 涉及模块/文件（预期）：
  - `modules/defaultset/defaultset.go`（或等价命名）
- 验收条件：
  - 默认集合包含与当前一致的 handlers + default forward。
  - `defaultset` 只依赖 `core` + `subproto/*`（不反向依赖 `modules`，避免 cycle）。
- 测试点：`go test ./...`。
- 回滚点：revert。

### DS2 - modules.DefaultHub 委托 defaultset
- 目标：`modules.DefaultHub` 不再直接 import 具体子协议包。
- 涉及文件（预期）：
  - `modules/hub.go`
- 验收条件：
  - 对外函数签名不变；
  - `modules/hub.go` 中不再直接出现 `subproto/*` import。
- 回滚点：revert。

### DS3 - 全量回归
- 目标：确保装配层调整不破坏编译/测试。
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过（Windows）。

### DS4 - Code Review + 归档变更
- 目标：按模板完成审查与归档。
- 涉及文件：
  - `docs/change/YYYY-MM-DD_modules-defaultset.md`
- 验收条件：归档包含任务映射、关键决策、测试结果与回滚方案。

## 注意事项
- 禁止计划外改动：若需要引入 module registry/build tags 等更大调整，必须回到 3.1 更新计划并重新确认。

## 执行记录
- 2026-02-16：创建本 workflow worktree 与计划文档（待确认后进入 3.2）。

