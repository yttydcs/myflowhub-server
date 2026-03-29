# Stream Control Plane Validation Lessons

## 背景

- `stream` 的 Server 发布链验证与后续 Win 联调暴露了三个非显而易见的前提：
  - `myflowhub-core v0.4.8` 的 `auth.role_perms` 默认值为空
  - `stream` coordinator 的私有控制往返在单 worker dispatcher 下会自阻塞
  - public `subscribe/connect` 不能因为 producer / consumer 落在同一个 downstream child，就盲转发给该 child；下游宿主可能只实现 catalog + private `delivery_*` contract

## 适用范围

- `MyFlowHub-Server`
- `stream` 多节点集成测试
- 任何直接使用 `config.NewMap(...)` 组装 dispatcher / handler 的测试或轻量运行时
- Win local-owner 联调

## 症状

- `undefined: coreconfig.DefaultAuthRolePerms`
- `permission denied`，即使 `auth.node_roles` 已给节点配置了 `superadmin`
- `connect_resp code=408 msg=context deadline exceeded`
- `drop frame: target not found`
- `stream subscribe: request timed out`
- `stream connect: request timed out`
- `announce` / `list_sources` / `announce_consumer` / `list_consumers` 成功，但建链动作单独超时

## 触发条件

- Server 在 `GOWORK=off` 下拉取 `myflowhub-core v0.4.8`
- 测试直接使用 `config.NewMap(...)`，但没有显式设置 `auth.role_perms`
- 协调节点配置 `process.workers_per_channel=1`
- 只给连接 `SetMeta("nodeID")`，没有更新 `ConnManager` 的 node index
- producer 与 consumer 位于同一个 downstream child，且该 child 宿主没有实现 public `subscribe/connect`
- `routeCoordinatorRequest(...)` 仍保留 same-child blind-forward 分支

## 关键词

- `DefaultAuthRolePerms`
- `auth.role_perms`
- `stream connect 408`
- `stream subscribe timeout`
- `stream connect timeout`
- `same child`
- `same downstream`
- `routeCoordinatorRequest`
- `delivery_prepare`
- `GetByNode`
- `process.workers_per_channel`

## 快速检查

1. 检查运行时或测试配置里是否显式提供 `auth.role_perms`
2. 检查 `hubruntime` 是否仍保留 Server 本地默认角色权限
3. 检查协调节点的 `process.workers_per_channel` 是否至少为 `2`
4. 检查所有测试 helper 在 `SetMeta("nodeID")` 后是否调用 `ConnManager().UpdateNodeIndex(...)`
5. 如果 catalog 请求成功但 `subscribe/connect` 超时，检查 `stream/handler.go` 的 `routeCoordinatorRequest(...)` 是否把 same-child public 请求直接下发给 child
6. 检查 downstream 宿主是否只实现 private `delivery_prepare/activate/abort/close`，而没有实现 public `subscribe/connect`

## 处理建议

- 对 Server 运行时：
  - 不要依赖 `myflowhub-core v0.4.8` 注入 `auth.role_perms` 默认值
  - 由 Server 侧保留本地默认角色权限常量，直到依赖链统一
- 对多节点 `stream` 测试：
  - 显式设置 `auth.role_perms=superadmin:*` 或等价权限映射
  - 对协调节点显式设置 `process.workers_per_channel >= 2`
- 对按 node 路由的测试：
  - 不要只改连接 metadata
  - 必须同步刷新 `ConnManager` 的 node index
- 对 `subscribe/connect` 联调：
  - 只要当前 coordinator 同时可达 producer 和 consumer，就应当本地协调
  - 下游 owner 若只实现 catalog + private `delivery_*`，不要向其盲转发 public `subscribe/connect`

## 回链

- [`docs/change/2026-03-28_stream-server-release.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/change/2026-03-28_stream-server-release.md)
- [`docs/change/2026-03-29_stream-coordinator-same-child-routing.md`](D:/project/MyFlowHub3/worktrees/fix-stream-coordinator-docs/docs/change/2026-03-29_stream-coordinator-same-child-routing.md)
