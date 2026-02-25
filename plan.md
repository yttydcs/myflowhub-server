# Plan - Android：Hub + UI（M0 可行性验证）

> 位置：`d:\project\MyFlowHub3\worktrees\android-hub-m0\MyFlowHub-Server\plan.md`  
> 目标：用**最小改动复用现有 Go Hub 栈**，在 Android 上跑通“后台常驻 Hub + 局域网可见 + 最小 UI + 最小冒烟验证”。  
> 说明：本 plan 只覆盖 M0（可行性验证）；M1（全量能力/体验完善）另起 workflow。

## 0. Workflow 信息

- Workflow 名称：`android-hub-m0`
- 分支（本仓）：`feat/android-hub-m0`
- Worktree（本仓）：`d:\project\MyFlowHub3\worktrees\android-hub-m0\MyFlowHub-Server`
- Base：`origin/main`
- 关联仓库（同一 workflow，预计新增）：
  - `MyFlowHub-Android`（新仓，用于 Android App + gomobile 绑定）
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）

## 1. 需求确认（已确认）

- 手机也是 Hub（可连接父节点 + 承载子树连接），后台长期运行（接受 Foreground Service 常驻通知）。
- 局域网可见（非回环监听）。
- auth 策略对齐现状：开放注册（不做配对/审批）；默认角色/权限对齐 `hub_server` 默认值（含 `file/flow/exec`）。
- 第一个版本先做 **M0 验证可行性**；后续再补齐全量体验与能力。
- 分发：仅 APK 自用；用户从 GitHub 手工下载（不考虑应用商店审核）。

## 2. 约束（边界）

- wire 不改：SubProto 值 / Action 字符串 / JSON payload / HeaderTcp 语义保持不变。
- 不在 `d:\project\MyFlowHub3\repo\MyFlowHub-Server`（控制面）做实现改动；只在本 worktree 内改。
- 若需要新建 Android 仓库：必须说明 `repo/` 控制面目录下“新仓初始化”的例外原因，并在阶段 4 归档。

## 3. 当前事实（用于设计/验证）

- `cmd/hub_server/main.go` 目前是 CLI 入口（flag/env + signal），不适合 gomobile/Android 宿主直接调用。
- Core/Dispatcher 有 `sourceMismatch` 校验：**子连接若未绑定 meta(nodeID)，非 AllowSourceMismatch 的子协议会被丢弃**。  
  => Hub 上联父节点时，需要在父侧把“该 child 连接的 meta(nodeID)”绑定好，否则树路由无法工作。
- auth 子协议（SubProto=2）允许 SourceMismatch，且 register 不需要签名（现状是开放注册）。

## 4. M0 目标（验收口径）

### 4.1 必须达成
1) Go 侧提供“可嵌入”的 Hub Runtime（Start/Stop/Status），且 `cmd/hub_server` 复用同一装配逻辑（避免漂移）。
2) Android 侧（新仓）能：
   - 以 Foreground Service 启动/停止 Hub；
   - 展示 Hub 监听地址、NodeID、父链状态、最近错误、日志路径；
3) 冒烟验证至少包含：
   - `management node_echo` 对本机 Hub 成功（可用 Go 自测函数或 LAN 另一设备验证）。

### 4.2 暂不要求（M1）
- UI 全量功能对齐 Win。
- 对开放注册/高权限的安全加固（仅做明显风险提示，不做改造）。
- 自动更新/Release 流水线/可裁切 module set 产品化。

## 5. 关键设计决策（M0）

1) **不重写 Hub**：复用 `myflowhub-core + myflowhub-server/modules + myflowhub-subproto/*`。
2) **Android 集成方式**：优先 `gomobile bind` 产出 AAR（目标至少 `android/arm64`）。
3) **父链可用性（关键）**：采用“两段式”：
   - 启动前（若启用 parent）先通过独立 TCP 连接对父节点 `register` 获取/确认 NodeID（用于本节点 NodeID 与父侧一致）。
   - parent link 建立后，在该持久连接上再发送一次 `register`（fire-and-forget），让父侧对该 child 连接写入 meta(nodeID)，从而通过 `sourceMismatch` 校验。
4) **状态/文件目录**：M0 允许在 Go Runtime 启动时设置工作目录（或等价的 baseDir）指向 Android App 私有目录，确保相对路径（如 `config/node_keys.json`）可写。

## 6. 计划拆分（Checklist）

> 约定：每个任务必须有可回滚点；不得引入计划外改动。需要新增任务时，回到本 plan 更新并请你确认。

### SRV0 - 归档旧 plan（已执行）
- 目标：保留历史 plan，避免覆盖。
- 已执行：`plan.md` → `docs/plan_archive/plan_archive_2026-02-20_server-subproto-modules-all.md`
- 回滚点：撤销 `git mv`。

### SRV1 - 新增可嵌入 Hub Runtime（Go）
- 目标：新增一个不依赖 signal 的 runtime 包，供 Android/gomobile 调用；并保证 `cmd/hub_server` 复用该 runtime。
- 涉及模块/文件（预期）：
  - `hubruntime/options.go`（配置结构，支持 addr/node_id/parent/auth/perf 参数 + workdir/self_id）
  - `hubruntime/hub.go`（Start/Stop/Status）
  - `cmd/hub_server/main.go`（重构为调用 hubruntime，保留原 flags/env 行为）
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过
  - `GOWORK=off go test ./... -count=1 -p 1`（仅当依赖均为 semver 时执行；M0 可暂用 go.work 联调）
- 回滚点：revert 提交。

### SRV2 - 父链 bootstrap（确保树路由可用）
- 目标：在启用 parent 的场景下，自动完成“父侧 child 连接 meta(nodeID) 绑定”，避免 `sourceMismatch` 丢帧。
- 设计要点：
  - 启动前：如 `parent.enable=true` 且 `self_id` 存在，先对 `parent.addr` 发起一次 `register` 获取 node_id（必要时覆盖本地 node_id）。
  - 启动后：检测到 parent 连接建立后，在该连接上再次发送 `register`（无需等待响应）以绑定父侧连接元数据。
- 涉及模块/文件（预期）：
  - `hubruntime/bootstrap.go`
  - `hubruntime/hub.go`
- 验收条件：
  - 新增/更新一条 integration test：不需要手工 `SetMeta(\"nodeID\")` 也能通过跨父链的最小请求（见 SRV3）。
- 回滚点：revert 提交。

### SRV3 - 集成测试：Root ↔ Hub（覆盖 bootstrap）
- 目标：用桌面集成测试验证“父链 + register bootstrap”在框架约束下真实可用。
- 涉及文件（预期）：
  - `tests/integration_root_hub_ping_test.go`（新增或改造：通过 hubruntime 启动 hub，并依赖 bootstrap，而非手工 SetMeta）
- 验收条件：
  - `go test ./tests -run TestRootHubPing -count=1` 通过
- 回滚点：revert 提交。

### AND0 - 新建 Android 仓库 + worktree（需你确认是否允许例外）
- 目标：新增 `MyFlowHub-Android`（Android App + gomobile bind）。
- 说明（例外点）：因为是**新仓**，需要在 `repo/` 下创建其控制面 git 仓库；会产生一次“初始化提交”以便 `git worktree add`（后续实现仍在 worktree 内完成）。
- 验收条件：
  - `repo/MyFlowHub-Android` 作为控制面存在
  - `worktrees/android-hub-m0/MyFlowHub-Android` 存在且分支为 `feat/android-hub-m0`
- 回滚点：
  - 删除新仓目录（若未推送远端）
  - 移除 worktree（`git worktree remove`）

### AND1 - Android App 骨架（Compose + Foreground Service）
- 目标：能在真机启动前台服务，并展示/控制 Hub 的启停。
- 涉及模块/文件（预期，位于新仓）：
  - `app/`（Kotlin + Compose）
  - `HubService`（Foreground Service）
  - 最小页面：配置（addr/parent/self_id）+ 状态页 + 日志路径展示
- 验收条件：
  - 安装 APK 后可 Start/Stop，常驻通知可见
- 回滚点：revert 提交。

### AND2 - gomobile 绑定（AAR）+ 集成到 App
- 目标：`gomobile bind` 产出 AAR（至少 arm64），并在 App 内调用 `Start/Stop/Status`。
- 验收条件：
  - AAR 构建成功
  - App 运行后 Hub 真正开始监听（LAN 可连接）
- 回滚点：revert 提交。

### AND3 - M0 冒烟脚本/文档
- 目标：把验证步骤写成可复现文档（含依赖安装、命令与期望结果）。
- 验收条件：
  - 文档包含：构建 AAR、构建 APK、启动 Hub、LAN 设备验证 node_echo。
- 回滚点：revert 提交。

### 3.3 - Code Review（强制）
- 按全局 3.3 清单逐项输出结论（通过/不通过）。
- 不通过：返回对应任务修正，重新 review。

### 4 - 归档变更（强制）
- 在各自 worktree 根目录创建 `docs/change/` 并新增归档文档：
  - `MyFlowHub-Server/docs/change/YYYY-MM-DD_android-hubruntime-m0.md`
  - `MyFlowHub-Android/docs/change/YYYY-MM-DD_android-hub-m0.md`
- 必须包含：任务映射、关键决策与权衡、验证方式与结果、回滚方案、以及（如有）控制面例外说明。

## 7. 验证命令（Go 侧统一）

```powershell
$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
go test ./... -count=1 -p 1
go test ./tests -run TestRootHubPing -count=1
```

