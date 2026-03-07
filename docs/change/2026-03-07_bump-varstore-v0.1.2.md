# 2026-03-07 bump VarStore v0.1.2（Server 依赖升级）

## 背景 / 目标

- 背景：`myflowhub-subproto/varstore` 已发布 `v0.1.2`（逐跳回程对齐：`*_resp` 统一 `MajorCmd`、SourceID 保留为原始 actor、并发写 1:1 匹配等）。
- 目标：Server 侧依赖升级到 `varstore v0.1.2`，确保编译与运行时使用新协议行为，便于 Android/Win 等客户端联调验证。

## 具体变更

- `go.mod`：`github.com/yttydcs/myflowhub-subproto/varstore v0.1.1 -> v0.1.2`
- `go.sum`：随依赖求解更新

## 影响范围

- 编译期：Server 构建会拉取 `varstore v0.1.2`（需要网络访问 GitHub tag）。
- 运行期：VarStore 子协议响应帧 Major 语义变化为 `MajorCmd`，中间节点逐跳可见；旧客户端若只接受 `OKResp/ErrResp` 可能 await 超时（需配合 SDK `v0.1.2` 及以上）。

## 测试与验证

- `GOWORK=off go mod tidy`（已执行）
- `GOWORK=off go test ./...`（已执行，通过）

## 回滚方案

- 将 `go.mod` 中 `varstore` 版本回退到 `v0.1.1` 并重新 `GOWORK=off go mod tidy`。
