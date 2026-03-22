# 2026-03-22 Hubruntime Layered Config Persistence

## 变更背景 / 目标

让 `hubruntime` 支持 management `config_set` 的持久化默认层，同时保留现有运行期 `Set` 语义，形成以下配置优先级：

- persistent default
- env / flags / caller explicit override
- runtime-only overlay

## 具体变更内容

- `hubruntime/options.go`
  - 新增 `DefaultOptions()`
  - 为 env / flags / caller 增加 `ConfigOverrideKeys` 跟踪能力
  - `Normalize()` 只在未显式 override 时补默认值，避免无意压过持久化层
- `hubruntime/layered_config.go`
  - 新增 `layeredConfig`
  - 持久化文件路径：`config/runtime_config.json`
  - `Set()` 保持运行期覆盖
  - `SetPersistent()` 以原子写方式更新持久化层
  - `Get()` / `Keys()` 返回 effective config
- `hubruntime/runtime.go`
  - 启动时先加载分层配置，再将 effective config 回灌到 runtime `Options`
  - 使重启后可从持久化层恢复 `addr` / `parent.*` / `auth.*` / `process.*` / `send.*`
- `cmd/hub_server/main.go`
  - 通过 `flag.Visit` 记录显式 flag 覆盖键
- `hubruntime/layered_config_test.go`
  - 覆盖优先级、运行期 `Set`、持久化写入与 effective option 回灌

## plan.md 任务映射

- `SERVER1 - Layered Persistent Config For Hubruntime`

## 关键设计决策与权衡

- 不修改 Core `IConfig`，由具体实现 `SetPersistent` 提供可选持久化能力
- 将 `Set` 保持为运行期覆盖，避免破坏 auth/trusted 等现有启动期注入路径
- 对持久化层写入使用临时文件 + rename，降低部分写入导致的损坏风险
- 延续现有 `parent target => parent enable` 语义，避免 `config_set(parent.addr)` 重启后静默不生效

## 需求 / 规范影响检查

- 控制面 requirement 已记录在 `D:\project\MyFlowHub3\docs\requirements\management-node-display-name.md`
- 控制面 spec 已记录在 `D:\project\MyFlowHub3\docs\specs\management-config-layering.md`
- 本仓 repo-local `requirements/specs` 无新增长期真相；变更已由控制面 requirement/spec 承载
- lessons 无新增
- 需要更新 `docs/change/README.md` 索引

## 测试与验证方式 / 结果

基础验证：

```powershell
$env:GOWORK='off'
go test ./hubruntime ./cmd/hub_server -count=1
```

结果：通过。

联编验证（使用临时 `go.work` 让当前 Server worktree 对齐本地 Core / Proto）：

```powershell
@'
go 1.25.0

use (
	.
	../MyFlowHub-Proto-feat-management-node-display-name
	../../repo/MyFlowHub-Core
)
'@ | Set-Content go.work
$code = 0
try {
	go test ./... -count=1
	if ($LASTEXITCODE -ne 0) { $code = $LASTEXITCODE }
} finally {
	Remove-Item go.work -ErrorAction SilentlyContinue
	Remove-Item go.work.sum -ErrorAction SilentlyContinue
}
exit $code
```

结果：通过。

## 潜在影响与回滚方案

### 潜在影响

- `config_get` 现在返回 effective value，而不是原始持久化层值
- 显式 env / flag override 会继续压住持久化层，这是预期行为

### 回滚

- 回退 `hubruntime/options.go`、`hubruntime/runtime.go`、`hubruntime/layered_config.go`、`cmd/hub_server/main.go` 与相关测试

## 子 Agent 执行轨迹

- `SERVER1` -> `Main Agent` -> `D:\project\MyFlowHub3\worktrees\MyFlowHub-Server-feat-management-node-display-name`
  - 文件：`hubruntime/options.go`、`hubruntime/layered_config.go`、`hubruntime/layered_config_test.go`、`hubruntime/runtime.go`、`cmd/hub_server/main.go`
  - 验收：分层配置测试通过，Server 全量测试在临时 workspace 下通过
