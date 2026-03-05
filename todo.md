# TODO - Server：升级 subproto/varstore 到 v0.1.1（跨层转发修复）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`chore/server-bump-varstore-v0.1.1`
- Worktree：`d:\project\MyFlowHub3\worktrees\chore-server-bump-varstore-v0.1.1`
- 上游版本：`github.com/yttydcs/myflowhub-subproto/varstore v0.1.1`

## 项目目标与当前状态
- 目标：
  - 将 Server 的 `myflowhub-subproto/varstore` 依赖从 `v0.1.0` 升级到 `v0.1.1`，纳入跨层 target 转发与 owner 路由自愈修复。
- 当前状态：
  - `go.mod` 仍为 `varstore v0.1.0`。

## 可执行任务清单（Checklist）

- [x] SRVVAR-1：升级依赖版本
  - 目标：`go.mod/go.sum` 对齐 `varstore v0.1.1`。
  - 涉及文件：
    - `go.mod`
    - `go.sum`
  - 验收条件：
    - `go list -m github.com/yttydcs/myflowhub-subproto/varstore` 输出 `v0.1.1`。
  - 回滚点：
    - 回退 `go.mod/go.sum` 到 `v0.1.0`。

- [x] SRVVAR-2：最小回归验证
  - 目标：保证升级后 Server 模块可构建/测试。
  - 验收条件：
    - `GOWORK=off go test ./... -count=1 -p 1` 通过。
  - 回滚点：
    - 回退依赖升级提交。

- [x] SRVVAR-3：Code Review + 归档
  - 目标：完成审查闭环与变更归档。
  - 涉及文件：
    - `docs/change/2026-03-05_server-bump-subproto-varstore-v0.1.1.md`
  - 验收条件：
    - 文档包含任务映射、验证结果、影响与回滚。

## 依赖关系
- `SRVVAR-1 -> SRVVAR-2 -> SRVVAR-3`

## 风险与注意事项
- 仅升级依赖，不混入功能改动。
