# 2026-03-25 Server：defaultset 依赖链版本对齐

## 变更背景 / 目标
- 背景：
  - `MyFlowHub-Server/modules/defaultset` 已切到显式 `WithDeps` / `runtimedeps` 装配。
  - 现网 `go.mod` 仍锁定旧版 `myflowhub-proto` 与 `myflowhub-subproto` 模块，单仓构建时会报：
    - `undefined: filehandler.NewHandlerWithDeps`
    - `undefined: management.NewHandlerWithDeps`
  - 继续在 `GOWORK=off` 下追查后，又暴露出 `broker.SharedExecCapQueryBroker`、`protocol.ActionDelete` 和 `exec/runtimedeps` 的发布链缺口。
- 目标：
  - 将 `MyFlowHub-Server` 对齐到完整可发布的 `Proto -> SubProto -> Server` patch 版本链；
  - 确认在不依赖 workspace `go.work` 的情况下，`defaultset` 构建和测试恢复正常。

## 具体变更内容
- `go.mod`
  - `github.com/yttydcs/myflowhub-proto v0.1.2 -> v0.1.3`
  - `github.com/yttydcs/myflowhub-subproto/exec v0.1.0 -> v0.1.2`
  - `github.com/yttydcs/myflowhub-subproto/file v0.1.2 -> v0.1.4`
  - `github.com/yttydcs/myflowhub-subproto/flow v0.1.0 -> v0.1.2`
  - `github.com/yttydcs/myflowhub-subproto/management v0.1.2 -> v0.1.4`
  - `github.com/yttydcs/myflowhub-subproto/topicbus v0.1.0 -> v0.1.2`
  - `github.com/yttydcs/myflowhub-subproto/varstore v0.1.2 -> v0.1.4`
  - 间接依赖 `github.com/yttydcs/myflowhub-subproto/broker v0.1.0 -> v0.1.1`
- `go.sum`
  - 同步本次版本链对应的校验和。
- 本次不改 `modules/defaultset` 业务代码：
  - 正式修复路径是发布链和 semver 对齐，而不是回退 `WithDeps` / `runtimedeps` 设计。

## Requirements impact
`none`

## Specs impact
`none`

## Lessons impact
`updated`

## Related requirements
- `none`

## Related specs
- `none`

## Related lessons
- `D:\project\MyFlowHub3\docs\lessons\cross-repo-semver-release.md`

## 对应 plan.md 任务映射
- `SRVREL1` - 升级 `MyFlowHub-Server` 依赖到完整版本链
- `VAL1` - 在真实依赖解析模式下完成构建与测试验证
- `DOC1` - 归档本次 Server 侧对齐结果

## 经验 / 教训摘要
- `Server` 侧最初暴露的两个 `undefined` 只是入口症状，真正根因是整条上游 semver 发布链未收口。
- `go.mod` 版本对齐必须在 `GOWORK=off` 下验证，不能把根 workspace `go.work` 当作正式修复证据。
- 对 `defaultset` 这类跨多个 subproto module 的装配入口，升级时要同时检查 shared package、协议契约和 sibling module 的最小版本。

## 可复用排查线索
- 症状
  - `undefined: filehandler.NewHandlerWithDeps`
  - `undefined: management.NewHandlerWithDeps`
  - `undefined: broker.SharedExecCapQueryBroker`
  - `undefined: protocol.ActionDelete`
  - `no required module provides package github.com/yttydcs/myflowhub-subproto/exec/runtimedeps`
- 触发条件
  - `Server` 单仓构建或 CI 使用 `GOWORK=off`
  - 本地 worktree 已包含新 API，但远端 tag 或下游 `go.mod` 仍停在旧版本
- 关键词
  - `defaultset`
  - `NewHandlerWithDeps`
  - `SharedExecCapQueryBroker`
  - `ActionDelete`
  - `GOWORK=off`
- 快速检查
  - `go list -m github.com/yttydcs/myflowhub-proto`
  - `go list -m github.com/yttydcs/myflowhub-subproto/exec`
  - `go list -m github.com/yttydcs/myflowhub-subproto/file`
  - `go list -m github.com/yttydcs/myflowhub-subproto/management`
  - 检查 `go.mod` 是否已升级到本次 patch 版本链

## 关键设计决策与权衡
- 采用“补齐发布链”而不是“回退 `defaultset` 代码”：
  - 可以保持显式 runtime deps 设计，不引入架构倒退。
- 验证口径使用真实依赖解析：
  - `go list -m` 用于确认版本落点；
  - `GOWORK=off go build/go test` 用于确认下游真实可消费。
- 本地 Go 1.25.0 下载工具链缓存损坏时，使用定点重拉恢复验证环境：
  - 不改仓库代码，只修复本机校验环境。

## 测试与验证方式 / 结果
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-proto`
  - 结果：`v0.1.3`
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/broker`
  - 结果：`v0.1.1`
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/exec`
  - 结果：`v0.1.2`
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/file`
  - 结果：`v0.1.4`
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/flow`
  - 结果：`v0.1.2`
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/topicbus`
  - 结果：`v0.1.2`
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/varstore`
  - 结果：`v0.1.4`
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/management`
  - 结果：`v0.1.4`
- `GOWORK=off go build ./...`
  - 结果：通过
- `GOWORK=off go test ./... -count=1 -p 1`
  - 结果：通过

## 潜在影响与回滚方案
- 潜在影响
  - 外部下游若继续锁在旧 patch 版本，仍会复现同类未定义符号问题。
  - 只在本地存在而未推送的 tag 不能算正式修复。
- 回滚方案
  - 若 `Server` 分支尚未合并：回退本次 `go.mod/go.sum` 与归档文档。
  - 若上游 tag 已公开：不重写已发布 tag，改发更高 patch 版本修复。

## 子Agent执行轨迹
- 本轮未使用子 Agent。
