# 2026-03-21 Flow 协议契约补齐（Server 文档）

## 变更背景 / 目标
- 背景：
  - `flow` 的 `run/status/list/get` 已有稳定 proto/handler，但 `docs/specs/flow.md` 仍未完整描述。
  - DAG 节点的正式写入契约已收敛到 `kind=call`，文档仍把 `local/exec` 当作正式格式。
- 目标：
  - 补齐 `run/status/list/get` 契约。
  - 将 DAG 节点写入模型与实现对齐，并明确历史兼容说明。

## 具体变更内容（新增 / 修改 / 删除）
### 修改
- `docs/specs/flow.md`
  - 补充 `run/status/list/get` 的请求字段、响应字段、执行语义与错误说明。
  - 明确 `flow_id` 为 UUID，并要求通过格式校验。
  - 将 DAG 节点的正式写入契约改为 `kind=call`。
  - 补充 `local/exec` 仅为历史兼容运行格式的说明。
  - 补充 `call` 节点在本地 / 远程场景下的调用分发语义。

### 新增
- 无。

### 删除
- 无。

## 对应 plan.md 任务映射
- `DOC-1`：完成（`run/status/list/get` 契约已补齐）。
- `DOC-2`：完成（正式写入契约与历史兼容边界已明确）。
- `DOC-3`：完成（已核对 `docs/specs/protocol_map.md`，本次无需改动）。

## 关键设计决策与权衡（尤其性能 / 扩展性）
1. 文档以当前 proto/handler 的稳定行为为准，不额外发明新语义。
   - 这样可以直接指导客户端和接手者，不再需要再去反推代码。
2. 正式写入契约与历史兼容语义分开写。
   - 优点：既不误导新客户端继续写旧格式，也不会让接手者误判历史数据不可运行。
3. `protocol_map.md` 只做核对，不做无必要改动。
   - 这样保持本次文档变更最小化，避免把纯契约补丁扩大成文档重排。

## 测试与验证方式 / 结果
- 人工一致性审阅：
  - 对照 `MyFlowHub-Proto/protocol/flow/types.go`
  - 对照 `MyFlowHub-SubProto/flow/handler.go`
- 结果：
  - `run/status/list/get` action、payload 与文档已对齐。
  - `kind=call` 正式写入契约与运行期历史兼容说明已对齐。
  - `protocol_map.md` 当前条目已覆盖本次范围，无需改动。

## 潜在影响与回滚方案
### 潜在影响
- 新客户端会按文档收敛到 UUID `flow_id` 和 `kind=call`；如果仍沿用旧格式，请求会被实现侧拒绝。
- 文档现在显式写出 `run/status/list/get` 无额外独立权限，这与当前实现一致；若未来权限模型收紧，需要再同步文档。

### 回滚方案
- 直接回滚 `docs/specs/flow.md` 本次改动即可。

## 子Agent执行轨迹
- 无子Agent。
- Task ID → Agent → Worktree → 文件 → 验收结果：
  - `DOC-1` → 主Agent → `D:\project\MyFlowHub3\worktrees\server-subproto-contract-spec` → `docs/specs/flow.md` → 通过
  - `DOC-2` → 主Agent → `D:\project\MyFlowHub3\worktrees\server-subproto-contract-spec` → `docs/specs/flow.md` → 通过
  - `DOC-3` → 主Agent → `D:\project\MyFlowHub3\worktrees\server-subproto-contract-spec` → `docs/specs/protocol_map.md`（核对，无变更） → 通过

