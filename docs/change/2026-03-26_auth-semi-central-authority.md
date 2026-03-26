# 2026-03-26 Server docs：auth semi-central authority contract

## 变更背景 / 目标
- 同步半中心 auth authority 方案的稳定文档，避免实现已变而 `docs/specs/auth.md` 仍停留在“直接父节点就是 authority”的旧表述。

## 具体变更内容
- 修改 [`docs/specs/auth.md`](D:/project/MyFlowHub3/worktrees/server-auth-semi-central-authority/docs/specs/auth.md)
  - 新增 `auth.authority_mode=semi-central` 的权威选择规则
  - 新增 `authority_policy_sync` 数据契约
  - 记录“lease 缺失/过期但父链在线时仍允许 bootstrap 路由”的边界
  - 明确断链后只允许本地已知身份登录
  - 明确 approve / reject / permit 仍是 authority 本地操作
- 同步 [`docs/specs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/server-auth-semi-central-authority/docs/specs/protocol_map.md)
  - 从 Proto canonical 文档复制最新协议映射副本

## Impact
- Requirements impact: `none`
- Specs impact: `updated`
- Lessons impact: `none`
- Related requirements: `none`
- Related specs:
  - [`docs/specs/auth.md`](D:/project/MyFlowHub3/worktrees/server-auth-semi-central-authority/docs/specs/auth.md)
  - [`docs/specs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/server-auth-semi-central-authority/docs/specs/protocol_map.md)
  - [`docs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/proto-auth-semi-central-authority/docs/protocol_map.md)
- Related lessons: `none`

## 对应 plan.md 任务映射
- `AUTHPOL-SRV-1`

## 经验 / 教训摘要
- 这类“运行时策略 + 多跳转发”改动，如果只改实现不改稳定 spec，后续会很容易重新掉回“直接父就是 authority”的旧理解。

## 可复用排查线索
- 症状
  - 文档仍写“父节点就是唯一 authority”，但实现已经按 root lease / edge hub response 工作
- 触发条件
  - 新增 `authority_policy_sync` 后未同步稳定文档
- 关键词
  - `semi-central`
  - `effective_authority_id`
  - `authority unavailable`
- 快速检查
  - 查看 [`docs/specs/auth.md`](D:/project/MyFlowHub3/worktrees/server-auth-semi-central-authority/docs/specs/auth.md) 是否包含 `authority_policy_sync`
  - 查看 [`docs/specs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/server-auth-semi-central-authority/docs/specs/protocol_map.md) 是否出现 `ActionAuthorityPolicySync`

## 关键设计决策与权衡
- 只更新稳定 contract，不在 Server 仓额外扩展审批远程化或 UI 行为。
- 协议映射副本继续以 Proto canonical 文档为准，Server 只保留同步副本。

## 测试与验证方式 / 结果
- 文档与实现对齐检查：完成
- Server 代码测试：本轮未执行
  - 原因：Server 仓本轮只有稳定文档更新，没有运行时代码变更

## 潜在影响与回滚方案
- 潜在影响
  - 若下游仍按旧 spec 理解 admission authority，会看到与实际实现不一致的行为差异
- 回滚方案
  - 回退 [`docs/specs/auth.md`](D:/project/MyFlowHub3/worktrees/server-auth-semi-central-authority/docs/specs/auth.md)
  - 回退 [`docs/specs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/server-auth-semi-central-authority/docs/specs/protocol_map.md)

## 子Agent执行轨迹
- 本轮未使用子Agent
