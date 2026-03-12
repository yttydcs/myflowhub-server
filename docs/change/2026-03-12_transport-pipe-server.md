# Server：多 Listener 开关 + ParentEndpoint（适配 Core Pipe，为 RFCOMM 接入预留配置骨架）

## 变更背景 / 目标
现状（变更前）：
- `hubruntime` 只能启动一个 TCP listener，无法同时启用多个协议入口。
- 父链路拨号隐含 TCP 语义，难以扩展到 `bt+rfcomm://...` 等 endpoint。
- Core 的连接抽象从 `RawConn()` 迁移到 `Pipe()` 后，Server 测试 stub/mock 需要同步适配。

目标（本次变更后）：
- 支持“多 listener 装配 + 分别开关（重启生效）”的框架能力（本轮 RFCOMM 仅保留配置位，不落地实现）。
- 引入 `ParentEndpoint`（支持 `tcp://...`），并通过注入 dialer 方式消除“父链路只能 TCP”的硬编码。
- 保持默认行为不变：不配置新字段时仍以 TCP 工作。

## 具体变更内容
### 修改
- `hubruntime.Options`：
  - 新增 `TCPEnable`（默认 true）、`RFCOMMEnable`（默认 false）、`RFCOMMUUID`（默认固定 MyFlowHub UUID）；
  - 新增 `ParentEndpoint`，并与 `ParentAddr` 保持兼容：两者任一非空都会使 `ParentEnable=true`。
- `hubruntime.Runtime.Start`：
  - 根据开关构建 listener 列表；多个 listener 时使用 Core `multi_listener.MultiListener` 组合；
  - `RFCOMMEnable=true` 时显式返回 “not implemented”，避免“假启动/隐性重连”。
  - 父链路拨号通过 `server.Options.ParentDialer` 注入 `dialParentEndpoint`：
    - 兼容 `host:port`（隐含 TCP）与 `tcp://host:port`；
    - 不支持的 scheme 会在启动前校验并报错（当前仅实现 TCP）。
- `buildConfig`：
  - 将父链路地址写入 Core 配置的 `parent.addr` 为“有效 parent target”（优先 `ParentEndpoint`）。
- `cmd/hub_server`：
  - 新增 flags：`-tcp-enable`、`-parent-endpoint`、`-rfcomm-enable`、`-rfcomm-uuid`。
- `tests/*`：
  - 测试 stub/mock 适配 `core.IConnection` 新接口：实现 `Pipe()`，移除 `RawConn()`。

## plan.md 任务映射
- SRV1：适配 Core Pipe 抽象（包含 tests stub/mock 更新）
- SRV2：多 listener 装配 + 开关（重启生效；RFCOMM 仅预留）
- SRV3：ParentEndpoint + dialer 注入（当前仅支持 TCP endpoint）

## 关键设计决策与权衡
- **兼容优先**：保留 `ParentAddr` 与 `-parent`，新增 `ParentEndpoint`/`-parent-endpoint`；未配置新字段时行为不变。
- **可维护性**：父链路“拨号策略”通过 dialer 注入集中管理，未来新增 RFCOMM 仅需扩展 dialer/listener，不需要改 Core 业务逻辑。
- **失败策略**：未实现的 RFCOMM/未知 endpoint scheme 直接启动失败（比后台无限重连更可审计、可定位）。
- **开关策略**：本轮开关变更以“重启生效”为准，避免引入运行期热切换复杂度与竞态。

## 测试与验证
在 workflow-local `go.work`（`worktrees/refactor-transport-pipe/go.work`）下验证：
- `cd MyFlowHub-Server; go test ./... -count=1 -p 1` ✅

## 潜在影响
- 直接以 struct literal 构造 `hubruntime.Options{...}` 的调用方需要显式设置 `TCPEnable=true`（本仓集成测试已更新）。
- 若配置了 `ParentEndpoint=tcp://...`，将走 endpoint dialer；若配置了非 tcp scheme，会在启动时报错（当前预期行为）。

## 回滚方案
- 回滚本次提交（或整体 revert）：
  - 恢复单 TCP listener 装配；
  - 移除 `ParentEndpoint`/dialer 注入，回退到旧的 `ParentAddr` 语义；
  - tests stub/mock 同步回滚到 `RawConn()` 版本（需同时回滚 Core）。

