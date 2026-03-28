# 2026-03-28 Server：stream 真实依赖链收口与最小多节点验证

## 变更背景 / 目标

- 把 `stream` 从“本地 `go.work` 联调可用”推进到“`GOWORK=off` 可拉取、可编译、可执行最小 root / hub 集成验证”。
- 收口 Server 对 `Proto v0.1.4` 与 `SubProto stream v0.1.0` 的真实依赖链。
- 解决 `myflowhub-core v0.4.8` 下 `hubruntime` 对 `DefaultAuthRolePerms` 的编译兼容问题，并保持 Server 既有 auth 默认角色层级不回退。

## 具体变更内容

### 1) 依赖版本收口

- [`go.mod`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/go.mod)
  - `github.com/yttydcs/myflowhub-proto` 升级到 `v0.1.4`
  - 新增 `github.com/yttydcs/myflowhub-subproto/stream v0.1.0`
- [`go.sum`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/go.sum)
  - 同步写入 `Proto v0.1.4` 与 `stream v0.1.0` 的依赖校验

### 2) hubruntime 兼容修复

- [`hubruntime/options.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/hubruntime/options.go)
  - 去掉对 `coreconfig.DefaultAuthRolePerms` 的编译期直接依赖
  - 在 Server 侧保留本地 `defaultAuthRolePerms`，避免 `myflowhub-core v0.4.8` 的空默认值把 auth 默认层级退化掉
- [`hubruntime/layered_config.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/hubruntime/layered_config.go)
  - 继续把 `AuthRolePerms` 写入 effective config，保证运行时与测试配置口径一致
- [`hubruntime/options_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/hubruntime/options_test.go)
- [`hubruntime/layered_config_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/hubruntime/layered_config_test.go)
  - 覆盖默认角色层级与显式清空 override 语义

### 3) 最小多节点 stream 集成测试

- [`tests/integration_stream_root_hub_connect_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/tests/integration_stream_root_hub_connect_test.go)
  - 新增 root / hub 拓扑下的最小控制面验证：
    - `announce`
    - `announce_consumer`
    - `list_sources`
    - `connect`
    - `disconnect`
    - 二次 `disconnect -> 404`
  - 测试中显式补齐：
    - `auth.role_perms=superadmin:*`
    - `process.workers_per_channel=2`
- [`tests/integration_varstore_end_to_end_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/tests/integration_varstore_end_to_end_test.go)
  - 修复测试 helper：`bindConnNodeID` 与 `bindParentChildNodeIDs` 在 `SetMeta("nodeID")` 后同步更新 `ConnManager` 的 `nodeIndex`
  - 避免 `GetByNode()` 命中不到刚绑定的测试连接，导致按 node 回包时被误丢

## Requirements impact

- `none`

## Specs impact

- `none`

## Lessons impact

- `updated`

## Related requirements

- [`docs/requirements/stream.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/requirements/stream.md)

## Related specs

- [`docs/specs/stream.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/specs/stream.md)
- [`docs/specs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/specs/protocol_map.md)

## Related lessons

- [`docs/lessons/stream-control-plane-validation.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/lessons/stream-control-plane-validation.md)
- `D:\project\MyFlowHub3\docs\lessons\cross-repo-semver-release.md`

## 对应 `plan.md` 任务映射

- `STRM-REL-1`：完成
- `STRM-COMP-1`：完成
- `STRM-IT-1`：完成
- `STRM-VAL-1`：完成
- `STRM-REL-2`：本次归档对应 tag `v0.0.12`
- `STRM-DOC-1`：完成

## 经验 / 教训摘要

- 不能把当前本地仓库里的 Core 常量当成“已发布依赖一定存在”的事实；`myflowhub-core v0.4.8` 既不导出 `DefaultAuthRolePerms`，也不在 `config.NewMap()` 里补默认角色层级。
- `stream` 的 coordinator 私有控制往返是同步等待模型；若协调节点 dispatcher 只有 1 个 worker，私有响应会因为同一 worker 被阻塞而超时。
- 测试里单纯 `SetMeta("nodeID")` 不足以支撑按 node 路由；凡是依赖 `GetByNode()` 的协议，都要同步更新 `ConnManager` 的 node index。

## 可复用排查线索

- 症状
  - `hubruntime/options.go: undefined: coreconfig.DefaultAuthRolePerms`
  - `connect_resp code=408 msg=context deadline exceeded`
  - `drop frame: target not found`
  - `permission denied`，即使测试里已声明 `superadmin`
- 触发条件
  - Server 在 `GOWORK=off` 下消费 `myflowhub-core v0.4.8`
  - `stream` 集成测试或部署配置使用原始 `config.NewMap(...)` 默认值
  - 协调节点 `process.workers_per_channel=1`
- 关键词
  - `DefaultAuthRolePerms`
  - `auth.role_perms`
  - `stream connect 408`
  - `delivery_prepare`
  - `GetByNode`
  - `process.workers_per_channel`
- 快速检查
  - 检查 [`hubruntime/options.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/hubruntime/options.go) 是否使用 Server 本地 `defaultAuthRolePerms`
  - 检查测试或部署配置是否显式提供 `auth.role_perms`
  - 检查协调节点的 `process.workers_per_channel` 是否至少为 `2`
  - 检查测试 helper 是否调用 `ConnManager().UpdateNodeIndex(...)`

## 关键设计决策与权衡

- 不新增 Core patch release：
  - 好处：避免再拉起一条跨仓 patch 发布链，最小化本轮改动面
  - 代价：Server 需要临时保留一份本地默认角色权限常量
- 保持多节点测试锁定在控制面最小可观察链路：
  - 好处：能证明 root / hub 路由、目录查询和 connect / disconnect 收敛
  - 代价：DATA / ACK 端到端观测仍留待后续 workflow
- 测试中显式设置 `process.workers_per_channel=2`：
  - 好处：与 Server 默认运行时口径一致，能稳定验证当前同步私有控制模型
  - 代价：暴露了 `stream` 在单 worker dispatcher 下的实现约束，需要单独沉淀 lesson

## 测试与验证方式 / 结果

- 依赖核对
  - `$env:GOWORK='off'; go list -m github.com/yttydcs/myflowhub-proto github.com/yttydcs/myflowhub-subproto/stream github.com/yttydcs/myflowhub-core`
  - 结果：
    - `github.com/yttydcs/myflowhub-proto v0.1.4`
    - `github.com/yttydcs/myflowhub-subproto/stream v0.1.0`
    - `github.com/yttydcs/myflowhub-core v0.4.8`
- hubruntime 定向验证
  - `$env:GOWORK='off'; go test ./hubruntime -count=1 -p 1`
  - 结果：通过
- stream 定向集成
  - `$env:GOWORK='off'; go test ./tests -run TestStreamRootHubConnectDisconnect -count=1`
  - 结果：通过
- Server 模块与测试回归
  - `$env:GOWORK='off'; go test ./modules/... ./tests/... -count=1 -p 1`
  - 结果：通过
- 发布动作
  - 本次 worktree 提交对应 Server patch tag：`v0.0.12`
  - 推送策略：不使用主仓 `main`，直接从本 worktree 分支提交推送 branch + tag

## 3.3 Code Review 结论

- 需求覆盖：通过。依赖对齐、编译兼容、最小多节点测试和发布准备都已覆盖。
- 架构合理性：通过。兼容修复收敛在 `hubruntime`，不改 `stream` 业务合同。
- 性能风险：通过。运行时仅保留本地默认常量；控制面同步私有往返的单 worker 风险已记录为 lesson。
- 可读性与一致性：通过。命名与既有 `hubruntime` / `tests` 风格一致。
- 可扩展性与配置化：通过。显式 override 仍可覆盖或清空 `auth.role_perms`；测试 helper 修复也复用到其它多节点用例。
- 稳定性与安全：通过。`GOWORK=off` 下不再依赖未发布 Core 符号；按 node 路由测试不再依赖隐式索引。
- 测试覆盖情况：通过。完成依赖核对、`hubruntime` 定向、`stream` 定向和模块回归。
- 子Agent治理与审计：通过。本轮未使用子Agent。

## 潜在影响与回滚方案

- 潜在影响
  - `hubruntime` 现在在 Server 侧自带默认角色权限字符串；若未来 Core 默认层级继续演进，需要同步评估是否还要保留本地常量
  - `stream` 当前实现仍依赖协调节点至少有两个 dispatcher workers 才能安全完成同步私有控制往返
- 回滚方案
  - 回退：
    - [`go.mod`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/go.mod)
    - [`go.sum`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/go.sum)
    - [`hubruntime/options.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/hubruntime/options.go)
    - [`hubruntime/layered_config.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/hubruntime/layered_config.go)
    - [`hubruntime/options_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/hubruntime/options_test.go)
    - [`hubruntime/layered_config_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/hubruntime/layered_config_test.go)
    - [`tests/integration_varstore_end_to_end_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/tests/integration_varstore_end_to_end_test.go)
    - [`tests/integration_stream_root_hub_connect_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/tests/integration_stream_root_hub_connect_test.go)
  - 若 tag `v0.0.12` 发布后发现问题，不删除 tag，直接追加更高 patch 修复

## 子Agent执行轨迹

- 本轮未使用子Agent
