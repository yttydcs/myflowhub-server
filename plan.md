# Plan - Flow 迁移到 subproto/flow（PR2-Flow）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-subproto-orchestration`（分支名避免包含 `flow`，遵守命名规则）
- Worktree：`d:\project\MyFlowHub3\worktrees\pr2-server-orchestration\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`
- 约束：commit 信息使用中文（前缀如 `refactor:` 可英文）

## 当前状态
- Flow 子协议实现位于 `internal/handler/flow/`，只能在本仓库内部引用。
- `modules/hub.go` 默认装配直接 import `internal/handler/flow`。
- Flow 目录内包含单测 `graph_test.go`（当前在 `internal/handler/flow` 包内）。
- Flow 的协议常量/类型当前由 `internal/handler/flow/types.go` 间接引用 Server 的兼容壳 `protocol/flow`；该兼容壳已委托到 `MyFlowHub-Proto`（wire 不变）。
- Flow handler 实现了 `BindServer(core.IServer)`，并由 `modules.BindServerHooks` 在启动期完成绑定（已在 `modules` 地基中固化）。

> 环境备注（不进 git）：本仓库 `go.mod` 使用 `replace ../MyFlowHub-Core`、`../MyFlowHub-Proto`。  
> 在本 worktree 布局下，需要在 `d:\project\MyFlowHub3\worktrees\pr2-server-orchestration\` 下存在同名目录。  
> 当前通过 Junction 指向 `d:\project\MyFlowHub3\repo\MyFlowHub-Core` / `d:\project\MyFlowHub3\repo\MyFlowHub-Proto` 满足。

## 1) 需求分析
### 目标
1) 将 Flow handler 从 `internal/handler/flow` 迁移到 `subproto/flow`（公开可装配的子协议模块），为后续拆库/裁切做准备。
2) 更新 `modules` 与引用点使用新 import path，移除对 `internal/handler/flow` 的直接依赖。
3) 保持行为与 wire 不变：不调整路由规则/权限语义/action 名称与 payload 结构，不引入新的调度语义。

### 范围
#### 必须（本 PR）
- 新增 `subproto/flow/`，承载 `Handler` 的全部实现（由 `internal/handler/flow` 迁移，包含 `graph_test.go`）。
- `subproto/flow/types.go` 直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/flow`（减少对 Server 兼容壳的耦合；wire 不变）。
- `modules/hub.go` 默认集合改用 `subproto/flow.NewHandlerWithConfig`。
- 清理：删除 `internal/handler/flow` 目录，确保仓库内不再引用该路径。
- 回归：`go test ./... -count=1 -p 1` 通过（Windows）。

#### 可选（本 PR，如不增加风险）
- 若存在文档对实现路径有明确引用，则同步更新（以保持文档与代码一致）。

#### 不做（本 PR）
- 修改 `flow.set/run/status/list/get` 的业务语义、落盘目录与文件格式、权限点与错误码等行为。
- 调整 `BindServer` 绑定时机/方式或引入新的启动顺序依赖。
- 将 flow 的发送/响应构造统一迁移到 `subproto/kit`（避免引入行为差异）。
- Linux 构建验收。

### 使用场景
- hub_server 启动时由 `modules.DefaultHub` 装配并注册 SubProto=6 的 handler。
- 运行期处理 `set/run/status/list/get` 等动作（本 PR 不改行为，仅迁移位置）。

### 功能需求（保持既有约定）
- action、payload、权限点、逐级授权/裁决、落盘与调度行为保持当前实现不变。
- 启动期 `BindServerHooks` 仍能正确调用 flow handler 的 `BindServer`（不改变启动顺序与依赖关系）。

### 非功能需求
- 性能：仅包路径迁移与 import 调整，不引入热路径额外开销；避免额外 I/O、重复计算。
- 可维护性：变更最小化、可回滚、文档与代码一致。

### 输入输出
- 输入：`OnReceive(ctx, conn, hdr, payload)`（payload 为 flow JSON envelope）。
- 输出：通过 `srv.Send` 或 `conn.SendWithHeader` 发送响应帧（Major/SubProto/Source/Target 规则保持既有实现）。

### 边界异常
- 非法 JSON / unknown action：告警/调试日志并丢弃（保持当前处理方式）。
- 缺少 server context：走直连发送分支（保持）。

### 验收标准
- `modules/hub.go` 不再 import `github.com/yttydcs/myflowhub-server/internal/handler/flow`。
- `rg "github.com/yttydcs/myflowhub-server/internal/handler/flow" ./` 在仓库内无命中（历史归档 `docs/change/*` 可能仍包含 `internal/handler/flow` 文字描述，不作为本 PR 验收对象）。
- `go test ./... -count=1 -p 1` 通过（Windows）。

### 风险
- 漏改 import 导致编译失败（`go test` 可覆盖）。
- 迁移时误改 handler 逻辑导致行为差异（本 PR 坚持最小迁移，避免额外重构）。
- 迁移目录包含测试文件，若包名/依赖关系处理不当会导致测试不再被执行（回归测试可覆盖）。

## 2) 架构设计（分析）
### 总体方案（含选型理由 / 备选对比）
- 方案 A（采用）：将 `internal/handler/flow` 通过 `git mv` 迁移到 `subproto/flow`，并将装配层引用切换到新路径。
  - 优点：最小 diff、行为稳定、符合“子协议模块可装配/可裁切”的方向。
  - 缺点：仍在 Server 仓库内；后续若要独立为单独库，再做下一轮拆分。
- 方案 B（不选）：保留 internal 实现，在 `subproto/flow` 做 wrapper。
  - 缺点：增加维护成本与间接层，不利于后续拆库。

### 模块职责
- `subproto/flow`：Flow 子协议处理（SubProto=6），包含 action 分发表、权限校验、落盘与调度等逻辑（本 PR 不改逻辑，仅迁移位置）。
- `modules`：装配入口，负责创建 handler 并注册到 dispatcher；并在启动期通过 `BindServerHooks` 完成 `BindServer` 调用。
- `protocol/flow`：兼容壳（保留旧 import path），本 PR 内不再被 `subproto/flow` 使用，但可继续保留给历史代码。

### 数据 / 调用流
1) `modules.DefaultHub` 构造 `flow.NewHandlerWithConfig(cfg, log)` 放入 `Set.Handlers`
2) `modules.RegisterAll` -> `dispatcher.RegisterHandler(handler)`
3) dispatcher 按 `SubProto()==6` 分发到 handler，handler 内部按 `msg.Action` 找到 action entry 并处理
4) 启动期 `modules.BindServerHooks` 若发现 handler 实现 `BindServer(core.IServer)` 则调用绑定（flow 依赖该绑定以访问 server 能力/调度组件）

### 接口草案
- 对外构造：
  - `NewHandler(log *slog.Logger) *Handler`
  - `NewHandlerWithConfig(cfg core.IConfig, log *slog.Logger) *Handler`
- 关键方法：
  - `AcceptCmd() bool`
  - `SubProto() uint8`
  - `BindServer(core.IServer)`

### 错误与安全
- 不改变既有权限模型与逐级裁决规则（详见 `docs/6-flow.md`）。
- 维持“不引入新权限点/不改默认路由语义”的安全默认。

### 性能与测试策略
- 性能：仅包路径迁移与 import 调整，无额外热路径开销；不引入额外锁竞争与重复计算。
- 测试：
  - `subproto/flow/graph_test.go`（随迁移保留）
  - 全量回归：`$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`

### 可扩展性设计点
- `subproto/<name>` 目录作为子协议模块统一落点，后续可对齐其它子协议迁移，并逐步拆分为独立 module/library。

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（范围与路径明确，且本 PR 坚持最小迁移，wire/行为不变）。

### FL0 - 归档旧 plan.md 并准备本 workflow 文档
- 目标：保留历史 plan 的可审计性，不影响当前 workflow。
- 涉及文件：
  - `plan_archive_2026-02-16_exec-subproto.md`
  - `plan.md`
- 验收条件：新 `plan.md` 仅描述本次 flow 迁移。
- 回滚点：revert 文档提交。

### FL1 - 迁移 flow 到 subproto/flow
- 目标：`Handler` 及其实现从 `internal/handler/flow` 迁移到 `subproto/flow`。
- 涉及模块/文件（预期）：
  - `internal/handler/flow/*` → `subproto/flow/*`
- 验收条件：
  - 包名保持 `flow`，对外构造函数签名不变（如 `NewHandlerWithConfig`）。
  - `SubProto()==6`、`AcceptCmd()`、`BindServer()` 等关键声明保持不变。
- 测试点：`go test ./...`。
- 回滚点：revert 本迁移提交。

### FL2 - subproto 直连 MyFlowHub-Proto 协议包
- 目标：`subproto/flow` 直接 import `github.com/yttydcs/myflowhub-proto/protocol/flow`。
- 涉及文件：
  - `subproto/flow/types.go`
- 验收条件：仅 import 路径变化，常量/类型一致，wire 不变。
- 回滚点：revert。

### FL3 - modules 装配切换到新路径
- 目标：`modules/hub.go` 使用 `subproto/flow`。
- 验收条件：默认装配集合仍启用 flow。
- 回滚点：revert。

### FL4 - 清理 internal/handler/flow 残留
- 目标：移除旧目录，确保仓库内无引用（历史归档除外）。
- 验收条件：
  - `rg "github.com/yttydcs/myflowhub-server/internal/handler/flow"` 无命中（排除 `plan_archive_*`；`docs/change/*` 的历史描述不参与验收）。
- 回滚点：revert。

### FL5 - 全量回归
- 目标：确保迁移不破坏编译/测试。
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过（Windows）。

### FL6 - Code Review + 归档变更
- 目标：按模板完成审查与归档。
- 涉及文件：
  - `docs/change/YYYY-MM-DD_flow-subproto.md`
- 验收条件：归档包含任务映射、关键决策、测试结果与回滚方案。

## 注意事项
- 禁止计划外改动：若迁移过程中发现必须同步修改其它子协议或引入 wire 变更，必须回到 3.1 更新计划并重新确认。

## 执行记录
- 2026-02-16：完成 FL1-FL5；回归 `go test ./... -count=1 -p 1` 通过。
