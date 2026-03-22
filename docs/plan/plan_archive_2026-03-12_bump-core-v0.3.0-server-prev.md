# Plan - Server：多 Listener 开关 + Parent Endpoint（适配 Core Pipe，为 RFCOMM 接入做准备）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/transport-pipe`
- Worktree：`d:\project\MyFlowHub3\worktrees\refactor-transport-pipe\MyFlowHub-Server`
- Base：`origin/main`
- 关联仓库（同一 workflow）：
  - `MyFlowHub-Core`：Pipe/MultiListener/ParentDialer
  - `MyFlowHub-SubProto`：适配 `core.IConnection` 变更（主要为测试 stub/mock）
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）

## 背景 / 问题陈述（事实，可审计）
- `hubruntime` 目前仅支持 1 个 TCP listener（`opts.Addr`），无法同时启用多个协议入口。
- 父链路（Parent link）拨号逻辑在 Core 层硬编码为 TCP（`net.Dial("tcp")`），阻碍未来使用 RFCOMM 拨号父节点。
- 用户期望：
  - TCP 与 Bluetooth Classic RFCOMM 都可作为 listen/dial（由参数配置）；
  - 可同时启用多个 listener，且可分别开关；开关变更以“重启生效”为第一版策略。

## 目标
1) 适配 Core 的 Pipe/MultiListener/ParentDialer 变更，使 Server 编译与测试通过。
2) `hubruntime.Options` 扩展为“多 listener 可配置开关”，并保持当前 TCP 行为默认不变。
3) 引入 `ParentEndpoint`（或等价机制）以支持未来 `tcp://` 与 `bt+rfcomm://` 等拨号方式（本轮只落地 TCP 实现与配置骨架）。

## 非目标
- 本仓不实现 RFCOMM 具体 listener/dialer（平台差异大，另起 workflow 在 Win/Android 落地）。
- 不调整子协议 wire/语义，不改 Major 路由规则。
- 不做运行期动态开关 listener（仅启动参数控制，需重启生效）。

## 约束（边界）
- Default RFCOMM UUID：固定一个默认值（由 workflow 生成），允许通过参数覆盖：
  - `0eef65b8-9374-42ea-b992-6ee2d0699f5c`

## 验收标准
- 使用 workflow-local `go.work` 联调：
  - `go test ./... -count=1 -p 1` 通过；
  - `cmd/hub_server` 仍可启动并监听 TCP（默认行为不变）。
- 在 Core 发布新 tag（若本次为破坏性变更）后，Server 能切换到新版本并在 `GOWORK=off` 下通过测试（作为发布验收）。

## 3.1) 计划拆分（Checklist）

### SRV0 - 归档旧 plan（已执行）
- 已执行：`git mv plan.md docs/plan/plan_archive_2026-03-11_transport-pipe-prev.md`
- 回滚点：撤销该 `git mv`。

### SRV1 - 适配 Core 连接/读写抽象变更（Pipe）
**目标**
- 修复因 `core.IConnection`/Reader/SendDispatcher 变更导致的编译错误与测试失败。

**涉及模块 / 文件（预期）**
- `hubruntime/runtime.go`（listener 装配保持兼容）
- `tests/*`（stub/mock 若依赖旧接口需适配）

**验收条件**
- `go test ./... -count=1 -p 1` 通过（在 workflow-local go.work 下）。

**回滚点**
- revert 该提交。

### SRV2 - 多 listener 装配：支持分别开关（重启生效）
**目标**
- 在 `hubruntime` 根据 Options 开关创建 listener 列表，并用 Core `MultiListener` 组合：
  - TCP listener：保持现有行为（默认启用）
  - RFCOMM listener：本轮仅预留配置与装配插槽（实现另起 workflow）

**涉及模块 / 文件（预期）**
- `hubruntime/options.go`（新增开关字段，保持 gomobile 友好）
- `hubruntime/runtime.go`（根据开关构建 listener 列表）
- `cmd/hub_server/main.go`（新增 flags/env 映射）

**验收条件**
- 只启用 TCP 时行为与当前一致；
- 禁用 TCP 且无其他 listener 时：启动失败并给出明确错误（避免“假启动”）。

**测试点**
- 最小单测/集成：Options 组合覆盖（TCP on/off）。

**回滚点**
- revert 该提交。

### SRV3 - Parent Endpoint + dialer 注入（为 BT dial 做准备）
**目标**
- 将父链路地址从“隐含 TCP 的 `ParentAddr`”升级为“可带协议的 endpoint”（建议格式：`tcp://127.0.0.1:9000`；未来扩展 `bt+rfcomm://<addr>?uuid=...`）。
- 将 Core 的父链路拨号替换为可注入 dialer；本仓负责根据配置注入 TCP dialer。

**涉及模块 / 文件（预期）**
- `hubruntime/options.go`（新增 `ParentEndpoint`；兼容旧 `ParentAddr`）
- `hubruntime/runtime.go`（组装 dialer 注入）

**验收条件**
- 兼容旧配置：仅设置 `ParentAddr` 时仍按 TCP 工作；
- 设置 `ParentEndpoint=tcp://...` 时工作一致。

**回滚点**
- revert 该提交。

### SRV4 - Code Review（阶段 3.3）+ 归档变更（阶段 4）
- 输出 `docs/change/2026-03-11_transport-pipe-server.md`，映射 SRV1~SRV3，并包含验证与回滚方案。

### SRV5 - 合并 / push（需你确认 workflow 结束后执行）
- 在 `repo/MyFlowHub-Server` 合并到 `main` 并 push（如需发布 tag 另起 workflow 或在归档中说明）。

---

## 验证命令（建议）
```powershell
$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
go test ./... -count=1 -p 1
```


