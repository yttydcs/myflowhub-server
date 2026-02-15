# 2026-02-15 - 子协议去 internal + subproto 基础（PR2-2a）

## 变更背景 / 目标
为后续“彻底重构（Core/Server/子协议解耦、可裁切/可组装、逐步拆库）”铺路，需要先把子协议实现从 `internal/handler/*` 中逐步迁移出来，形成可复用的公开实现包，并沉淀共享能力到可复用位置，降低子协议间耦合。

本次 PR 目标（保持行为不变）：
1) 引入 `subproto/kit`：承载子协议实现可复用的通用能力（响应发送、头部克隆等）。
2) 将 **management 子协议** 去 internal 化：`internal/handler/management` → `subproto/management`。
3) **彻底移除 LoginServer**：删除 `cmd/login_server` 与 `internal/login_server` 以及对应文档入口，减少维护面与暴露面。

## 具体变更内容
### 删除
- 删除 LoginServer 相关代码与入口：
  - `cmd/login_server/**`
  - `internal/login_server/**`
  - `docs/2-login-server.md`

### 新增
- `subproto/kit/kit.go`
  - `CloneRequest` / `CloneWithTarget`：封装 HeaderTcp 头部克隆与改写目标节点
  - `BuildResponse`：根据请求构建响应头
  - `SendResponse`：优先通过 `srv.Send` 走发送管线；无 `srv` 时回退 `conn.SendWithHeader`

### 修改
- `internal/handler/common.go`
  - 保持原对内 API 不变，但实现委托到 `subproto/kit`，减少重复实现，为后续迁移提供统一底座。
- management 子协议迁移并解耦 internal 依赖：
  - `internal/handler/management/**` → `subproto/management/**`（迁移实现）
  - `subproto/management` 改为使用 `subproto/kit.SendResponse`
  - `subproto/management/types.go` 依赖 `github.com/yttydcs/myflowhub-proto/protocol/management`（为未来拆库降低对本仓库兼容壳依赖）
- `modules/hub.go`
  - 默认启用集合中 management handler 的 import 切换为 `subproto/management`（其余子协议仍保持 `internal/handler/*`，遵循小步迁移策略）。
- `tests/integration_root_hub_ping_test.go`
  - 更新 management import path 以覆盖迁移后的 `node_echo` 链路。
- `docs/2-auth.md`
  - 更新说明：login_server 旧流程已移除（避免误导）。

## plan.md 任务映射
- S1 - 彻底移除 LoginServer ✅
- S2 - 新增 `subproto/kit`（共享工具）✅
- S3 - 迁移 management 子协议到 `subproto/management` ✅
- S4 - 全量回归 + 冒烟说明 ✅（回归已执行；冒烟步骤在 plan.md 中给出）

## 关键设计决策与权衡
- **小步迁移**：仅先迁移 management，其他子协议仍留在 `internal/handler/*`，避免单个 PR 过大、便于回滚与审计。
- **共享能力前置**：先落地 `subproto/kit`，并让旧 `internal/handler/common.go` 委托到 kit，确保后续迁移时共享能力不重复实现。
- **依赖方向**：`subproto/management` 优先依赖 `myflowhub-proto` 的协议定义包，减少对本仓库 `protocol/*` 兼容壳的耦合，便于未来抽离为独立库。

## 测试与验证方式 / 结果
- 回归测试（通过）：
  - `GOTMPDIR=d:\project\MyFlowHub3\.tmp\gotmp`
  - `go test ./... -count=1 -p 1`
- 冒烟验证（建议执行）：
  - 启动 `hub_server` 后执行 management `node_echo`（详见 `plan.md`）。

## 潜在影响与回滚方案
### 潜在影响
- 迁移 management 后，任何仍引用旧路径 `internal/handler/management` 的代码会编译失败；本 PR 已同步更新 `modules` 与集成测试。
- LoginServer 删除后，若外部仍依赖其二进制/流程，需要另行迁移（本仓库已不再提供）。

### 回滚方案
- 可按提交粒度回滚：
  1) 回滚 “management 子协议迁移到 subproto”
  2) 回滚 “新增 subproto/kit”
  3) 回滚 “移除 login_server”
  - 或直接整体 revert 本 PR 合并提交

