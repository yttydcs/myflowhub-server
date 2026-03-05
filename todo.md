# TODO - Server：升级 subproto/auth 到 v0.1.1（修复跨级路由传播）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`chore/server-bump-auth-v0.1.1`
- Worktree：`d:\project\MyFlowHub3\repo\MyFlowHub-Server\worktrees\chore-server-bump-auth-v0.1.1\MyFlowHub-Server`
- 上游版本：`github.com/yttydcs/myflowhub-subproto/auth v0.1.1`

## 项目目标与当前状态
- 目标：
  - 将 Server 的 `myflowhub-subproto/auth` 依赖从 `v0.1.0` 升级到 `v0.1.1`，纳入 `up_login SenderPub` 修复。
- 当前状态：
  - `go.mod` 仍为 `auth v0.1.0`。

## 可执行任务清单（Checklist）

- [x] SRVAUTH-1：升级依赖版本
  - 目标：`go.mod/go.sum` 对齐 `auth v0.1.1`。
  - 涉及文件：
    - `go.mod`
    - `go.sum`
  - 验收条件：
    - `go list -m github.com/yttydcs/myflowhub-subproto/auth` 输出 `v0.1.1`。
  - 测试点：
    - `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/auth`
  - 回滚点：
    - 回退 `go.mod/go.sum` 到 `v0.1.0`。

- [x] SRVAUTH-2：最小回归验证
  - 目标：保证升级后 Server 模块可构建/测试。
  - 涉及文件：
    - 无新增功能文件。
  - 验收条件：
    - `GOWORK=off go test ./... -count=1 -p 1` 通过。
  - 回滚点：
    - 回退依赖升级提交。

- [x] SRVAUTH-3：Code Review + 归档
  - 目标：完成审查闭环与变更归档。
  - 涉及文件：
    - `docs/change/2026-03-05_server-bump-subproto-auth-v0.1.1.md`
  - 验收条件：
    - 文档包含任务映射、验证结果、影响与回滚。
  - 回滚点：
    - 回滚文档提交。

## 依赖关系
- `SRVAUTH-1 -> SRVAUTH-2 -> SRVAUTH-3`

## 风险与注意事项
- 若远端 tag 解析异常，会阻塞升级并需回到上游发布流程核对。
- 仅升级依赖，不混入功能改动。

## 当前执行状态
- 已完成：SRVAUTH-1、SRVAUTH-2、SRVAUTH-3
- 进行中：无
- 待完成：无
