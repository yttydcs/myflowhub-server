# 2026-04-04_hubruntime-bootstrap-endpoint-dialer

## 变更背景 / 目标
- `hubruntime.Runtime.Start(...)` 虽然已经支持通用 parent endpoint 解析和持久父链拨号，但启动前 bootstrap 仍硬编码为 TCP-only。
- 本次目标是让启动前 bootstrap 复用现有 parent endpoint dialer，而不改变“启动前 SelfRegister + 启动后持久连接 register/rebind”的两段式时序。

## 具体变更内容
- `hubruntime/runtime.go`
  - 启动前只保留 `parseParentEndpoint(...)` 的 scheme 校验，不再额外拒绝非 TCP endpoint。
  - `selfRegisterNodeID(...)` 改为接收 `parentTarget + dialer`，并通过 `bootstrap.SelfRegisterOptions.Dial` 复用 `dialParentEndpoint(...)`。
  - 启动后 `sendRegisterOnConn(...)` / watcher 逻辑保持不变。
- `hubruntime/runtime_test.go`
  - 新增用例，验证 `selfRegisterNodeID(...)` 会把 `quic://...` 这类 generic target 原样交给注入 dialer，并完成同步 register 响应处理。

## Requirements impact
- none

## Specs impact
- none

## Lessons impact
- none

## Related requirements
- none

## Related specs
- `../specs/auth.md`
- `../specs/core.md`

## Related lessons
- none

## 对应 plan.md 任务映射
- `SRV-BOOT-1`：启动前 bootstrap 走通用 endpoint dialer
- `SRV-BOOT-2`：补 runtime 单测
- `SRV-BOOT-3`：Server 验证

## 经验 / 教训摘要
- 启动前 bootstrap 不需要自己理解 transport；Server 已经拥有 endpoint 解析和 dial 分发，就应该直接复用。
- 这类多 repo 验证不要盲跑 `go test`；先确认临时 `go.work` 是否覆盖了 Core / Proto / 必要 subproto 模块，否则很容易把版本矩阵问题误判成当前改动回归。

## 可复用排查线索
- 症状：`ParentEndpoint` 支持 `quic://` / 其他 scheme，但配置了 `self_id` 后启动前 bootstrap 直接报 only supports tcp。
- 触发条件：`parent.enable=true`、`self_id` 非空、`parentTarget` 不是裸 TCP。
- 关键词：`selfRegisterNodeID`, `dialParentEndpoint`, `parseParentEndpoint`, `SelfRegister`
- 快速检查：
  - 看 `Runtime.Start(...)` 是否还保留 `parentScheme != "tcp"` 的显式拒绝
  - 看 `selfRegisterNodeID(...)` 是否仍只接收 `parentAddr`

## 关键设计决策与权衡
- 不把 `Runtime.Start(...)` 做成更大的 test seam；只把 dialer 依赖收敛到 `selfRegisterNodeID(...)`，变更面最小。
- 不改 watcher / post-start register 路径，避免把“helper 改造”和“bootstrap 时序重构”混在一次 workflow 里。

## 测试与验证方式 / 结果
- Core helper 验证：
  - `cd D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/MyFlowHub-Core`
  - `$env:GOWORK='off'; go test ./bootstrap -count=1`
  - `$env:GOWORK='off'; go test ./... -count=1`
- Server 验证使用 workflow-local 临时 `go.work`，包含：
  - `D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/MyFlowHub-Core`
  - `D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/MyFlowHub-Server`
  - `D:/project/MyFlowHub3/repo/MyFlowHub-Proto`
  - `D:/project/MyFlowHub3/repo/MyFlowHub-SubProto/flow`
  - `D:/project/MyFlowHub3/repo/MyFlowHub-SubProto/stream`
- 在该临时 workspace 下执行：
  - `go test ./hubruntime -count=1`
  - `go test ./tests -run TestIntegrationVarStoreSetGetAcrossHub -count=1`
- 结果：通过

## 潜在影响
- 一旦 runtime 新增更多 parent endpoint scheme，启动前 bootstrap 会自动继承这些能力；相关 dialer 实现必须保证返回的 `core.IConnection` 可用于同步 request/reply。
- 若 workspace 没把 `Proto/flow/stream` 模块放进同一个 `go.work`，Server 侧验证可能先挂在版本矩阵而不是本次改动。

## 回滚方案
- 回退：
  - `hubruntime/runtime.go`
  - `hubruntime/runtime_test.go`
- 恢复启动前 TCP-only gate 和旧的 `selfRegisterNodeID(...)` 签名即可。

## 子Agent执行轨迹
- 未使用子Agent。
