# Plan - Server：升级 SubProto Management 至 v0.1.2（children-only）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`chore/bump-management-v0.1.2`
- Worktree：`d:\project\MyFlowHub3\worktrees\chore-server-bump-management-v0.1.2`
- Base：`origin/main`
- 关联发布：
  - `myflowhub-subproto`：发布 tag `management/v0.1.2`
- 参考：`d:\project\MyFlowHub3\guide.md`（commit 信息中文）

## 背景 / 问题陈述（事实，可审计）
- `myflowhub-subproto/management v0.1.1` 的 `list_nodes` 在存在 parent link 的节点上会枚举到 upstream(parent) 连接，导致设备树出现回指（例如 `5 -> 1`）。
- `myflowhub-subproto` 已在 `main` 修复为 children-only（过滤 `role=parent`），但尚未发布新版本。
- `MyFlowHub-Server` 当前依赖 `myflowhub-subproto/management v0.1.1`，因此仍不会获得修复。

## 目标
1) 发布 `github.com/yttydcs/myflowhub-subproto/management v0.1.2`（tag：`management/v0.1.2`）。
2) 将 `MyFlowHub-Server` 的依赖从 `management v0.1.1` 升级到 `v0.1.2`，并更新 `go.sum`。
3) 归档变更，保证可审计、可回滚。

## 非目标
- 不修改 `management` 的 wire schema（已在 SubProto 仓库完成行为修复，本次仅做发布与依赖升级）。
- 不发布 `myflowhub-server` 新版本 tag（如需对外发布另起 workflow）。
- 不触发 Android CI 构建（若需要触发 `debug-latest`，应在 Android 仓库另起提交或另起 workflow）。

## 约束（边界）
- 必须以 `GOWORK=off` 方式验证（避免 go.work 掩盖依赖问题）。
- 变更最小化：只改 `go.mod/go.sum` 与文档。

## 验收标准
- `myflowhub-subproto`：
  - tag `management/v0.1.2` 存在且已 push 到 `origin`；
  - `go list -m github.com/yttydcs/myflowhub-subproto/management@v0.1.2` 可解析。
- `myflowhub-server`：
  - `go.mod` 中 `github.com/yttydcs/myflowhub-subproto/management` 版本为 `v0.1.2`；
  - `GOWORK=off go test ./... -count=1 -p 1` 通过。

---

## 3.1) 计划拆分（Checklist）

### SVRMG0 - 归档旧 plan.md
- 目标：避免历史 plan 覆盖本次任务。
- 已执行：`plan.md` → `docs/plan/plan_archive_2026-03-03_server-bump-management-v0.1.2-prev.md`
- 验收条件：归档文件存在且可阅读。
- 回滚点：撤销该移动提交。

### SVRMG1 - 发布 `management/v0.1.2` tag
- 目标：让上游可通过 semver 依赖拉取修复版本。
- 涉及仓库：`repo/MyFlowHub-SubProto`
- 操作：
  - `git fetch --tags`（确认 tag 未占用）
  - `git tag -a management/v0.1.2 <commit> -m ...`
  - `git push origin management/v0.1.2`
- 验收条件：远端可见 tag，且 `go list -m ...@v0.1.2` 成功。
- 回滚点：删除 tag（高风险，需谨慎；仅在确认未被消费时执行）。

### SVRMG2 - 升级 Server 依赖到 `management v0.1.2`
- 目标：Server 编译/运行使用新版本 management 行为（children-only）。
- 涉及文件：
  - `go.mod`
  - `go.sum`
- 操作：
  - `GOWORK=off go get github.com/yttydcs/myflowhub-subproto/management@v0.1.2`
  - `GOWORK=off go mod tidy`
- 验收条件：`go.mod/go.sum` 更新且最小化；`go test` 通过。
- 回滚点：revert 提交。

### SVRMG3 - Code Review（阶段 3.3）
- 目标：确认只发生依赖升级与可审计文档变更；风险可控。

### SVRMG4 - 归档变更（阶段 4）
- 目标：记录发布与依赖升级过程、测试与回滚方式。
- 涉及文件：
  - `docs/change/2026-03-03_server-bump-management-v0.1.2.md`


