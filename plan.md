# Plan - Server：改用 semver 依赖并移除 replace（PR19-SERVER-SemVer）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`chore/server-semver-deps`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr19-semver-deps\MyFlowHub-Server`
- Base：`main`
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）
- 目标：`go.mod` 移除本地 `replace`，依赖改为：
  - `github.com/yttydcs/myflowhub-core@v0.2.0`
  - `github.com/yttydcs/myflowhub-proto@v0.1.0`

## 约束（边界）
- 仅做依赖版本化（go.mod/go.sum）与归档文档；不改业务逻辑、不改 wire。
- 验收必须使用 `GOWORK=off`。

## 当前状态（事实，可审计）
- Server 当前使用 `replace ../MyFlowHub-Core` / `replace ../MyFlowHub-Proto` 本地联调。
- 本 workflow 要保证 Server 单仓 clone 时可直接 `go test ./...`（不依赖 go.work/replace）。

---

## 1) 需求分析

### 目标
1) Server 移除 `replace`，并锁定到 `core v0.2.0`、`proto v0.1.0`。
2) `GOWORK=off go test ./...` 通过（真实拉取依赖）。
3) 归档变更说明与回滚策略。

### 验收标准
- `go.mod` 无 `replace github.com/yttydcs/myflowhub-* => ../...`
- `GOWORK=off go test ./... -count=1 -p 1` 通过

---

## 2) 架构设计（分析）

### 总体方案（采用）
- 依赖通过 semver tag 引用，配合 go.sum 锁定：
  - core：`v0.2.0`
  - proto：`v0.1.0`
- 本地多仓联调通过 `d:\project\MyFlowHub3\go.work`（不提交）实现；发布验收用 `GOWORK=off`。

---

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 已确认依赖版本与验收方式：core=`v0.2.0`、proto=`v0.1.0`，验收使用 `GOWORK=off`。

### SRVSEM1 - 调整 go.mod/go.sum（移除 replace + 固定版本）
- 目标：Server 可在无 go.work/无 replace 下构建与测试。
- 涉及文件：
  - `go.mod`
  - `go.sum`
- 验收条件：
  - 移除 `replace`；
  - `require` 指向 `core v0.2.0`、`proto v0.1.0`；
  - `go mod tidy` 后工作区干净。
- 测试点：
  - `GOWORK=off go test ./...`
- 回滚点：
  - revert 该提交。

### SRVSEM2 - 回归测试（GOWORK=off）
- 目标：确保依赖真实拉取且 `go test` 通过。
- 命令：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'`
  - `New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null`
  - `$env:GOWORK='off'`
  - `go test ./... -count=1 -p 1`
- 验收条件：通过。

### SRVSEM3 - 归档变更
- 目标：记录依赖版本化变更、验证方式与回滚策略。
- 涉及文件：
  - `docs/change/2026-02-18_server-semver-deps.md`
- 验收条件：文档可独立复现。

### SRVSEM4 - Code Review（阶段 3.3）+ 归档（阶段 4）
- 验收条件：Review 结论为“通过”。

### SRVSEM5 - 合并（你确认结束 workflow 后执行）
- 目标：合并到 `main` 并 push。
- 步骤（在 `repo/MyFlowHub-Server` 执行）：
  1) `git merge --ff-only origin/chore/server-semver-deps`
  2) `git push origin main`
- 回滚点：
  - revert 合并提交（或 revert 分支提交）。
