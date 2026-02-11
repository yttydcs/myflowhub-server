# Plan - 协议仓库拆分（Proto）+ 子协议解耦地基（Server）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/proto-extract`
- Worktree：`d:\project\MyFlowHub3\worktrees\proto-extract\MyFlowHub-Server`
- 目标 PR：PR1（跨多个 repo 同步提交/合并）

## 项目目标（PR1）
1) 将协议定义从 `MyFlowHub-Server/protocol/*` 抽离到新仓库 **MyFlowHub-Proto**（wire 不变）。
2) Server 继续提供 `protocol/*` **兼容壳**（用户确认选 A）：外部依赖不立刻断。
3) 移除子协议之间的“实现层耦合”：至少消除 `flow -> internal/handler/exec` 的直接 import（改为协议层依赖 + 通用 broker）。
4) 为后续“子协议可裁切/可组装（编译期裁切）”铺路，但本 PR 不做大规模模块重排。

## 已确认信息
- `MyFlowHub-Proto` module：`github.com/yttydcs/myflowhub-proto`
- 兼容策略：保留 `myflowhub-server/protocol/*` 作为兼容壳（const/type alias 指向 Proto）
- wire：action 名称/消息结构/SubProto 值均不变（策略 A）

## 范围
### 必须（PR1）
- 新增依赖：MyFlowHub-Proto（go.mod require + replace）
- 将 `protocol/auth|varstore|topicbus|file|flow|management` 改为兼容壳（指向 Proto）
- 新增 `protocol/exec`（与 Proto 对齐，供 Win/Server 统一引用）
- 重构：将 `internal/handler/exec` 中的 Broker 抽到独立包（避免 flow 依赖 exec 实现包）
- 全量回归：`go test ./...`

### 不做（本 PR）
- 改 wire（action 名称/消息结构/子协议号）
- assist/up/notify 的 wire 收敛（本轮策略 A：仅做代码层抽象，wire 保持不变）
- 子协议按库拆分为多仓（PR2+）
- build tags/模块裁切（PR2+，待模块注册方式收敛后再做）

## 问题清单（阻塞：否）
- 无

## 任务清单（Checklist）

### S1 - 接入 Proto 依赖（go.mod）
- 目标：Server 可以引用 Proto 包，但保持本仓库对外 API 稳定。
- 涉及模块/文件：
  - `go.mod`（新增 require；新增 replace 指向本地 `../MyFlowHub-Proto`）
- 验收条件：
  - `go test ./...` 编译通过（后续任务完成后整体验证）。
- 回滚点：
  - revert `go.mod`。

### S2 - protocol/* 兼容壳改造（指向 Proto）
- 目标：保留 import path `github.com/yttydcs/myflowhub-server/protocol/...`，其内容转为 alias 到 Proto。
- 涉及模块/文件：
  - `protocol/*/types.go`（auth/varstore/topicbus/file/flow/management）
  - 新增 `protocol/exec/types.go`
- 验收条件：
  - `internal/**` 在不大改的情况下可继续编译。
  - 外部依赖若仍 import `myflowhub-server/protocol/*` 不破坏。
- 测试点：
  - `go test ./... -count=1`
- 回滚点：
  - revert `protocol/**`。

### S3 - 解耦 flow 与 exec 实现包（移除跨 handler import）
- 目标：消除 `internal/handler/flow` 对 `internal/handler/exec` 的直接 import。
- 涉及模块/文件：
  - 新增通用 broker 包：`internal/broker`（共享 Exec call_resp 的投递能力）
  - `internal/handler/exec/*`（改用 `protocol/exec` + `internal/broker`）
  - `internal/handler/flow/handler.go`（改用 `protocol/exec` + `internal/broker`）
- 验收条件：
  - `rg "internal/handler/exec" internal/handler/flow` 无结果。
  - 功能行为不回退（至少单测/集成测覆盖的路径不回退）。
- 测试点：
  - `go test ./... -count=1`
- 回滚点：
  - broker 抽取独立提交，可单独 revert。

### S4 - 全量回归
- 目标：确保 Server 行为在协议拆分后保持稳定。
- 验收条件：
  - `go test ./...` 通过。
- 回滚点：
  - 按提交粒度回滚（优先回滚 protocol 壳与 broker 变更）。

## 依赖关系
- S1/S2 依赖 Proto 仓库完成基础结构（见 Proto 侧 plan）。
- Win 将在本 PR 同步切换 import（避免继续依赖 Server repo 取协议类型）。

## 风险与注意事项
- 兼容壳（A）策略的风险：import 链会多一层，但可控；后续可逐步引导外部迁移到 Proto。
- 严禁 Proto 反向依赖 Server/Core（Proto 只能依赖标准库）。
