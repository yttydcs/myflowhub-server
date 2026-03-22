# Plan - Server：File CTRL 响应继承 MsgID/TraceID（支持 SDK v1 Awaiter）（PR18-SERVER-FileCtrl-Ids）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`fix/file-ctrl-inherit-ids`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr18-file-ctrl-await\MyFlowHub-Server`
- Base：`main`
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）
- 依赖（本地 replace/junction）：
  - `..\MyFlowHub-Core` / `..\MyFlowHub-Proto`

## 约束（边界）
- 仅改 `subproto/file` 的 CTRL 响应帧（`read_resp/write_resp`）：
  - 继承请求头的 `MsgID/TraceID`；
  - `Major` 保持 `MajorOKResp`（PR12 已落地）。
- 不改 wire（SubProto/action/JSON/Kind 前缀均不变）。
- 不改 File DATA/ACK 帧与传输状态机。
- 不改其它子协议；不改 Core/Proto/SDK/Win。
- 所有实现性改动只在本 worktree 内完成；`repo/` 仅用于合并/推送/集成验证。

## 当前状态（事实，可审计）
- `file.read_resp/write_resp` 已使用 `MajorOKResp`（Core 可按 TargetID 快速转发）。
- 但当前 `sendCtrlToNode` 从零构造响应头：未设置 `MsgID/TraceID`，导致 SDK v1 Awaiter 无法按 `MsgID+SubProto+Action` 匹配。
- 请求帧逐跳转发使用 `header.CloneToTCPForForward`，会保留 `MsgID/TraceID`，因此只需补齐响应继承即可闭环。

---

## 1) 需求分析

### 目标
1) `read_resp/write_resp` 的响应 HeaderTcp 继承请求的 `MsgID/TraceID`（成功/失败均如此）。
2) wire 不变；`MajorOKResp` 语义不变。

### 验收标准
- 任意 `read_resp/write_resp`：`MsgID == req.MsgID` 且 `TraceID == req.TraceID`。
- `go test ./... -count=1 -p 1` 通过。

---

## 2) 架构设计（分析）

### 总体方案（采用）
- 在 `subproto/file` 发送 CTRL 响应处引入“基于请求头构造响应头”的逻辑：
  - `MsgID/TraceID` 从请求头复制；
  - `Major/SubProto/SourceID/TargetID` 保持既有语义（Source=本节点；Target=请求方）。

### 备选对比
- 备选 A：仅改 SDK（不采用）
  - Server 侧不继承 `MsgID` 时 Awaiter 无法匹配，SDK 单改无解。
- 备选 B：改 wire（不采用）
  - 与策略 A（wire 不变）冲突。

### 错误与安全
- 不改变权限/裁决链路；错误仍通过 payload `code/msg` 表达。
- `TraceID` 仅做继承，不新增敏感信息。

### 性能与测试策略
- 性能：构造响应头时复制两个 `uint32` 字段，开销可忽略。
- 测试：单测断言响应继承；回归 `go test`。

---

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 已确认：本 PR 纳入 Server 修复 `read_resp/write_resp` 继承 `MsgID/TraceID`，以支持 SDK v1 Awaiter。

### FCI1 - 实现：响应继承 ID
- 目标：`read_resp/write_resp` 的 HeaderTcp 继承请求 `MsgID/TraceID`。
- 涉及文件：
  - `subproto/file/handler.go`
- 验收条件：
  - 响应 `MsgID/TraceID` 与请求一致；`MajorOKResp` 不变。
- 回滚点：
  - revert 本提交。

### FCI2 - 单测：断言继承 + 回归 Major
- 目标：覆盖非法 `read/write` 触发 `*_resp`，断言 `MajorOKResp` 且继承 `MsgID/TraceID`。
- 涉及文件：
  - `tests/test_stubs.go`
  - `tests/file_handler_test.go`
- 验收条件：
  - `go test ./...` 通过。

### FCI3 - 回归测试
- 命令：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'`
  - `New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null`
  - `go test ./... -count=1 -p 1`
- 验收条件：通过。

### FCI4 - Code Review（阶段 3.3）+ 归档变更（阶段 4）
- 归档文件：
  - `docs/change/2026-02-18_file-ctrl-inherit-ids.md`
- 验收条件：
  - Review 覆盖：需求/架构/性能/安全/测试；
  - 归档包含：任务映射、关键决策、测试命令与回滚方案。

