# 2026-02-10 HeaderTcp v2（32B）+ Cmd 逐跳路由适配（Server）

## 背景 / 目标
本仓库配合 Core 完成 **HeaderTcp v2（32B）big-bang** 升级，并适配新的路由框架规则：
- `MajorCmd`：必须进入 handler（逐跳可见），Core 不再自动转发
- `MajorMsg/OK/Err`：由 Core 快速转发

目标是保证 server 侧各子协议在新规则下仍能正确“逐跳转发/裁决/响应”，并保持测试通过。

## 具体变更内容
### 新增 / 修改
- `internal/handler/file/handler.go`
  - CTRL 转发路径改为使用 `header.CloneToTCPForForward()`，确保转发时 `hop_limit` 递减。
- `internal/handler/exec/handler.go`
  - `call_resp`（返回帧）在 `target!=local` 时按 `header.TargetID` 逐跳转发（避免因 Core 不再转发 Cmd 而中断返回路径）。
  - 未知 action 在 `target!=local` 时也按 `TargetID` 逐跳转发（兼容新/旧节点动作集合差异）。
- `internal/handler/flow/handler.go`
  - 未知 action 在 `target!=local` 时按 `TargetID` 逐跳转发（用于 `_resp` 等返回帧逐跳回传）。
- `internal/handler/management/management.go`
  - `target!=local` 的管理类 Cmd 现在会按 `TargetID` 逐跳转发（Win 端可对远端节点执行 management 操作）。
  - 转发失败（无路由 / hop_limit 耗尽 / send 失败）时，返回对应 action 的 *_resp（code=404/500）以避免 UI 侧长时间等待。
- `docs/core.md`
  - 同步文档：移除 Core 内 “file CTRL 特判” 的描述，明确 Major 分流与 `hop_limit/TargetID=0` 语义。

### 删除
- 无

## plan.md 任务映射
- S1：适配 Core HeaderTcp v2 / IHeader 变更 ✅
- S2：路由语义与 Major 使用自检 ✅（补齐 Cmd 返回路径逐跳转发与 hop_limit 递减）
- S3：文档同步（core.md）✅

## 关键设计决策与权衡
- big-bang：v1 不再兼容，必须三端同步发布/部署。
- `hop_limit`：在 handler 执行“逐跳转发”时统一递减，避免环路。
- 对于 Cmd 返回帧（如 exec/flow 的 *_resp），选择在 handler 层按 `TargetID` 逐跳转发，保持协议格式不变、避免本 PR 引入大规模 Major 重构。

## 测试与验证
- 单元 / 集成：
  - `go test ./...`（包含 `tests/`）通过
- 跨端冒烟（与同批次 Win/Core）：
  - register/login + management node_echo 端到端收发可跑通（HeaderTcp v2）。

## 潜在影响
- 破坏性 wire 变更：旧 Win/Core 将无法与新 Server 互通。
- 由于 Core 不再自动转发 Cmd，任何仍依赖旧行为的子协议/动作必须在 handler 层补齐逐跳转发（本次已覆盖 exec/flow/management 关键路径；其他子协议后续可按需要补齐统一 fallback）。

## 回滚方案
- 以提交为单位 `git revert` 回退本 PR（需三端同步回退）。

