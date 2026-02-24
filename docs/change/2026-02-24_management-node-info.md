# 变更说明：Server 接入 management/node_info（集成测试）

## 变更背景 / 目标
management 子协议新增 `node_info` 后，需要在 Server 仓补充最小集成验证，确保节点能够正确响应并返回基础信息 KV。

## 具体变更内容
- 修改：
  - `protocol/management/types.go`：兼容壳补充 `ActionNodeInfo/Resp`、`NodeInfoReq/Resp` 的常量与类型委托
- 新增：
  - `tests/integration_management_node_info_test.go`：新增集成测试，发送 `node_info` 并断言返回包含 `platform/node_id`

## 对应 plan.md 任务映射
- `worktrees/node-info/MyFlowHub-Win/plan.md`
  - T4. Server：接入新能力 + 集成测试 + 发布 `v0.0.2`

## 关键设计决策与权衡
- 采用“最小断言”策略：
  - `platform/node_id` 是稳定且跨构建环境可得的字段；
  - 不强依赖 semver（测试二进制通常为 `(devel)`），避免测试在 CI/本地差异下不稳定。

## 测试与验证方式 / 结果
- 已在本地执行：`go test ./...`

## 潜在影响与回滚方案
- 影响：仅新增测试与兼容壳常量/类型委托，不改变运行逻辑。
- 回滚：revert 本提交。

