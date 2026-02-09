# 变更归档：MyFlowHub-Server 公共协议包（protocol/*）

日期：2026-02-09

## 变更背景 / 目标
为 `MyFlowHub-Win` 等客户端复用服务端协议模型，避免在客户端重复定义 request/response 与常量，将原先位于 `internal/handler/*/types.go` 中的协议类型抽出为公共包 `protocol/*`，并将 handler 引用迁移到新包中，**保持运行时行为不变**。

## 具体变更内容（新增 / 修改 / 删除）
### 新增
- `protocol/auth`：认证相关协议类型
- `protocol/file`：文件传输/浏览相关协议类型
- `protocol/flow`：工作流相关协议类型
- `protocol/management`：管理相关协议类型
- `protocol/topicbus`：TopicBus 相关协议类型
- `protocol/varstore`：VarStore 相关协议类型

### 修改
- `internal/handler/*`：将原先本地 types 定义替换为 `protocol/*` 中的导出类型/常量引用（仅类型迁移与引用调整）。

### 删除
- 无（未删除功能模块；仅在 handler 内部减少重复 types 定义）。

## 对应 plan.md 任务映射
- S1：迁移主 worktree 的 WIP 变更到本 worktree（含新增 `protocol/*`）
- S2：审核并最小化变更范围（确保无行为修改）
- S3：提交并可审计化（形成可回放提交）

## 关键设计决策与权衡（性能 / 扩展性）
- 决策：仅导出“协议模型/常量/必要校验”，不把 handler 运行时逻辑暴露为可复用包，保持边界清晰。
- 扩展性：`protocol/*` 作为稳定依赖边界后，客户端与服务端可共享同一套模型；后续新增子协议/字段时可按包扩展。
- 权衡：外部项目开始依赖 `protocol/*` 后，需要更明确的兼容性策略；开发期先通过 `replace` 本地联调，后续建议 tag/版本固化。

## 测试与验证方式 / 结果
- `go test ./...`：通过。

## 潜在影响与回滚方案
### 潜在影响
- 协议模型成为公共依赖后，字段变更需要谨慎（建议：新增字段优先、避免破坏性重命名；必要时通过新版本包或版本字段处理）。

### 回滚方案
- 回滚对应提交（`git revert`）即可恢复到 handler 内部 types 定义方式；运行时行为预期不受影响。

