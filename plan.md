# Plan - MyFlowHub-Server：升级 Auth 以修复多 hop 路由索引缺失

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`fix/auth-route-index-heal`
- Worktree：`d:\project\MyFlowHub3\worktrees\fix-auth-route-index-heal\MyFlowHub-Server`
- Base：`main`
- 关联仓库：`MyFlowHub-SubProto`（发布 `auth/v0.1.2`）

## 项目目标与当前状态
- 目标：解决 Root Hub 在多 hop 场景中因 `sourceMismatch` 丢弃后代节点帧，导致 VarStore `list/get` 返回 `not found (code=4)` 的回归问题。
- 当前状态：已定位根因为 Auth 模块的 trusted/binding 公钥毒化 + `up_login` 不具备自愈能力；本仓需升级依赖并修订文档以匹配新行为。

## 依赖关系
- 依赖 `MyFlowHub-SubProto` 发布 `auth/v0.1.2`。
- Android 构建流（CI/Release）会在构建时 checkout `myflowhub-server` main 作为 hubmobile replace 源，因此本仓 main 合入后即可被 Android 构建消费（无需 Android 仓改动）。

## 风险与注意事项
- 依赖升级必须指向可解析版本（tag 已发布）；否则 `go mod tidy`/CI 会失败。
- 文档需与实现一致，避免继续误导“register 缺省 pubkey 会填本机公钥”的旧语义。

## 可执行任务清单（Checklist）

### SRV-AUTH-1 更新 auth 文档（与实现一致）
- 目标：更新 `docs/2-auth.md`：
  - 修正 `register`：`pubkey` 缺失不再自动填本节点公钥；
  - 说明 `up_login` sender 公钥自愈策略（触发条件/约束/审计日志）；
  - 明确 `auth.disable_persist=true` 的读写语义（如 SubProto 已对齐）。
- 涉及模块/文件：`docs/2-auth.md`
- 验收条件：文档描述与 `myflowhub-subproto/auth` 行为一致。
- 测试点：人工审阅 + 结合关键代码路径交叉检查。
- 回滚点：回退文档提交。

### SRV-AUTH-2 升级依赖：auth v0.1.2
- 目标：将 `go.mod` 中 `github.com/yttydcs/myflowhub-subproto/auth` 从 `v0.1.1` 升级到 `v0.1.2`，并更新 `go.sum`。
- 涉及模块/文件：`go.mod`、`go.sum`
- 验收条件：`GOWORK=off go test ./... -count=1 -p 1` 通过。
- 测试点：`go list -m github.com/yttydcs/myflowhub-subproto/auth` 输出版本为 `v0.1.2`。
- 回滚点：回退依赖升级提交。

### SRV-AUTH-3（可选）发布 Server tag
- 目标：视需要为本仓打新 tag（建议 `v0.0.6`），便于下游固定版本。
- 涉及模块/文件：无（git tag）
- 验收条件：tag 存在且 CI 可通过。
- 测试点：按 SRV-AUTH-2 测试项。
- 回滚点：未推送前删除 tag；已推送则改用新 patch tag。

### SRV-AUTH-4 归档变更
- 目标：按要求在本 worktree 根下创建 `docs/change/YYYY-MM-DD_auth-route-index-heal.md`，记录背景、变更、决策、测试与回滚。
- 涉及模块/文件：`docs/change/2026-03-08_auth-route-index-heal-server.md`
- 验收条件：文档映射 SRV-AUTH-1~SRV-AUTH-3，且包含验证步骤。
- 测试点：文档中的验证命令可执行。
- 回滚点：文档可独立回退，不影响功能代码。
