# 2026-03-22 Hubruntime Node Display Name Follow-up

## 变更背景 / 目标

`hubruntime` 父链持久连接上的 auth register 之前只发送 `device_id`。这会让 direct-child 名称 bootstrap 缺少发送端输入，父节点只能回退 `node_id`。

本次目标是在不改 bootstrap 时序的前提下，让父链 register payload 在当前 `node.display_name` 非空时带出该字段。

## 具体变更内容

- `hubruntime/runtime.go`
  - `startParentBootstrapWatcher` 接收当前 runtime `cfg`
  - 发送 register 前读取 `node.display_name`
  - `sendRegisterOnConn` 对显示名做 `TrimSpace`，仅在非空时下发 `display_name`
- `hubruntime/runtime_test.go`
  - 覆盖“带显示名”“空白值省略”“配置读取裁剪”路径
- `docs/specs/auth.md`
  - 补充 auth `register/login/resp` 中可选 `display_name` 的长期说明

## 对应计划任务映射

- `SRV1`
- `SRV2`

## 关键设计决策与权衡

- 读取当前 runtime `cfg`，而不是依赖 `Options` 快照，确保 register 发送值和实际配置一致。
- 空白名称不发送，保持旧 bootstrap 行为和兼容性。
- 本次不改变运行中配置变更的传播时机；名称更新仍依赖后续重连或 direct-child rename refresh。

## Requirements / Specs 影响检查

- Requirements impact：`none`
- Specs impact：`updated`
- Related requirements：
  - [management-node-display-name.md](/D:/project/MyFlowHub3/worktrees/MyFlowHub3-feat-node-display-name-followup/docs/requirements/management-node-display-name.md)
- Related specs：
  - [management-config-layering.md](/D:/project/MyFlowHub3/worktrees/MyFlowHub3-feat-node-display-name-followup/docs/specs/management-config-layering.md)
  - [auth.md](/D:/project/MyFlowHub3/worktrees/MyFlowHub-Server-feat-node-display-name-followup/docs/specs/auth.md)
- Lessons：`none`

## 测试与验证方式 / 结果

- `GOWORK=off go test ./hubruntime -count=1`：通过
- `GOWORK=off go test ./... -count=1`：失败
  - 阻塞点：`protocol/exec/types.go` 依赖的 Proto 类型在当前仓内即不匹配
  - 结论：与本次 `hubruntime` / auth spec 改动无关

## 潜在影响与回滚方案

### 潜在影响

- 运行中若只改 `node.display_name` 而不触发父链重连，bootstrap payload 不会立即重发；即时刷新由上游 SubProto rename 回程承担。

### 回滚方案

- 回退 `hubruntime/runtime.go`
- 回退 `hubruntime/runtime_test.go`
- 回退 `docs/specs/auth.md`

## 子 Agent 执行轨迹

- `SRV1` -> `Copernicus (019d1618-11ec-7632-82b9-8bd406d8b22a)` -> `D:\project\MyFlowHub3\worktrees\MyFlowHub-Server-feat-node-display-name-followup`
  - 文件：`hubruntime/runtime.go`、`hubruntime/runtime_test.go`
  - 验收：`go test ./hubruntime -count=1` 通过；整仓失败点已明确为既有 `protocol/exec` 构建问题
