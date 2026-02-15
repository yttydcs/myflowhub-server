# Plan - 子协议去 internal + subproto 基础（PR2-2a）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-subproto-public`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr2-subproto-foundation\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`
- 约束：commit 信息使用中文（前缀如 `refactor:` 可英文）

## 当前状态
- `cmd/hub_server` 已通过 `modules` 统一装配默认启用集合（PR2-1a 已合并到 `main`）。
- 子协议实现仍主要位于 `internal/handler/*`（auth/varstore/topicbus/file/flow/exec/management + default）。
- LoginServer（`cmd/login_server` 与 `internal/login_server`）已确认可以移除（本 PR S1）。
- `protocol/*` 为兼容壳，实际委托到 `github.com/yttydcs/myflowhub-proto`（wire 不变）。

> 环境备注（不进 git）：本仓库 `go.mod` 使用 `replace ../MyFlowHub-Core`、`../MyFlowHub-Proto`。
> 在本 worktree 布局下，需要在 `d:\project\MyFlowHub3\worktrees\pr2-subproto-foundation\` 下存在同名目录。
> 当前通过 Junction 指向 `d:\project\MyFlowHub3\repo\MyFlowHub-Core` / `repo\MyFlowHub-Proto` 满足。

## 1) 需求分析
### 目标
1) 将子协议实现从 `internal/handler/*` 逐步迁移为公开可复用实现包：`subproto/<name>`，共享能力沉到 `subproto/kit`。
2) 保持 `modules` 为 `hub_server` 唯一装配入口；本轮不改 wire/SubProto/权限语义/handler 业务逻辑。
3) **彻底移除 LoginServer**：删除 `cmd/login_server` 与 `internal/login_server`（以及对应文档/引用）。

### 范围
#### 必须（本 PR，小步）
- 删除 LoginServer 入口与实现：
  - 删除 `cmd/login_server/**`
  - 删除 `internal/login_server/**`
  - 删除/更新仓库内相关文档（至少 `docs/2-login-server.md`）
- 新增 `subproto/kit`：承载跨子协议可复用的响应发送/头部克隆等工具。
- 迁移最小一批子协议到 `subproto/*`：优先迁移 `management`（用于冒烟 `node_echo`）。
- 更新 `modules.DefaultHub` 使用 `subproto/management`（默认集合不变）。
- 回归：`go test ./... -count=1 -p 1` 通过。

#### 可选（后续 PR）
- 继续迁移 `auth/varstore/topicbus/file/flow/exec` 到 `subproto/*`。
- 将 `internal/broker` 去 internal 化（为后续 flow/exec 拆库铺路）。
- 最终移除 `internal/handler/**`。

#### 不做（本 PR）
- 改 wire（action 名称/消息结构/SubProto 值）。
- 改 HeaderTcp、路由规则、权限语义。
- 改 Win / SDK（本 PR 仅 Server）。

### 验收标准
- 仓库不再包含 `cmd/login_server` 与 `internal/login_server`，且全仓可编译/测试通过。
- `hub_server` 仍可启动并注册 management；`management node_echo` 仍可走通。
- `go test ./... -count=1 -p 1` 通过。

### 风险
- 删除 + `git mv` 造成 diff 噪音：严格控制在本计划范围。
- 共享工具抽取引入循环依赖：`subproto/kit` 必须保持无子协议依赖。

## 2) 架构设计（分析）
### 总体方案（选型理由 / 备选对比）
- 采用你确认的方案 A：`subproto/<name>` + `subproto/kit`
  - 优点：实现可复用、未来可拆库；共享能力不再被 `internal` 限制。
  - 缺点：迁移需要多 PR 渐进推进（符合“小步多 PR”）。

### 模块职责
- `subproto/kit`：子协议共享工具（响应发送、头部克隆等），不依赖具体子协议。
- `subproto/management`：management 子协议实现（node_echo/config/list 等），依赖 `subproto/kit` + `core` + `proto`。
- `modules`：hub 默认启用集合与装配入口（逐步把 internal handler 替换为 subproto）。
- LoginServer：本轮删除。

### 数据 / 调用流
1) `cmd/hub_server` 构建 `cfg/log/dispatcher`
2) `modules.DefaultHub(cfg, log)` 构造 handler 集合（management 由 `subproto/management` 提供）
3) `dispatcher.RegisterHandler` 触发 `Init`（生命周期保持不变）
4) `srv.Start` 后 `modules.BindServerHooks`（保持 flow 的 BindServer 语义）

### 错误与安全
- 保持现有 hop_limit 与 parent 保护逻辑。
- 删除 LoginServer 同时删除文档入口，避免误用。

### 性能与测试策略
- 回归：`GOTMPDIR=d:\project\MyFlowHub3\.tmp\gotmp; go test ./... -count=1 -p 1`
- 冒烟（手工）：
  1) `go run .\\cmd\\hub_server -addr :9000 -node-id 1`
  2) 执行 management `node_echo`（或运行 `go test .\\tests -run TestRootHubPing -count=1` 覆盖 echo 链路）

### 可扩展性设计点
- `modules` 后续可扩展多套 set（minimal/full）或 build tags，完成编译期裁切。
- `subproto/*` 按“未来可独立 go module”原则组织依赖。

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（目录方案/PR 粒度/移除 LoginServer 已确认）。

### S1 - 彻底移除 LoginServer
- 目标：仓库不再包含 LoginServer 入口/实现，降低维护面与暴露面。
- 涉及模块/文件（预期）：
  - 删除 `cmd/login_server/**`
  - 删除 `internal/login_server/**`
  - 删除 `docs/2-login-server.md`
  - 清理相关引用（如存在）
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过。
- 回滚点：
  - revert 删除提交（恢复目录与入口）。

### S2 - 新增 `subproto/kit`（共享工具）
- 目标：提供子协议实现可复用的共享能力，避免依赖 `internal/handler/common.go`。
- 涉及模块/文件（预期）：
  - 新增 `subproto/kit/**`（CloneRequest/CloneWithTarget/BuildResponse/SendResponse 等）
  - 可选：让 `internal/handler/common.go` 委托到 `subproto/kit`（降低重复实现）
- 验收条件：
  - `subproto/kit` 不依赖任何具体子协议包。
  - 行为与 `internal/handler/common.go` 一致（优先走 `srv.Send`，否则回退直写连接）。
- 测试点：
  - `go test ./... -count=1 -p 1`
- 回滚点：
  - revert `subproto/kit/**`。

### S3 - 迁移 management 子协议到 `subproto/management`
- 目标：将 management 实现去 internal 化，并改用 `subproto/kit` 发送响应。
- 涉及模块/文件（预期）：
  - `internal/handler/management/**` → `subproto/management/**`（git mv）
  - `modules/hub.go`：management import/装配切换到 `subproto/management`
  - `tests/**`：更新 import path（尤其 `tests/integration_root_hub_ping_test.go`）
- 验收条件：
  - management 行为不变，且不再 import `internal/handler`。
  - `modules.DefaultHub` 默认启用集合不变。
- 测试点：
  - `go test ./... -count=1 -p 1`
- 回滚点：
  - revert management 迁移提交。

### S4 - 全量回归 + 冒烟说明
- 目标：保证结构调整不影响现有可用性。
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过。
  - 冒烟步骤可执行（见 2) 性能与测试策略）。
- 回滚点：
  - revert 本 PR 全部提交。

## 注意事项
- 禁止计划外改动：若迁移过程中发现必须同时迁移其它子协议或修改其它仓库，必须回到 3.1 更新计划并重新确认。
