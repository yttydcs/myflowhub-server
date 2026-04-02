# 2026-04-02_server-flow-run-archive-contract

## 变更背景 / 目标

- `flow` 之前的 retained run 仅停留在内存窗口；执行器重启后，`status/detail/list_runs` 无法继续读取这些已结束 run。
- 本轮目标是在稳定 requirements/specs 中补齐可选 run archive 契约，让 retained window 内的终态 run 可以被持久化并在重启后继续查询。

## 具体变更内容

### 修改

- `docs/requirements/flow_data_dag.md`
  - 增加 retained run archive 背景、目标、功能需求和验收标准
  - 明确 `flow.run_archive_enabled` 与 `flow.max_retained_runs` 的关系
- `docs/specs/flow.md`
  - 明确 delete 后 retained archive 仍可查询
  - 在持久化与结果保留章节增加 run archive sidecar 语义
  - 补充建议配置项 `flow.run_archive_enabled`

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

- `RC-P1-4`
  - `Server docs`: retained run archive contract

## 经验 / 教训摘要

- retained run archive 不需要新增 wire；复用现有 `status/detail/list_runs` 即可承载最近窗口内的已归档 run。
- archive 默认关闭更安全，避免给未显式配置的环境引入额外本地 I/O。
- delete 语义要明确区分“删除定义”与“删除 retained archive”，否则历史查询行为会继续漂移。

## 可复用排查线索

- 症状：
  - 执行器重启后 recent run 丢失
  - delete 后 retained run 无法再查
  - archive 开启后仍只保留内存窗口
- 触发条件：
  - requirements/specs 没写清 archive 默认值或 retained window 语义
  - 只做了运行时实现，没有更新 delete / retention 契约
- 关键词 / 错误文本：
  - `flow.run_archive_enabled`
  - retained archive
  - `flow.max_retained_runs`
- 快速检查：
  1. 看 `docs/specs/flow.md` 是否说明 archive sidecar 和 delete 后查询语义
  2. 看 `docs/requirements/flow_data_dag.md` 是否把 retained archive 写成显式需求
  3. 看文档是否明确 archive 默认关闭

## 关键设计决策与权衡

- 复用 `flow.max_retained_runs` 作为 retained archive window，而不是新增新的 archive count
  - 好处：窗口语义一致，避免双上限配置漂移
  - 代价：若未来要把内存 cache 与 archive 历史分离，需要再引入独立配置
- archive 使用 local JSON sidecar
  - 好处：最小可执行，不改 wire，不碰 definition persistence 接口
  - 代价：未来若要接 PG，需要再单独抽象 archive backend

## 测试与验证方式 / 结果

- 文档联动验收依赖对应实现验证：
  - `D:\project\MyFlowHub3`
    - `GOWORK=D:\project\MyFlowHub3\.tmp\verify-run-control-phase3\go.work go test github.com/yttydcs/myflowhub-subproto/flow/... -count=1 -p 1`
- 结果：通过

## 潜在影响

- 启用 `flow.run_archive_enabled` 后，retained window 内的 recent run 在重启后仍可查询。
- 未启用时继续保持当前仅内存 retained window 语义。

## 回滚方案

1. 回退 `docs/requirements/flow_data_dag.md`
2. 回退 `docs/specs/flow.md`
3. 恢复“retained run 不承诺重启后仍可查询”的稳定文档口径

## 子Agent执行轨迹

- 本轮未使用子Agent
