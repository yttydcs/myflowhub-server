# 变更记录：File 控制响应帧统一为 MajorOKResp

日期：2026-02-17  
Repo：`MyFlowHub-Server`  
分支：`fix/file-resp-major-okresp`  
Worktree：`d:\project\MyFlowHub3\worktrees\file-resp-major-okresp`

## 背景 / 目标
- Core 已确立统一路由规则：
  - `MajorCmd`：必须进入 handler（逐跳可见），Core 不做自动转发/广播。
  - `MajorMsg/MajorOKResp/MajorErrResp`：按 `TargetID` 由 Core 快速转发/广播（转发时递减 `hop_limit`）。
- 目前 `file` 子协议的 CTRL 响应帧（`read_resp/write_resp`）仍使用 `MajorCmd`，导致跨节点返回路径需要依赖 file handler 逐跳转发，不符合统一框架规则。
- 目标：将 `file.read_resp/write_resp` 的 HeaderTcp `Major` 统一为 `MajorOKResp`，使其按 `TargetID` 走 Core 快速转发，中间节点不需要进入 file handler。

## 具体变更内容
### 修改
- `subproto/file/handler.go`
  - `sendCtrlToNode` 构造的 HeaderTcp：`Major` 从 `MajorCmd` 改为 `MajorOKResp`（覆盖 `read_resp/write_resp`）。
- `docs/5-file.md`
  - 补齐 Major 约定：请求（CTRL `read/write`）为 `MajorCmd`；响应（CTRL `read_resp/write_resp`）为 `MajorOKResp`；`DATA/ACK` 为 `MajorMsg`；失败响应仍使用 `MajorOKResp`（错误在 payload `code/msg`）。
- `plan.md`
  - 本 workflow 的需求/架构/拆分计划（便于审计与接手）。
- `plan_archive_2026-02-16_exec-flow-resp-major-ok.md`
  - 归档上一轮计划（避免覆盖历史）。

### 新增
- `tests/file_handler_test.go`
  - 覆盖：非法 `file.read` 触发 `read_resp`、非法 `file.write` 触发 `write_resp`，断言响应帧 `Major==MajorOKResp`。

## plan.md 任务映射
- FMO1：File `read_resp/write_resp` → `MajorOKResp`
- FMO2：单测断言 File 响应帧 Major
- FMO3：更新 `docs/5-file.md` 的 Major 约定与转发边界
- FMO4：Windows 回归 `go test ./...`
- FMO5：Code Review + 归档

## 关键设计决策与权衡
- big-bang：不提供兼容开关（已确认：不考虑兼容旧客户端/旧行为）。
- 失败响应仍使用 `MajorOKResp`：
  - 保持既有“错误在 payload `code/msg`”语义；
  - 暂不引入 `MajorErrResp` 的跨协议标准化，避免同步成本扩大。
- 依赖 Core 的快速转发（`MajorOKResp`）：
  - 收敛中间节点职责为“只转发，不理解”，减少逐跳 handler 解包/分发开销；
  - 注意：若运行时关闭 Core 转发能力（`routing.forward_remote=false`），跨节点的响应返回路径将被 Core 丢弃（与统一框架规则一致）。

## 测试与验证
- Windows：
  - `$env:GOTMPDIR='d:\project\MyFlowHub3\.tmp\gotmp' ; New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null ; go test ./... -count=1 -p 1`
  - 结果：通过。

## 潜在影响与回滚方案
- 潜在影响：
  - mixed-version 网络中，旧节点若仍发送 `MajorCmd` 的 `read_resp/write_resp`，其跨节点转发行为与新节点不一致（本变更按 big-bang 执行）。
  - 若中间节点关闭 Core 快速转发（`routing.forward_remote=false`），响应帧将无法跨节点返回（符合该配置语义，但可能影响联调/排障）。
- 回滚：
  - revert 本次提交即可恢复 `MajorCmd` 行为。

