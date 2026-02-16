# Plan - Auth assist 响应收敛（assist_*_resp 支持）（PR7-Auth-AssistResp）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`fix/server-auth-assist-resp`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr7-auth-assist-resp\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`
- 约束：commit 信息使用中文（前缀如 `fix:` 可英文）

## 当前状态（事实）
- 文档 `docs/2-auth.md` 定义了 `assist_register_resp`、`assist_login_resp` 等响应动作。
- 代码 `subproto/auth` 当前会发送 `assist_register_resp` / `assist_login_resp`（权威节点处理 `assist_*` 时）。
- 但 `subproto/auth` 的 action 注册仅包含 `register_resp` / `login_resp`，缺少对上述 `assist_*_resp` 的接收处理：在“非权威节点向上 assist，请求返回时”的链路下，响应会被当作 unknown action 丢弃，导致下游设备/子节点得不到 `*_resp`。

## 1) 需求分析
### 目标
1) 在 `subproto/auth` 中补齐 `assist_register_resp`、`assist_login_resp` 的 action 处理，使其与普通 `register_resp`、`login_resp` 复用同一处理路径。
2) 统一规则（你已确认）：
   - `assist_*_resp` 在中间节点必须被消费（不继续转发），并映射为下游的普通 `*_resp`（避免 assist 语义泄漏给客户端）。
3) 仅针对 Auth 协议修复；不扩到其它子协议；wire 不改。

### 范围
#### 必须（本 PR）
- `MyFlowHub-Server`：
  - `subproto/auth`：注册并处理 `assist_register_resp`、`assist_login_resp`（复用现有 resp 逻辑；对下游仍发送 `register_resp` / `login_resp`）。
  - `tests`：新增单测覆盖 `assist_*_resp` 回落行为。
- 回归：
  - `go test ./... -count=1 -p 1`（Windows）

#### 不做（本 PR）
- 抽象出跨协议通用的 “*_resp/assist_*_resp” 注册器（后续再做）。
- 修改 `assist_*` 的 wire（action 名称、JSON struct）。
- 调整 auth 的头部 Major/Target 规则（保持现有行为，降低风险）。

### 使用场景
- 节点 A 非权威：收到设备 `register/login` → 向父/权威发送 `assist_*` → 收到权威回的 `assist_*_resp` → A 正确消费并向设备回 `*_resp`。

### 功能需求
- 收到 `assist_register_resp` 时：
  - 能 pop pending 的 device_id，更新绑定/路由必要元数据，并向 pending conn 发送 `register_resp`。
- 收到 `assist_login_resp` 时：
  - 同上，向 pending conn 发送 `login_resp`。
- 行为与收到普通 `*_resp` 一致（复用逻辑）。

### 非功能需求
- 性能：仅多两个 action 注册；不增加热路径额外 marshal/unmarshal。
- 可维护性：尽量复用现有代码，不引入重复分支。

### 输入输出
- 输入：上游连接发来的 `{"action":"assist_*_resp","data":...}`。
- 输出：下游 pending 连接收到 `{"action":"*_resp","data":...}`。

### 边界异常
- data.device_id 为空：忽略（与现有 resp 行为一致）。
- pending 不存在/连接不存在：忽略。
- code != 1：仍应回落发送（与现有 resp 行为一致）。

### 验收标准
- 新增测试覆盖 `assist_register_resp`/`assist_login_resp` 回落逻辑。
- `go test ./... -count=1 -p 1` 通过（Windows）。

### 风险
- 若未来某端确实需要把 `assist_*_resp` 透传给客户端，本 PR 的“回落为 *_resp”将与其冲突；目前文档与现有 handler 设计倾向于中间节点消费，因此风险可控。

## 2) 架构设计（分析）
### 总体方案（含选型理由 / 备选对比）
- 方案 A（采用）：在 `subproto/auth` action 注册中，为 `assist_register_resp`/`assist_login_resp` 增加 action entry，并复用现有 resp handler（最终对下游发送普通 `*_resp`）。
  - 优点：wire 不变；行为符合文档；与 `varstore` 对 `assist_*_resp` 的处理方式一致；变更最小、风险低。
- 方案 B（不选）：让权威节点改为回复普通 `register_resp/login_resp`。
  - 缺点：改变既有 wire 语义/文档；对已存在的客户端/节点兼容性不明；需要跨端协调。

### 模块职责
- `subproto/auth`：维护 Auth 协议的 action 注册表与处理逻辑；本 PR 仅补齐响应动作接收。
- `tests/auth_handler_test.go`：覆盖关键链路回归。

### 数据 / 调用流（简化）
1) device → A：`register` / `login`
2) A → authority：`assist_register` / `assist_login`（A 记录 pending）
3) authority → A：`assist_*_resp`
4) A：消费 `assist_*_resp` → 对 device 发送 `*_resp`

### 接口草案
- 不新增对外 API；仅新增 action 注册项。

### 错误与安全
- 不新增权限绕过：仅让既有响应可被正确处理；仍沿用原有签名/白名单逻辑。

### 性能与测试策略
- 性能：常量级注册；无额外 I/O。
- 测试：
  - 单测模拟 A 有 parent 连接，触发 pending + 收到 `assist_*_resp`，断言 device 收到 `*_resp`。

### 可扩展性设计点
- 后续可抽象统一工具：例如 `subproto/kit` 提供 “resp alias 注册” 帮助函数，减少各协议重复实现（另起 workflow）。

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（规则与范围已确认：仅 auth；assist_*_resp 回落为 *_resp）。

### AR0 - 归档旧 plan.md 并准备本 workflow 文档
- 目标：保留历史 plan 可审计性，避免混淆。
- 涉及文件：
  - `plan_archive_2026-02-16_defaultset-buildtags.md`
  - `plan.md`
- 验收条件：新 `plan.md` 仅描述本次 Auth assist 响应收敛。
- 回滚点：revert 文档提交。

### AR1 - 补齐 auth 的 assist_*_resp action 注册
- 目标：`subproto/auth` 能接收 `assist_register_resp` 与 `assist_login_resp` 并走与 `*_resp` 一致的处理。
- 涉及文件：
  - `subproto/auth/actions_register.go`
  - `subproto/auth/actions_login.go`
- 验收条件：
  - `assist_*_resp` 不再被当作 unknown action。
  - 对下游发送的 action 仍为 `register_resp/login_resp`。
- 测试点：见 AR2。
- 回滚点：revert 代码提交。

### AR2 - 新增单测覆盖 assist_*_resp 回落
- 目标：用测试锁住行为。
- 涉及文件：
  - `tests/auth_handler_test.go`
- 验收条件：
  - 模拟链路：device->A register/login → A forward assist → A 收到 assist_*_resp → device 收到 *_resp。
- 回滚点：revert 测试提交。

### AR3 - 回归测试（Windows）
- 目标：确保改动不会破坏现有功能。
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过（Windows）

### AR4 - Code Review + 归档变更
- 目标：完成强制 Review 与归档。
- 涉及文件：
  - `docs/change/YYYY-MM-DD_auth-assist-resp.md`
- 验收条件：归档包含任务映射、关键决策、测试命令与回滚方案。

## 注意事项
- 禁止计划外扩散：本 PR 不触及其它协议的 `assist/up/notify` 收敛；若需要统一抽象，另起 workflow。

## 执行记录
- 2026-02-16：创建本 workflow 分支与计划文档（待确认后进入 3.2）。
- 2026-02-16：确认 plan.md，进入 3.2；环境准备（不进 git）：在 `worktrees/pr7-auth-assist-resp/` 下创建 `MyFlowHub-Core`、`MyFlowHub-Proto` Junction 指向 `repo/`，满足 `go.mod replace ../MyFlowHub-*`。
- 2026-02-16：完成 AR1/AR2（补齐 `assist_*_resp` action + 单测）；提交：`5094e26`。
- 2026-02-16：完成 AR3（Windows 回归通过）：`go test ./... -count=1 -p 1`。
