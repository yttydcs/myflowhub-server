# 2026-04-02_server-flow-permission-refinement

## 变更背景 / 目标

- `flow` 已补齐 `cancel_run` 与 `list_runs`，但运行控制面和只读观测面仍缺少稳定权限边界。
- 本轮目标是在 Server 稳定文档和 Hub 默认配置层明确：
  - `run` / `cancel_run` -> `flow.run`
  - `status` / `detail` / `list_runs` / `list` / `get` -> `flow.read`
  - 默认 `admin/node` 开箱配置继续具备上述能力

## 具体变更内容

### 修改

- `docs/requirements/flow_data_dag.md`
  - 补充 `flow.run` / `flow.read` 的需求、验收和权限隔离边界。
- `docs/specs/flow.md`
  - 明确稳定权限常量、动作映射、`flow::run` capability 权限要求和 `403` 响应语义。
- `docs/specs/auth.md`
  - 更新默认 `admin/node` 角色权限集合，纳入 `flow.run` / `flow.read`。
- `hubruntime/options.go`
  - `defaultAuthRolePerms` 改为直接引用 `coreconfig.DefaultAuthRolePerms`，减少与 Core 默认值漂移。
- `hubruntime/options_test.go`
  - 锁定 `flow.run` / `flow.read` 出现在默认角色权限中。

### 删除

- 无

## Requirements impact

- `updated`
  - `docs/requirements/flow_data_dag.md`

## Specs impact

- `updated`
  - `docs/specs/flow.md`
  - `docs/specs/auth.md`

## Lessons impact

- `none`

## Related requirements

- `D:\project\MyFlowHub3\worktrees\server-run-control-phase1\docs\requirements\flow_data_dag.md`

## Related specs

- `D:\project\MyFlowHub3\worktrees\server-run-control-phase1\docs\specs\flow.md`
- `D:\project\MyFlowHub3\worktrees\server-run-control-phase1\docs\specs\auth.md`
- `D:\project\MyFlowHub3\worktrees\proto-run-control-phase1\docs\protocol_map.md`

## Related lessons

- 无

## 对应 plan.md 任务映射

- `RC-P0-3`
  - `Server docs`: requirements/specs/auth 对齐
  - `Server runtime`: `hubruntime` 默认角色权限收口

## 经验 / 教训摘要

- `flow` 的动作面一旦拆出 `cancel_run` 与 `list_runs`，就不适合继续沿用“run/read 默认无权限”模型。
- 默认角色权限字符串不应在 `Core` 和 `Server` 维护两份独立拷贝。
- capability 描述、稳定 specs 和 runtime 默认值需要同时更新，才能避免权限行为和控制面展示漂移。

## 可复用排查线索

- 症状：
  - `run` 或 `status` 在默认部署下突然返回 `403 permission denied`
  - `exec.cap.query` 看不到 `flow::run` 的权限要求
  - `Server` 与 `Core` 的默认 `auth.role_perms` 不一致
- 触发条件：
  - 只更新了 SubProto 或 Proto，没有同步默认角色值
  - `hubruntime` 继续维护独立默认权限字符串
- 关键词 / 错误文本：
  - `flow.run`
  - `flow.read`
  - `permission denied`
  - `defaultAuthRolePerms`
  - `flow::run`
- 快速检查：
  1. 看 `docs/specs/flow.md` 中动作到权限的映射是否完整
  2. 看 `docs/specs/auth.md` 中 `admin/node` 是否包含 `flow.run` / `flow.read`
  3. 看 `hubruntime/options.go` 是否继续直接复用 `coreconfig.DefaultAuthRolePerms`

## 关键设计决策与权衡

- 保持现有逐级授权与父节点信任模型
  - 好处：不改变现有 LCA 裁决链路
  - 代价：新增权限时必须确保文档、默认值和运行时 helper 同步更新
- `Server` 默认值直接引用 `Core`
  - 好处：减少后续权限新增时的字符串漂移
  - 代价：`Server` 默认值现在显式依赖 `Core` 常量定义

## 测试与验证方式 / 结果

- `D:\project\MyFlowHub3\worktrees\proto-run-control-phase1`
  - `$env:GOWORK='off'; go test ./... -count=1 -p 1`
- `D:\project\MyFlowHub3`
  - `GOWORK=D:\project\MyFlowHub3\.tmp\verify-run-control-phase3\go.work go test github.com/yttydcs/myflowhub-subproto/flow/... -count=1 -p 1`
  - `GOWORK=D:\project\MyFlowHub3\.tmp\verify-run-control-phase3\go.work go test github.com/yttydcs/myflowhub-server/hubruntime/... -count=1 -p 1`
- 结果：通过

## 潜在影响

- 只依赖默认角色配置的部署现在会显式看到 `flow.run` / `flow.read`。
- 自定义 `auth.role_perms` 的部署若未补齐新权限，对应动作会按预期返回 `403`。

## 回滚方案

1. 回退 `docs/requirements/flow_data_dag.md`
2. 回退 `docs/specs/flow.md` 与 `docs/specs/auth.md`
3. 回退 `hubruntime/options.go` 与 `hubruntime/options_test.go`
4. 重新按旧权限模型验证默认部署行为

## 子Agent执行轨迹

- 本轮未使用子Agent
