# 2026-04-02_server-flow-active-run-limit-contract

## 变更背景 / 目标

- `flow` 之前已经补齐了显式取消、历史摘要查询、权限边界和 retry backoff，但“同一 flow 允许多少个活动 run”仍依赖入口差异形成隐式行为。
- 本轮目标是在稳定 requirements/specs 中补齐 `max_active_runs` 契约，把“手动 run 可并发、trigger 默认单飞”的现状收敛为可声明、可审计的兼容规则。

## 具体变更内容

### 修改

- `docs/requirements/flow_data_dag.md`
  - 补充 active-run limit 背景、兼容默认值、手动冲突与 trigger 跳过语义
  - 增加 `max_active_runs` 的功能需求、边界条件和验收标准
- `docs/specs/flow.md`
  - 在 `set/get` 契约中新增 `max_active_runs`
  - 明确 `nil` / `0` / `>0` 的语义
  - 明确手动 `run` 超限返回 `409`、trigger 超限跳过且不生成新 run

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

- `RC-P1-2`
  - `Server docs`: `max_active_runs` requirement/spec alignment

## 经验 / 教训摘要

- active-run limit 不能只写成“默认 1”，否则会破坏现有手动 `run` 的兼容行为。
- `nil` 与 `0` 必须明确区分：前者表示沿用 legacy 行为，后者表示显式无限制。
- “手动返回 `409`、trigger 跳过”必须进入稳定文档，否则后续实现者很容易把两条路径重新做成不一致。

## 可复用排查线索

- 症状：
  - 手动 `run` 被活动 run 挡住时没有明确冲突响应
  - trigger 超限时仍继续生成重叠 run
  - 读回 flow 定义时看不到 `max_active_runs`
- 触发条件：
  - requirements/specs 没写清 `nil` / `0` / `>0` 三种语义
  - `get` 契约遗漏了新字段
- 关键词 / 错误文本：
  - `max_active_runs`
  - `active run limit reached`
  - `409`
- 快速检查：
  1. 看 `docs/specs/flow.md` 的 `set/get` 是否都列出 `max_active_runs`
  2. 看 `docs/requirements/flow_data_dag.md` 是否明确 legacy 兼容行为
  3. 看文档是否区分“手动冲突返回”与“trigger 跳过”

## 关键设计决策与权衡

- 保留 legacy 默认值，而不是把未设置字段解释成 `1`
  - 好处：旧 flow 不会被静默收紧
  - 代价：文档和运行时都需要同时维护“未设置”和“显式 0”两套语义
- 当前只做 active-run cap，不引入 queue / cancel_previous
  - 好处：边界清晰、测试面小
  - 代价：更复杂的重入策略需要后续任务继续补

## 测试与验证方式 / 结果

- 文档联动验收依赖对应实现验证：
  - `D:\project\MyFlowHub3\worktrees\proto-run-control-phase1`
    - `$env:GOWORK='off'; go test ./... -count=1 -p 1`
  - `D:\project\MyFlowHub3`
    - `GOWORK=D:\project\MyFlowHub3\.tmp\verify-run-control-phase3\go.work go test github.com/yttydcs/myflowhub-subproto/flow/... -count=1 -p 1`
- 结果：通过

## 潜在影响

- 新 flow 可以显式声明活动 run 上限。
- 旧 flow 未设置该字段时继续保持原有手动/trigger 差异行为，不会被静默改变。

## 回滚方案

1. 回退 `docs/requirements/flow_data_dag.md`
2. 回退 `docs/specs/flow.md`
3. 恢复“active-run limit 仅为隐式行为”的稳定文档口径

## 子Agent执行轨迹

- 本轮未使用子Agent
