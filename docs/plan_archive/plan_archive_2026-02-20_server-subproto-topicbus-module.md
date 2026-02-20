# Plan - PR3：拆 `topicbus` 子协议到 `myflowhub-subproto/topicbus`（Server 侧收敛依赖）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/subproto-topicbus-module`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr3-topicbus-subproto\MyFlowHub-Server`
- Base：`refactor/subproto-management-module`（允许堆叠：需先合 PR2 再合本 PR3）
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）

## 约束（边界）
- wire 不改：SubProto 值 / Action 字符串 / JSON payload struct / HeaderTcp 语义均保持不变。
- 本 PR 只做“代码归属与依赖边界调整”：
  - Server 删除 `subproto/topicbus` 实现；
  - Server 改为依赖 `github.com/yttydcs/myflowhub-subproto/topicbus`。
- 验收测试必须使用 `GOWORK=off`（避免本地 `go.work` 干扰审计）。

## 当前状态（事实，可审计）
- `subproto/topicbus` 当前不依赖 Server 私有包（适合拆为独立 module）。
- 本 PR 将依赖 SubProto 仓发布的新版本：`myflowhub-subproto/topicbus@v0.1.0`（tag：`topicbus/v0.1.0`）。

---

## 目标
1) Server 侧将 topicbus 子协议改为“独立 Go module 依赖”（A2：单仓多 module）。
2) 延续 management 的拆库模式，逐步把子协议实现从 Server 装配层剥离出去。

## 非目标
- 不调整 topicbus 行为/语义；
- 不拆其它子协议；
- 不做 minimal/full 变体产品化（按当前决策 deferred）。

---

## 3.1) 计划拆分（Checklist）

### SRVTB0 - 归档旧 plan
- 目标：归档上一轮 PR2 的 `plan.md`，避免覆盖。
- 涉及文件：
  - `docs/plan_archive/plan_archive_2026-02-19_server-subproto-management-module.md`
- 验收条件：旧 plan 已归档且可阅读。
- 回滚点：撤销本次 `git mv`。

### SRVTB1 - 切换到 subproto module（import + go.mod）
> 依赖：先完成 `myflowhub-subproto/topicbus` 的发布与 tag（见 SubProto 仓 plan）。
- 目标：
  - 引用 `github.com/yttydcs/myflowhub-subproto/topicbus`；
  - Server 不再包含 `subproto/topicbus` 实现目录。
- 涉及模块/文件（预期）：
  - `modules/defaultset/topicbus_enabled.go`
  - `modules/defaultset/hub.go`（若间接引用）
  - `tests/topicbus_handler_test.go`（如引用）
  - `go.mod` / `go.sum`
  - `subproto/topicbus/*`（删除）
- 验收条件：
  - `rg \"myflowhub-server/subproto/topicbus\" -n --glob \"*.go\"` 无命中
  - `Test-Path subproto/topicbus` 为 `False`
- 回滚点：revert 提交。

### SRVTB2 - 回归验证（命令级）
- 命令：
```powershell
$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
GOWORK=off go test ./... -count=1 -p 1
```
- 验收条件：命令通过。
- 回滚点：revert 提交。

### SRVTB3 - Code Review（阶段 3.3）
- 按 3.3 清单输出结论（通过/不通过）；不通过则回到 SRVTB1 修正。

### SRVTB4 - 归档变更（阶段 4）
- 新增文档：
  - `docs/change/2026-02-19_server-use-subproto-topicbus.md`
- 需包含：
  - 背景/目标、变更范围（仅归属调整）、对外影响（新增依赖）、验证方式/结果、回滚方案。

### SRVTB5 - push 分支（便于 PR）
- `git push -u origin refactor/subproto-topicbus-module`

---

## 依赖关系 / 风险 / 注意事项
- 依赖：
  - `github.com/yttydcs/myflowhub-subproto/topicbus@v0.1.0` 必须已可拉取（tag 已 push），否则 `GOWORK=off` 无法通过验收。
- 风险：
  - 堆叠 PR：本 PR3 基于 PR2 分支，合并顺序必须是 PR2 → PR3。
- 注意：
  - commit 信息使用中文（允许 `refactor:`/`docs:` 等英文前缀）。

