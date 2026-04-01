# 2026-04-02 Server Flow Local Vars Docs

## 变更背景 / 目标

- `MyFlowHub-Server` 主线稳定文档仍停留在 `call/compose + status` 视角，未覆盖 `set_var`、`flow_var`、`detail`。
- 本轮目标是在 clean branch 上把 local vars 与 detail 的长期 requirements/specs 补回主线，作为 Proto/SubProto/Win 的稳定真相。

## 具体变更内容

### 修改

- `docs/requirements/flow_data_dag.md`
  - 增加 run-local vars 背景、目标、范围、场景、验收标准
  - 明确 `set_var` / `flow_var` / `detail` 与 `varstore` / `status` 的边界
- `docs/specs/flow.md`
  - 增加 `detail/detail_resp` action 契约
  - 增加 `set_var` 正式节点类型
  - 增加 `flow_var` binding source 与 `RunContext.vars`
  - 明确图校验、结果保留与 detail 读取边界

### 删除

- 无

## Requirements impact

- `updated`

## Specs impact

- `updated`

## Lessons impact

- `none`

## Related requirements

- `D:\project\MyFlowHub3\worktrees\server-local-vars-clean\docs\requirements\flow_data_dag.md`

## Related specs

- `D:\project\MyFlowHub3\worktrees\server-local-vars-clean\docs\specs\flow.md`

## Related lessons

- 无

## 对应 plan.md 任务映射

- `SERV-DOC-1`
  - `docs/requirements/flow_data_dag.md`
- `SERV-DOC-2`
  - `docs/specs/flow.md`

## 经验 / 教训摘要

- 对 `flow` 这类跨仓能力，长期真相必须先回到 `MyFlowHub-Server/docs`，不能把 dirty worktree 的 change 归档当真相来源。
- `flow` 局部变量和 `varstore` 必须从文档层面明确区分，否则运行时和编辑器都会出现语义漂移。

## 可复用排查线索

- 症状：
  - 主线文档查不到 `set_var`
  - 主线文档查不到 `flow_var`
  - 调用方误以为 `status` 会直接返回完整节点结果
- 触发条件：
  - 功能先在代码或 UI 落地，但稳定 requirements/specs 未同步
- 关键词 / 错误文本：
  - `set_var`
  - `flow_var`
  - `detail`
  - `status`
- 快速检查：
  1. 看 `docs/requirements/flow_data_dag.md` 是否包含 run-local vars 生命周期
  2. 看 `docs/specs/flow.md` 是否包含 `detail/detail_resp`

## 关键设计决策与权衡

- 把 `detail` 定义成与 `status` 分离的正式 action
  - 好处：保留轻量状态接口
  - 代价：调用方需要显式发起第二次查询
- 把局部变量定义成 run-local 而非 `varstore` 映射
  - 好处：边界清晰，不引入跨 run 持久化歧义
  - 代价：默认摘要接口不会直接看到 vars 内容

## 测试与验证方式 / 结果

- 文档自洽校验
  - 结果：通过
- 下游联调参考
  - `Proto clean worktree`：`go test ./...`
  - `SubProto clean worktree` 目标测试：通过
- `git diff --check`
  - 结果：通过

## 潜在影响

- 之后主线 `flow` 的稳定文档将正式包含 local vars 与 detail
- 下游消费方不再需要从旧 change 或 UI 代码反推协议边界

## 回滚方案

1. 回退 `docs/requirements/flow_data_dag.md`
2. 回退 `docs/specs/flow.md`

## 子Agent执行轨迹

- 本轮未使用子Agent
