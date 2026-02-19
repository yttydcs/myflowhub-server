# 2026-02-19 - Server：topicbus 子协议改为依赖 `myflowhub-subproto/topicbus`

## 变更背景 / 目标
为实现 “子协议可裁切/可组装”，Server 应仅承担装配与编排职责，子协议实现逐步迁移为独立 Go module。

本次将 topicbus 子协议从 Server 仓库中抽离，改为依赖独立 module：
- `github.com/yttydcs/myflowhub-subproto/topicbus v0.1.0`（tag：`topicbus/v0.1.0`）

目标（保持行为不变）：
1) Server 默认启用集合仍注册 topicbus handler；
2) 删除 Server 内部 `subproto/topicbus`，避免双实现漂移；
3) `GOWORK=off` 方式回归通过，便于审计与可复现。

## 具体变更内容（新增 / 修改 / 删除）

### 修改
- `go.mod/go.sum`：新增依赖 `github.com/yttydcs/myflowhub-subproto/topicbus v0.1.0`
- `modules/defaultset/topicbus_enabled.go`：topicbus handler import 指向 subproto module
- `tests/topicbus_handler_test.go`：改为使用 subproto module 的 `SubProtoTopicBus` 与 handler 类型

### 删除
- `subproto/topicbus/*`：删除 Server 内部 topicbus 实现目录

## 对应 plan.md 任务映射
- SRVTB0 - 归档旧 plan ✅
- SRVTB1 - 切换到 subproto module（import + go.mod）✅
- SRVTB2 - 回归验证（命令级）✅
- SRVTB3 - Code Review ✅
- SRVTB4 - 归档变更（本文档）✅

## 关键设计决策与权衡
1) **只做归属调整，wire 不变**
   - 不改 SubProto 值、Action 字符串、HeaderTcp 语义与 topicbus 行为，降低回归风险。

2) **删除 Server 内部实现**
   - 通过删除 `subproto/topicbus` 避免“双路径/双实现并存”导致的长期漂移风险。

## 测试与验证方式 / 结果
- `GOWORK=off go test ./... -count=1 -p 1`
- 结果：通过。

## 潜在影响与回滚方案

### 潜在影响
- Server 增加新依赖：`myflowhub-subproto/topicbus`；若拉取版本受网络/缓存影响，会导致 `go mod tidy` 或构建失败。

### 回滚方案
- `git revert` 本次切换提交即可回滚（恢复为 Server 自带 topicbus 实现的方式）。

