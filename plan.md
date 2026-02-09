# Plan - MyFlowHub-Server 公共协议包收敛（protocol/*）

## 项目目标
将当前开发中的 `protocol/*` 公共协议包（对外可复用的请求/响应模型、常量、必要校验）纳入版本控制，并将 `internal/handler/*` 的类型定义与引用收敛到 `protocol/*`（保持行为不变），形成可审计、可回放的提交链，供 `MyFlowHub-Win` 通过 `replace` 复用。

## 当前状态（事实）
- 本 worktree 分支：`feat/public-protocol`（从 `main` 创建）。
- 原主 worktree（`d:\project\MyFlowHub3\repo\MyFlowHub-Server`）存在：
  - 已跟踪文件的未提交修改（多处 `internal/handler/*/types.go`）。
  - 未跟踪新增目录 `protocol/`（多处 `protocol/*/types.go`）。
- 当前 `go test ./...` 在主 worktree 可通过（作为回归基线）。

## 非目标 / 约束
- 不新增或改变运行时行为（仅类型/常量/校验辅助导出 + handler 引用迁移）。
- 不在主 worktree（`repo/MyFlowHub-Server`）直接做实现性改动；所有改动在本 worktree 完成。

## 任务清单（Checklist）

### S1 - 将主 worktree 的 WIP 迁移到本 worktree
- 目标：把主 worktree 的未提交修改与未跟踪 `protocol/*` 文件完整迁入本 worktree，且不引入 `node_modules/`、构建产物等无关文件。
- 涉及模块/文件：
  - `internal/handler/**/types.go`（以及可能的关联引用文件）
  - `protocol/**`（新增）
- 验收条件：
  - 本 worktree 的 `git status` 能看到与主 worktree 同等的变更集合（内容一致）。
  - `go test ./...` 通过。
- 测试点：
  - `go test ./...`
  - `go test ./tests -run Test -count=1`（如存在用例）
- 回滚点：
  - 直接删除本 worktree 目录 + 删除分支 `feat/public-protocol`。

### S2 - 审核与最小化变更范围（保持行为不变）
- 目标：确认迁移仅为“类型定义迁移/导出/引用调整”，不改变 handler 逻辑、路由规则、权限校验等行为。
- 涉及模块/文件：
  - `internal/handler/**`（重点检查：除了 import 与类型名外是否有逻辑变化）
  - `protocol/**`
- 验收条件：
  - `git diff` 中除类型定义移动、包名引用、必要导出符号外，不出现行为相关修改（分支、条件、路由、权限）。
  - `go test ./...` 通过。
- 测试点：
  - `go test ./...`
- 回滚点：
  - 若发现行为修改，回退到 S1 并重新迁移/修剪。

### S3 - 提交与可审计化
- 目标：将变更拆成清晰提交（至少 1 个），便于 Win 端跟随。
- 验收条件：
  - `git status` 干净。
  - 提交信息能反映“新增 protocol 包 + handler 引用迁移（无行为变更）”。
- 测试点：
  - `go test ./...`（提交前后各跑一次，记录结果）
- 回滚点：
  - `git revert <commit>` 或直接删除分支。

### S4 - Code Review（阶段 3.3）与归档（阶段 4）
- 目标：按要求输出 Review 结论，并在本 worktree 根目录创建 `docs/change/YYYY-MM-DD_public-protocol.md` 归档。
- 验收条件：
  - Review 清单逐项“通过/不通过”结论明确；不通过则回到阶段 3.2 修正。
  - `docs/change` 文档包含：背景、具体变更、任务映射（S1-S3）、关键决策/权衡、测试结果、影响与回滚。
- 回滚点：
  - 删除归档文档并回退提交（仅当 workflow 未结束）。

## 依赖关系
- S1 完成后才能进行 S2/S3。
- S3 完成后才能进入阶段 3.3 与阶段 4。

## 风险与注意事项
- 行尾 LF/CRLF 可能导致噪音 diff；原则上本 workflow 不做全局格式化，仅在必要时做最小化处理。
- 若 `protocol/*` 与 Win 端存在生成/依赖关系（如代码生成产物），需在 S2 中明确是否应提交生成文件（默认：仅提交源码与稳定产物）。

