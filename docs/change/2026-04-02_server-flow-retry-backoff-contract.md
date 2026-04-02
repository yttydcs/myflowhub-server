# 2026-04-02_server-flow-retry-backoff-contract

## 变更背景 / 目标

- `flow` 节点目前虽然支持 `retry`，但失败后仍会立即重试，缺少最小可控的 backoff 语义。
- 本轮目标是在稳定 requirements/specs 中补齐 `retry_backoff_ms`，把 `RC-P1-1` 收敛为“固定间隔重试”。

## 具体变更内容

### 修改

- `docs/requirements/flow_data_dag.md`
  - 补充 retry backoff 背景、固定间隔语义、取消边界和验收条件。
- `docs/specs/flow.md`
  - 在 `graph.nodes[]` 中新增 `retry_backoff_ms`
  - 明确固定间隔等待、默认 `0`、负值非法、等待期间响应取消的规则

### 删除

- 无

## Requirements impact

- `updated`
  - `docs/requirements/flow_data_dag.md`

## Specs impact

- `updated`
  - `docs/specs/flow.md`

## Lessons impact

- `none`

## Related requirements

- `D:\project\MyFlowHub3\worktrees\server-run-control-phase1\docs\requirements\flow_data_dag.md`

## Related specs

- `D:\project\MyFlowHub3\worktrees\server-run-control-phase1\docs\specs\flow.md`

## Related lessons

- 无

## 对应 plan.md 任务映射

- `RC-P1-1`
  - `Server docs`: retry backoff requirement/spec alignment

## 经验 / 教训摘要

- `retry` 次数和 `retry_backoff_ms` 间隔必须分开表达，否则“等待策略”会继续隐藏在实现里而不是进入稳定契约。
- 第一版先收敛为固定间隔，比一次性上指数退避更容易维持协议和测试边界。

## 可复用排查线索

- 症状：
  - 节点失败后立刻再次打远程能力
  - 接手者只看文档无法判断 retry 之间是否有等待
- 触发条件：
  - graph 里只写了 `retry`，没有显式 backoff 字段
  - requirements/specs 没把等待语义写出来
- 关键词 / 错误文本：
  - `retry_backoff_ms`
  - `retry`
  - `timeout_ms`
- 快速检查：
  1. 看 `docs/specs/flow.md` 的 `graph.nodes[]` 是否列出 `retry_backoff_ms`
  2. 看 `docs/requirements/flow_data_dag.md` 是否要求等待期间响应取消

## 关键设计决策与权衡

- 采用固定间隔字段 `retry_backoff_ms`
  - 好处：协议扩展最小，兼容旧 graph
  - 代价：暂不支持指数退避和 jitter

## 测试与验证方式 / 结果

- 文档联动验收依赖对应实现验证：
  - `D:\project\MyFlowHub3\worktrees\proto-run-control-phase1`
    - `$env:GOWORK='off'; go test ./... -count=1 -p 1`
  - `D:\project\MyFlowHub3`
    - `GOWORK=D:\project\MyFlowHub3\.tmp\verify-run-control-phase3\go.work go test github.com/yttydcs/myflowhub-subproto/flow/... -count=1 -p 1`
- 结果：通过

## 潜在影响

- 新 graph 可显式声明固定重试间隔；未声明时保持旧行为不变。

## 回滚方案

1. 回退 `docs/requirements/flow_data_dag.md`
2. 回退 `docs/specs/flow.md`
3. 恢复“retry 仅有次数语义、无显式间隔策略”的稳定文档口径

## 子Agent执行轨迹

- 本轮未使用子Agent
