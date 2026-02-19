# Plan - Server：action 注册模板化补齐（exec/flow/topicbus/management → kit.NewAction）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-action-kit`
- Worktree：`d:\project\MyFlowHub3\worktrees\server-action-kit\MyFlowHub-Server`
- Base：`origin/main`
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）

## 约束（边界）
- 仅改 `MyFlowHub-Server`；不改 Core/Proto/SDK/Win。
- wire 不改：SubProto 值 / Action 字符串 / JSON payload struct / HeaderTcp 语义均保持不变。
- 仅做“注册方式/样板代码”收敛：从 `subproto.BaseAction` 包装类型/结构体，迁移到 `subproto/kit.NewAction(...)`。
- Management 采用 **方案 A**：保留现有文件拆分（`action_*.go`），不合并成一个巨大的 `actions.go`。
- 验收测试必须使用 `GOWORK=off`（避免本地 `go.work` 干扰审计）。

## 当前状态（事实，可审计）
- `subproto/kit.NewAction(...)` 已存在并在部分子协议（如 `auth/varstore`）使用。
- `subproto/exec`、`subproto/flow`、`subproto/topicbus` 仍使用“包装类型 + BaseAction”的注册样式（`actions.go`）。
- `subproto/management` 仍使用“每个 action 一个结构体 + BaseAction”，并在 `management.go:initActions()` 显式 `RegisterAction(&xxxAction{h:h})`。

---

## 1) 需求分析

### 目标
1) 将 `exec/flow/topicbus/management` 的 action 注册方式统一为 `kit.NewAction(...)`，做到风格一致、减少样板。
2) 不改变任何 wire/行为：名称、路由、转发、返回、鉴权语义保持不变。
3) 让后续新增 action 的成本更低、可读性更强（注册处“名称 + handler 绑定”一眼可见）。

### 范围（必须 / 可选 / 不做）
- 必须：
  - `subproto/exec/actions.go` 迁移到 `kit.NewAction`。
  - `subproto/flow/actions.go` 迁移到 `kit.NewAction`。
  - `subproto/topicbus/actions.go` 迁移到 `kit.NewAction`。
  - `subproto/management/*`（`action_echo.go` / `action_nodes.go` / `action_config.go` / `management.go`）迁移到 `kit.NewAction`（方案 A：保留拆分文件）。
- 可选（仅当阻塞测试/验收时才做）：
  - 为缺失覆盖的边界补充极小单测（优先复用现有 tests）。
- 不做：
  - 修改 action 名称、消息结构、字段、SubProto 编号、HeaderTcp v2 规则。
  - 调整 handler 的转发策略、鉴权策略、错误码语义。

### 使用场景
- 每个子协议 handler 在 `Init()` 内 `ResetActions()` 后注册 action map；收到帧后按 action 名称查表并调用 `Handle(...)`。

### 功能需求
- 每个原有 action 必须在 initActions 后可被查到（无漏注册）。
- `RequireAuth()` 语义保持与原实现一致（本次涉及子协议当前均为 `false`）。

### 非功能需求
- 性能：仅改注册期代码，不引入 `OnReceive` 热路径的额外开销。
- 可读性：减少“每个 action 一个结构体”的样板；保持各子协议内部结构清晰。
- 可扩展性：未来新增 action 只需要新增 `kit.NewAction(name, handler)` 一行（或极少量辅助函数）。

### 输入输出
- 输入：action name（string）+ data（json.RawMessage）+ conn/hdr。
- 输出：调用原有 handler 逻辑，产生与当前一致的响应/转发结果。

### 边界异常
- action name 为空：`kit.NewAction` 返回 `nil`（调用方不应注册 `nil`）。
- action 重名：保持 `RegisterAction` 现有覆盖语义（不在本 PR 引入额外告警/强校验）。

### 验收标准
- 代码层面：上述 4 个子协议的 action 注册全部使用 `kit.NewAction`（不再出现 BaseAction wrapper/结构体 action 注册）。
- 测试层面（必须）：
  - `GOWORK=off go test ./... -count=1 -p 1`
  - `GOWORK=off go test ./tests -run TestRootHubPing -count=1`

### 风险
- 漏注册 action 导致 runtime unknown action：通过回归测试 + 冒烟测试降低风险。
- 迁移时误改业务逻辑：坚持只做“注册样式”变更，handler 业务函数不动。

## 问题清单
- 阻塞：否

---

## 2) 架构设计（分析）

### 总体方案（采用）
- 使用 `github.com/yttydcs/myflowhub-server/subproto/kit.NewAction(name, handler, opts...)` 替代当前的：
  - `type xxxAction struct { subproto.BaseAction; name string; fn func(...) }`
  - 或 `type xxxAction struct { subproto.BaseAction; h *Handler }` + `RegisterAction(&xxxAction{h:h})`
- 迁移只影响“action 对象的构造方式”，不改变 handler 的调用链：
  - `ResetActions()` → `RegisterAction(act)` → `LookupAction(name)` → `act.Handle(...)`

### 模块职责
- `subproto/kit`：提供 action 构造模板（函数式 action），减少样板；可选提供 kind（仅用于组织/可观测，不参与 wire）。
- `subproto/*`：各子协议定义 action 常量、消息结构与业务 handler；仅在 `registerActions` 层绑定 action → handler。

### 数据 / 调用流
1) Handler.Init → initActions → ResetActions → RegisterAction(kit.NewAction(...))
2) OnReceive 解包出 `Action` + `Data`
3) LookupAction(Action) → Handle(ctx, conn, hdr, Data) → 进入既有业务函数

### 接口草案（本次不新增对外接口）
- 保持现有 `registerActions(h) []core.SubProcessAction`（或拆分为 `registerXxxActions(h)`）返回 action 列表，统一由 `initActions()` 注册。

### 错误与安全
- 不改变鉴权：`RequireAuth()` 默认 `false`；若未来某 action 需鉴权，使用 `kit.WithRequireAuth(true)` 显式声明（本 PR 不引入）。
- 不改变错误码/返回内容：仍由原业务函数构造并发送。

### 性能与测试策略
- `kit.NewAction` 只在初始化期构造闭包对象；`OnReceive` 仍为一次查表 + 一次函数调用。
- 使用既有测试覆盖“漏注册/行为回退”风险，并补充必要的冒烟测试命令。

### 可扩展性设计点
- 未来可在不改 wire 的前提下，为 action 增加统一的可观测/统计（必须另起 PR，避免语义漂移）。

## 问题清单
- 阻塞：否

---

## 3.1) 计划拆分（形成文档）

> 说明：每个任务都必须做到“可审计、可回滚、可验证”。未确认本计划前禁止进入 3.2 写代码。

### Checklist

#### ACT1 - exec：迁移 action 注册到 kit.NewAction
- 目标：移除 `execAction` wrapper，注册处直接绑定 `actionCall/actionCallResp` → handler 方法。
- 涉及文件：
  - `subproto/exec/actions.go`
- 验收条件：
  - `registerActions()` 返回的 action 列表与迁移前一致（2 个 action，名称不变）。
- 测试点：
  - 走 `TEST1/SMOKE1` 覆盖。
- 回滚点：
  - revert 本任务提交。

#### ACT2 - flow：迁移 action 注册到 kit.NewAction
- 目标：移除 `flowAction` wrapper；`set/run/status/list/get` 绑定到既有 handler 方法。
- 涉及文件：
  - `subproto/flow/actions.go`
- 验收条件：
  - action 名称与数量与迁移前一致。
- 测试点：
  - 走 `TEST1/SMOKE1` 覆盖。
- 回滚点：
  - revert 本任务提交。

#### ACT3 - topicbus：迁移 action 注册到 kit.NewAction
- 目标：移除 `topicAction` wrapper；订阅/退订/列表/发布等绑定到既有 handler 方法。
- 涉及文件：
  - `subproto/topicbus/actions.go`
- 验收条件：
  - action 名称与数量与迁移前一致。
- 测试点：
  - 走 `TEST1/SMOKE1` 覆盖。
- 回滚点：
  - revert 本任务提交。

#### ACT4 - management：迁移 action 注册到 kit.NewAction（方案 A：保留拆分文件）
- 目标：
  - 用 `kit.NewAction(...)` 闭包替换结构体 action（`node_echo`、`list_nodes`、`list_subtree`、`config_get`、`config_set`、`config_list`）。
  - `initActions()` 改为统一注册 action 列表（与其他子协议一致），不再显式 new struct action。
- 涉及文件：
  - `subproto/management/management.go`
  - `subproto/management/action_echo.go`
  - `subproto/management/action_nodes.go`
  - `subproto/management/action_config.go`
  - （如需要）`subproto/management/actions.go`（仅做聚合注册，不承载业务逻辑）
- 验收条件：
  - 以上 action 全部可被 Lookup；行为保持不变（响应 code/msg/字段一致）。
- 测试点：
  - 走 `TEST1/SMOKE1` 覆盖。
- 回滚点：
  - revert 本任务提交。

#### TEST1 - 回归测试（强制：GOWORK=off）
- 命令（在本 worktree 根目录执行）：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'`
  - `New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null`
  - `$env:GOWORK='off'`
  - `go test ./... -count=1 -p 1`
- 验收条件：全部通过。
- 回滚点：revert 相关提交，或回退到通过测试的节点。

#### SMOKE1 - 冒烟验证（强制）
- 命令（在本 worktree 根目录执行）：
  - `$env:GOWORK='off'`
  - `go test ./tests -run TestRootHubPing -count=1`
- 验收条件：通过。
- 回滚点：同 `TEST1`。

#### CR1 - Code Review（阶段 3.3）
- 逐项结论（通过/不通过）：
  - 需求覆盖、架构合理性、性能风险、可读性与一致性、可扩展性、稳定性与安全、测试覆盖。

#### ARCH1 - 归档变更（阶段 4）
- 新增文档：
  - `docs/change/2026-02-18_server-action-kit-coverage.md`
- 必须包含：
  - 变更背景/目标
  - 具体变更（按 ACT1~ACT4 列出）
  - 关键设计决策与权衡（强调 wire/语义不变）
  - 测试与验证方式/结果（TEST1/SMOKE1 输出要点）
  - 潜在影响与回滚方案

## 问题清单（阻塞：是）
- 请你确认：本 `plan.md` 是否可以作为本 worktree 的执行计划？确认后我进入 3.2 开始改代码。

