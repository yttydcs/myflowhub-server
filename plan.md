# Plan - Server：接入 RFCOMM（Bluetooth Classic）Listener + ParentEndpoint Dial

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`feat/bluetooth-rfcomm-transport`
- Worktree：`d:\project\MyFlowHub3\worktrees\feat-bluetooth-rfcomm-transport\repo\MyFlowHub-Server`
- Base：`main`
- 依赖仓：
  - `MyFlowHub-Core`：本分支引入 RFCOMM transport（本 workflow 需联动编译）

## 背景 / 问题陈述（事实，可审计）
- Server 当前：
  - TCP listener 已可用；
  - RFCOMM 仅有 `RFCOMMEnable/RFCOMMUUID` 配置位与 CLI flag，但 `Start()` 中直接返回 “not implemented”；
  - ParentEndpoint 目前仅支持 `tcp://` 或裸 `host:port`。
- 需求：Server 需要能够：
  - 在同一套 Process/Codec 下复用 RFCOMM 字节流 Pipe；
  - 同时启用 TCP + RFCOMM（可分别开关）；
  - ParentEndpoint 支持 `bt+rfcomm://...` 的 dial（与 TCP 功能对齐）。

## 目标
1) 使 `RFCOMMEnable` 真正生效：装配 RFCOMM listener（与 TCP 可并存）。
2) 扩展 `ParentEndpoint`：在不破坏现有 TCP 行为的前提下支持 `bt+rfcomm://...`。
3) 配置/参数可控：listen 与 dial 所需参数均可通过 Options/flags 配置（至少 uuid；可扩展 channel/adapter/secure）。

## 非目标
- 不改协议语义与子协议处理逻辑（仅扩展 transport 能力）。
- 不在 Server 内实现蓝牙扫描/按设备名解析（由 endpoint 扩展点与平台实现负责）。

## 验收标准
- `GOWORK=on`（workflow-local go.work）下：
  - `go test ./...` 通过。
- `GOWORK=off` 下：
  - 仅当 Core 已发布并在 go.mod 升级后，`go test ./...` 才要求通过（该动作由本 workflow 后续任务明确执行）。
- 运行期（手工冒烟）：
  - 启用 `-rfcomm-enable` 时不再返回 “not implemented”；
  - `-parent-endpoint bt+rfcomm://...` 能进入拨号路径并给出可定位错误/成功连接（依赖真实环境）。

## 3.1) 计划拆分（Checklist）

### SRV-BT0 - 归档旧 plan（已执行）
- 已执行：`git mv plan.md docs/plan_archive/plan_archive_2026-03-12_bluetooth-rfcomm-transport-server-prev.md`

### SRV-BT1 - RFCOMM Listener 装配（多入口可并存）
- 目标：将 runtime 中 RFCOMM 从 “not implemented” 改为真实 listener，并与 TCP 通过 `multi_listener` 组合。
- 涉及模块/文件（预期）：
  - `hubruntime/runtime.go`
  - `hubruntime/options.go`（如需补充 channel/adapter/secure 等字段）
  - `cmd/hub_server/main.go`（如需补充 flags）
- 验收条件：
  - TCP-only、RFCOMM-only、TCP+RFCOMM 三种组合均可启动（RFCOMM 真实环境不足时至少能返回明确错误）。
- 回滚点：revert。

### SRV-BT2 - ParentEndpoint 支持 `bt+rfcomm://`（dial）
- 目标：扩展 `normalize*/dialParentEndpoint` 支持多 scheme：
  - `tcp://host:port` / 裸 `host:port`（保持兼容）
  - `bt+rfcomm://<bdaddr>?uuid=...&channel=...`（新）
- 涉及模块/文件（预期）：
  - `hubruntime/runtime.go`（normalize + dial 分发）
- 验收条件：
  - endpoint 解析失败时错误可读；
  - tcp 行为不回归。
- 回滚点：revert。

### SRV-BT3 - 依赖/版本对齐（确保 GOWORK=off 可用）
- 目标：在 Core 具备可用发布版本后，升级 `go.mod` 中 Core 依赖，确保 `GOWORK=off go test ./...` 通过。
- 涉及文件：`go.mod`、`go.sum`
- 验收条件：`GOWORK=off go test ./... -count=1 -p 1` 通过。
- 回滚点：revert。

### SRV-BT4 - Code Review（强制）
- 审查项：需求覆盖/架构/性能/可读性/扩展性/稳定性与安全/测试覆盖。

### SRV-BT5 - 归档变更（强制）
- 输出：`docs/change/2026-03-12_bluetooth-rfcomm-transport-server.md`
- 标注：重大变更（新增 transport 能力 + endpoint 扩展）。

### SRV-BT6 - 合并 / push（需 workflow 结束后执行）
- 在 `repo/MyFlowHub-Server` 合并到 `main` 并 push。

