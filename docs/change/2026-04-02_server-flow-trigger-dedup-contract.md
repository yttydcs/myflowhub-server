# 2026-04-02_server-flow-trigger-dedup-contract

## 变更背景 / 目标

- `flow` 已经补齐 active-run limit，但 `event/var_changed` trigger 在重复通知场景下仍会直接尝试再次启动。
- 本轮目标是在稳定 requirements/specs 中补齐 `trigger.dedup_window_ms` 契约，把“默认不去重、显式开启短窗口去重”的行为固定下来。

## 具体变更内容

### 修改

- `docs/requirements/flow_data_dag.md`
  - 增加 trigger dedup 背景、目标、功能需求、边界条件和验收标准
- `docs/specs/flow.md`
  - 在 trigger/set/get 契约中增加 `dedup_window_ms`
  - 明确其仅支持 `event/var_changed`，默认关闭，`interval` 不支持
  - 明确 dedup 为内存态、窗口内重复 trigger 不生成新 run

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

- `RC-P1-3`
  - `Server docs`: trigger dedup requirement/spec alignment

## 经验 / 教训摘要

- dedup 必须明确为显式 opt-in；如果文档只写“trigger 会自动去重”，会直接改变旧 flow 的预期。
- `interval` 需要明确排除，否则调用方会把调度频率控制和重复事件抑制混成一类语义。
- “窗口内重复 trigger 不生成新 run”必须写入稳定文档，否则实现侧很容易退化成生成失败 run 或静默覆盖。

## 可复用排查线索

- 症状：
  - 重复事件仍连续触发多个 run
  - 配置 `dedup_window_ms` 后不同 payload 也被误杀
  - `interval` trigger 写了 `dedup_window_ms` 却没有被拒绝
- 触发条件：
  - requirements/specs 没写清支持范围和规范化 trigger 语义
  - 只更新了实现，没有补稳定文档
- 关键词 / 错误文本：
  - `dedup_window_ms`
  - `event`
  - `var_changed`
  - `interval`
- 快速检查：
  1. 看 `docs/specs/flow.md` 是否只允许 `event/var_changed` 使用 dedup
  2. 看 `docs/requirements/flow_data_dag.md` 是否明确 dedup 默认关闭且仅内存态
  3. 看文档是否写明窗口内重复 trigger 不生成新 run

## 关键设计决策与权衡

- dedup 仅做内存态，不跨重启持久化
  - 好处：实现简单，语义聚焦在短窗口抑制
  - 代价：重启后无法延续 dedup 记忆
- 将 run archive 从本轮拆到 `RC-P1-4`
  - 好处：本轮只处理 trigger dedup，一次只收敛一个行为面
  - 代价：归档/历史观测能力要在下一轮继续补齐

## 测试与验证方式 / 结果

- 文档联动验收依赖对应实现验证：
  - `D:\project\MyFlowHub3\worktrees\proto-run-control-phase1`
    - `$env:GOWORK='off'; go test ./... -count=1 -p 1`
  - `D:\project\MyFlowHub3`
    - `GOWORK=D:\project\MyFlowHub3\.tmp\verify-run-control-phase3\go.work go test github.com/yttydcs/myflowhub-subproto/flow/... -count=1 -p 1`
- 结果：通过

## 潜在影响

- 新 flow 可以显式配置 trigger dedup 窗口，减少重复通知导致的重复 run。
- 旧 flow 不写该字段时继续保持当前行为，不会被静默改变。

## 回滚方案

1. 回退 `docs/requirements/flow_data_dag.md`
2. 回退 `docs/specs/flow.md`
3. 恢复“trigger 无显式 dedup window” 的稳定文档口径

## 子Agent执行轨迹

- 本轮未使用子Agent
