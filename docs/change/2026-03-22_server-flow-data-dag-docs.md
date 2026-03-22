# 2026-03-22 Server Flow 数据流 DAG 文档

## 变更背景 / 目标

- Flow 运行时与 Win 编辑器都要升级到“祖先结果绑定 + `compose` 节点”模型，但长期 requirements/specs 里还没有这套正式描述。
- 本轮目标是在 `MyFlowHub-Server/docs` 中补齐长期真相，避免实现完成后继续依赖对话或源码猜测协议。

## Requirements / Specs Impact

- Requirements impact：`updated`
- Specs impact：`updated`
- Related requirements：
  - `docs/requirements/flow_data_dag.md`
- Related specs：
  - `docs/specs/flow.md`

## 具体变更内容（新增 / 修改 / 删除）

### 新增

- `docs/requirements/flow_data_dag.md`
  - 新增 Flow 数据流 DAG 的长期需求文档。

### 修改

- `docs/requirements/README.md`
  - 新增 requirement 索引入口。
- `docs/specs/flow.md`
  - 重写 Flow spec，补齐：
    - `call` / `compose`
    - 运行期 `RunContext`
    - trigger 上下文规范化
    - `InputBinding`
    - 祖先可见性与图校验要求
- `docs/change/README.md`
  - 更新 Server change 索引，登记本次归档。

### 删除

- 无。

## 对应 `plan.md` 任务映射

- `DAG-DOC-1` → `docs/requirements/flow_data_dag.md`, `docs/requirements/README.md`
- `DAG-DOC-2` → `docs/specs/flow.md`

## 关键设计决策与权衡

- 长期真相放在 Server docs，而不是分散回 SubProto / Win：
  - 好处：协议入口稳定，后续实现侧 change 文档只记录结果；
  - 代价：跨仓阅读时需要明确 canonical docs 位置。
- 首版 spec 明确 `call + compose + InputBinding`，不继续沿用自由字符串模板：
  - 好处：前后端实现都可基于同一结构化契约收敛；
  - 代价：历史实现需要一次适配。

## 测试与验证方式 / 结果

- 方式：
  - 对照 requirement / spec / SubProto 当前实现 / Win 当前设计进行人工复核
- 结果：
  - 通过。文档已能直接指导运行时与编辑器实现。

## 3.3 Code Review 结论

- 需求覆盖：通过。需求文档覆盖目标、范围、场景、验收标准。
- 架构合理性：通过。requirements 与 specs 的边界清晰，未把实现细节重复写入 requirement。
- 性能风险：通过。spec 明确保持 `status_resp` 轻量，不鼓励大结果回传。
- 可读性与一致性：通过。术语统一为 `call / compose / InputBinding / RunContext`。
- 可扩展性与配置化：通过。`source.kind` 与 trigger context 为后续扩展保留结构位。
- 稳定性与安全：通过。spec 明确非法 Pointer、非祖先引用、必填缺失的失败语义。
- 测试覆盖情况：通过。文档审阅与实现对照完成。
- 子Agent治理与审计：通过。本轮未使用子Agent。

## 潜在影响与回滚方案

- 潜在影响：
  - `docs/specs/flow.md` 已成为后续实现的正式技术契约，旧的口头约定不再应继续使用。
- 回滚方案：
  - 回退 `docs/requirements/flow_data_dag.md`
  - 回退 `docs/requirements/README.md`
  - 回退 `docs/specs/flow.md`

## 子Agent执行轨迹

- 本轮未使用子Agent。
- Task ID → Agent → Worktree → 文件 → 验收结果
  - `DAG-DOC-1` → 主Agent → `D:\project\MyFlowHub3\worktrees\server-data-dag-specs` → `docs/requirements/flow_data_dag.md`, `docs/requirements/README.md` → 通过
  - `DAG-DOC-2` → 主Agent → `D:\project\MyFlowHub3\worktrees\server-data-dag-specs` → `docs/specs/flow.md` → 通过
