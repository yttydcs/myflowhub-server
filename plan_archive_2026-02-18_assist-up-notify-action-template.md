# Plan - Server：assist_* / up_* / notify_* 代码层收敛（wire 不改）（PR2-3）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-action-template`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr2-3-action-template\MyFlowHub-Server`
- Base：`main`
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`（PR2-3：L230）
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）

## 约束（边界）
- 仅改 `MyFlowHub-Server`（本 PR 不改 Core/Proto/SDK/Win）。
- wire 不改：SubProto/Action 字符串/JSON struct 保持不变。
- 不改变既有 send/forward/header 语义（尤其 assist/up/notify 的逐跳链路）。
- 优先落地 `subproto/auth` + `subproto/varstore`，其余子协议不要求本 PR 覆盖。
- 执行顺序：先更新文档（`target.md`、`repos.md`）再改代码。
- 验收必须使用 `GOWORK=off`（避免本地 `go.work` 干扰审计）。

## 当前状态（事实，可审计）
- `subproto/auth`、`subproto/varstore` 的 action 注册方式存在多种写法（struct/包装类型/assisted bool），导致：
  - 增加 action 时容易漏项或写出风格不一致的实现；
  - `assist_* / up_* / notify_*` 的工程承载不统一，未来拆库/裁切更难维护。
- 本 PR 的目标是：**不改变行为** 的前提下，提供统一的 action 分类/注册模板，并将 auth+varstore 收敛到同一写法。

---

## 1) 需求分析

### 目标
1) Server 引入统一 action 模板：用最少样板完成 `Name/RequireAuth/Handle`。
2) 提供 action 分类（Assist/Up/Notify/Local…）用于工程组织与可观测（不改 wire）。
3) 将 `auth` 与 `varstore` 的 `assist_* / up_* / notify_*` 注册方式收敛到同一模板。
4) 保持现有测试与端到端最小链路不回退。

### 范围（必须 / 可选 / 不做）
- 必须：
  - `subproto/kit` 新增 action 模板与 kind 分类能力。
  - `subproto/auth`：迁移 action 注册（至少覆盖 assist_* 与 up_* 及其 resp）。
  - `subproto/varstore`：迁移 action 注册（覆盖 assist_* / up_* / notify_* / var_changed / var_deleted）。
  - 更新文档：`d:\project\MyFlowHub3\target.md`、`d:\project\MyFlowHub3\repos.md`（写清 PR2-3 的目标/边界/验收/风险/回滚）。
- 可选（本 PR 不做，除非阻塞验收）：
  - topicbus/file/flow/exec 的同步迁移到模板（后续拆分 PR 分阶段做）。
  - 启动期重复 action 检测/告警（不影响现有语义）。
- 不做：
  - 修改 action 名称、payload 结构、SubProto 值、HeaderTcp 编解码/major 路由规则。
  - 基于 kind 的统一转发/发送策略（触及语义，必须另起 PR）。

### 使用场景
- 子协议 handler 初始化期注册 action map（`Init()` → `ResetActions()` → `RegisterAction(...)`）。
- 未来新增子协议时复用相同模板，避免每个协议都发明一套 `assist/up/notify` 写法。

### 功能需求
- action 模板需支持：
  - 指定 `name`。
  - 指定 `requireAuth`。
  - 指定 `kind`（默认从 name 推导，允许覆盖）。
  - 绑定 `Handle` 函数（闭包/函数指针）。
- 不引入运行期热路径额外开销（kind 推导/校验仅在注册期完成）。

### 非功能需求
- 可读性：注册处一眼能看出 action 名称与 handler 绑定关系。
- 可扩展性：未来可在不改现有 action 的前提下增加更多 kind（例如 Resp/Local）。
- 安全：不改变现有鉴权/permission 校验语义（`RequireAuth()`、逐跳裁决等）。
- 性能：不在 `OnReceive` 每帧增加字符串判断或额外 marshal/unmarshal。

### 输入输出
- 输入：action name（string），data（json.RawMessage），conn/hdr。
- 输出：调用原有 handler 逻辑，产生相同的响应/转发行为。

### 边界异常
- action name 为空 → 不注册。
- action 重名 → 按现有 `RegisterAction` 语义覆盖（如需告警，仅做启动期 log，不影响行为）。

### 验收标准
- 文档：
  - `d:\project\MyFlowHub3\target.md` / `d:\project\MyFlowHub3\repos.md` 更新并与当前事实一致（例如 PR2-5/semver 不再写“未来/待合并”）。
- 代码：
  - `subproto/auth`、`subproto/varstore` 的 action 注册方式收敛到统一模板（本 PR 范围内）。
  - wire 行为无变化（现有测试覆盖通过）。
- 测试：
  - `GOWORK=off go test ./... -count=1 -p 1` 通过。
  - 冒烟步骤可执行（hub_server + management node_echo）。

### 风险
- 漏注册 action → runtime unknown action；依赖现有单测/集成测试降低风险。
- 过度抽象 → 可读性下降；模板保持“薄”，只封装样板，不封装业务语义。

---

## 2) 架构设计（分析）

### 总体方案（采用：方案 A）
- 在 `subproto/kit` 增加 `ActionKind` + `FuncAction`（实现 `core.SubProcessAction`），提供统一 `NewAction(...)` 构造。
- kind 推导规则（默认）：
  - 前缀 `assist_` → Assist
  - 前缀 `up_` → Up
  - 前缀 `notify_` → Notify
  - 其它 → Local/Other
  - 允许显式覆盖（用于 `var_changed/var_deleted` 这类无前缀但语义为通知的 action）。
- 迁移方式：
  - `auth`：将 assisted bool/resp/up 相关 action 的注册改为 `kit.NewAction(...)` 列表（业务逻辑不变）。
  - `varstore`：移除 `varAction` wrapper，改为 `kit.NewAction(...)` 列表（业务逻辑不变）。
- 不改：
  - `subproto/kit` 现有 header/response helper 的语义（仅追加新文件/新能力）。
  - Core 的路由规则/Dispatcher 行为。

### 模块职责
- `subproto/kit`：action 模板 + kind 工具（仅承载“样板”与分类元信息）。
- `subproto/auth`：保留现有业务实现；替换 action 注册样板。
- `subproto/varstore`：保留现有业务实现；替换 action 注册样板。

### 数据 / 调用流（不变）
1) `OnReceive` 解包 `message{action,data}`
2) `LookupAction(action)`
3) `action.Handle(ctx, conn, hdr, data)` 执行业务逻辑

### 接口草案（拟在 kit 内提供）
- `type ActionKind uint8`
- `type ActionHandler func(ctx context.Context, conn core.IConnection, hdr core.IHeader, data json.RawMessage)`
- `func NewAction(name string, h ActionHandler, opts ...ActionOption) core.SubProcessAction`
- `func KindFromName(name string) ActionKind`
- （可选）`func (a *Action) Kind() ActionKind`（用于调试/可观测，不参与路由语义）

### 错误与安全
- 模板层不吞/不替换业务错误；业务仍通过现有 `respData/varResp` 的 `code/msg` 表达。
- `RequireAuth()` 值保持原有定义，不改变 source mismatch 等安全语义。

### 性能与测试策略
- kind 推导仅在注册期（Init）执行；运行期不增加额外判断。
- 覆盖：复用现有 `tests/auth_handler_test.go`、`tests/varstore_handler_test.go`、集成测试，降低“漏注册”风险。

### 可扩展性设计点
- 其它子协议（file/flow/exec/topicbus）后续可逐步迁移到同一模板（不要求本 PR 一次完成）。
- kind 后续可用于统一日志/metrics/调试输出，但必须另起 PR（避免本 PR 引入语义漂移）。

---

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 已确认：方案 A、wire 不改、仅改 Server、优先 auth+varstore、不做兼容开关、验收命令与冒烟方式。

### DOC1 - 更新全局文档（先做）
- 目标：`d:\project\MyFlowHub3\target.md` 与 `d:\project\MyFlowHub3\repos.md` 反映当前真实状态，并补充 PR2-3 细节与验收。
- 涉及文件：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
- 验收条件：
  - PR2-5/semver 等状态不再写“未来/待合并”，改为已完成。
  - PR2-3 有明确：范围/不做/验收/风险/回滚。
- 测试点：无。
- 回滚点：手工恢复文件历史内容（建议先本地备份）。

### KIT1 - kit：新增 action 模板与 kind
- 目标：提供统一 `NewAction` + `ActionKind` + `KindFromName`（支持显式覆盖 kind）。
- 涉及文件：
  - `subproto/kit/kit.go`（如需，仅追加；不改变现有函数语义）
  - `subproto/kit/action.go`（新增）
- 验收条件：
  - `go test` 编译通过。
  - kind 推导仅在注册期执行（不进入热路径）。
- 测试点（可选其一）：
  - 新增轻量单测覆盖 kind 推导（推荐）。
- 回滚点：revert 该提交。

### AUTH1 - auth：迁移 assist/up 注册到 kit 模板
- 目标：不改业务逻辑，仅收敛 action 注册样板（assist_* / up_* 及对应 resp）。
- 涉及文件（预期）：
  - `subproto/auth/actions_register.go`
  - `subproto/auth/actions_login.go`
  - `subproto/auth/actions_up_login.go`
  - `subproto/auth/actions_up_login_register.go`（可能合并/删除）
  - 其它 auth actions：按“可读性优先”决定是否一起迁移（不强制）。
- 验收条件：
  - 所有 action 仍被 `initActions()` 注册（无漏项）。
  - `go test ./...` 通过（含 auth 相关测试）。
- 测试点：
  - `tests/auth_handler_test.go`
- 回滚点：revert 该提交。

### VSTORE1 - varstore：迁移 assist/up/notify 注册到 kit 模板
- 目标：移除 `varAction` wrapper，用 `kit.NewAction(...)` 显式注册（含 `var_changed/var_deleted`）。
- 涉及文件：
  - `subproto/varstore/actions.go`
- 验收条件：
  - `go test ./...` 通过（含 varstore 相关测试/集成）。
- 测试点：
  - `tests/varstore_handler_test.go`
  - `tests/integration_varstore_end_to_end_test.go`
- 回滚点：revert 该提交。

### TEST1 - 回归测试（GOWORK=off）
- 命令：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'`
  - `New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null`
  - `$env:GOWORK='off'`
  - `go test ./... -count=1 -p 1`
- 验收条件：全部通过。

### SMOKE1 - 冒烟步骤（hub_server + node_echo）
- 目标：提供可执行步骤（写入 `docs/change`），用于人工验证最小链路。
- 建议步骤（可二选一，优先 1 便于自动化）：
  1) 跑自建 root+hub 的集成测试：`go test ./tests -run TestRootHubPing -count=1`
  2) 手动启动 `cmd/hub_server`，再用最小 client 发送 management `node_echo`（如后续补充脚本）。
- 验收条件：收到 `node_echo_resp` 且 `code=1`、`echo=ping`。

### CR1 - Code Review（阶段 3.3）
- 逐项审查：需求覆盖/架构/性能/可读性/可扩展性/稳定性与安全/测试覆盖。

### ARCH1 - 归档（阶段 4）
- 新增：`docs/change/YYYY-MM-DD_assist-up-notify-action-template.md`
- 内容必须包含：背景、具体变更、任务映射（DOC1/KIT1/AUTH1/VSTORE1/TEST1/SMOKE1）、关键决策与权衡、验证方式与结果、影响与回滚方案。
- 验收条件：文档可独立复现。

### SRVSEM4 - Code Review（阶段 3.3）+ 归档（阶段 4）
- 验收条件：Review 结论为“通过”。

### SRVSEM5 - 合并（你确认结束 workflow 后执行）
- 目标：合并到 `main` 并 push。
- 步骤（在 `repo/MyFlowHub-Server` 执行）：
  1) `git merge --ff-only origin/chore/server-semver-deps`
  2) `git push origin main`
- 回滚点：
  - revert 合并提交（或 revert 分支提交）。
