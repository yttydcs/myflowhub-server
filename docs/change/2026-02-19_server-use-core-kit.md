# 2026-02-19 - Server：切换到 MyFlowHub-Core/subproto/kit（依赖 core@v0.2.1）

## 变更背景 / 目标
此前 `MyFlowHub-Server` 仓库内包含 `subproto/kit`，用于承载子协议实现复用的“样板/工具”：
- action 注册模板（`NewAction/FuncAction`）与 action 语义分类（`ActionKind`，仅用于工程组织/可观测）；
- 常用的响应构造/发送辅助（`BuildResponse/SendResponse/Clone*`）。

随着后续“子协议实现拆成独立 Go module（A2：单仓多 module）”推进，子协议实现应尽量只依赖：
- `myflowhub-core`（框架与运行时通用能力）
- `myflowhub-proto`（协议字典）

因此本次将 `subproto/kit` 的归属收敛到 Core：Server 侧删除自带实现，统一依赖 `myflowhub-core@v0.2.1` 提供的 `github.com/yttydcs/myflowhub-core/subproto/kit`。

本次目标（保持行为不变）：
1) Server 全仓切换 import 到 Core 的 `subproto/kit`；
2) 删除 Server 内部的 `subproto/kit`，避免双实现漂移风险；
3) `GOWORK=off` 方式回归通过，便于审计与可复现。

## 具体变更内容（新增 / 修改 / 删除）

### 修改
- `go.mod`：依赖 `github.com/yttydcs/myflowhub-core v0.2.1`
- 多个子协议实现：import 从 `github.com/yttydcs/myflowhub-server/subproto/kit` 切换到 `github.com/yttydcs/myflowhub-core/subproto/kit`

### 删除
- `subproto/kit/*`：删除 Server 内部实现（统一以 Core 提供为准）

## plan.md 任务映射
- SRVKIT0 - 归档旧 plan ✅
- SRVKIT1 - 切换 import 到 Core kit ✅
- SRVKIT2 - 删除 Server 内 `subproto/kit` ✅
- SRVKIT3 - 回归验证（命令级）✅
- SRVKIT4 - Code Review ✅
- SRVKIT5 - 归档变更（本文档）✅

## 关键设计决策与权衡
1) **kit 归属 Core**
   - `kit` 属于“运行时模板/工具”，不应绑定 Server；上移 Core 后，未来子协议独立 module 可直接复用。

2) **仅做归属调整，保持 wire 与行为不变**
   - 不改 SubProto 值、Action 字符串、HeaderTcp 语义与发送策略，降低回归风险与回滚成本。

## 测试与验证方式 / 结果
- 回归测试（通过）：
  - `GOWORK=off go test ./... -count=1 -p 1`

## 潜在影响与回滚方案
### 潜在影响
- 若上游/本仓库存在遗漏的旧 import 路径，将导致编译失败或出现“双路径并存”；本次已全量切换并删除 Server 侧实现以消除漂移面。
- 本仓库开始显式依赖 `myflowhub-core@v0.2.1`；若外部环境无法拉取该版本，需要同步更新依赖缓存或网络策略。

### 回滚方案
- `git revert` 本次切换提交即可回滚（恢复为 Server 自带 `subproto/kit` 的方式）。

