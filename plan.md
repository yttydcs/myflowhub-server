# Plan - Server：升级依赖到 Core v0.3.0（对齐 Pipe 抽象重大变更）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`chore/bump-core-v0.3.0`
- Worktree：`d:\project\MyFlowHub3\worktrees\chore-bump-core-v0.3.0\MyFlowHub-Server`
- Base：`main`
- 关联仓库：
  - `MyFlowHub-Core`：已发布 `v0.3.0`（重大变更：`IConnection.RawConn()` → `IConnection.Pipe()`，并新增 `listener/multi_listener`）

## 背景 / 问题陈述（事实，可审计）
- Server `main` 已合入 Pipe 抽象相关改动，并开始依赖 Core 新增包：`github.com/yttydcs/myflowhub-core/listener/multi_listener`。
- 但 `go.mod` 仍依赖 `github.com/yttydcs/myflowhub-core v0.2.1`，导致在 `GOWORK=off`（CI/用户默认）下无法编译与运行测试。

## 目标
1) 将 `github.com/yttydcs/myflowhub-core` 依赖升级到 `v0.3.0`。
2) 执行 `go mod tidy` 并确保 `GOWORK=off go test ./...` 通过。

## 非目标
- 不改任何业务逻辑/协议语义（仅做依赖升级与必要的 go.mod/go.sum 更新）。
- 不发布新 tag（如需发布由后续 workflow 决策）。

## 验收标准
- `cd d:\project\MyFlowHub3\worktrees\chore-bump-core-v0.3.0\MyFlowHub-Server`
  - `GOWORK=off go test ./... -count=1 -p 1` 通过。
- 合并到 `main` 并 push。

## 3.1) 计划拆分（Checklist）

### SRVDEP0 - 归档旧 plan（已执行）
- 已执行：`git mv plan.md docs/plan_archive/plan_archive_2026-03-12_bump-core-v0.3.0-prev.md`

### SRVDEP1 - 升级 Core 依赖到 v0.3.0
- 目标：`go.mod` 中 `github.com/yttydcs/myflowhub-core` 从 `v0.2.1` 升级到 `v0.3.0`。
- 涉及文件：`go.mod`、`go.sum`
- 验收条件：`GOWORK=off go test ./...` 编译通过。
- 回滚点：revert 本任务提交。

### SRVDEP2 - 回归测试（GOWORK=off）
- 目标：确保 CI/用户默认模式可运行。
- 测试点：
  - `GOWORK=off go test ./... -count=1 -p 1`

### SRVDEP3 - Code Review + 归档变更
- 输出：`docs/change/2026-03-12_bump-core-v0.3.0.md`

### SRVDEP4 - 合并 / push（需 workflow 结束后执行）
- 在 `repo/MyFlowHub-Server` 合并到 `main` 并 push。

