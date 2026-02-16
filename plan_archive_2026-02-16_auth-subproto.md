# Plan - Auth 迁移到 subproto/auth（PR4-Auth）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-subproto-auth`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr4-server-auth-subproto\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`
- 约束：commit 信息使用中文（前缀如 `refactor:` 可英文）

## 当前状态
- Auth 子协议实现位于 `internal/handler/auth/`，只能在本仓库内部引用。
- `modules/hub.go` 默认装配仍直接 import `internal/handler/auth`。
- `tests/*` 多处 import `internal/handler/auth`，用于构造并验证 `LoginHandler` 的权限与路由行为。
- `docs/2-auth.md` 明确以 `internal/handler/auth` 为实现路径描述。
- Auth 的协议常量/类型当前由 `internal/handler/auth/types.go` 通过 Server 兼容壳 `protocol/auth` 间接引用；该兼容壳已委托到 `MyFlowHub-Proto`（wire 不变）。

> 环境备注（不进 git）：本仓库 `go.mod` 使用 `replace ../MyFlowHub-Core`、`../MyFlowHub-Proto`。  
> 在本 worktree 布局下，需要在 `d:\project\MyFlowHub3\worktrees\pr4-server-auth-subproto\` 下存在同名目录。  
> 当前通过 Junction 指向 `d:\project\MyFlowHub3\repo\MyFlowHub-Core` / `d:\project\MyFlowHub3\repo\MyFlowHub-Proto` 满足。

## 1) 需求分析
### 目标
1) 将 Auth handler 从 `internal/handler/auth` 迁移到 `subproto/auth`（公开可装配的子协议模块），为后续拆库/裁切做准备。
2) 更新 `modules`、`tests` 与文档使用新 import path，移除对 `internal/handler/auth` 的直接依赖。
3) 保持行为与 wire 不变：不调整 SubProto=2 的动作集合、payload 结构、路由/权威选择、签名与持久化语义。

### 范围
#### 必须（本 PR）
- 新增 `subproto/auth/`，承载 Auth 子协议实现（由 `internal/handler/auth` 迁移而来）。
- `subproto/auth/types.go` 直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/auth`（减少对 Server 兼容壳的耦合；wire 不变）。
- `modules/hub.go` 默认集合改用 `subproto/auth.NewLoginHandlerWithConfig`。
- `tests/*` 引用点切换为 `github.com/yttydcs/myflowhub-server/subproto/auth`。
- 文档 `docs/2-auth.md` 更新为 `subproto/auth` 路径描述（保持内容语义不变）。
- 清理：删除 `internal/handler/auth` 目录，确保仓库内不再引用该路径。
- 回归：`go test ./... -count=1 -p 1` 通过（Windows）。

#### 可选（本 PR，如不增加风险）
- 若存在其它文档对实现路径有明确引用，则同步更新（以保持文档与代码一致）。

#### 不做（本 PR）
- 修改 register/login/revoke/offline 等动作语义、错误码、权限点与授权裁决行为。
- 调整密钥生成/加载/持久化策略（`node_keys.json`、`trusted_nodes.json` 结构与写回逻辑保持）。
- 将 auth 的发送/响应构造统一迁移到 `subproto/kit`（避免引入行为差异）。
- Linux 构建验收。

### 使用场景
- hub_server 启动时由 `modules.DefaultHub` 装配并注册 SubProto=2 的 handler。
- 运行期处理 `register/login/assist_*`、`get_perms/list_roles/perms_*` 等动作（本 PR 不改行为，仅迁移位置与依赖）。

### 功能需求（保持既有约定）
- 消息格式：`{"action":"<name>","data":{...}}`；响应 action=`<req>_resp`；状态码写在 data.code（保持）。
- 签名/密钥：ES256（P256+SHA256）及节点密钥/信任节点加载持久化逻辑保持。
- 路由/权威选择：`authority.node_id` 优先，否则父链接；无父则本地即权威（保持）。
- 权限：角色/权限配置键、缓存与 invalidate/snapshot 行为保持。

### 非功能需求
- 性能：仅包路径迁移与 import 调整，不引入热路径额外开销；避免额外 I/O、重复计算与锁竞争。
- 可维护性：变更最小化、可回滚、文档与代码一致。

### 输入输出
- 输入：`OnReceive(ctx, conn, hdr, payload)`（payload 为 auth JSON envelope）。
- 输出：通过 `srv.Send` / `conn.SendWithHeader` 发送响应帧与转发帧（Major/SubProto/Source/Target 规则保持既有实现）。

### 边界异常
- 非法 JSON / unknown action：保持当前处理方式（丢弃或返回 4xx；不引入新行为）。
- 缺少 server context：保持当前处理方式（无法转发则返回）。

### 验收标准
- `modules/hub.go` 不再 import `github.com/yttydcs/myflowhub-server/internal/handler/auth`。
- `tests` 不再 import `github.com/yttydcs/myflowhub-server/internal/handler/auth`。
- `docs/2-auth.md` 不再以 `internal/handler/auth` 作为实现路径描述。
- `rg "github.com/yttydcs/myflowhub-server/internal/handler/auth" ./` 在仓库内无命中（历史归档 `docs/change/*` 可能仍包含文字描述，不作为本 PR 验收对象）。
- `go test ./... -count=1 -p 1` 通过（Windows）。

### 风险
- 漏改 import 导致编译失败（`go test` 可覆盖）。
- 迁移时误改 handler 逻辑导致行为差异（本 PR 坚持“最小迁移”，避免额外重构）。

## 2) 架构设计（分析）
### 总体方案（含选型理由 / 备选对比）
- 方案 A（采用）：将 `internal/handler/auth` 通过 `git mv` 迁移到 `subproto/auth`，并将装配层/测试/文档引用切换到新路径。
  - 优点：最小 diff、行为稳定、符合“子协议模块可装配/可裁切”的目标架构方向。
  - 缺点：仍在 Server 仓库内；后续若要独立为单独库，再做下一轮拆分。
- 方案 B（不选）：保留 internal 实现，在 `subproto/auth` 做 wrapper。
  - 缺点：增加维护成本与间接层，不利于后续拆库。

### 模块职责
- `subproto/auth`：Auth 子协议处理（SubProto=2），包含动作分发表、签名校验、角色/权限查询与路由索引维护等逻辑（本 PR 不改逻辑，仅迁移位置）。
- `modules`：装配入口，负责创建 handler 并注册到 dispatcher。
- `protocol/auth`：兼容壳（保留旧 import path）；本 PR 内 `subproto/auth` 将直接依赖 Proto，但该兼容壳可继续保留给历史代码。

### 数据 / 调用流
1) `modules.DefaultHub` 构造 `auth.NewLoginHandlerWithConfig(cfg, log)` 放入 `Set.Handlers`
2) `modules.RegisterAll` -> `dispatcher.RegisterHandler(handler)`
3) dispatcher 按 `SubProto()==2` 分发到 handler，handler 内部按 `msg.Action` 找到 action entry 并处理
4) handler 通过 `ServerFromContext(ctx)` 获取路由/连接管理能力并转发/响应（保持既有实现）

### 接口草案
- 对外构造：
  - `NewLoginHandler(log *slog.Logger) *LoginHandler`
  - `NewLoginHandlerWithConfig(cfg core.IConfig, log *slog.Logger) *LoginHandler`
- 关键方法：
  - `SubProto() uint8`
  - `Init() bool`
  - `OnReceive(ctx, conn, hdr, payload)`

### 错误与安全
- 不改变既有签名校验、可信节点、公钥冲突处理与默认拒绝策略（安全默认保持）。
- 不新增权限点、不改变 `auth.revoke` 等权限裁决路径。

### 性能与测试策略
- 性能：仅包路径迁移与 import 调整，无额外热路径开销。
- 测试：
  - 全量回归：`$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`

### 可扩展性设计点
- `subproto/<name>` 目录作为子协议模块统一落点；`subproto/auth` 直连 Proto 协议包后，更易进一步抽离为可复用库。

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（范围与路径明确；本 PR 坚持最小迁移，wire/行为不变）。

### AU0 - 归档旧 plan.md 并准备本 workflow 文档
- 目标：保留历史 plan 的可审计性，不影响当前 workflow。
- 涉及文件：
  - `plan_archive_2026-02-16_file-subproto.md`
  - `plan.md`
- 验收条件：新 `plan.md` 仅描述本次 auth 迁移。
- 回滚点：revert 文档提交。

### AU1 - 迁移 auth 到 subproto/auth
- 目标：`LoginHandler` 及其实现从 `internal/handler/auth` 迁移到 `subproto/auth`。
- 涉及模块/文件（预期）：
  - `internal/handler/auth/*` → `subproto/auth/*`
- 验收条件：
  - 包名保持 `auth`，对外构造函数签名不变（如 `NewLoginHandlerWithConfig`）。
  - `SubProto()==2`、动作分发与路由行为保持不变。
- 测试点：`go test ./...`。
- 回滚点：revert 本迁移提交。

### AU2 - subproto 直连 MyFlowHub-Proto 协议包
- 目标：`subproto/auth` 直接 import `github.com/yttydcs/myflowhub-proto/protocol/auth`。
- 涉及文件：
  - `subproto/auth/types.go`
- 验收条件：仅 import 路径变化，常量/类型一致，wire 不变。
- 回滚点：revert。

### AU3 - modules + tests 引用切换到新路径
- 目标：装配层与测试不再依赖 `internal/handler/auth`。
- 涉及文件（预期）：
  - `modules/hub.go`
  - `tests/*`
- 验收条件：编译与测试通过；无旧 import 残留（历史归档除外）。
- 回滚点：revert。

### AU4 - 文档路径同步
- 目标：避免文档继续描述已不存在的实现路径。
- 涉及文件（预期）：
  - `docs/2-auth.md`
- 验收条件：文档中实现路径更新为 `subproto/auth`，语义不变。
- 回滚点：revert 文档提交。

### AU5 - 全量回归
- 目标：确保迁移不破坏编译/测试。
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过（Windows）。

### AU6 - Code Review + 归档变更
- 目标：按模板完成审查与归档。
- 涉及文件：
  - `docs/change/YYYY-MM-DD_auth-subproto.md`
- 验收条件：归档包含任务映射、关键决策、测试结果与回滚方案。

## 注意事项
- 禁止计划外改动：若迁移过程中发现必须同步修改其它子协议或引入 wire 变更，必须回到 3.1 更新计划并重新确认。

## 执行记录
- 2026-02-16：创建本 workflow worktree 与计划文档。
- 2026-02-16：完成 AU1-AU5；回归 `go test ./... -count=1 -p 1` 通过（Windows）。
- 2026-02-16：完成 AU6（Code Review 通过；归档文档补齐）。
