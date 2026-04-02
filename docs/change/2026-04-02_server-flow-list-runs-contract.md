# 2026-04-02_server-flow-list-runs-contract

## 变更背景 / 目标

- `flow` 已有 `status/detail`，但它们分别只解决“最新摘要”和“单节点重查询”，缺少按 `flow_id` 查看保留窗口内历史 run 的入口。
- 本轮目标是在稳定 requirements/specs 中补齐 `list_runs` 契约，明确它和 `status/detail` 的边界。

## 具体变更内容

### 修改

- `docs/requirements/flow_data_dag.md`
  - 补充 `list_runs` 的背景、目标、功能要求和验收条件。
- `docs/specs/flow.md`
  - 新增 `list_runs/list_runs_resp` 动作说明。
  - 明确 `flow_id + limit` 请求格式、最新到最旧的排序规则、retained run 语义和摘要字段。

## Requirements impact

- update
  - `docs/requirements/flow_data_dag.md`

## Specs impact

- update
  - `docs/specs/flow.md`

## Lessons impact

- none

## Related requirements

- `D:\project\MyFlowHub3\worktrees\server-run-control-phase1\docs\requirements\flow_data_dag.md`

## Related specs

- `D:\project\MyFlowHub3\worktrees\server-run-control-phase1\docs\specs\flow.md`

## 对应 plan / todo 任务映射

- `RC-P0-2`
  - Server requirements/specs 对齐

## 关键设计决策与权衡

- `list_runs` 只返回 retained run 摘要，不扩成 archive 接口
  - 好处：能直接复用当前 retention 模型和运行时索引
  - 代价：窗口外历史仍不可查询
- `list_runs` 与 `list` 分离
  - 好处：避免污染现有 flow 列表语义
  - 代价：调用方需要显式增加一次 history 查询

## 测试与验证方式 / 结果

- 文档联动验收依赖对应实现验证：
  - `D:\project\MyFlowHub3\worktrees\proto-run-control-phase1`
    - `$env:GOWORK='off'; go test ./... -count=1 -p 1`
  - `D:\project\MyFlowHub3`
    - `GOWORK=D:\project\MyFlowHub3\.tmp\verify-run-control-phase2\go.work go test github.com/yttydcs/myflowhub-subproto/flow/... -count=1 -p 1`
- 结果：通过

## 潜在影响

- 上层若需要“永久历史”或“跨 flow 聚合”，仍需后续单独引入 archive 能力；`list_runs` 不承担该职责。

## 回滚方案

1. 回退 `docs/requirements/flow_data_dag.md`
2. 回退 `docs/specs/flow.md`
3. 重新以 `status/detail/list` 为现有唯一观测面
