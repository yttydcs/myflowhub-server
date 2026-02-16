# Plan - File 迁移到 subproto/file（PR3-File）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-subproto-file`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr3-server-file-subproto\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`
- 约束：commit 信息使用中文（前缀如 `refactor:` 可英文）

## 当前状态
- File 子协议实现位于 `internal/handler/file/`，只能在本仓库内部引用。
- `modules/hub.go` 默认装配仍直接 import `internal/handler/file`。
- File 协议常量/类型当前由 `internal/handler/file/types.go` 通过 Server 兼容壳 `protocol/file` 间接引用；该兼容壳已委托到 `MyFlowHub-Proto`（wire 不变）。
- File handler 目前未包含 `_test.go` 单测文件；依赖全量回归覆盖编译与包级集成。

> 环境备注（不进 git）：本仓库 `go.mod` 使用 `replace ../MyFlowHub-Core`、`../MyFlowHub-Proto`。  
> 在本 worktree 布局下，需要在 `d:\project\MyFlowHub3\worktrees\pr3-server-file-subproto\` 下存在同名目录。  
> 当前通过 Junction 指向 `d:\project\MyFlowHub3\repo\MyFlowHub-Core` / `d:\project\MyFlowHub3\repo\MyFlowHub-Proto` 满足。

## 1) 需求分析
### 目标
1) 将 File handler 从 `internal/handler/file` 迁移到 `subproto/file`（公开可装配的子协议模块），为后续拆库/裁切做准备。
2) 更新 `modules` 与引用点使用新 import path，移除对 `internal/handler/file` 的直接依赖。
3) 保持行为与 wire 不变：不调整 `KindCtrl/KindData/KindAck` 二进制帧格式、不调整 action/op 名称与 payload 结构、不引入新的传输语义。

### 范围
#### 必须（本 PR）
- 新增 `subproto/file/`，承载 File 子协议实现（由 `internal/handler/file` 迁移而来）。
- `subproto/file/types.go` 直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/file`（减少对 Server 兼容壳的耦合；wire 不变）。
- `modules/hub.go` 默认集合改用 `subproto/file.NewHandlerWithConfig`。
- 清理：删除 `internal/handler/file` 目录，确保仓库内不再引用该路径。
- 回归：`go test ./... -count=1 -p 1` 通过（Windows）。

#### 可选（本 PR，如不增加风险）
- 若存在文档对实现路径有明确引用，则同步更新（以保持文档与代码一致）。

#### 不做（本 PR）
- 修改 file 的落盘目录规则、覆盖策略、分片/ACK/超时/并发策略等业务语义。
- 调整权限点（`file.read` / `file.write`）与授权裁决行为。
- Linux 构建验收。

### 使用场景
- hub_server 启动时由 `modules.DefaultHub` 装配并注册 File 的 handler。
- 运行期处理 file 的控制帧（read/write 及其 resp）与数据帧（DATA/ACK），完成节点间文件传输（本 PR 不改行为，仅迁移位置与依赖）。

### 功能需求（保持既有约定）
- `payload[0]` 作为帧类型：`KindCtrl/KindData/KindAck`（保持）。
- 控制帧：`ActionRead/ActionWrite` 及其响应（保持）。
- 数据/ACK 帧：沿用 `binHeaderV1`（session/offset/fin 等语义保持）。
- 权限：`file.read` / `file.write` 保持既有校验路径与默认拒绝策略。

### 非功能需求
- 性能：仅包路径迁移与 import 调整，不引入热路径额外开销；避免额外 I/O、重复计算与锁竞争。
- 可维护性：变更最小化、可回滚、文档与代码一致。

### 输入输出
- 输入：`OnReceive(ctx, conn, hdr, payload)`；其中：
  - `payload[0]==KindCtrl` 时，`payload[1:]` 为 JSON message；
  - `payload[0]==KindData/KindAck` 时，`payload[1:]` 为二进制 header + body（KindData）。
- 输出：通过 `srv.Send` / `conn.SendWithHeader` 等现有路径发送 ctrl resp / data / ack（路由规则保持既有实现）。

### 边界异常
- 非法 payload / 非法 JSON：丢弃或返回 4xx（保持当前行为）。
- `ServerFromContext(ctx)==nil`：无法转发时直接返回（保持当前行为）。
- 路径清洗失败（非法 dir/name、目录穿越）：按现有错误码返回（保持）。

### 验收标准
- `modules/hub.go` 不再 import `github.com/yttydcs/myflowhub-server/internal/handler/file`。
- `rg "github.com/yttydcs/myflowhub-server/internal/handler/file" ./` 在仓库内无命中（历史归档 `docs/change/*` 可能仍包含 `internal/handler/file` 文字描述，不作为本 PR 验收对象）。
- `go test ./... -count=1 -p 1` 通过（Windows）。

### 风险
- 漏改 import 导致编译失败（`go test` 可覆盖）。
- 迁移时误改 handler 逻辑导致行为差异（本 PR 坚持“最小迁移”，避免额外重构）。

## 2) 架构设计（分析）
### 总体方案（含选型理由 / 备选对比）
- 方案 A（采用）：将 `internal/handler/file` 通过 `git mv` 迁移到 `subproto/file`，并将装配层引用切换到新路径。
  - 优点：最小 diff、行为稳定、符合“子协议模块可装配/可裁切”的目标架构方向。
  - 缺点：仍在 Server 仓库内；后续若要独立为单独库，再做下一轮拆分。
- 方案 B（不选）：保留 internal 实现，在 `subproto/file` 做 wrapper。
  - 缺点：增加维护成本与间接层，不利于后续拆库与裁切。

### 模块职责
- `subproto/file`：File 子协议处理（节点间文件传输），包含 ctrl/data/ack 处理、会话管理、落盘与转发逻辑（本 PR 不改逻辑，仅迁移位置）。
- `modules`：装配入口，负责创建 handler 并注册到 dispatcher。
- `protocol/file`：兼容壳（保留旧 import path）；本 PR 内 `subproto/file` 将直接依赖 Proto，但该兼容壳可继续保留给历史代码。

### 数据 / 调用流
1) `modules.DefaultHub` 构造 `file.NewHandlerWithConfig(cfg, log)` 放入 `Set.Handlers`
2) `modules.RegisterAll` -> `dispatcher.RegisterHandler(handler)`
3) dispatcher 按 `SubProto()` 分发到 handler
4) handler 内部按 `payload[0]` 分流：ctrl/data/ack，并通过 `ServerFromContext(ctx)` 做逐跳路由/转发（保持既有实现）

### 接口草案
- 对外构造：
  - `NewHandler(log *slog.Logger) *Handler`
  - `NewHandlerWithConfig(cfg core.IConfig, log *slog.Logger) *Handler`
- 关键方法：
  - `SubProto() uint8`
  - `Init() bool`
  - `OnReceive(ctx, conn, hdr, payload)`

### 错误与安全
- 维持 `file.read` / `file.write` 的权限校验（不新增权限点）。
- 维持 dir/name 清洗与目录穿越防护策略（不改变错误处理分支）。

### 性能与测试策略
- 性能：仅包路径迁移与 import 调整，无额外热路径开销。
- 测试：
  - 全量回归：`$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`

### 可扩展性设计点
- `subproto/<name>` 目录作为子协议模块统一落点；后续可继续迁移剩余 internal handler，并逐步演进为可拆分库。

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（范围与路径明确；本 PR 坚持最小迁移，wire/行为不变）。

### FI0 - 归档旧 plan.md 并准备本 workflow 文档
- 目标：保留历史 plan 的可审计性，不影响当前 workflow。
- 涉及文件：
  - `plan_archive_2026-02-16_flow-subproto.md`
  - `plan.md`
- 验收条件：新 `plan.md` 仅描述本次 file 迁移。
- 回滚点：revert 文档提交。

### FI1 - 迁移 file 到 subproto/file
- 目标：`Handler` 及其实现从 `internal/handler/file` 迁移到 `subproto/file`。
- 涉及模块/文件（预期）：
  - `internal/handler/file/*` → `subproto/file/*`
- 验收条件：
  - 包名保持 `file`，对外构造函数签名不变（如 `NewHandlerWithConfig`）。
  - `SubProto()==SubProtoFile`、`OnReceive` 分流行为保持不变。
- 测试点：`go test ./...`。
- 回滚点：revert 本迁移提交。

### FI2 - subproto 直连 MyFlowHub-Proto 协议包
- 目标：`subproto/file` 直接 import `github.com/yttydcs/myflowhub-proto/protocol/file`。
- 涉及文件：
  - `subproto/file/types.go`
- 验收条件：仅 import 路径变化，常量/类型一致，wire 不变。
- 回滚点：revert。

### FI3 - modules 装配切换到新路径
- 目标：`modules/hub.go` 使用 `subproto/file`。
- 验收条件：默认装配集合仍启用 file。
- 回滚点：revert。

### FI4 - 清理 internal/handler/file 残留
- 目标：移除旧目录，确保仓库内无引用（历史归档除外）。
- 验收条件：
  - `rg "github.com/yttydcs/myflowhub-server/internal/handler/file"` 无命中（排除 `plan_archive_*`；`docs/change/*` 的历史描述不参与验收）。
- 回滚点：revert。

### FI5 - 全量回归
- 目标：确保迁移不破坏编译/测试。
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过（Windows）。

### FI6 - Code Review + 归档变更
- 目标：按模板完成审查与归档。
- 涉及文件：
  - `docs/change/YYYY-MM-DD_file-subproto.md`
- 验收条件：归档包含任务映射、关键决策、测试结果与回滚方案。

## 注意事项
- 禁止计划外改动：若迁移过程中发现必须同步修改其它子协议或引入 wire 变更，必须回到 3.1 更新计划并重新确认。

## 执行记录
- 2026-02-16：创建本 workflow worktree 与计划文档（待确认后进入 3.2）。
