# 变更归档：action 注册模板化补齐（exec/flow/topicbus/management → kit.NewAction）

## 背景 / 目标
此前 `MyFlowHub-Server/subproto/kit.NewAction(...)` 已用于部分子协议（例如 `auth/varstore`），用于将 action 的 `Name/RequireAuth/Handle` 样板收敛为“函数式 action”，降低新增 action 的样板成本并提升可读性。

但 `exec/flow/topicbus/management` 仍存在旧式注册写法（`subproto.BaseAction` + 包装结构体 / 每个 action 一个结构体），导致同仓内 action 注册风格不一致，也不利于后续继续收敛与维护。

本次目标：**只收敛注册方式，不改 wire 与行为**，将上述 4 个子协议的 action 注册统一迁移到 `kit.NewAction(...)`。

## 约束与边界（确认项）
- 仅改 `MyFlowHub-Server`。
- wire 不改：SubProto 编号 / Action 字符串 / JSON payload struct / HeaderTcp 语义不变。
- 不调整转发、鉴权、错误码、响应结构等运行时语义。
- Management 采用方案 A：保留 `action_*.go` 文件拆分，不合并成单一大文件。

## 具体变更内容（新增 / 修改 / 删除）

### 修改：exec
- `subproto/exec/actions.go`
  - 移除 `execAction + subproto.BaseAction` 包装类型。
  - 改为使用 `kit.NewAction(actionCall, h.handleCall)` 与 `kit.NewAction(actionCallResp, h.handleCallResp)` 注册。

### 修改：flow
- `subproto/flow/actions.go`
  - 移除 `flowAction + subproto.BaseAction` 包装类型。
  - 改为使用 `kit.NewAction(...)` 直接绑定 `set/run/status/list/get` 到既有 handler 方法。

### 修改：topicbus
- `subproto/topicbus/actions.go`
  - 移除 `topicAction + subproto.BaseAction` 包装类型。
  - 改为使用 `kit.NewAction(...)` 直接绑定 subscribe/unsubscribe/list/publish 等 handler 方法。

### 修改：management
- `subproto/management/management.go`
  - `initActions()` 改为“循环注册 action 列表”的统一模式（与其它子协议一致）。
- `subproto/management/action_echo.go` / `action_nodes.go` / `action_config.go`
  - 移除每个 action 一个结构体（含 `subproto.BaseAction`）的写法。
  - 改为使用 `kit.NewAction(...)` + 闭包，直接复用原业务逻辑与响应构造方式（`sendActionResp`、错误码、字段保持一致）。
- 新增 `subproto/management/actions.go`
  - 聚合注册入口：按原有顺序提供 `registerActions(h)` 列表（echo/config/nodes）。

## 对应 plan.md 任务映射
- ACT1：`subproto/exec/actions.go`
- ACT2：`subproto/flow/actions.go`
- ACT3：`subproto/topicbus/actions.go`
- ACT4：`subproto/management/*`
- TEST1 / SMOKE1：见下方验证

## 关键设计决策与权衡
- 采用 `kit.NewAction` 的原因：
  - 降低样板：避免为“仅转调 handler 方法”的 action 定义包装结构体。
  - 可读性更强：注册处可直接看到 `action 名称 → handler` 的绑定关系。
  - 可扩展：未来需要 `RequireAuth` 或 kind（仅用于组织/可观测）时可通过 option 方式扩展，避免在各子协议重复发明写法。
- 为什么不引入更多运行期校验/告警：
  - 本次目标严格限定为“注册方式收敛”，避免引入任何可能影响行为的逻辑；重复注册检测、指标统计等另起 PR 更安全。

## 测试与验证方式 / 结果
在 worktree（`GOWORK=off`）执行：
- `go test ./... -count=1 -p 1`：通过
- `go test ./tests -run TestRootHubPing -count=1`：通过

## 潜在影响与回滚方案
- 潜在影响：
  - 预期无运行时行为变化；风险主要来自“漏注册/写错 action 名称”导致 unknown action。
  - 通过上述回归测试与冒烟测试覆盖降低风险。
- 回滚：
  - 直接 revert 本次提交（或回退到迁移前 commit），即可恢复旧注册方式。

