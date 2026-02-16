# 变更记录：Exec / Flow 响应帧统一为 MajorOKResp

日期：2026-02-16  
Repo：`MyFlowHub-Server`  
分支：`fix/resp-major-okresp`  
Worktree：`d:\project\MyFlowHub3\worktrees\resp-major-okresp`

## 背景 / 目标
- Core 已确立统一路由规则：
  - `MajorCmd`：必须进入 handler（逐跳可见），Core 不做自动转发/广播。
  - `MajorMsg/MajorOKResp/MajorErrResp`：按 `TargetID` 由 Core 快速转发/广播（转发时递减 `hop_limit`）。
- 但 Server 侧仍有两类“返回帧/响应帧”在 header 上使用了 `MajorCmd`：
  - `exec.call_resp`
  - `flow.*_resp`（`set_resp/run_resp/status_resp/list_resp/get_resp`）
- 目标：将上述响应帧统一为 `MajorOKResp`，使其在跨节点链路中由 Core 按 `TargetID` 快速转发，中间节点无需进入子协议 handler 转发，从而降低耦合并减少不必要的 JSON 解包/动作分发开销。

## 具体变更内容
### 修改
- `subproto/exec/handler.go`
  - `sendCallRespToNode` 构造的 HeaderTcp：`Major` 从 `MajorCmd` 改为 `MajorOKResp`。
- `subproto/flow/handler.go`
  - `sendCtrlToNode` 构造的 HeaderTcp：`Major` 从 `MajorCmd` 改为 `MajorOKResp`（覆盖所有 `flow.*_resp` 的发送点）。
- `tests/test_stubs.go`
  - `stubServer.Send(...)` 记录发送帧的 `hdr.Major()`，用于断言响应帧 Major。
- `docs/6-flow.md`、`docs/7-exec.md`
  - 补充 Major 约定：请求帧 `MajorCmd`、响应帧 `MajorOKResp`，并说明失败也使用 `MajorOKResp`（错误在 payload 的 `code/msg`）。

### 新增
- `tests/exec_handler_test.go`
  - 覆盖：本地裁决拒绝（permission denied）触发 `call_resp`，断言其 `Major==MajorOKResp`。
- `tests/flow_handler_test.go`
  - 覆盖：非法 `flow.set` 触发 `set_resp`，断言其 `Major==MajorOKResp`（并使用 `flow.base_dir=t.TempDir()` 避免写入 `./flows`）。

## plan.md 任务映射
- RMO1：Exec `call_resp` → `MajorOKResp`
- RMO2：Flow `*_resp` → `MajorOKResp`
- RMO3：单测断言 Major 变化
- RMO4：更新协议文档（可选项，本次已完成）
- RMO5：Windows 回归 `go test ./...`

## 关键设计决策与权衡
- 采用 big-bang：不提供兼容开关（已确认）。
- 失败响应仍使用 `MajorOKResp`：
  - 兼容既有“错误在 payload `code/msg`”的语义与上层处理方式；
  - 暂不引入 `MajorErrResp` 的跨协议标准化，避免同步成本扩大。
- 保留 handler 内逐跳转发的兼容逻辑不删除：
  - 减少改动面与回归风险；
  - 在 mixed-version 网络中仍可作为兜底路径（尽管新规则下正常链路应由 Core 快转发）。

## 测试与验证
- Windows：
  - `$env:GOTMPDIR='d:\project\MyFlowHub3\.tmp\gotmp' ; New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null ; go test ./... -count=1 -p 1`
  - 结果：通过。

## 潜在影响与回滚方案
- 潜在影响：
  - mixed-version 网络中，旧节点若仍发送 `MajorCmd` 的 `*_resp/call_resp`，其跨节点转发行为与新节点不一致（本变更按 big-bang 执行）。
- 回滚：
  - 直接 revert 本次提交即可恢复 `MajorCmd` 行为。

