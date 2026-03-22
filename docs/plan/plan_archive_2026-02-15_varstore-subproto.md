# Plan - VarStore 迁移到 subproto/varstore（PR2-Varstore）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-subproto-varstore`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr2-server-varstore\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`
- 约束：commit 信息使用中文（前缀如 `refactor:` 可英文）

## 当前状态
- VarStore 子协议实现位于 `internal/handler/varstore/`，仅能在本仓库内部引用。
- `modules/hub.go` 与部分测试直接 import `internal/handler/varstore`。
- 目标架构方向：Server 作为“中间层装配”，子协议尽量解耦、可裁切；因此逐步将子协议 handler 从 `internal/handler/*` 迁移到 `subproto/*`。

> 环境备注（不进 git）：本仓库 `go.mod` 使用 `replace ../MyFlowHub-Core`、`../MyFlowHub-Proto`。  
> 在本 worktree 布局下，需要在 `d:\project\MyFlowHub3\worktrees\pr2-server-varstore\` 下存在同名目录。  
> 当前通过 Junction 指向 `d:\project\MyFlowHub3\repo\MyFlowHub-Core` / `repo\MyFlowHub-Proto` 满足。

## 1) 需求分析
### 目标
1) 将 VarStore handler 迁移到 `subproto/varstore`（公开可装配的子协议模块），为后续拆库/裁切做准备。
2) 更新 `modules` 与测试使用新路径，移除对 `internal/handler/varstore` 的直接依赖。
3) 保持行为不变：不改 wire/SubProto/权限语义，不调整 action 名称（assist_* / up_* / notify_* 收敛不在本 PR）。

### 范围
#### 必须（本 PR）
- 新增 `subproto/varstore/`，承载 `VarStoreHandler` 与其全部实现（由 `internal/handler/varstore` 迁移）。
- 将 varstore 协议常量/类型引用改为直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/varstore`（减少对 Server 兼容壳的耦合；wire 不变）。
- `modules/hub.go` 默认集合改用 `subproto/varstore.NewVarStoreHandlerWithConfig`。
- 更新测试 import：
  - `tests/varstore_handler_test.go`
  - `tests/integration_varstore_end_to_end_test.go`
  - `tests/integration_root_hub_ping_test.go`
- 清理：删除 `internal/handler/varstore` 目录，确保仓库内不再引用该路径。
- 回归：`go test ./... -count=1 -p 1` 通过（Windows）。

#### 可选（本 PR，如不增加风险）
- 无（本 PR 坚持最小迁移，避免额外重构）。

#### 不做（本 PR）
- 重写 varstore 的发送/响应构造逻辑到 `subproto/kit`。
- varstore action 命名收敛/协议升级（wire 变更）。
- Linux 构建验收。

### 使用场景
- hub_server 启动时由 `modules.DefaultHub` 装配并注册 SubProto=3 的 handler。
- 运行期处理 varstore 的 set/get/list/revoke/subscribe 及其 assist/up/notify 行为（本 PR 不改行为，仅迁移位置）。

### 输入输出
- 输入：`OnReceive(ctx, conn, hdr, payload)`（payload 为 varstore JSON message）。
- 输出：通过 `srv.Send` 或 `conn.SendWithHeader` 发送 OK/Err/Cmd 响应（保持既有 Major/SubProto/Source/Target 规则）。

### 边界异常
- 非法 JSON / unknown action：仅告警/调试日志，保持当前处理方式。
- owner 不在子树且无 parent：返回 not found（保持）。
- 权限不足：返回对应 code/msg（保持）。

### 验收标准
- `modules/hub.go`、测试不再 import `github.com/yttydcs/myflowhub-server/internal/handler/varstore`。
- `rg \"internal/handler/varstore\" ./` 在仓库内无命中。
- `go test ./... -count=1 -p 1` 通过（Windows）。

### 风险
- 迁移漏改 import 导致编译失败（`go test` 可覆盖）。
- 迁移过程中误改业务逻辑导致行为差异（本 PR 只做目录迁移 + 引用更新，尽量避免）。

## 2) 架构设计（分析）
### 总体方案（含备选对比）
- 方案 A（采用）：将 `internal/handler/varstore` 通过 `git mv` 迁移到 `subproto/varstore`，并将装配层/测试切换到新 import path。
  - 优点：最小 diff、行为稳定、符合“子协议可裁切/可复用”的方向。
  - 缺点：仍在 Server 仓库内，尚未独立成单独 module/library（后续再拆）。
- 方案 B（不选）：保留 internal 实现，仅在 `subproto/varstore` 写一层 wrapper。
  - 缺点：增加维护成本与间接层，不利于后续拆库。

### 模块职责
- `subproto/varstore`：VarStore 子协议处理（SubProto=3），包含 action 分发表、权限校验、订阅/缓存/转发等逻辑。
- `modules`：装配入口，负责创建 handler 并注册到 dispatcher。
- `protocol/varstore`：兼容壳（保留旧 import path），本 PR 内不再被 `subproto/varstore` 使用，但可继续保留给历史代码。

### 数据 / 调用流
1) `modules.DefaultHub` 构造 `varstore.NewVarStoreHandlerWithConfig(cfg, log)` 放入 `Set.Handlers`
2) `modules.RegisterAll` -> `dispatcher.RegisterHandler(handler)`
3) 运行期 dispatcher 按 `SubProto()==3` 分发到 handler，handler 内部按 `msg.Action` 找到 action entry 并处理

### 错误与安全
- 不改变权限模型：仍通过 `permission.SharedConfig(cfg)` + `permission.SourceNodeID(hdr, conn)` 判定。
- 不改变转发边界：仍通过 parent 连接与子树判断决定本地/上行 assist。

### 性能与测试策略
- 性能：仅包路径迁移 + import 调整，无额外热路径开销；避免引入重复 marshal/unmarshal 逻辑变更。
- 测试：
  - 现有单测与集成测试回归（见上文 1) 验收标准）。
  - 执行：`$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`

### 可扩展性设计点
- `subproto/<name>` 目录作为“子协议模块”的统一落点，后续可对齐其它子协议迁移，并逐步拆分为独立 module/library。

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（变更范围明确，且本 PR 坚持最小迁移，wire/行为不变）。

### V0 - 归档旧 plan.md 并准备本 workflow 文档
- 目标：保留历史 plan 的可审计性，不影响当前 workflow。
- 涉及文件：
  - `plan_archive_2026-02-15_default-forward-subproto-forward.md`
  - `plan.md`
- 验收条件：新 `plan.md` 仅描述本次 varstore 迁移。
- 回滚点：revert 文档提交。

### V1 - 迁移 varstore 到 subproto/varstore
- 目标：`VarStoreHandler` 及其实现从 `internal/handler/varstore` 迁移到 `subproto/varstore`。
- 涉及模块/文件（预期）：
  - `internal/handler/varstore/*` → `subproto/varstore/*`
- 验收条件：
  - 包名保持 `varstore`，对外构造函数签名不变（`NewVarStoreHandlerWithConfig` 等）。
  - `SubProto()==3`、`AcceptCmd()==true` 等关键声明保持不变。
- 测试点：`go test ./...`。
- 回滚点：revert 本迁移提交。

### V2 - 降低协议耦合（subproto 直连 MyFlowHub-Proto）
- 目标：`subproto/varstore` 直接 import `github.com/yttydcs/myflowhub-proto/protocol/varstore`。
- 涉及文件：
  - `subproto/varstore/types.go`
- 验收条件：仅 import 路径变化，常量/类型一致，wire 不变。
- 回滚点：revert 本提交。

### V3 - modules 装配切换到新路径
- 目标：`modules/hub.go` 使用 `subproto/varstore`。
- 涉及文件：
  - `modules/hub.go`
- 验收条件：
  - `DefaultHub` 装配不变（仍启用 varstore）。
- 回滚点：revert 本提交。

### V4 - 测试切换到新路径
- 目标：测试 import 统一改为 `subproto/varstore`。
- 涉及文件：
  - `tests/varstore_handler_test.go`
  - `tests/integration_varstore_end_to_end_test.go`
  - `tests/integration_root_hub_ping_test.go`
- 验收条件：测试编译通过并运行通过。
- 回滚点：revert。

### V5 - 清理 internal/handler/varstore 残留
- 目标：移除旧目录，确保仓库内无引用。
- 验收条件：
  - `rg \"internal/handler/varstore\"` 无命中。
- 回滚点：revert。

### V6 - 全量回归
- 目标：确保迁移不破坏编译/测试。
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过（Windows）。

### V7 - 归档变更
- 目标：形成可交接的变更记录。
- 涉及文件：
  - `docs/change/2026-02-15_varstore-subproto.md`
- 验收条件：包含任务映射、设计权衡、测试结果、回滚方案。
- 回滚点：revert。

## 注意事项
- 禁止计划外改动：若迁移过程中发现必须同步修改其它子协议或引入 wire 变更，必须回到 3.1 更新计划并重新确认。

## 执行记录
- 2026-02-15：完成 V1-V6；回归 `go test ./... -count=1 -p 1` 通过。
