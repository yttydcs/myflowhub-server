# 2026-02-18 - File CTRL 响应继承 MsgID/TraceID（read_resp/write_resp）

Repo：`MyFlowHub-Server`  
分支：`fix/file-ctrl-inherit-ids`  
Worktree：`d:\\project\\MyFlowHub3\\worktrees\\pr18-file-ctrl-await\\MyFlowHub-Server`

## 背景 / 目标

SDK v1 Awaiter 的匹配维度为：`MsgID + SubProto + Action`。

File 子协议的 CTRL payload 形态为：`KindCtrl + JSON(message{action,data})`，其中 `read_resp/write_resp` 属于“控制类响应”。此前 Server 在 `subproto/file` 发送 `read_resp/write_resp` 时从零构造响应 HeaderTcp，**未继承请求头的 `MsgID/TraceID`**，导致：

- 上层（Win 等）即使使用 `SendCommandAndAwait(..., expectAction=read_resp)` 发送请求，也无法在 SDK 层按 `MsgID` 匹配到响应；
- 进而表现为 await 超时（但实际响应可能已经到达并被当作 unmatched）。

本次变更目标：

1. 仅针对 File CTRL 响应 `read_resp/write_resp`：响应头 **继承请求头的 `MsgID/TraceID`**（成功/失败均如此）。
2. wire 不变（SubProto/action/KindCtrl 前缀/JSON 结构不变）。
3. `Major` 保持 `MajorOKResp`（此前已按统一框架规则落地）。

## 具体变更内容

### 修改
- `subproto/file/handler.go`
  - `sendCtrlToNode` 在构造响应头时，从请求头复制 `MsgID/TraceID`：
    - `resp.MsgID = req.MsgID`
    - `resp.TraceID = req.TraceID`
  - `sendReadResp/sendWriteResp` 增加请求头入参，以便在 `read_resp/write_resp` 生成处完成继承。

### 测试增强
- `tests/test_stubs.go`
  - stubServer 记录 Send 时的 `msg_id/trace_id`，便于断言。
- `tests/file_handler_test.go`
  - 增强既有用例：在“非法 read/write”触发 `*_resp` 的场景下，断言响应 `MajorOKResp` 且 `MsgID/TraceID` 与请求一致。

## 任务映射（plan.md）
- FCI1 - 实现：响应继承 ID ✅
- FCI2 - 单测：断言继承 + 回归 Major ✅
- FCI3 - 回归测试 ✅
- FCI4 - Code Review + 归档变更 ✅（本文 + Review 结论）

## 关键设计决策与权衡
1. **仅继承 `MsgID/TraceID`，不复用整份请求 Header**：
   - 优点：最小化语义绑定，避免把请求侧的 `Flags/HopLimit/RouteFlags/Timestamp` 等字段含义“复制固化”到响应；
   - 代价：若未来需要继承更多字段，应另起 workflow 明确范围与兼容策略。
2. **不改 wire**：
   - 保持 `KindCtrl + JSON(message)` 形态不变，避免影响现有节点与客户端。
3. **性能**：
   - 仅增加两个 `uint32` 字段复制，对 CPU/内存开销可忽略。

## 测试与验证方式 / 结果

Windows：

```powershell
$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
go test ./... -count=1 -p 1
```

结果：通过。

## Code Review（结论：通过）
- 需求覆盖：通过（`read_resp/write_resp` 继承 `MsgID/TraceID`；wire 不变）
- 架构合理性：通过（仅在 File 子协议响应生成点补齐；不引入跨模块耦合）
- 性能风险：通过（常量级字段复制；无额外 I/O）
- 可读性与一致性：通过（改动点集中；测试断言清晰）
- 可扩展性与配置化：通过（为 SDK v1 Awaiter 闭环铺路；后续若推广到其它 CTRL 响应需另起 workflow）
- 稳定性与安全：通过（不改变权限与路由策略；仅补齐头字段）
- 测试覆盖情况：通过（单测覆盖“失败响应也继承 ID”路径）

## 潜在影响与回滚方案

### 潜在影响
- 对忽略 `MsgID/TraceID` 的客户端无影响。
- 对基于 `MsgID` 做请求-响应匹配的客户端（SDK v1 Awaiter）属于必要修复。

### 回滚方案
- `git revert f07b21f`（撤销 File CTRL 响应继承逻辑 + 对应测试增强）
