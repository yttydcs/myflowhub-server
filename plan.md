# Plan - Pending 回包继承 MsgID/TraceID（Auth/VarStore）（PR8-PendingIDs）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`fix/pending-msgid-traceid`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr8-pending-ids\MyFlowHub-Server`
- Base：`main`
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）
- 约束：
  - 仅改 `MyFlowHub-Server`；仅 `auth/varstore`；仅 pending 回落链路；wire 不改。
  - 所有实现性改动只在本 worktree 内完成；`repo/` 仅用于合并/推送/集成验证。

## 当前状态（事实）
- 后续 `MyFlowHub-SDK v1` 计划按 HeaderTcp v2 的 `MsgID` 做请求-响应等待；但 Server 在“中间节点 pending 回落响应给下游”的路径里，构造响应头时未继承原请求的 `MsgID/TraceID`，导致客户端无法用 `MsgID` 匹配响应。
- 现状证据（不做引用外链，便于本地审计）：
  - Auth pending 仅记录 `device_id -> connID`：`subproto/auth/auth.go`、`subproto/auth/transport.go`。
  - Auth 多处回落响应使用 `sendResp(..., nil, ...)`：`subproto/auth/actions_register.go`、`subproto/auth/actions_login.go`、`subproto/auth/actions_query.go`。
  - VarStore pending 仅记录 connIDs，回落响应通过 `broadcastPendingResp -> sendResp(..., nil, ...)`：`subproto/varstore/varstore.go`。

---

## 1) 需求分析

### 目标
- 仅在 `MyFlowHub-Server` 内，把 **Auth/VarStore 的 pending（中间节点代回包）场景** 的响应头补齐为：继承原始下游请求的 `MsgID` 与 `TraceID`。
- 为下一步 SDK v1 的 “按 MsgID Awaiter” 扫清阻塞点。

### 范围（必须 / 可选 / 不做）
#### 必须（本 PR）
- `subproto/auth`：
  - pending 从 `device_id -> connID` 扩展为 `device_id -> {conn_id,msg_id,trace_id}`。
  - 在以下 pending 链路回落响应时，响应头 `MsgID/TraceID` 必须等于原始请求：
    - `assist_register_resp` → 下游 `register_resp`
    - `assist_login_resp` → 下游 `login_resp`
    - `assist_query_credential_resp`（补齐 credential）→ 下游 `login_resp`
- `subproto/varstore`：
  - pending 从 “仅 connIDs” 扩展为 “每个等待者均记录 `{conn_id,msg_id,trace_id}`”。
  - 在上游 `assist_*_resp` 到来后，对每个等待者回包时写回其各自的 `MsgID/TraceID`。
- 单测覆盖关键链路：断言 pending 回落响应头继承 `MsgID/TraceID`。
- 回归：`go test ./... -count=1 -p 1`（Windows）。

#### 可选（本 PR 如无额外风险）
- 在归档文档中明确：本 PR 只补齐 pending 场景；其它直接响应（例如 auth 的 get_perms/list_roles）仍保持现状，避免范围外扩。

#### 不做（本 PR）
- 不改 wire：action 名称 / JSON 结构 / SubProto 编号不变。
- 不扩到其它子协议（topicbus/file/flow/exec/management 等）。
- 不统一其它头字段语义（Major/Source/Target/TS 等）；本 PR 只关心 `MsgID/TraceID`。

### 使用场景
- device/子节点 → 中间节点 → authority/父节点（assist 上送）→ 中间节点收到上游响应后 **代回包给原发起连接**：
  - 客户端可用 `MsgID`（配合 `TraceID` 诊断）可靠匹配响应。

### 验收标准
- `auth/varstore` 的 pending 回落响应头：
  - `MsgID == 原请求 MsgID`
  - `TraceID == 原请求 TraceID`
- 单测覆盖并通过。
- `go test ./...` 通过（Windows）。

### 风险
- 若某些客户端历史上默认 `MsgID/TraceID=0`，SDK 仍无法等待；但该问题属于调用方未按协议生成 header，不由本 PR 扩展处理。

---

## 2) 架构设计（分析）

### 总体方案（采用）
- pending 结构额外记录 `msg_id/trace_id`：
  - Auth：单等待者（device_id 唯一）→ 保存一份 `{conn_id,msg_id,trace_id}`
  - VarStore：多等待者 → 保存 `[]waiter{conn_id,msg_id,trace_id}`
- 回包时不改变 payload/action，仅在最终写出 header 前写回 `MsgID/TraceID`。

### 为什么不存整份 header
- 本 PR 只为 SDK 等待语义清障；存整头会引入额外语义绑定与后续演进成本。

---

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（范围/方案/验收已确认：仅 pending 场景补齐 MsgID/TraceID）。

### PID0 - 归档旧 plan.md（PR7）
- 目标：保留已完成 workflow 的 plan 文档，避免被覆盖后无法审计。
- 涉及文件：
  - `plan_archive_2026-02-16_auth-assist-resp.md`
  - `plan.md`
- 验收条件：
  - 旧内容完整保存在 archive 文件中；
  - 新 `plan.md` 只描述本 workflow。
- 回滚点：revert 本提交。

### PID1 - Auth：pending 记录 msg_id/trace_id，并在回落响应写回
- 目标：Auth pending 回落响应头继承原请求的 `MsgID/TraceID`。
- 涉及模块 / 文件：
  - `subproto/auth/auth.go`
  - `subproto/auth/transport.go`
  - `subproto/auth/actions_register.go`
  - `subproto/auth/actions_login.go`
  - `subproto/auth/actions_query.go`
- 验收条件：
  - `setPending(...)` 会记录原请求 header 的 `MsgID/TraceID`；
  - `assist_*_resp` / `assist_query_credential_resp` 回落给下游时，响应头写回相同 `MsgID/TraceID`；
  - 不影响非 pending 的直接响应路径。
- 测试点：见 PID3。
- 回滚点：revert 本提交。

### PID2 - VarStore：pending/pendingSubs 记录 msg_id/trace_id，并在回落响应写回
- 目标：VarStore pending 回落响应头继承原请求的 `MsgID/TraceID`（支持多等待者）。
- 涉及模块 / 文件：
  - `subproto/varstore/types.go`
  - `subproto/varstore/varstore.go`
- 验收条件：
  - `addPending(...)` 为每个等待者记录 `{conn_id,msg_id,trace_id}`；
  - `broadcastPendingResp(...)` 逐等待者回包时写回对应 `MsgID/TraceID`；
  - subscribe pending（如涉及）同理。
- 测试点：见 PID3。
- 回滚点：revert 本提交。

### PID3 - 单测：断言 pending 回落响应头继承 msg_id/trace_id
- 目标：锁定行为，避免未来回归。
- 涉及文件：
  - `tests/auth_handler_test.go`
  - `tests/varstore_handler_test.go`
- 验收条件：
  - Auth：在 `assist_register_resp` / `assist_login_resp` 回落用例中，device 收到的响应头 `MsgID/TraceID` 与原请求一致。
  - VarStore：在 “get miss → forward → assist_get_resp → get_resp 回落” 用例中，child 收到的响应头 `MsgID/TraceID` 与原请求一致。
- 回滚点：revert 本提交。

### PID4 - 回归测试（Windows）
- 目标：确保改动不会破坏现有功能。
- 命令（建议统一，避免临时目录权限/并发问题）：
  - `$env:GOTMPDIR='d:\project\MyFlowHub3\.tmp\gotmp'`
  - `New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null`
  - `go test ./... -count=1 -p 1`
- 验收条件：测试通过。
- 回滚点：无（仅验证）。

### PID5 - Code Review（阶段 3.3）+ 归档变更（阶段 4）
- 目标：完成强制 Code Review 与 docs/change 归档。
- 归档文件：
  - `docs/change/YYYY-MM-DD_pending-msgid-traceid.md`
- 验收条件：
  - Review 覆盖：需求/架构/性能/安全/测试；
  - 归档包含：任务映射、关键设计决策、测试命令与回滚方案（`git revert <sha>`）。

## 注意事项（避免范围外扩）
- 本 PR 不处理 Auth 的 `get_perms/list_roles` 等直接响应的 `MsgID/TraceID` 继承问题；若 SDK v1 需要对这些动作也支持 Awaiter，需另起 workflow 明确范围后再做。
