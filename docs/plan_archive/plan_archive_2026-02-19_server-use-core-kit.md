# Plan - Server：切换到 MyFlowHub-Core/subproto/kit（依赖 core@v0.2.1）（PR1）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/subproto-kit-core`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr1-kit-core\MyFlowHub-Server`
- Base：`origin/main`
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）

## 约束（边界）
- wire 不改：SubProto 值 / Action 字符串 / JSON payload struct / HeaderTcp 语义均保持不变。
- 本 PR 仅做“工具包归属调整”：
  - `MyFlowHub-Server/subproto/kit` 删除；
  - Server 内所有引用改为 `github.com/yttydcs/myflowhub-core/subproto/kit`。
- 验收测试必须使用 `GOWORK=off`（避免本地 `go.work` 干扰审计）。

## 当前状态（事实，可审计）
- Server 当前存在 `subproto/kit` 包（action 模板 + response send helper）。
- 多个子协议（`subproto/auth/varstore/exec/flow/topicbus/management/forward`）已依赖 `subproto/kit`。
- 本次将依赖 Core 发布的新版本 `myflowhub-core@v0.2.1`（由 Core PR1 先完成并打 tag）。

---

## 目标
1) Server 不再承载 `subproto/kit` 实现，改为依赖 Core 的 `subproto/kit`（保持行为不变）。
2) 为后续“子协议实现拆成独立 Go module”（A2：单仓多 module）清理依赖边界：子协议实现不再被 Server 仓库绑定。

## 非目标
- 不做任何 handler 行为重构；
- 不做 broker/flow/exec 的进一步解耦（后续单独 workflow 处理）。

---

## 3.1) 计划拆分（Checklist）

### SRVKIT0 - 归档旧 plan
- 目标：避免覆盖上一轮 action-kit workflow 的 `plan.md`，保留可审计回放。
- 涉及文件：
  - `docs/plan_archive/plan_archive_2026-02-19_server-action-kit.md`
- 验收条件：旧 plan 已归档且可阅读。
- 回滚点：撤销本次 `git mv`。

### SRVKIT1 - 切换 import 到 Core kit
- 目标：全仓不再引用 `github.com/yttydcs/myflowhub-server/subproto/kit`。
- 涉及模块/文件（预期）：
  - `subproto/*`（所有 import `.../subproto/kit` 的文件）
  - `modules/defaultset/hub.go`、`tests/*`（如有间接引用）
- 验收条件：
  - `rg \"myflowhub-server/subproto/kit\" ./` 无命中（允许 `docs/plan_archive` 与 `docs/change` 历史文本不参与验收）。
  - `go test` 可通过（见 SRVKIT3）。
- 回滚点：revert 提交。

### SRVKIT2 - 删除 Server 内 `subproto/kit`
- 目标：避免双实现与未来漂移风险。
- 涉及文件：
  - `subproto/kit/*`（删除）
- 验收条件：
  - Server 编译/测试通过；`subproto/kit` 目录不存在。
- 回滚点：revert 提交。

### SRVKIT3 - 回归验证（命令级）
> 注意：本步骤依赖 Core 已发布 tag `v0.2.1`，否则 `GOWORK=off` 无法拉取新包路径。
- 命令：
```powershell
$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
GOWORK=off go test ./... -count=1 -p 1
```
- 验收条件：命令通过。
- 回滚点：revert 提交。

### SRVKIT4 - Code Review（阶段 3.3）
- 按 3.3 清单输出结论（通过/不通过）；不通过则回到 SRVKIT1/2 修正。

### SRVKIT5 - 归档变更（阶段 4）
- 新增文档：
  - `docs/change/2026-02-19_server-use-core-kit.md`
- 需包含：
  - 背景/目标、变更范围（仅归属调整）、对外影响（依赖 core@v0.2.1）、验证方式/结果、回滚方案。

### SRVKIT6 - 合并与 push（需你确认 workflow 结束后执行）
- 在 `repo/MyFlowHub-Server` 执行：
  1) `git merge --ff-only origin/refactor/subproto-kit-core`
  2) `git push origin main`

---

## 依赖关系 / 风险 / 注意事项
- 依赖：
  - 必须先完成 Core PR1（发布并 push tag `myflowhub-core@v0.2.1`），否则本仓库无法用 `GOWORK=off` 通过验收。
- 风险：
  - 若遗漏某处 import，将导致 “同时存在旧包路径/新包路径” 或编译失败；需用 `rg` 全量检查。
- 注意：
  - commit 信息使用中文（允许 `refactor:`/`docs:` 等英文前缀）。

