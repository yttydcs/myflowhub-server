# 2026-03-29 Server：stream 主线合并

## 变更背景 / 目标

- root `run-dev.ps1` 之前之所以会自动切到 `worktrees/server-stream-subproto-design`，根因不是脚本本身，而是主线 `repo/MyFlowHub-Server` 尚未接入 `stream`。
- 同时，根 workspace 的 `go.work` 固定命中 `repo/MyFlowHub-Proto/main`，因此只有把 `Proto` 的 `protocol/stream` 也合回主线，`Server` 主线路径才可能真正恢复可用。
- 本次目标是把 `stream` 的 Server 集成结果正式合回 `repo/MyFlowHub-Server/main`，并保持当前主线已有的 auth/authority 更新不回退。

## 具体变更内容

- 合并 `chore/stream-subproto-design` 到当前 Server 主线语境。
- 保留当前主线已有依赖版本，并在 [`go.mod`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/go.mod) / [`go.sum`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/go.sum) 中新增：
  - `github.com/yttydcs/myflowhub-subproto/stream v0.1.0`
- 更新默认集合接线：
  - [`modules/defaultset/hub.go`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/modules/defaultset/hub.go)
  - [`modules/defaultset/stream_enabled.go`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/modules/defaultset/stream_enabled.go)
  - [`modules/defaultset/stream_disabled.go`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/modules/defaultset/stream_disabled.go)
- 新增 Server compat 协议壳：
  - [`protocol/stream/types.go`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/protocol/stream/types.go)
- 合并 `stream` 文档与索引：
  - [`docs/requirements/stream.md`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/docs/requirements/stream.md)
  - [`docs/specs/stream.md`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/docs/specs/stream.md)
  - [`docs/specs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/docs/specs/protocol_map.md)
  - [`docs/lessons/stream-control-plane-validation.md`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/docs/lessons/stream-control-plane-validation.md)
- 合并最小控制面集成测试：
  - [`tests/integration_stream_root_hub_connect_test.go`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/tests/integration_stream_root_hub_connect_test.go)
- 删除来源分支带入的根 [`plan.md`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/plan.md)
  - 改由 `todo.md` 作为 worktree 控制文档，并将可保留信息归档到 `docs/plan/`

## Impact

- Requirements impact: `updated`
- Specs impact: `updated`
- Lessons impact: `updated`
- Related requirements:
  - [`docs/requirements/stream.md`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/docs/requirements/stream.md)
- Related specs:
  - [`docs/specs/stream.md`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/docs/specs/stream.md)
  - [`docs/specs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/docs/specs/protocol_map.md)
- Related lessons:
  - [`docs/lessons/stream-control-plane-validation.md`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/docs/lessons/stream-control-plane-validation.md)

## 对应 plan.md / todo.md 任务映射

- `SVR-STRM-P1`
- `SVR-STRM-M1`
- `SVR-STRM-M2`
- `SVR-STRM-M3`
- `SVR-STRM-M4`

## 经验 / 教训摘要

- 当根 `go.work` 固定引用某个上游仓库主线时，业务仓主线恢复前必须先确认上游协议仓主线也已经同步，否则“代码合回去了但主线路径仍不可用”会继续发生。
- 对不在根 `go.work` 列表中的 merge worktree，可以用临时未跟踪 `go.work` 做集成验证，避免为了预合并测试去污染正式依赖文件。

## 可复用排查线索

- 症状
  - Win `Stream` 页请求 `announce / list_sources / list_consumers` 统一超时
  - root `run-dev.ps1` 被迫长期切到 `server-stream-subproto-design`
- 触发条件
  - Server 主线缺少 `newStreamHandler(...)`
  - Proto 主线缺少 `protocol/stream`
- 关键词
  - `newStreamHandler`
  - `protocol/stream`
  - `stream announce: request timed out`
  - `go.work`
- 快速检查
  - 查看 [`modules/defaultset/hub.go`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/modules/defaultset/hub.go) 是否包含 `newStreamHandler`
  - 查看 [`protocol/stream/types.go`](D:/project/MyFlowHub3/worktrees/fix-server-stream-mainline-merge/protocol/stream/types.go) 是否存在
  - 执行 `go test ./tests -run TestStreamRootHubConnectDisconnect -count=1 -p 1`

## 关键设计决策与权衡

- 保留主线 `auth v0.1.5` / `proto v0.1.5` 版本组合，不回退当前已合入的 authority 相关能力。
- 采用“先把 Proto 主线补齐，再在本地 workspace 模块图下验证 Server merge”方案，而不是继续强降 `proto` 版本或扩大到新的发布链调整。
- 本轮不承担新的 `myflowhub-proto` 远端 semver tag 发布；如需纯 `GOWORK=off` 的 semver 消费链，需要后续单独补发布。

## 测试与验证方式 / 结果

- `Proto merge worktree`
  - 执行：`$env:GOWORK='off'; go test ./... -count=1 -p 1`
  - 结果：通过
- `Server merge worktree`
  - 执行：临时创建未跟踪 `go.work` 后，`go test ./tests -run TestStreamRootHubConnectDisconnect -count=1 -p 1`
  - 结果：通过
- `Server merge worktree`
  - 执行：临时创建未跟踪 `go.work` 后，`go test ./... -count=1 -p 1`
  - 结果：通过

## 潜在影响与回滚方案

- 潜在影响
  - 本地主线路径现在会直接暴露 `stream` 能力；若依赖方默认假设主线不含 `stream`，其测试基线需要同步更新
  - 纯 semver 的 `GOWORK=off` Server 构建仍依赖后续 Proto 发布链补齐
- 回滚方案
  - 在主线对本次 merge commit 执行 `git revert`

## 子Agent执行轨迹

- 本轮未使用子Agent
