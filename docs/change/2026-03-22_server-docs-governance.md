# 2026-03-22 Server Docs Governance

## Background
- `MyFlowHub-Server/docs` 之前将长期 spec、历史计划归档和完成结果归档混放在顶层，缺少统一 taxonomy 和入口索引。
- 顶层稳定文档使用 `2-auth.md`、`6-flow.md`、`权限.md` 等 legacy 命名，当前路径语义不清晰，也不利于后续新增规范继续扩展。

## Changes
- 创建 Server 仓库标准分类目录与索引：
  - `docs/README.md`
  - `docs/requirements/README.md`
  - `docs/specs/README.md`
  - `docs/plan/README.md`
  - `docs/change/README.md`
  - `docs/lessons/README.md`
- 将稳定长期规范迁移到 `docs/specs/` 并规范命名：
  - `docs/specs/auth.md`
  - `docs/specs/varstore.md`
  - `docs/specs/topicbus.md`
  - `docs/specs/file.md`
  - `docs/specs/flow.md`
  - `docs/specs/exec.md`
  - `docs/specs/core.md`
  - `docs/specs/permission.md`
  - `docs/specs/protocol_map.md`
- 将 `docs/plan_archive/` 纯迁移到 `docs/plan/`。
- 生成 Server 仓库新的 `docs/change/README.md` 与 `docs/plan/README.md`。
- 批量更新仓库内 markdown 对旧 doc 路径的 canonical 引用。

## Related Plan
- `D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\plan\plan_archive_2026-03-22_server-docs-governance.md`

## Related Requirements
- 无现存 server 级叶子需求文档需要改写

## Related Specs
- `D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\specs\README.md`
- `D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\specs\protocol_map.md`
- `D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\specs\flow.md`
- `D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\specs\auth.md`

## Requirements Impact
- none

## Specs Impact
- updated

## Plan Task Mapping
- `SRV-DOC-001`：完成。Server taxonomy 与 category README 已补齐。
- `SRV-DOC-002`：完成。顶层长期 spec 已迁到 `docs/specs/` 并规范命名。
- `SRV-DOC-003`：完成。`docs/plan_archive/` 已迁到 `docs/plan/`。
- `SRV-DOC-004`：完成。Server `docs/change/README.md` 已重建。
- `SRV-DOC-005`：完成。仓库内对旧 doc canonical 路径的显式引用已迁到新路径。
- `SRV-DOC-006`：完成。generated 边界、入口层和目标文件存在性已验证。

## Design Decisions And Tradeoffs
- 选择“纯净优先”策略：直接迁移 canonical 目录与稳定 spec 文件，不保留旧顶层 stub。
- legacy 数字命名只保留在历史归档正文中，不再继续作为当前规范入口。
- `protocol_map.md` 不在本仓库声明本地生成命令，而是明确标注 canonical 生成源位于 `MyFlowHub-Proto`，避免继续误导读者在错误仓库执行生成。

## Validation
- 目录与关键文件存在性检查：
  - `docs/README.md`
  - `docs/requirements/README.md`
  - `docs/specs/README.md`
  - `docs/plan/README.md`
  - `docs/change/README.md`
  - `docs/lessons/README.md`
  - `docs/specs/protocol_map.md`
  - `docs/specs/auth.md`
  - `docs/specs/core.md`
  - `docs/specs/flow.md`
- 入口层残留检查：
  - `rg -n "docs/plan_archive|docs/[2-7]-|docs/core\\.md|docs/权限\\.md|docs/protocol_map\\.md" docs/README.md docs/specs/README.md docs/change/README.md docs/plan/README.md -S`
  - 结果：无命中
- generated 区块完整性检查：
  - `rg -n "BEGIN GENERATED|END GENERATED" docs/specs/protocol_map.md -S`
  - 结果：区块边界保留
- 顶层旧 spec 清理检查：
  - `Get-ChildItem docs -File`
  - 结果：仅保留 `README.md`

## Docs Governor Check
- Requirements impact: none
- Specs impact: updated
- Lessons impact: none
- Index update required: yes
- 结论：
  - 本次无需新增 `requirements` 叶子文档
  - `specs` taxonomy 和长期规范入口已更新
  - 无独立 `lessons` 产出
  - 根索引与分类索引已同步更新

## Rollback
- 回退本次 Server docs governance 相关提交即可恢复：
  - `docs/README.md`
  - `docs/requirements/`
  - `docs/specs/`
  - `docs/plan/`
  - `docs/change/README.md`
  - `docs/lessons/`
  - 仓库内针对旧路径的 markdown 引用改写

## Agent Trace
- 本轮未使用子 Agent。
- 执行轨迹：
  - `SRV-DOC-001` -> MainAgent -> `D:\project\MyFlowHub3\worktrees\server-docs-governance` -> `docs/README.md`, category README files -> 通过
  - `SRV-DOC-002` -> MainAgent -> `D:\project\MyFlowHub3\worktrees\server-docs-governance` -> `docs/specs/*` -> 通过
  - `SRV-DOC-003` -> MainAgent -> `D:\project\MyFlowHub3\worktrees\server-docs-governance` -> `docs/plan/*` -> 通过
  - `SRV-DOC-004` -> MainAgent -> `D:\project\MyFlowHub3\worktrees\server-docs-governance` -> `docs/change/README.md` -> 通过
  - `SRV-DOC-005` -> MainAgent -> `D:\project\MyFlowHub3\worktrees\server-docs-governance` -> server markdown refs -> 通过
  - `SRV-DOC-006` -> MainAgent -> `D:\project\MyFlowHub3\worktrees\server-docs-governance` -> validation evidence -> 通过
