# Plan Archive - 2026-03-29 stream 主线合并

## Workflow 信息
- Repo：`MyFlowHub-Server`
- Branch：`fix/server-stream-mainline-merge`
- Base：`main @ bbdb4fc`
- Worktree：`D:\project\MyFlowHub3\worktrees\fix-server-stream-mainline-merge`
- 关联 Repo：
  - `MyFlowHub-Proto`
    - merge branch：`fix/proto-stream-mainline-merge`
    - merge worktree：`D:\project\MyFlowHub3\worktrees\fix-proto-stream-mainline-merge`

## 目标
- 将 `stream` 的 Proto + Server 集成结果正式合回各自主线。
- 恢复根 workspace 默认主线路径对 `stream` 的原生支持。

## 已确认约束
- 必须保留当前主线已有的 `auth v0.1.5` / remote authority 相关更新。
- `Server` merge worktree 不在根 `go.work` 列表中，预合并验证需要临时未跟踪 `go.work`。
- 来源 worktree 的 `.gitignore` 脏改动与根 `plan.md` 不能带入主线。
- 本轮不承担新的远端 semver tag 发布。

## 任务映射
- `SVR-STRM-P1`
  - 先把 `Proto` 的 `protocol/stream` 合回主线，恢复 workspace 上游协议路径。
- `SVR-STRM-M1`
  - 合并 `chore/stream-subproto-design` 并解决 docs / deps 冲突。
- `SVR-STRM-M2`
  - 用临时 `go.work` 验证 `go test ./tests -run TestStreamRootHubConnectDisconnect -count=1 -p 1`
  - 用临时 `go.work` 验证 `go test ./... -count=1 -p 1`
- `SVR-STRM-M3`
  - 归档 `docs/change/2026-03-29_stream-mainline-merge.md`
- `SVR-STRM-M4`
  - 将验证通过的分支合回 `repo/MyFlowHub-Server/main`

## 验收标准
- 主线包含 `newStreamHandler(...)` 与 `protocol/stream/types.go`。
- 根 workspace 主线路径不再依赖 `server-stream-subproto-design` 才能获得 `stream`。
- `Server` merge 结果在本地 workspace 模块图下 `go test ./... -count=1 -p 1` 通过。
