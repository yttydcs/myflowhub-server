# TODO - Server 文档勘误：VarStore `not found` 错误码示例对齐

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`chore/server-varstore-docs-errorcode`
- Worktree：`d:\project\MyFlowHub3\worktrees\chore-server-varstore-docs-errorcode`
- Base：`main`

## 项目目标与当前状态

### 目标
1) 将 VarStore 规范文档中的错误码示例与“错误码约定”对齐，避免客户端误读（`4` vs `404`）。
2) 不改变协议实现，仅修正文档描述。

### 当前状态（事实）
- 文档错误码约定定义：`4` = 未找到。
- 同一文档示例处出现：`{"code":404,"msg":"not found"}`。
- 当前实现侧在 `list/get` 等场景使用 `code=4`（非 `404`）。

## 可执行任务清单（Checklist）

- [x] `DOCERRATA-1`：修正规范文档错误码示例
  - 目标：将 `docs/3-varstore.md` 的 not found 示例码从 `404` 调整为 `4`。
  - 涉及文件：
    - `docs/3-varstore.md`
  - 验收条件：
    - 文档示例与错误码约定一致。
  - 回滚点：
    - revert 本任务提交。

- [x] `DOCERRATA-2`：文档一致性复查
  - 目标：确认同文档无其它 `404` 残留语义冲突。
  - 验收条件：
    - `docs/3-varstore.md` 内相关示例与约定一致。
  - 回滚点：
    - revert 本任务提交。

- [x] `DOCERRATA-3`：Code Review + 归档
  - 目标：完成审查闭环与变更归档。
  - 涉及文件：
    - `docs/change/2026-03-06_server-varstore-doc-errorcode-align.md`
  - 验收条件：
    - 审查结论完整；
    - 归档内容包含任务映射、影响评估、回滚方案。

## 依赖关系
- `DOCERRATA-1 -> DOCERRATA-2 -> DOCERRATA-3`

## 风险与注意事项
- 本 workflow 仅修文档，不改协议 wire 与实现行为；避免与功能修复混杂。
