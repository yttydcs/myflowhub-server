# 2026-02-15 - hub_server 模块装配层（modules）地基（PR2-1a）

## 变更背景 / 目标
在彻底重构（Core/Server/子协议解耦）路线中，Server 需要先把“启用哪些子协议 handler”的装配逻辑从 `cmd/hub_server/main.go` 抽离出来，形成统一入口，以便后续：
- 逐步将各子协议迁移为可复用包/独立库（保持互相解耦）
- 支持按需裁切/组装（例如多套 set、build tags）

本次变更目标（保持行为不变）：
1) 为 `cmd/hub_server` 引入 `modules` 装配层，集中管理默认启用的子协议 handler 集合
2) 保持 wire/SubProto/权限语义/现有 handler 业务逻辑不变
3) 兼容 `flow` 的 `BindServer`（启动后绑定发送能力）

## 具体变更内容
### 新增
- `modules/hub.go`
  - 定义 `modules.Set`：默认 handler 集合 + 默认 fallback
  - 提供 `modules.DefaultHub(cfg, log)`：构造 hub_server 默认启用集合
  - 提供 `modules.RegisterAll(dispatcher, set)`：集中注册 handlers + fallback（不触发 Init）
  - 提供 `modules.BindServerHooks(srv, set)`：对实现 `BindServer(core.IServer)` 的 handler 执行启动后绑定（当前用于 flow）
  - 启动期校验：nil handler、重复 SubProto 直接报错
- `modules/hub_test.go`
  - 覆盖：默认集合非空、SubProto 唯一性、`BindServerHooks` 仅对实现接口者生效

### 修改
- `cmd/hub_server/main.go`
  - 替换手写的 `dispatcher.RegisterHandler(...)` 列表：改为 `modules.DefaultHub` + `modules.RegisterAll`
  - 替换 `flowH.BindServer(srv)`：改为 `modules.BindServerHooks(srv, set)`（保持“Start 后绑定”的时机不变）

## plan.md 任务映射
- M1：新增 `modules` 包（默认启用集合） ✅
- M2：hub_server main 改为通过 modules 装配 ✅
- M3：全量回归（`go test ./...`）✅

## 关键设计决策与权衡
- **薄装配层**：`modules` 只做“构造/校验/注册/启动后 hook”，不迁移 handler 位置、不修改 wire，降低风险、缩小 PR diff。
- **生命周期保持一致**：不在 `DefaultHub` 内调用 `Init`；仍由 `Dispatcher.RegisterHandler` 触发初始化，避免改变既有生命周期假设。
- **扩展点**：未来可在 `modules` 下新增 `MinimalHub()` / `FullHub()` 等多套 set，或配合 build tags 做编译期裁切；`cmd/hub_server` 不需要再维护大量注册代码。
- **性能**：装配与校验只发生在启动期，不引入运行期热路径额外开销。

## 测试与验证方式 / 结果
- 单测：`go test ./... -count=1 -p 1`（通过）
- 冒烟（建议后续由使用者执行）：启动 `hub_server` 并用 `management node_echo` 验证基本链路

> 说明：本工作区的 `go.mod` 使用相对 `replace ../MyFlowHub-Core`、`../MyFlowHub-Proto`。在 MyFlowHub3 的 worktree 布局下，需要确保这些路径存在；本次通过在 `worktrees/pr2-server-modules/` 下建立到 `repo/` 的 Junction 来满足（不进入 git 变更）。

## 潜在影响与回滚方案
### 潜在影响
- 若默认启用集合漏项/顺序差异导致某子协议未注册，将在启动期报错或功能缺失；已通过“集合非空 + SubProto 唯一性”与 `go test` 降低风险。

### 回滚方案
- 回滚 `modules/**`
- 回滚 `cmd/hub_server/main.go` 到手写注册方式
- 删除本变更文档（如需）

