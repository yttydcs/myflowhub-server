# 2026-03-28 Server docs：auth register permit list contract

## 变更背景 / 目标

- `MyFlowHub-Server` 需要为新的 `list_register_permits` action 提供稳定兼容壳和对外文档，否则 Win / SubProto 已落地后，Server docs 仍会停留在旧 admission contract。

## 具体变更内容

- 修改 `protocol/auth/types.go`
  - 重新导出 `list_register_permits` action 常量
  - 重新导出 `RegisterPermitInfo`、`ListRegisterPermitsReq/Resp`
  - 顺手补齐 auth admission 相关缺失的类型导出，保持兼容壳完整
- 修改 `docs/specs/auth.md`
  - 新增 `list_register_permits` 契约说明
  - 明确 `meta.register_permits` 只代表当前活动 permit
  - 明确列表权限沿用 `auth.permit.issue` 或 `auth.permit.revoke`
- 修改 `docs/specs/protocol_map.md`
  - 同步 `list_register_permits` 的 action / payload 映射
  - 明确 Proto generator 才是 canonical 来源，Server 侧仅保留同步副本

## Impact

- Requirements impact: `none`
- Specs impact: `updated`
- Lessons impact: `none`
- Related requirements: `none`
- Related specs:
  - `docs/specs/auth.md`
  - `docs/specs/protocol_map.md`
- Related lessons: `none`

## 对应 plan.md 任务映射

- `SERVER-PERMIT-1`
- `SERVER-PERMIT-2`
- `REVIEW-SERVER-PERMIT-1`
- `ARCHIVE-SERVER-PERMIT-1`

## 经验 / 教训摘要

- Server 仓的 auth protocol 壳和 stable spec 必须跟 Proto canonical wire 一起推进，否则 UI/运行时已经可用，文档侧却还停在旧 contract。
- 对这类镜像型 `protocol_map` 文档，必须继续把 Proto 生成产物视为唯一 canonical 来源。

## 可复用排查线索

- 症状
  - Win/SubProto 已经能调用 permit list，但 Server stable docs 找不到 `list_register_permits`
  - Server protocol 壳缺少对应 req/resp alias
- 触发条件
  - Proto 新增 admission action 后，只改了下游实现，没有同步 Server 壳和文档
- 关键词
  - `list_register_permits`
  - `RegisterPermitInfo`
  - `docs/specs/auth.md`
  - `docs/specs/protocol_map.md`
- 快速检查
  - 查看 `protocol/auth/types.go` 是否导出 permit list 常量和结构
  - 查看 `docs/specs/auth.md` 是否已有 permit list action 和权限说明
  - 查看 `docs/specs/protocol_map.md` 是否已出现 permit list 映射

## 关键设计决策与权衡

- 不在 Server 仓引入新的业务权限，只文档化“issue 或 revoke 任一权限即可列出”
  - 优点：和当前 Core 权限模型保持一致
  - 代价：如果后续要拆分独立 list 权限，需要再补文档与壳层
- `docs/specs/protocol_map.md` 继续保持镜像副本角色
  - 优点：Server 仓仍可本地查阅稳定映射
  - 代价：同步时需要显式说明 canonical 来源

## 测试与验证方式 / 结果

- 临时 workspace 下执行 `go test ./protocol/... -count=1 -p 1`
  - 环境：临时 `go.work` 指向本地 `MyFlowHub-Core`、Proto worktree 和 Server worktree
  - 结果：通过

## 潜在影响与回滚方案

- 潜在影响
  - 若下游继续读取旧 Server stable docs，会遗漏 permit list 的稳定 contract
- 回滚方案
  - 回退 `protocol/auth/types.go`
  - 回退 `docs/specs/auth.md`
  - 回退 `docs/specs/protocol_map.md`

## 子Agent执行轨迹

- 本轮未使用子Agent
