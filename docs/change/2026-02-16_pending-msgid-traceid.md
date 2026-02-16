# 2026-02-16 Pending 回包继承 MsgID/TraceID（Auth/VarStore）

## 背景 / 目标

后续 `MyFlowHub-SDK v1` 计划基于 HeaderTcp v2 的 `MsgID` 实现通用的请求-响应等待（Awaiter）。但当前 Server 在“中间节点 pending 回落响应给下游”的链路中：

- pending 只记录了 `connID`，未记录原始请求的 `MsgID/TraceID`
- 回落响应发送时多处使用 `sendResp(..., nil, ...)`（即不携带原请求头），导致最终下游收到的响应头 `MsgID/TraceID=0`

本次变更目标：

1. 仅针对 **Auth / VarStore** 的 **pending 回落链路**，让响应头 **继承原始下游请求的 `MsgID/TraceID`**。
2. **wire 不改**（action 名称 / JSON 结构 / SubProto 不变）。
3. 不扩大范围：不统一其它头字段语义（Major/Source/Target 等），仅补齐 `MsgID/TraceID`。

## 具体变更内容

### 1) Auth：pending 记录 msg_id/trace_id，并在回落响应写回

- pending 结构从：
  - `device_id -> connID`
  - 调整为：`device_id -> {conn_id,msg_id,trace_id}`
- 在以下“上游响应 → 中间节点回落 → 下游响应”的路径中补齐头字段：
  - `assist_register_resp` → `register_resp`
  - `assist_login_resp` → `login_resp`
  - `assist_query_credential_resp` → `login_resp`

### 2) VarStore：pending / pendingSubs 记录 msg_id/trace_id，并在回落响应写回

- pending 结构从：
  - `(owner,name,kind) -> []connID`
  - 调整为：`(owner,name,kind) -> []{conn_id,msg_id,trace_id}`
- subscribe pending 从：
  - `(owner,name,subscribe) -> []{conn_id,subscriber}`
  - 调整为：`(owner,name,subscribe) -> []{conn_id,subscriber,msg_id,trace_id}`
- 上游 `assist_*_resp` 到来后，对每个等待者回包时写回其各自 `MsgID/TraceID`。

## 任务映射（plan.md）

- PID0 - 归档旧 plan.md（PR7）
  - 对应提交：`6ff7984`
- PID1 - Auth pending 继承 MsgID/TraceID
  - 对应提交：`1b7106d`
- PID2 - VarStore pending 继承 MsgID/TraceID
  - 对应提交：`1b7106d`
- PID3 - 单测覆盖（断言 MsgID/TraceID）
  - 对应提交：`1b7106d`
- PID4 - 回归测试（Windows）
  - 通过：见下方“测试与验证”
- PID5 - Code Review + 归档变更
  - 本文 + Review 结论（见下方）

## 关键设计决策与权衡

1. **只记录 `MsgID/TraceID`，不存整份 header**：
   - 优点：变更面最小；不引入额外语义绑定；避免把历史不一致的头字段语义“固化”进 pending。
   - 代价：若未来需要更多字段（例如 timestamp），需再扩展（应另起 workflow 明确范围）。
2. **仅修复 pending 回落链路，不动其它直接响应**：
   - 例如 Auth 的 `get_perms/list_roles` 仍使用 `sendResp(..., nil, ...)`，本 PR 不顺手统一，避免范围外扩。
3. **性能**：
   - 仅增加两个 `uint32` 的存取；pending map 的 key 与查找逻辑不变；
   - VarStore 回包原本就会循环发送，多一个 header 字段设置不改变复杂度阶。

## 测试与验证

Windows：

- `$env:GOTMPDIR='d:\project\MyFlowHub3\.tmp\gotmp'`
- `go test ./... -count=1 -p 1`

结果：通过。

新增/增强断言的测试：

- Auth：
  - `TestLoginHandlerAssistRegisterRespFallback`（断言回落响应头 `MsgID/TraceID`）
  - `TestLoginHandlerAssistLoginRespFallback`（断言回落响应头 `MsgID/TraceID`）
  - `TestLoginHandlerAssistQueryCredentialRespFallbackPreservesHeader`（覆盖 `assist_query_credential_resp` 回落）
- VarStore：
  - `TestVarStoreGetMissForwardAndCache`（断言 `assist_get_resp` 回落 `get_resp` 的头字段）

## Code Review 结论（3.3）

- 需求覆盖：通过（pending 回落链路补齐 `MsgID/TraceID`；仅 auth/varstore；wire 不改）
- 架构合理性：通过（pending 元信息局部增强；不改变 Core 路由与子协议职责边界）
- 性能风险：通过（常量级字段存取；无额外 I/O；复杂度阶不变）
- 可读性与一致性：通过（命名明确；辅助函数/结构简单；测试断言清晰）
- 可扩展性与配置化：通过（为 SDK v1 Awaiter 铺路；后续可按需抽象通用 pending 工具，需另起 workflow）
- 稳定性与安全：通过（不放开权限、不改 wire；pending 不命中仍静默，保持现有行为）
- 测试覆盖情况：通过（新增断言覆盖关键 pending 回落链路）

## 潜在影响与回滚方案

### 潜在影响

- pending 回落响应头现在携带原请求的 `MsgID/TraceID`：
  - 对忽略这些字段的客户端无影响；
  - 对基于 `MsgID` 做匹配/等待的客户端是必要增强。

### 回滚方案

- 功能回滚：`git revert 1b7106d`（撤销 pending 记录与回落写回逻辑 + 单测）
- 文档回滚（可选）：revert 本归档提交（`docs(change)`）
