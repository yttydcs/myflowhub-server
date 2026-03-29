# 2026-03-29_stream-coordinator-same-child-routing

## 变更背景 / 目标

- Win 端当前已经实现本地 `stream` owner 的 catalog 动作和私有 `delivery_*` 生命周期动作，但没有实现公开 `subscribe/connect` 协调动作。
- 现网问题表现为：
  - 本地 Source / Consumer 创建成功
  - `list_sources` / `list_consumers` 成功
  - 右侧详情正常展示
  - 但点击 `Subscribe` / `Connect` 时前端报：
    - `stream subscribe: request timed out`
    - `stream connect: request timed out`
- 本次目标是在不改 wire / requirements / specs 的前提下，修复 `MyFlowHub-SubProto/stream` 中 coordinator 的同子树路由决策，让当前 coordinator 在“两端都可达”时直接本地建链，而不是把原始 public 请求盲转发给同一个下游 child。

## 具体变更内容

- 修改 `D:\project\MyFlowHub3\worktrees\fix-stream-coordinator-routing\stream\handler.go`
  - 删除 `routeCoordinatorRequest(...)` 中“producer / consumer 都在同一 downstream child 时，把原始 public `connect/subscribe` 请求直接下发给该 child”的特殊分支。
  - 保留原有语义：
    - 两端任一不可达：继续向 parent 上送
    - 两端都可达：当前节点执行权限检查并本地协调
- 修改 `D:\project\MyFlowHub3\worktrees\fix-stream-coordinator-routing\stream\handler_test.go`
  - 新增 same-child 路由回归测试，证明当前 coordinator 不再盲转发 public 请求。
  - 新增 same-child 权限测试，证明修复后仍先做本地权限检查。
  - 新增不可达上送测试，证明 parent-forward 语义未回退。

## Requirements impact

- `none`

## Specs impact

- `none`

## Lessons impact

- `updated`

## Related requirements

- `D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\requirements\stream.md`

## Related specs

- `D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\specs\stream.md`

## Related lessons

- `D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\lessons\stream-control-plane-validation.md`

## 对应 `plan.md` 任务映射

- `STRCOORD-1`：完成
- `STRCOORD-2`：完成
- `STRCOORD-3`：完成

## 经验 / 教训摘要

- `stream` 的 coordinator 判断“谁来建链”时，不能把“同一 downstream child”直接等价成“该 child 必定支持 public coordinator 动作”。
- 当前 Win local-owner 的最小 contract 是：
  - 支持 public catalog / owner 动作
  - 支持 private `delivery_prepare/activate/abort/close`
  - 不支持 public `subscribe/connect`
- 因此对 public `subscribe/connect` 来说，只要当前节点同时可达 producer 和 consumer，就应当本地协调，再用 private helper 建两端 delivery。

## 可复用排查线索

- 症状
  - `stream subscribe: request timed out`
  - `stream connect: request timed out`
  - `announce` / `list_sources` / `announce_consumer` / `list_consumers` 都成功，只有建链超时
- 触发条件
  - producer 和 consumer 位于同一个 Win 节点，或至少位于同一个 downstream child
  - downstream owner 只实现 catalog + private `delivery_*` contract
  - coordinator 存在“same child -> blind forward public request”特殊分支
- 关键词
  - `stream subscribe timeout`
  - `stream connect timeout`
  - `same child`
  - `same downstream`
  - `routeCoordinatorRequest`
  - `delivery_prepare`
  - `Win local owner`
- 快速检查
  - 检查 `stream/handler.go` 的 `routeCoordinatorRequest(...)` 是否仍存在 same-child blind-forward 分支
  - 检查 downstream host 是否真的实现 public `subscribe/connect`
  - 若 catalog 请求成功但建链超时，优先怀疑 coordinator/public-vs-private action contract 不一致

## 关键设计决策与权衡

- 采用“当前 coordinator 只要能同时路由到两端，就本地协调”的最小修复，而不是扩展 Win local-owner 支持 public `subscribe/connect`。
- 这样做的好处：
  - 更符合 `stream` spec 的 coordinator 语义
  - 不扩大 Win host contract
  - 修复面只在 SubProto 一处路由判断
- 代价：
  - 失去把 public 协调动作进一步下沉到同 child 的潜在优化
  - 如果未来真要做 lower coordinator 优化，必须先定义清晰的 private coordination contract，而不是复用 public action 盲转发

## 测试与验证方式 / 结果

- `D:\project\MyFlowHub3\worktrees\fix-stream-coordinator-routing\stream`
  - `$env:GOWORK='off'; go test ./... -count=1 -p 1`
  - 结果：通过
- `D:\project\MyFlowHub3\repo\MyFlowHub-Server`
  - 使用临时 `go.work` 把 Server 指向修复后的本地 `stream` module
  - `go test ./tests -run TestStreamRootHubConnectDisconnect -count=1 -p 1`
  - 结果：通过
- `D:\project\MyFlowHub3\repo\MyFlowHub-Win`
  - `$env:GOWORK='D:\project\MyFlowHub3\go.work'; go test ./internal/services/stream -count=1 -p 1`
  - 结果：通过

## 潜在影响

- 当前修复会让更高一层 coordinator 在“同 child 可达”时保留本地建链职责。
- 这可能减少未来某些 lower coordinator 下沉机会，但不会改变现有 wire 和权限合同。

## 回滚方案

- 回退：
  - `D:\project\MyFlowHub3\worktrees\fix-stream-coordinator-routing\stream\handler.go`
  - `D:\project\MyFlowHub3\worktrees\fix-stream-coordinator-routing\stream\handler_test.go`
- 若后续发现更适合由下游节点协调，应先设计新的私有协调 contract，再做新的 workflow，而不是恢复 blind-forward。

## 子Agent执行轨迹

- 本轮未使用子Agent。
