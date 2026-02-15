# 2026-02-15 - DefaultForwardHandler 去 internal（subproto/forward）（PR2-3a）

## 变更背景 / 目标
`DefaultForwardHandler` 是 Dispatcher 的默认 fallback：当收到未知子协议帧时，按配置进行转发（到指定节点/父节点）或丢弃。

为后续“子协议可裁切/可组装、逐步拆库”的彻底重构铺路，需要把该 fallback 从 `internal/handler` 顶层包迁移到可复用的公开实现包，并让装配层与测试不再依赖 `internal/handler` 顶层 import 路径。

本次 PR 目标（保持行为不变）：
1) 将 `DefaultForwardHandler` 迁移到 `subproto/forward`。
2) `modules` 与测试切换到新路径。
3) 如无引用，删除 `internal/handler` 顶层残留（仅保留 `internal/handler/<sub>` 子目录）。

## 具体变更内容
### 删除
- 删除 `internal/handler/common.go`
  - 原文件仅为 `subproto/kit` 的薄封装；在 `DefaultForwardHandler` 迁移后不再需要，避免继续维持 `internal/handler` 顶层包。

### 迁移 / 新增
- `internal/handler/default_handler.go` → `subproto/forward/forward.go`
  - 包名调整为 `forward`
  - 头部克隆等通用操作改为直接复用 `subproto/kit`（行为不变）

### 修改
- `modules/hub.go`
  - 默认 fallback 处理器切换为 `subproto/forward.NewDefaultForwardHandler`
- 测试更新
  - `tests/default_handler_test.go`
  - `tests/integration_root_hub_ping_test.go`

## plan.md 任务映射
- F1 - 迁移 DefaultForwardHandler 到 `subproto/forward` ✅
- F2 - modules 与测试切换到新路径 ✅
- F3 - 清理 `internal/handler` 顶层包残留 ✅（删除 `internal/handler/common.go`）
- F4 - 全量回归 ✅

## 关键设计决策与权衡
- **共享能力复用优先**：forward 直接使用 `subproto/kit`，避免重复实现 Header 克隆/响应发送等工具。
- **保持行为不变**：不改 wire、不改 SubProto/Action 语义；仅调整包路径与依赖组织方式。
- **最小迁移面**：本 PR 仅处理 default forward；其他子协议仍保留在 `internal/handler/<sub>`，遵循“小步多 PR”。

## Code Review（结论：通过）
- 需求覆盖：通过（DefaultForwardHandler 已去 internal；modules/测试已切换新路径）
- 架构合理性：通过（`subproto/forward` + `subproto/kit` 依赖方向清晰，不再依赖 `internal/handler` 顶层包）
- 性能风险：通过（无新增热路径开销；仅包路径调整与工具复用）
- 可读性与一致性：通过（命名清晰；modules 装配入口保持不变）
- 可扩展性与配置化：通过（fallback handler 具备独立包落点，为后续裁切/拆库预留扩展点）
- 稳定性与安全：通过（保留既有 parent forward 与目标节点查找逻辑；未放宽 source mismatch 等校验）
- 测试覆盖情况：通过（默认 forward 的关键单测 + 集成用例均已覆盖并通过）

## 测试与验证方式 / 结果
- 回归测试（通过）：
  - `GOTMPDIR=d:\\project\\MyFlowHub3\\.tmp\\gotmp`
  - `go test ./... -count=1 -p 1`

## 潜在影响与回滚方案
### 潜在影响
- 若存在其它代码仍 import `github.com/yttydcs/myflowhub-server/internal/handler` 顶层包，将在编译期失败；本 PR 已同步更新 `modules` 与相关测试。

### 回滚方案
- 可直接 revert 本 PR 的提交；或按需要回退到迁移前的 handler 位置。

