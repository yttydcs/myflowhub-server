# 2026-03-21_server-docs-flow-delete

## 变更背景 / 目标
根据协议变更要求，同步 Server 文档中 Flow 协议的删除部署能力说明，确保文档与 Proto/SubProto 实现一致。

## 具体变更内容（新增 / 修改 / 删除）
- 修改 `docs/specs/flow.md`：
  - 新增 `action=delete` 请求/响应字段说明。
  - 新增权限 `flow.delete`。
  - 明确删除语义：删除后立即中断该 flow 的运行中 run。
  - 更新 Major 请求动作集合（含 `delete`）。
  - 补充 delete 相关错误语义（403/404 等）。
- 新增 `docs/specs/protocol_map.md`：
  - 新增 Flow action 映射：`delete`、`delete_resp`。
  - 新增 payload 类型映射：`DeleteReq`、`DeleteResp`。
  - 新增常量映射：`PermFlowDelete`。

## 对应 plan.md 任务映射
- `SRV-DOC-1`：完成（6-flow.md 更新）。
- `SRV-DOC-2`：完成（protocol_map 同步）。
- `SRV-DOC-3`：完成（一致性自检）。

## 关键设计决策与权衡（性能 / 扩展性）
- 保持与现有文档结构一致，避免新增并行文档体系。
- 明确“delete 会中断 run”语义，减少前后端对删除副作用的理解偏差。

## 测试与验证方式 / 结果
- 文档一致性人工审阅：通过。
- 协议映射条目与 Proto 变更对齐检查：通过。

## 潜在影响与回滚方案
- 潜在影响：若未同步客户端，旧客户端可能仍只认 `set/run/status/list/get`。
- 回滚方案：回退 `docs/specs/flow.md` 和 `docs/specs/protocol_map.md` 本次改动。 

