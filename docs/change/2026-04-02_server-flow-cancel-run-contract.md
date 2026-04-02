# 2026-04-02_server-flow-cancel-run-contract

## 变更背景 / 目标

- `flow` 已有 `run/status/detail`，但缺少显式的单 run 取消控制面。
- 本轮目标是在稳定 requirements/specs 中补齐 `cancel_run` 的行为边界，避免实现继续依赖“delete 才能中断 run”的隐式语义。

## 具体变更内容

### 修改

- `docs/requirements/flow_data_dag.md`
  - 补充 `cancel_run` 的目标、范围、功能要求与验收标准。
  - 明确 `cancel_run` 只取消指定 run，不删除 flow 定义。
- `docs/specs/flow.md`
  - 在动作总览中加入 `cancel_run/cancel_run_resp`。
  - 新增 `cancel_run` 请求/响应契约、`404/409` 语义和 `status` 回显要求。
  - 明确 `detail` 在 run 已取消时可通过 `msg` 与节点状态体现取消结果。

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

- `RC-P0-1`
  - Server requirements/specs 对齐

## 关键设计决策与权衡

- 保持 `cancel_run` 为最小增量动作，不在本轮同步引入 history 或新的权限常量。
- `detail` 不新增整体 run 状态字段，而是沿用现有响应结构，通过 `msg` 和节点摘要反映取消结果，避免首轮控制面扩张过大。

## 测试与验证方式 / 结果

- 文档联动验收依赖对应实现验证：
  - `D:\project\MyFlowHub3\worktrees\proto-run-control-phase1`
    - `$env:GOWORK='off'; go test ./... -count=1 -p 1`
  - `D:\project\MyFlowHub3`
    - `GOWORK=D:\project\MyFlowHub3\.tmp\verify-run-control-phase1\go.work go test github.com/yttydcs/myflowhub-subproto/flow/... -count=1 -p 1`
- 结果：通过

## 潜在影响

- 调用方现在可以对活动 run 发起显式取消，前端和上层编排需要区分 `cancel_run` 与 `delete` 的不同语义。

## 回滚方案

1. 回退 `docs/requirements/flow_data_dag.md`
2. 回退 `docs/specs/flow.md`
3. 重新按旧动作集合评估实现与调用方是否仍使用 `delete` 作为唯一取消入口
