# Plan - DefaultForwardHandler 去 internal（subproto/forward）（PR2-3a）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-default-forward`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr2-default-forward\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`
- 约束：commit 信息使用中文（前缀如 `refactor:` 可英文）

## 当前状态
- `subproto/kit` 已落地：共享响应发送、头部克隆等通用能力。
- management 已迁移到 `subproto/management`，并由 `modules` 通过公开包装配。
- `DefaultForwardHandler`（dispatcher 默认 fallback）已迁移到 `subproto/forward`，`modules`/测试已切换到新包路径；`internal/handler` 顶层包已移除（仅保留 `internal/handler/<sub>`）。

> 环境备注（不进 git）：本仓库 `go.mod` 使用 `replace ../MyFlowHub-Core`、`../MyFlowHub-Proto`。  
> 在本 worktree 布局下，需要在 `d:\project\MyFlowHub3\worktrees\pr2-default-forward\` 下存在同名目录。  
> 当前通过 Junction 指向 `d:\project\MyFlowHub3\repo\MyFlowHub-Core` / `repo\MyFlowHub-Proto` 满足。

## 1) 需求分析
### 目标
1) 将 `DefaultForwardHandler` 去 internal 化：迁移到 `subproto/forward`（作为可复用 fallback handler，为后续拆库/裁切做准备）。
2) 更新 `modules` 与测试使用新路径，避免继续依赖 `internal/handler` 顶层包。
3) 保持行为不变：不改 wire/SubProto/权限语义，不改变默认转发/丢弃策略。

### 范围
#### 必须（本 PR）
- 新增/迁移 `subproto/forward`（承载 `DefaultForwardHandler` 与 `NewDefaultForwardHandler`）。
- `DefaultForwardHandler` 内部依赖统一改为使用 `subproto/kit`（clone/response 等共享能力）。
- `modules/hub.go` 默认 fallback 改用 `subproto/forward.NewDefaultForwardHandler`。
- 更新测试：
  - `tests/default_handler_test.go`
  - `tests/integration_root_hub_ping_test.go`
- 清理：若 `internal/handler` 顶层包不再被引用，可删除其残留文件（可选，见后述）。
- 回归：`go test ./... -count=1 -p 1` 通过。

#### 不做（本 PR）
- 迁移其它子协议（auth/varstore/topicbus/file/flow/exec）。
- 改路由规则、HeaderTcp、权限校验逻辑。

### 使用场景
- `Dispatcher.RegisterDefaultHandler` 的默认处理器：未知子协议按配置转发到指定节点/父节点，或丢弃。
- 未来在 Server 之外复用该 fallback（例如拆库、最小 hub 组合）。

### 边界异常
- 未配置/未命中路由时的默认行为必须保持一致（默认尝试转发父节点，显式关闭 forward 时丢弃）。
- parent 连接不存在、target 不可达等情况保持现有日志与行为。

### 验收标准
- `modules` 不再 import `github.com/yttydcs/myflowhub-server/internal/handler` 顶层包。
- `go test ./... -count=1 -p 1` 通过。
- 默认 forward 的两条关键测试仍通过（转发父节点默认行为、关闭 forward 丢弃）。

### 风险
- 迁移过程中若遗漏 import 更新，会导致编译失败；通过 `go test ./...` 可快速发现。

## 2) 架构设计（分析）
### 总体方案
- 采用 `subproto/forward` 作为 fallback handler 的公开落点：
  - 仍实现 `core.ISubProcess`（SubProto=0），用于 `RegisterDefaultHandler`。
  - 通过 `subproto/kit` 复用头部克隆等工具，避免重复实现。
- `internal/handler` 顶层包逐步收缩直至移除（本 PR 若无引用，可删除残留）。

### 模块职责
- `subproto/forward`：默认 fallback 转发/丢弃逻辑（启动期构造、运行期处理未知子协议）。
- `subproto/kit`：共享工具（clone/response/send）。
- `modules`：装配入口，负责选择并注入 fallback。

### 数据 / 调用流
1) `modules.DefaultHub` 构造 handlers + `forward.NewDefaultForwardHandler(cfg, log)`
2) `modules.RegisterAll` 调用 `dispatcher.RegisterDefaultHandler(fallback)`
3) 未命中 handler 的帧进入 fallback：按配置转发到指定节点/父节点或丢弃

### 错误与安全
- 不改变既有路由与转发安全防护（如 parent 连接判断、node 索引查找）。

### 性能与测试策略
- 性能：仅包路径迁移与复用工具，不引入额外热路径开销。
- 回归：`GOTMPDIR=d:\project\MyFlowHub3\.tmp\gotmp; go test ./... -count=1 -p 1`

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（范围与路径已明确，行为保持不变）。

### F1 - 迁移 DefaultForwardHandler 到 subproto/forward
- 目标：`DefaultForwardHandler` 不再位于 `internal/handler` 顶层包。
- 涉及模块/文件（预期）：
  - `internal/handler/default_handler.go` → `subproto/forward/forward.go`（或同名）
  - `subproto/forward` 引用 `subproto/kit` 的 clone 能力
- 验收条件：
  - 构造函数与对外行为保持一致（配置键/默认策略不变）。
- 回滚点：
  - revert 本迁移提交。

### F2 - modules 与测试切换到新路径
- 目标：`modules` 与测试不再依赖 `internal/handler` 顶层包。
- 涉及模块/文件：
  - `modules/hub.go`
  - `tests/default_handler_test.go`
  - `tests/integration_root_hub_ping_test.go`
- 验收条件：
  - `go test ./...` 通过，且 default forward 关键测试仍覆盖。
- 回滚点：
  - revert 本提交。

### F3 - 清理 internal/handler 顶层包残留（可选）
- 目标：若 `internal/handler` 顶层包不再被引用且无必要保留，则删除其残留文件（例如 `internal/handler/common.go`）。
- 验收条件：
  - `rg \"myflowhub-server/internal/handler\"` 不再命中顶层包路径（子目录 `internal/handler/<sub>` 仍允许存在）。
  - `go test ./...` 通过。
- 回滚点：
  - revert 清理提交。

### F4 - 全量回归
- 目标：确保迁移不破坏编译/测试。
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过。

## 执行记录
- 2026-02-15：完成 F1/F2/F3；回归 `go test ./... -count=1 -p 1` 通过。

## 注意事项
- 禁止计划外改动：若迁移过程中发现必须同步迁移其它子协议或修改其它仓库，必须回到 3.1 更新计划并重新确认。
