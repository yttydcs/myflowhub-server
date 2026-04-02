# 2026-04-02_server-flow-run-archive-backend-contract

## 变更背景 / 目标

- retained run archive 一期只定义了 local JSON sidecar 语义，尚未明确“PG 是可选增强、无 PG 默认继续工作”的稳定口径。
- 本轮目标是在 `Server` 装配层新增可选 PG archive backend，并把 requirements/specs 收敛到 `off/file/pg + legacy bool` 的稳定契约。

## 具体变更内容

### 修改

- `modules/defaultset/state_backends.go`
  - 新增 `flow.run_archive.backend`
  - 新增 `state.pg.flow_run_archive_table`
  - 新增 `newFlowRunArchiveStore(...)`
  - 新增 `pgFlowRunArchiveStore`
- `modules/defaultset/flow_enabled.go`
  - 将 archive store 注入 `flow.HandlerOptions`
- `modules/defaultset/state_backends_test.go`
  - 覆盖 unsupported backend、缺 DSN、PG connection error
- `docs/requirements/flow_data_dag.md`
  - 明确 archive backend 三态与“PG 可选、无 PG 默认可运行”的需求口径
- `docs/specs/flow.md`
  - 明确 `off/file/pg`
  - 明确 `flow.run_archive_enabled=true -> file`
  - 明确 `state.pg.flow_run_archive_table`
  - 明确查询面语义不变、backend 切换不自动迁移

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
  - 本轮没有新增需要长期查询的高成本排障路径

## Related requirements

- `docs/requirements/flow_data_dag.md`

## Related specs

- `docs/specs/flow.md`

## Related lessons

- `none`

## 对应 plan.md 任务映射

- `RA-SRV-1`
  - `modules/defaultset/state_backends.go`
  - `modules/defaultset/flow_enabled.go`
  - `modules/defaultset/state_backends_test.go`
- `RA-SRV-2`
  - `docs/requirements/flow_data_dag.md`
  - `docs/specs/flow.md`
- `RA-VER-1`
  - `modules/defaultset` / `hubruntime` 联调验证

## 经验 / 教训摘要

- 对 `Server` 来说，最小风险路径不是把 file backend 也搬进 `defaultset`，而是只在 `pg` 显式配置时注入外部 store，其余继续让 `SubProto` 走默认 `off/file`。
- 只要配置 key 足够明确，`PG 可选增强` 与 `无 PG 默认工作` 可以共存，不需要引入隐式降级。
- archive backend 也应沿用现有 `state.pg.*` 命名风格，避免把 definition persistence 和 archive persistence 的配置心智混乱。

## 可复用排查线索

- 症状：
  - `flow.run_archive.backend=pg` 启动时报缺 DSN
  - 配了 `backend=file` 却仍试图连 PG
  - `hubruntime` 能启动，但 retained run 仍只保留内存
- 触发条件：
  - `flow.run_archive.backend` 配置非法
  - `backend=pg` 但 `state.pg.dsn` 缺失或不可达
  - 文档仍把 archive 写死为 `flow.run_archive_enabled + local sidecar`
- 关键词 / 错误文本：
  - `flow.run_archive.backend`
  - `state.pg.flow_run_archive_table`
  - `unsupported flow.run_archive.backend`
  - `state.pg.dsn required`
- 快速检查：
  1. 看 `modules/defaultset/state_backends.go` 是否实现 `newFlowRunArchiveStore(...)`
  2. 看 `modules/defaultset/flow_enabled.go` 是否把 archive store 注入 `flow.HandlerOptions`
  3. 看 `docs/specs/flow.md` 是否明确 `off/file/pg` 与 legacy bool 映射
  4. 看 `modules/defaultset/state_backends_test.go` 是否覆盖 PG 错误路径

## 关键设计决策与权衡

- `Server` 只在 `backend=pg` 时注入 archive store
  - 好处：默认行为完全不变，部署侧不用为了 archive backend 重配无 PG 环境
  - 代价：`file` backend 逻辑仍保留在 `SubProto`
- 新增独立 `state.pg.flow_run_archive_table`
  - 好处：definition 与 archive 可独立管理
  - 代价：PG 配置项多一个，但边界更清晰
- `backend=pg` 配置错误时显式失败
  - 好处：避免无声降级造成审计与持久化预期偏差
  - 代价：错误配置会更早暴露

## 测试与验证方式 / 结果

- `D:\project\MyFlowHub3`
  - `GOWORK=D:\project\MyFlowHub3\.tmp\verify-run-archive-backend\go.work go test github.com/yttydcs/myflowhub-server/modules/defaultset/... github.com/yttydcs/myflowhub-server/hubruntime/... -count=1 -p 1`
- 结果：通过

## 潜在影响

- `PG` 现在是 archive 的显式可选增强，而不是默认前置条件。
- 配置 `backend=pg` 但 PG 不可用时，`Server` 初始化会直接失败。
- `backend=file/off` 场景不依赖 PG，仍按既有路径工作。

## 回滚方案

1. 回退 `modules/defaultset/state_backends.go`
2. 回退 `modules/defaultset/flow_enabled.go`
3. 回退 `modules/defaultset/state_backends_test.go`
4. 回退 `docs/requirements/flow_data_dag.md`
5. 回退 `docs/specs/flow.md`
6. 恢复“archive 仅 local sidecar”与旧文档口径

## 子Agent执行轨迹

- 本轮未使用子Agent
