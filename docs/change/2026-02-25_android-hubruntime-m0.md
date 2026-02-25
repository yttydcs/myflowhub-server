# 2026-02-25 Android Hub Runtime（M0）

## 背景 / 目标

为 Android 端提供“可嵌入的 Hub + 最小 UI”能力的 Server 侧支撑：
- 提供不依赖 OS signal 的可嵌入 runtime（供 gomobile/Android 宿主调用）
- 父链可用：Hub 上联父节点后，父侧能够把该连接绑定为该 Hub 的 node_id（否则树路由会丢帧）
- 保持 wire 协议不变（SubProto/Action/JSON/HeaderTcp 语义不改）

## 变更内容

### 新增
- `hubruntime/`：新增可嵌入的 Hub Runtime
  - `hubruntime/options.go`：配置结构 + env 默认值（与 `cmd/hub_server` 对齐）
  - `hubruntime/runtime.go`：`Start/Stop/Status` + 父链 bootstrap

### 修改
- `cmd/hub_server/main.go`：
  - 重构为复用 `hubruntime`（避免 CLI 与嵌入模式漂移）
  - 新增可选参数：
    - `-workdir` / `HUB_WORKDIR`：工作目录（Android 私有目录场景）
    - `-self-id` / `HUB_SELF_ID`：自注册与父链 bootstrap 的 self_id
- `tests/integration_root_hub_ping_test.go`：
  - 改为覆盖“Root → Hub management 命令转发 + parent bootstrap”的真实链路
  - 客户端侧改为走 `auth register`，不再手工 `SetMeta("nodeID")`

## 任务映射（plan.md）

- SRV1：新增可嵌入 Hub Runtime（Go）✅
- SRV2：父链 bootstrap（确保树路由可用）✅
- SRV3：集成测试覆盖 bootstrap ✅

## 关键设计决策与权衡

1) **不重写 Hub**：runtime 仍复用 `myflowhub-core` 的 `Server + Process/Dispatcher` 与 `myflowhub-server/modules` 默认模块装配。
2) **父链 bootstrap（两段式）**：
   - 启动前：当 `parent.enable + self_id` 存在时，先对父节点执行一次 `auth register` 获取/确认 `node_id`，避免本地配置与父侧分配不一致。
   - 启动后：监测 parent connection 建立后，在该持久连接上再发送一次 `auth register`，让父侧把该连接绑定为该 Hub 的 `meta(nodeID)`，从而：
     - 父侧可以按 node_id 路由/转发到该连接
     - 父侧 `sourceMismatch` 不会丢弃该 Hub 发来的非登录协议帧
3) **父连接本地侧 nodeID 兜底**：为避免 parent 连接 `meta(nodeID)==0` 导致子协议帧被 Dispatcher 直接丢弃（尚未进入“父连接免检”分支），runtime 会在本地将 parent conn 的 `meta(nodeID)` 初始化为非 0（当前固定为 1）。
   - 取舍：该值仅用于通过“已登录”门槛；真实的来源校验对 parent conn 仍由“父连接免检”规则处理。
4) **工作目录**：runtime 支持 `WorkDir` 并在 Start 时 `chdir`，以兼容现有相对路径配置文件（Android 私有目录写入）。

## 测试与验证

在本 worktree 下执行（由于 workspace `go.work` 未包含 worktree 路径，采用 `GOWORK=off`）：

```powershell
cd d:\project\MyFlowHub3\worktrees\android-hub-m0\MyFlowHub-Server
$env:GOWORK='off'
go test ./... -count=1 -p 1
go test ./tests -run TestRootHubPing -count=1
```

结果：通过。

## 潜在影响与回滚方案

### 潜在影响
- 新增 `cmd/hub_server` 参数（可选），默认行为保持兼容；仅当启用 `self-id + parent` 时会发生自注册/覆盖 node-id 的行为。
- `WorkDir` 使用 `os.Chdir` 为进程级别变更：嵌入宿主应避免同进程内同时运行多个依赖不同相对路径的实例（Android M0 场景为单实例可接受）。
- parent bootstrap 会额外发送一次 `auth register`（对开放注册的现状配置是预期行为）。

### 回滚
- 直接 revert 本次功能提交（引入 `hubruntime` 的提交），恢复原 `cmd/hub_server` 启动逻辑与旧测试。

