# 2026-03-05 - Server：升级 subproto/varstore 到 v0.1.1

## 变更背景 / 目标
- 背景：上游 `myflowhub-subproto/varstore` 发布 `v0.1.1`，修复跨层拓扑下 VarStore 的 target 转发与 owner 路由自愈。
- 目标：让 Server 编译与运行时使用该修复版本，不改业务行为与协议字段。

## 具体变更内容
- `go.mod`
  - `github.com/yttydcs/myflowhub-subproto/varstore v0.1.0 -> v0.1.1`
- `go.sum`
  - 同步校验和。
- `todo.md`
  - 记录本次 workflow 任务与验收。

## plan 任务映射
- SRVVAR-1：依赖升级 -> 完成
- SRVVAR-2：回归验证 -> 完成
- SRVVAR-3：归档 -> 完成

## 关键设计决策与权衡
- 采用最小变更策略：仅做 patch 版本依赖升级，不引入额外功能改动，降低回归风险。

## 测试与验证方式 / 结果
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/varstore`
  - 输出：`v0.1.1`
- `GOWORK=off go test ./... -count=1 -p 1`
  - 结果：通过。

## 潜在影响与回滚方案
- 潜在影响：VarStore 跨层场景行为将按上游修复版本执行。
- 回滚：将 `go.mod/go.sum` 回退到 `varstore v0.1.0` 并重新发布。
