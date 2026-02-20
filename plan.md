# Plan - PR4：Server 依赖拆分后的子协议 modules（broker/auth/varstore/file/forward/exec/flow）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/subproto-modules-all`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr4-subproto-modules\MyFlowHub-Server`
- Base：`origin/main`
- 关联仓库（同一 workflow）：`MyFlowHub-SubProto`（同名分支/独占 worktree）
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）

## 约束（边界）
- wire 不改：SubProto 值 / Action 字符串 / JSON payload struct / HeaderTcp 语义均保持不变。
- 本 PR 只做“归属与依赖边界调整”：
  - Server 删除 `subproto/*`（剩余的 auth/varstore/file/forward/exec/flow）实现目录；
  - Server 改为依赖 `github.com/yttydcs/myflowhub-subproto/<name>` 的 semver 版本；
  - `internal/broker` 删除（改由 `myflowhub-subproto/broker` 承载）。
- 验收测试必须使用 `GOWORK=off`（避免本地 `go.work` 干扰审计）。

## 当前状态（事实，可审计）
- Server 当前已依赖 `myflowhub-subproto/{management,topicbus}`；
- 仍保留实现目录：
  - `subproto/auth`、`subproto/varstore`、`subproto/file`、`subproto/forward`、`subproto/exec`、`subproto/flow`
- `exec/flow` 通过 `internal/broker` 做同进程投递（需随本轮拆分一起迁移到 SubProto 的 `broker` module）。

## 目标
1) Server 侧完成依赖切换：
   - `myflowhub-subproto/broker v0.1.0`
   - `myflowhub-subproto/auth v0.1.0`
   - `myflowhub-subproto/varstore v0.1.0`
   - `myflowhub-subproto/file v0.1.0`
   - `myflowhub-subproto/forward v0.1.0`
   - `myflowhub-subproto/exec v0.1.0`
   - `myflowhub-subproto/flow v0.1.0`
2) 删除 Server 内对应实现目录与 `internal/broker`，保持行为不变。
3) 文档策略（已确认）：
   - SubProto 仓输出完整 docs/change；
   - Server 仓输出短引用 docs/change（指向 SubProto 文档）。

## 非目标
- 不调整任何子协议行为/语义；
- 不推进 “minimal/full 变体产品化”（按当前决策 deferred）。

---

## 3.1) 计划拆分（Checklist）

### SRVALL0 - 归档旧 plan
- 目标：保留上一轮 topicbus 拆分 plan，避免覆盖。
- 已执行（可审计）：`git mv plan.md docs/plan_archive/plan_archive_2026-02-20_server-subproto-topicbus-module.md`
- 验收条件：归档文件存在且可阅读。
- 回滚点：撤销该 `git mv`。

### SRVALL1 - 更新 go.mod/go.sum（新增 module 依赖）
> 依赖：必须先完成 SubProto 仓的 tags 发布（见其 plan：SUBALL9）。
- 目标：
  - `go.mod` 增加对 `myflowhub-subproto/{broker,auth,varstore,file,forward,exec,flow}` 的 `v0.1.0` 依赖。
  - `go mod tidy` 后 `GOWORK=off` 可拉取对应版本。
- 涉及文件（预期）：
  - `go.mod`
  - `go.sum`
- 验收条件：`GOWORK=off go mod tidy` 无报错。
- 回滚点：revert 提交。

### SRVALL2 - 更新 import 与装配点（modules/defaultset + tests）
- 目标：将对 `myflowhub-server/subproto/*` 的引用替换为 `myflowhub-subproto/*`。
- 涉及模块/文件（预期）：
  - `modules/defaultset/*.go`
  - `tests/*.go`
- 验收条件：
  - `rg \"myflowhub-server/subproto/(auth|varstore|file|forward|exec|flow)\" -n --glob \"*.go\"` 无命中
- 回滚点：revert 提交。

### SRVALL3 - 删除 Server 内实现目录与 internal/broker
- 目标：
  - 删除：`subproto/auth`、`subproto/varstore`、`subproto/file`、`subproto/forward`、`subproto/exec`、`subproto/flow`
  - 删除：`internal/broker`
- 验收条件：
  - 目录不存在（`Test-Path` 为 False）
  - `go test` 仍通过（见 SRVALL4）
- 回滚点：revert 提交。

### SRVALL4 - 回归验证（命令级）
- 验收命令：
```powershell
$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
$env:GOWORK='off'
go test ./... -count=1 -p 1
go test ./tests -run TestRootHubPing -count=1
```
- 验收条件：命令通过。
- 回滚点：revert 提交。

### SRVALL5 - Code Review（阶段 3.3）
- 按 3.3 清单逐项审查并输出结论（通过/不通过）；不通过则回到对应任务修正。

### SRVALL6 - 归档变更（阶段 4：Server 短引用文档）
- 新增文档（短引用）：`docs/change/2026-02-20_server-use-subproto-remaining-modules.md`
- 必须包含：
  - 变更范围（仅依赖切换 + 删除旧目录）
  - go.mod 依赖列表（module + version）
  - 验收命令与结果
  - 指向 SubProto 完整文档：
    - `MyFlowHub-SubProto/docs/change/2026-02-20_subproto-split-remaining-modules.md`
  - 回滚方案

### SRVALL7 - push 分支（便于合并）
- `git push -u origin refactor/subproto-modules-all`

