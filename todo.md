# Todo - Server 下游依赖对齐 Core v0.4.5

## 目标与状态
- 目标：将 `github.com/yttydcs/myflowhub-core` 依赖升级到 `v0.4.5`，确保 Server 发布默认包含最新 RFCOMM 流式修复。
- 当前状态：Server `main` 位于 `v0.0.8`，Core 依赖为 `v0.4.4`。

## 任务清单
- [ ] SERVER-1 更新 `go.mod` 的 Core 版本到 `v0.4.5`
- [ ] SERVER-2 运行测试验证（`go test ./... -count=1`）
- [ ] SERVER-3 更新变更归档 `docs/change/2026-03-15_bump-core-v0.4.5-server.md`
- [ ] SERVER-4 提交、合并、打 tag（`v0.0.9`）

## 验收条件
- Server 在不改业务逻辑前提下完成 Core 版本对齐。
- 测试通过。
- 归档完整可审计。

## 回滚点
- 回退 `go.mod` 与归档文档。
