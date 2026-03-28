# Plan - Stream Release And Integration

## Workflow 信息
- 控制面 Repo：`MyFlowHub-Server`
- 控制面 Branch：`chore/stream-subproto-design`
- Base：`main`
- 控制面 Worktree：`D:\project\MyFlowHub3\worktrees\server-stream-subproto-design`
- 当前 Stage：`4`
- 关联 Repo：
  - `MyFlowHub-Proto`
    - Branch：`feat/proto-stream-subproto`
    - Worktree：`D:\project\MyFlowHub3\worktrees\proto-stream-subproto`
    - Plan：`D:\project\MyFlowHub3\worktrees\proto-stream-subproto\plan.md`
  - `MyFlowHub-SubProto`
    - Branch：`feat/subproto-stream-subproto`
    - Worktree：`D:\project\MyFlowHub3\worktrees\subproto-stream-subproto`
    - Plan：`D:\project\MyFlowHub3\worktrees\subproto-stream-subproto\plan.md`

## 当前状态
- 上一轮 `stream` 文档、Proto wire、SubProto module、Server compat/defaultset 已在各自 worktree 落地，并已归档：
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\docs\change\2026-03-28_stream-subproto-design.md`
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\docs\change\2026-03-28_stream-server-integration.md`
  - `D:\project\MyFlowHub3\worktrees\proto-stream-subproto\docs\change\2026-03-28_stream-wire.md`
  - `D:\project\MyFlowHub3\worktrees\subproto-stream-subproto\docs\change\2026-03-28_stream-module.md`
- 远端当前状态已核对：
  - Proto 最新 tag：`v0.1.3`
  - SubProto 尚无 `stream/v0.1.0`
  - Server 最新 tag：`v0.0.11`
- 当前 `Server/go.mod` 仍未依赖 `myflowhub-subproto/stream`，且 `myflowhub-proto` 仍停留 `v0.1.3`
- 上一轮只证明了本地 workspace 联调成立；本轮目标是形成真实 release chain，并补一条多节点最小集成验证
- 在真实依赖链验证中发现额外阻塞：
  - `hubruntime/options.go` 直接引用 `coreconfig.DefaultAuthRolePerms`
  - 已发布 `myflowhub-core v0.4.8` 不导出该符号
  - 补充核对结果：
    - `myflowhub-core v0.4.8` 的 `config.NewMap(...)` 对 `auth.role_perms` 默认值也是空字符串
    - 因此不能把默认角色权限继续委托给 Core 注入
  - 结论：本轮需要在 Server 内做最小兼容修复，去掉对该符号的编译期直接依赖，同时由 Server 本地保留默认角色权限常量，保持既有运行时语义

## Stage 1 - 需求分析

### 目标
- 把 `stream` 从“仅 worktree 联调可用”推进到“下游 `GOWORK=off` 可拉取、可编译、可执行最小多节点集成测试”。
- 在 Server 侧完成版本对齐、最小多节点 `stream` 集成用例，以及 `v0.0.12` 发布准备。

### 范围
- 必须：
  - 将 `Server` 依赖对齐到：
    - `github.com/yttydcs/myflowhub-proto v0.1.4`
    - `github.com/yttydcs/myflowhub-subproto/stream v0.1.0`
  - 消除 `Server` 对 `coreconfig.DefaultAuthRolePerms` 的编译期直接依赖，使 `GOWORK=off` 下可继续使用已发布 `myflowhub-core v0.4.8`
  - 保持 `Server` 现有 auth 默认角色层级不回退，即不因为 `core v0.4.8` 的空默认值而丢失 `superadmin:*`
  - 保持已有 compat wrapper / defaultset 装配不回退
  - 新增一条多节点最小集成测试，覆盖：
    - root / hub 拓扑
    - source 声明
    - consumer 声明
    - 跨节点 `list_sources`
    - 控制侧 `connect / disconnect`
  - 在 `GOWORK=off` 下执行 Server 验证
  - 形成本轮 change 归档，并准备发布 `v0.0.12`
- 可选：
  - 若不增加范围，可对 `go list -m` 增加版本核对
- 不做：
  - 不再改 `stream` 长期 requirements/specs
  - 不新增 plan 外的 `stream` 业务语义
  - 不把本轮多节点集成扩展为音视频 payload 完整端到端回放

### 使用场景
- 下游用户直接获取 `hub_server` 默认集合中的 `stream` 能力，不需要本地 `go.work`
- root 控制侧把 hub 上的 `source` 与 root 本地 `consumer` 连接，再断开
- CI 或本地审计环境在 `GOWORK=off` 下复现本轮 Server 集成结果

### 功能需求
- `go.mod` / `go.sum` 必须对齐 Proto 与 SubProto 新版本
- 多节点集成测试必须真实经过 root <-> hub 父子链，而不是单节点 mock
- 测试必须验证可观察的控制面行为：
  - `announce`
  - `announce_consumer`
  - `list_sources`
  - `connect`
  - `disconnect`
  - 重复 `disconnect` 返回 `404`
- Server 发布版本必须是新的 patch：`v0.0.12`

### 非功能需求
- 不扩大运行时改动面，优先做依赖对齐和最小验证
- 不为本轮额外引入新的 Core 发版链
- 验证必须使用 `GOWORK=off`
- tag 一旦 push 不得改写

### 输入输出
- 输入：
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\go.mod`
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\go.sum`
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\tests\integration_root_hub_ping_test.go`
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\modules\defaultset\*.go`
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\docs\requirements\stream.md`
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\docs\specs\stream.md`
- 输出：
  - `GOWORK=off` 可消费的 Server 依赖版本链
  - 新的 `stream` 多节点集成测试
  - 远端 tag：`v0.0.12`
  - 本轮 change 归档

### 边界异常
- 若 Proto / SubProto 新 tag 还不可见，Server 依赖升级会失败
- 若 `hubruntime` 仍直接引用未发布 Core 符号，`GOWORK=off` 编译会继续失败
- 若多节点测试只验证本地 direct path，不足以证明 root / hub 集成
- 若重复 `disconnect` 仍返回成功，说明 delivery route 清理不完整

### 验收标准
- `GOWORK=off go test ./modules/... ./tests/... -count=1 -p 1` 通过
- `GOWORK=off go test ./tests -run TestStreamRootHubConnectDisconnect -count=1` 通过
- `git ls-remote --tags origin refs/tags/v0.0.12` 能看到远端 tag

### 风险
- 若上游 tag 未完全传播，Server 侧版本解析会短暂失败
- 若兼容修复未补回 Server 本地默认角色权限，会影响 auth 默认角色权限与 `stream` 权限验证
- `release-hub-server.yml` 对 `v*` tag 自动发版，发布动作必须在验证后执行
- 若本轮把集成测试范围拉到 DATA/ACK 全链路观测，会显著增加实现面；本轮仅做最小可观察控制面集成

## Stage 2 - 架构设计

### 总体方案
- 方案 A：只改 `go.mod`，不补多节点集成测试
  - 不选：无法证明 `stream` 在 root / hub 场景下可用
- 方案 B：继续使用临时 `go.work` 做联调，不发布下游 patch
  - 不选：仍不满足真实消费与发版要求
- 方案 C：按发布链依赖顺序完成版本对齐，并新增最小多节点集成测试
  - 采用：既能收口 `GOWORK=off` 可复现性，也能控制变更面

### 模块职责
- `go.mod` / `go.sum`
  - 对齐上游版本链
- `hubruntime`
  - 去除对未发布 Core 默认常量的编译期耦合
  - 在 Server 侧本地保留默认角色权限常量，避免依赖旧 Core 的空默认值
- `tests/`
  - 新增 `stream` 多节点集成用例
  - 复用现有 root / hub 启动、连接 nodeID 绑定和帧发送 helper
- `modules/defaultset`
  - 维持上一轮已接入的 `stream` 装配，不再扩大改动

### 数据 / 调用流
1. Proto 发布 `v0.1.4`
2. SubProto 发布 `stream/v0.1.0`
3. Server 升级 `go.mod`
4. `hubruntime` 去掉对 `coreconfig.DefaultAuthRolePerms` 的编译期直接引用
5. `hubruntime` 使用 Server 本地默认角色权限常量维持 `auth.role_perms` 既有默认语义
6. layered config 继续透传该默认值或显式 override
7. 测试启动 root 与 hub 两节点
8. hub 本地声明 `source`
9. root 本地声明 `consumer`
10. root 控制侧执行 `list_sources -> connect -> disconnect`
11. 重复 `disconnect` 返回 `404`，证明 delivery route 已清理
12. Server 发布 `v0.0.12`

### 接口草案
- Server 依赖版本：
  - `github.com/yttydcs/myflowhub-proto v0.1.4`
  - `github.com/yttydcs/myflowhub-subproto/stream v0.1.0`
- 新增测试：
  - `tests/integration_stream_root_hub_connect_test.go`
- 发布版本：
  - `v0.0.12`

### 错误与安全
- 若上游 tag 未就绪，立即停止 Server 侧版本对齐
- 兼容修复只能去掉编译期常量耦合，不能把 existing auth 默认层级退化成空权限
- 不对 `stream` action / payload 做计划外修改
- 不把旧 `go.work` 结果当作本轮发布依据

### 性能与测试策略
- 依赖版本核对：
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-proto`
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/stream`
- `GOWORK=off go test ./hubruntime -count=1`
- 仓库验证：
  - `GOWORK=off go test ./modules/... ./tests/... -count=1 -p 1`
- 定向集成：
  - `GOWORK=off go test ./tests -run TestStreamRootHubConnectDisconnect -count=1`

### 可扩展性设计点
- 本轮集成测试先锁定 root / hub 控制面多节点链路；若后续要补 DATA/ACK 端到端观测，再以独立 workflow 增量推进

## Stage 3.1 - 计划
- Requirements impact：`none`
- Specs impact：`none`
- Related requirements：
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\docs\requirements\stream.md`
- Related specs：
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\docs\specs\stream.md`
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\docs\specs\protocol_map.md`
  - `D:\project\MyFlowHub3\worktrees\proto-stream-subproto\docs\protocol_map.md`
- Related lessons：
  - `D:\project\MyFlowHub3\docs\lessons\cross-repo-semver-release.md`

### 执行清单
- [x] `STRM-REL-1` 对齐 Server 依赖到 Proto `v0.1.4` 和 SubProto `stream v0.1.0`
- [x] `STRM-COMP-1` 修复 `hubruntime` 对已发布 Core 版本的编译兼容
- [x] `STRM-IT-1` 新增 root / hub `stream` 多节点集成测试
- [x] `STRM-VAL-1` 在 `GOWORK=off` 下完成版本核对与仓库验证
- [x] `STRM-REL-2` 创建并推送 `v0.0.12`
- [x] `STRM-DOC-1` 归档本轮 Server 发布与集成结果

### 任务明细

#### STRM-REL-1
- Owner：主 Agent
- Worktree：`D:\project\MyFlowHub3\worktrees\server-stream-subproto-design`
- Files：
  - `go.mod`
  - `go.sum`
- Goal：
  - 让 Server 在 `GOWORK=off` 下直接消费已发布 `stream` 依赖
- Acceptance：
  - `go.mod` 明确依赖 `myflowhub-proto v0.1.4`
  - `go.mod` 明确依赖 `myflowhub-subproto/stream v0.1.0`
  - `go.sum` 同步更新
- Tests：
  - `GOWORK=off go list -m github.com/yttydcs/myflowhub-proto`
  - `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/stream`
- Rollback：
  - 回退 `go.mod` / `go.sum`

#### STRM-COMP-1
- Owner：主 Agent
- Worktree：`D:\project\MyFlowHub3\worktrees\server-stream-subproto-design`
- Files：
  - `hubruntime/options.go`
  - `hubruntime/options_test.go`
- Goal：
  - 让 Server 在 `myflowhub-core v0.4.8` 下消除编译期符号依赖，同时保持 effective auth 默认角色权限不变
- Acceptance：
  - `hubruntime` 不再引用 `coreconfig.DefaultAuthRolePerms`
  - `DefaultOptions()` 仍提供原有默认 `auth.role_perms` 层级
  - 显式 override 仍能覆盖或清空 `auth.role_perms`
  - 相关测试更新后通过
- Tests：
  - `GOWORK=off go test ./hubruntime -count=1`
- Rollback：
  - 回退 `hubruntime/options.go` 与相关测试

#### STRM-IT-1
- Owner：主 Agent
- Worktree：`D:\project\MyFlowHub3\worktrees\server-stream-subproto-design`
- Files：
  - `tests/integration_stream_root_hub_connect_test.go`
- Goal：
  - 提供一条最小但真实的 root / hub 多节点 `stream` 集成验证
- Acceptance：
  - 测试覆盖 `announce -> announce_consumer -> list_sources -> connect -> disconnect`
  - 第二次 `disconnect` 返回 `404`
- Tests：
  - `GOWORK=off go test ./tests -run TestStreamRootHubConnectDisconnect -count=1`
- Rollback：
  - 回退新增测试文件

#### STRM-VAL-1
- Owner：主 Agent
- Worktree：`D:\project\MyFlowHub3\worktrees\server-stream-subproto-design`
- Files：
  - `none`
- Goal：
  - 以真实依赖解析模式完成 Server 侧验证
- Acceptance：
  - 版本核对通过
  - 模块和测试回归通过
- Tests：
  - `GOWORK=off go list -m github.com/yttydcs/myflowhub-proto`
  - `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/stream`
  - `GOWORK=off go test ./hubruntime -count=1`
  - `GOWORK=off go test ./modules/... ./tests/... -count=1 -p 1`
  - `GOWORK=off go test ./tests -run TestStreamRootHubConnectDisconnect -count=1`
- Rollback：
  - 若验证失败，先修正依赖或测试，再继续发版

#### STRM-REL-2
- Owner：主 Agent
- Worktree：`D:\project\MyFlowHub3\worktrees\server-stream-subproto-design`
- Files：
  - `none (git commit / tag / push)`
- Goal：
  - 创建并推送 `v0.0.12`
- Acceptance：
  - `git ls-remote --tags origin refs/tags/v0.0.12` 有结果
- Tests：
  - `git ls-remote --tags origin refs/tags/v0.0.12`
- Rollback：
  - 不删除 tag；若发布点有问题，追加更高 patch

#### STRM-DOC-1
- Owner：主 Agent
- Worktree：`D:\project\MyFlowHub3\worktrees\server-stream-subproto-design`
- Files：
  - `docs/change/2026-03-28_stream-server-release.md`
  - `docs/change/README.md`
- Goal：
  - 归档本轮依赖链收口、多节点集成测试和 Server patch 发布结果
- Acceptance：
  - change 文档记录版本链、测试结果、发布动作和回滚策略
- Tests：
  - 人工核对文档与实际 tag / 验证输出一致
- Rollback：
  - 回退本轮文档改动

### 依赖 / 风险 / 备注
- 发布顺序固定：
  - `Proto v0.1.4`
  - `SubProto stream/v0.1.0`
  - `Server v0.0.12`
- 兼容策略固定：
  - 不新增 Core patch release
  - `hubruntime` 使用 Server 本地默认角色权限常量
  - 不依赖 `core v0.4.8` 的空 `auth.role_perms` 默认值
- `release-hub-server.yml` 已确认对 `v*` tag 自动发版，因此打 tag 前必须完成全部 `GOWORK=off` 验证
- 本轮多节点集成测试只锁定最小控制面链路；DATA/ACK 端到端观测继续依赖 SubProto 模块级测试

阻塞：否
进入 3.2
