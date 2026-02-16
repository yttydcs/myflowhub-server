# 2026-02-16 Auth assist 响应收敛（assist_*_resp 回落为 *_resp）

## 背景 / 目标

Auth 协议文档 `docs/2-auth.md` 明确存在并使用以下动作：

- `assist_register_resp`
- `assist_login_resp`

当前 `subproto/auth` 的实现会在“权威节点处理 assist_* 请求”时发送 `assist_*_resp`，但在“中间节点（非权威）收到 `assist_*_resp`”的场景下，因缺少 action 注册，响应会被当作 unknown action 丢弃，导致：

- 中间节点无法 pop pending（device_id → connID）
- 下游设备/子节点收不到最终的 `register_resp/login_resp`

本次变更目标（你已确认）：

1. **补齐 Auth 对 `assist_register_resp` / `assist_login_resp` 的接收处理**。
2. 统一规则：`assist_*_resp` 在中间节点必须被消费，并**回落为普通 `*_resp`** 发给下游（避免 assist 语义泄漏给客户端）。
3. 仅针对 Auth 协议修复；**wire 不改**（action 名称/JSON 结构/SubProto 不变）。

## 具体变更内容

### 1) 新增 action：消费 `assist_register_resp` 并回落为 `register_resp`

- 新增 `assistRegisterRespAction`：
  - 触发动作：`assist_register_resp`
  - 行为：复用当前 `register_resp` 的 pending-pop 逻辑，将响应回落给 pending 连接，发送动作 **仍为 `register_resp`**

### 2) 新增 action：消费 `assist_login_resp` 并回落为 `login_resp`

- 新增 `assistLoginRespAction`：
  - 触发动作：`assist_login_resp`
  - 行为：与 `login_resp` 共享处理路径（新增 `handleLoginResp`），将响应回落给 pending 连接，发送动作 **仍为 `login_resp`**

### 3) 新增单测锁定回落行为

- `TestLoginHandlerAssistRegisterRespFallback`
- `TestLoginHandlerAssistLoginRespFallback`

覆盖“device → 中间节点 → authority → 中间节点 → device”的最小链路，验证 `assist_*_resp` 会被中间节点消费并回落为普通 `*_resp`。

## 任务映射（plan.md）

- AR1 - 补齐 auth 的 assist_*_resp action 注册
  - 对应提交：`5094e26`
- AR2 - 新增单测覆盖 assist_*_resp 回落
  - 对应提交：`5094e26`
- AR3 - 回归测试（Windows）
  - 通过：见下方“测试与验证”
- AR4 - Code Review + 归档变更
  - 本文 + Review 结论（见下方）

## 关键设计决策与权衡

1. 将 `assist_*_resp` 视为 `*_resp` 的“上送链路响应变体”，在中间节点消费并回落：
   - 优点：与文档一致；客户端只需理解 `*_resp`；避免 assist 语义外溢。
   - 代价：需要在每个协议内部补齐 alias action（本 PR 仅做 auth，后续可抽象）。
2. 不改权威节点的发送动作（不把 `assist_*_resp` 改为 `*_resp`）：
   - 避免 wire 语义漂移与跨端协同风险；变更范围最小。

## 测试与验证

Windows：

- `go test ./... -count=1 -p 1`

结果：通过。

## Code Review 结论（3.3）

- 需求覆盖：通过（补齐 `assist_*_resp` 接收；回落为 `*_resp`；仅 auth；wire 不改）
- 架构合理性：通过（子协议内部收敛；不改 Core 路由规则）
- 性能风险：通过（仅多两个 action 注册与一次解析路径复用）
- 可读性与一致性：通过（命名清晰；登录响应复用同一处理函数）
- 可扩展性与配置化：通过（后续可抽象为通用 alias 注册器，另起 workflow）
- 稳定性与安全：通过（仍依赖 pending 关联；不放开额外权限）
- 测试覆盖情况：通过（新增两条单测覆盖关键链路）

## 潜在影响与回滚方案

### 潜在影响

- 若未来某端“刻意”要求将 `assist_*_resp` 透传给下游客户端，本变更会改变其观测（客户端只会收到 `*_resp`）。当前文档与既定规则倾向于回落，因此风险可控。

### 回滚方案

- 回滚提交：`git revert 5094e26`（移除 `assist_*_resp` 的 action 注册与单测）

