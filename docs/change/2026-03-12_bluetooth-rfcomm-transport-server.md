# Server：接入 RFCOMM（Bluetooth Classic）Listener + ParentEndpoint Dial（重大变更）

## 背景 / 目标
- 背景：Core 已提供 Pipe 抽象与 RFCOMM transport（字节流），Server 需要将其装配为 listener，并让 parent-link 支持 `bt+rfcomm://...` 拨号。
- 目标：
  1) RFCOMM listener 可启用（不再返回 “not implemented”），并可与 TCP 并存、分别开关；
  2) ParentEndpoint 支持多 scheme：`tcp://` / 裸 `host:port`（保持兼容）+ `bt+rfcomm://`（新）。

## 变更内容
### 新增 / 修改
- `hubruntime/options.go`
  - 新增 RFCOMM 配置项：`RFCOMMEnable/RFCOMMUUID/RFCOMMChannel/RFCOMMAdapter/RFCOMMInsecure`
- `cmd/hub_server/main.go`
  - 增补 flags：
    - `-rfcomm-enable`
    - `-rfcomm-uuid`
    - `-rfcomm-channel`
    - `-rfcomm-adapter`
    - `-rfcomm-insecure`
  - `-parent-endpoint` 示例扩展：支持 `bt+rfcomm://...`
- `hubruntime/runtime.go`
  - 装配 RFCOMM listener（`rfcomm_listener.New(...)`）
  - 支持 TCP + RFCOMM 多 listener 组合（`multi_listener.New(...)`）
  - ParentEndpoint 拨号分发：`tcp` 与 `bt+rfcomm`（复用 Core dial）

## 关键设计决策与权衡
- **多 listener**：通过 Core 的 `multi_listener` 组合器实现，避免在 Server 内部写多套 goroutine/关闭逻辑。
- **endpoint 解析**：对未知 scheme / 非法参数尽早失败，返回可定位错误（避免后台无限重连造成不可审计行为）。
- **可扩展性**：新增承载协议只需扩展“listener + dial”装配，不需要改动业务协议与路由。

## 测试与验证
- `go test ./... -count=1`
- 手工冒烟（依赖真实环境）：
  - `-rfcomm-enable` 可进入 RFCOMM listener 启动路径；
  - `-parent-endpoint bt+rfcomm://...` 可进入 RFCOMM dial 路径并返回可诊断的成功/失败信息。

## 潜在影响
- RFCOMM 的实际可用性受平台蓝牙栈与权限影响（Linux/Windows/Android 行为可能不同）。

## 回滚方案
- revert 本次提交；或运行期关闭 `-rfcomm-enable` 回退到 TCP-only。

