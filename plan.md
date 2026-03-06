# Plan - MyFlowHub-Server：VarStore 规范文档与依赖联动

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`chore/varstore-hop-align-docs`
- Worktree：`d:\project\MyFlowHub3\worktrees\varstore-hop-align\server`
- Base：`main`
- 关联仓库：`MyFlowHub-SubProto`、`MyFlowHub-SDK`、`MyFlowHub-Win`

## 项目目标与当前状态
- 目标：把 `docs/3-varstore.md` 与已确认语义对齐，并准备 Server 侧对新 VarStore 版本的联动说明。
- 当前状态：已完成 SRV-1：`docs/3-varstore.md` 已按最终决策更新；SRV-2（依赖版本联动）待 SubProto 发布可解析版本后再评估执行。

## 依赖关系
- 文档语义依赖已确认的阶段 1/2 决策。
- 如需要更新 `go.mod` 的 VarStore 版本，依赖 SubProto 发布可解析版本。

## 风险与注意事项
- 文档必须明确 `MajorCmd` 与 `TargetID=0` 的关系，避免误导实现。
- 若版本未发布，避免提交不可解析的依赖版本号。

## 可执行任务清单（Checklist）

### SRV-1 更新 VarStore 规范文档
- 目标：按确认结论修订 `docs/3-varstore.md`：
  - `*_resp/assist_*_resp` 逐跳可见；
  - requester/owner 回程语义；
  - notify 下行“转发+本地处理”；
  - SourceID 端到端保留；
  - set_resp value、list 空集合、set.value、private 例外、subscriber 规则。
- 涉及模块/文件：`docs/3-varstore.md`
- 验收条件：与 `varstore_requirements.md`/`varstore_architecture.md` 无冲突。
- 测试点：人工审阅 + 与实现交叉核对。
- 回滚点：回退文档提交。

### SRV-2 评估并执行依赖联动（条件任务）
- 目标：在 SubProto 发布新版本后，评估是否升级 Server 的 `myflowhub-subproto/varstore` 版本。
- 涉及模块/文件：`go.mod`、`go.sum`（如执行）
- 验收条件：仅在版本可解析时提交依赖变更。
- 测试点：`go list -m` + `go test ./... -count=1 -p 1`。
- 回滚点：回退依赖版本提交。

### SRV-3 归档变更
- 目标：记录文档与依赖联动（若有）的最终结果。
- 涉及模块/文件：`docs/change/2026-03-06_varstore-hop-align-server.md`
- 验收条件：文档映射 SRV-1~SRV-2，说明未执行条件任务的原因（如适用）。
- 测试点：归档内容可供他人复核。
- 回滚点：文档可独立回退。
