# Plan - Server：升级 SubProto File 至 v0.1.1（修复 Hub File Console not found）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`fix/hub-file-console-base-dir`
- Worktree：`d:\project\MyFlowHub3\worktrees\fix-hub-file-console-base-dir\MyFlowHub-Server`
- Base：`main`
- 关联发布：
  - `myflowhub-subproto/file`：发布 tag `file/v0.1.1`
- 关联仓库（同一 workflow）：`MyFlowHub-SubProto`（同名分支/独占 worktree）
- 参考：
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）

## 背景 / 问题陈述（事实，可审计）
- Win 的 File Console 访问 Hub（node1）时提示：`not found`。
- 根因在 `myflowhub-subproto/file v0.1.0`：
  - root list 时 `BaseDir(默认 ./file)` 不存在会被映射为 `404 not found`；
  - 相对路径以 CWD 为基准导致目录落点不稳定。
- `myflowhub-subproto/file v0.1.1` 将修复上述问题（详见 SubProto 仓计划/变更文档）。

## 目标
1) 将 `MyFlowHub-Server` 依赖 `github.com/yttydcs/myflowhub-subproto/file` 从 `v0.1.0` 升级到 `v0.1.1`。
2) 更新 `go.sum` 并通过最小验证测试。
3) 归档变更，保证可审计、可回滚。

## 非目标
- 不修改 Server 侧 file handler 装配逻辑（仍由 `modules/defaultset/file_enabled.go` 装配）。
- 不发布 `myflowhub-server` 新版本 tag（如需对外发布另起 workflow）。

## 约束（边界）
- 变更最小化：只改 `go.mod/go.sum` 与文档。
- 验收测试必须使用 `GOWORK=off`（避免本地 `go.work` 掩盖依赖问题）。

## 验收标准
- `go.mod` 中 `github.com/yttydcs/myflowhub-subproto/file` 版本为 `v0.1.1`；
- `GOWORK=off go test ./... -count=1 -p 1` 通过；
- `go list -m github.com/yttydcs/myflowhub-subproto/file@v0.1.1` 可解析（证明 tag 可拉取）。

---

## 3.1) 计划拆分（Checklist）

### SRV0 - 归档旧 plan（已执行）
- 目标：避免历史 plan 覆盖本 workflow。
- 已执行：`git mv plan.md docs/plan_archive/plan_archive_2026-03-04_server-bump-management-v0.1.2.md`
- 验收条件：归档文件存在且可阅读。
- 回滚点：撤销该 `git mv`。

### SRV1 - 升级依赖：`myflowhub-subproto/file v0.1.1`
**目标**
- 获取 SubProto 修复（root list 自动创建 + BaseDir exeDir 解析）。

**涉及模块 / 文件**
- `go.mod`
- `go.sum`

**验收条件**
- `go.mod` 依赖版本更新为 `v0.1.1` 且 `go mod tidy`（如需要）后无额外无关 diff。

**测试点**
- `GOWORK=off go test ./... -count=1 -p 1`

**回滚点**
- revert 该提交（回到 `v0.1.0`）。

### SRV2 - 验收：依赖可解析与编译通过
**目标**
- 确保 tag 已发布且依赖可拉取（避免“本地 replace/缓存”假通过）。

**验收条件**
- `go list -m github.com/yttydcs/myflowhub-subproto/file@v0.1.1` 成功。

**回滚点**
- 若依赖无法解析：阻塞并回到 SubProto workflow 修复发布流程。

### SRV3 - Code Review（强制）+ 归档变更（强制）
**目标**
- 输出 `docs/change/YYYY-MM-DD_*.md`，记录依赖升级原因、验证与回滚方案。

