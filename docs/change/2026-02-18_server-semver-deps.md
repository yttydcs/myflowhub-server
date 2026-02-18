# 2026-02-18 Server：依赖 semver 化并移除 replace

## 变更背景 / 目标
Server 在开发期使用本地 `replace ../MyFlowHub-*` 便于联调，但会导致“单仓 clone 无法构建/不可审计”。本次变更将依赖切换为可拉取的 semver tag，保证发布可复现。

本次目标：
- 移除 `go.mod` 中对 `myflowhub-core` / `myflowhub-proto` 的本地 `replace`
- 固定依赖版本为：
  - `github.com/yttydcs/myflowhub-core@v0.2.0`
  - `github.com/yttydcs/myflowhub-proto@v0.1.0`
- 确保 `GOWORK=off go test ./...` 可通过（不依赖 go.work）

## 具体变更内容
### 修改
- `go.mod`
  - 删除 `replace github.com/yttydcs/myflowhub-core => ../MyFlowHub-Core`
  - 删除 `replace github.com/yttydcs/myflowhub-proto => ../MyFlowHub-Proto`
  - `require` 升级到 `core v0.2.0`、`proto v0.1.0`
- `go.sum`
  - 通过 `go mod tidy` 更新依赖校验和。

### 不变（保证）
- 不改 wire、不改业务逻辑；仅做依赖版本化与可拉取化。

## plan.md 任务映射
- SRVSEM1：调整 go.mod/go.sum（移除 replace + 固定版本）✅
- SRVSEM2：回归测试（`GOWORK=off go test ./...`）✅
- SRVSEM3：归档变更 ✅

## 测试与验证
- `GOWORK=off go test ./... -count=1 -p 1`

## 潜在影响
- 上游/本仓的本地联调方式从 `replace` 迁移为 `go.work`（不提交），发布验收必须以 `GOWORK=off` 为准，避免 go.work 掩盖问题。

## 回滚方案
- 通过 `git revert` 回退本次依赖变更提交即可恢复开发期 `replace`（不建议长期保留）。
