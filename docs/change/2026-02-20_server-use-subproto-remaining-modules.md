# 2026-02-20 - Server：剩余子协议改为依赖 `myflowhub-subproto/*` modules（broker/auth/varstore/file/forward/exec/flow）

## 变更背景 / 目标
为实现 “子协议可裁切/可组装”，Server 应仅承担装配与编排职责，子协议实现逐步迁移为独立 Go module。

此前 Server 仓库内仍包含以下子协议实现目录：
- `subproto/auth`、`subproto/varstore`、`subproto/file`、`subproto/forward`、`subproto/exec`、`subproto/flow`
并通过 `internal/broker` 承载 `exec/flow` 的“同进程 reqID -> resp 投递”能力，导致 `exec/flow` 难以彻底从 Server 解耦。

本次变更目标（保持行为不变）：
1) Server 改为依赖独立 subproto modules（semver）并删除内置实现目录；
2) 删除 `internal/broker`，由 `myflowhub-subproto/broker` 承载共享投递器；
3) 以 `GOWORK=off` 方式回归通过，确保可审计与可复现。

> 完整拆分设计与 module/tag 列表见：
> - `MyFlowHub-SubProto/docs/change/2026-02-20_subproto-split-remaining-modules.md`

## 具体变更内容（新增 / 修改 / 删除）

### 修改
- `go.mod/go.sum`：新增依赖（均为 `v0.1.0`）
  - `github.com/yttydcs/myflowhub-subproto/auth`
  - `github.com/yttydcs/myflowhub-subproto/varstore`
  - `github.com/yttydcs/myflowhub-subproto/file`
  - `github.com/yttydcs/myflowhub-subproto/forward`
  - `github.com/yttydcs/myflowhub-subproto/exec`
  - `github.com/yttydcs/myflowhub-subproto/flow`
  - `github.com/yttydcs/myflowhub-subproto/broker`（间接依赖：由 exec/flow 引入）
- `modules/defaultset/*`：对应 handler 的 import 切换到 subproto modules
- `tests/*`：对应 handler 测试 import 切换到 subproto modules

### 删除
- `subproto/auth|varstore|file|forward|exec|flow`：删除 Server 内置实现目录
- `internal/broker`：删除 Server 私有投递器实现（迁移至 `myflowhub-subproto/broker`）

## 对应 plan.md 任务映射
- SRVALL0 - 归档旧 plan ✅
- SRVALL1 - 更新 go.mod/go.sum ✅
- SRVALL2 - 更新 import 与装配点 ✅
- SRVALL3 - 删除 Server 内实现目录与 internal/broker ✅
- SRVALL4 - 回归验证 ✅
- SRVALL5 - Code Review ✅
- SRVALL6 - 归档变更（本文档）✅

## 关键设计决策与权衡
1) **只做归属/依赖边界调整，wire 不变**
   - 不改 SubProto 值、Action 字符串、JSON 结构、HeaderTcp 语义与路由规则，降低回归风险。

2) **删除 Server 内置实现，避免“双实现漂移”**
   - 同一子协议实现只保留在 `myflowhub-subproto/*`，Server 只装配引用，长期维护成本更低。

3) **跨子协议共享能力以显式 shared module 承载**
   - `exec/flow` 依赖的“同进程投递器”以 `myflowhub-subproto/broker` 承载，避免回退到 Server 私有包导致边界塌陷。

## 测试与验证方式 / 结果

统一约束：验收使用 `GOWORK=off`，避免本地 `go.work` 干扰审计。

```powershell
$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
$env:GOWORK='off'
go test ./... -count=1 -p 1
go test ./tests -run TestRootHubPing -count=1
```

结果：通过。

## 潜在影响与回滚方案

### 潜在影响
- Server 增加多项新的 semver 依赖；若网络/缓存异常，会导致 `go mod tidy` 或构建失败。
- 子协议实现不再位于 Server 仓库：调试时需跳转到 `MyFlowHub-SubProto` 对应 module 源码（但边界更清晰）。

### 回滚方案
- `git revert` 本次切换提交即可回滚（恢复为 Server 自带子协议实现的方式）。

