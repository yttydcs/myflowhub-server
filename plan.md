# Plan - Server Docs：同步 Flow 删除部署协议约定

## Workflow 信息
- Repo：`MyFlowHub-Server`
- Branch：`chore/protocol-delete-docs`
- Worktree：`D:/project/MyFlowHub3/repo/MyFlowHub-Server/repo/MyFlowHub-Server/worktrees/chore-protocol-delete-docs`
- Base：`main`

## 项目目标与当前状态
- 目标：在 Server 文档中同步新增 `flow.delete` 协议约定（你明确要求协议修改必须同步 docs）。
- 当前状态：`docs/6-flow.md` 与 `docs/protocol_map.md` 仅描述 `set/run/status/list/get`。

## 可执行任务清单（Checklist）
- [x] SRV-DOC-1 更新 `docs/6-flow.md`（delete 动作与语义）
- [x] SRV-DOC-2 更新 `docs/protocol_map.md`（如通过生成命令）
- [x] SRV-DOC-3 文档一致性自检

## 任务明细

### SRV-DOC-1 更新 Flow 协议文档
- 目标：补齐 `action=delete` 请求/响应、权限、错误码、运行中断语义。
- 涉及模块/文件：
  - `docs/6-flow.md`
- 验收条件：
  - 明确 `flow.delete` 权限。
  - 明确 `delete_req/delete_resp` 字段。
  - 明确“删除时中断运行中的 run（需求指定）”。
- 测试点：
  - 文档内容与实现计划一致。
- 回滚点：
  - 回退文档修改。

### SRV-DOC-2 更新协议映射文档
- 目标：确保映射文档列出 delete action/type。
- 涉及模块/文件：
  - `docs/protocol_map.md`
- 验收条件：
  - Flow action/type 列表包含 delete/delete_resp 与 DeleteReq/DeleteResp。
- 测试点：
  - 生成命令（若可用）可执行：`go run ./cmd/protocolmapgen -write -out docs/protocol_map.md`。
- 回滚点：
  - 回退该文档。

### SRV-DOC-3 文档一致性自检
- 目标：防止文档描述与代码计划冲突。
- 涉及模块/文件：
  - `docs/6-flow.md`
  - `docs/protocol_map.md`
- 验收条件：
  - 关键字段、权限、错误码、Major 约定无冲突。
- 测试点：
  - 人工审阅 + 差异检查。
- 回滚点：
  - 回退本 workflow 文档变更。

## 依赖关系
- 依赖 Proto workflow（新增 action/type）。
- 与 SubProto workflow 并行，但发布前应做一致性核对。

## 风险与注意事项
- 风险：若 `protocol_map` 生成依赖环境不完整，需手工最小变更并在归档记录原因。
- 注意：仅文档改动，不引入实现代码。

