# 2026-03-03 Server：升级 subproto/management 至 v0.1.2（children-only）

## 背景 / 目标
- 背景：`myflowhub-subproto/management v0.1.1` 的 `list_nodes` 会把 upstream(parent) 链接也枚举出来，导致设备树出现回指（例如 `5 -> 1`）。
- 目标：
  - 发布 `github.com/yttydcs/myflowhub-subproto/management v0.1.2`，让上游能通过 semver 拉取修复；
  - 将 `MyFlowHub-Server` 的依赖升级到 `management v0.1.2`。

## 变更内容
- 发布：
  - 新增并推送 tag：`management/v0.1.2`
  - tag 指向 SubProto 仓库提交：`d512b41`（children-only 行为修复已在该仓落地）
- Server 依赖：
  - `go.mod`：`github.com/yttydcs/myflowhub-subproto/management v0.1.1` → `v0.1.2`
  - `go.sum`：同步更新 checksum
- 审计与交接：
  - 归档旧 `plan.md`：`docs/plan_archive/plan_archive_2026-03-03_server-bump-management-v0.1.2-prev.md`
  - 更新 `plan.md` 为本次发布/升级 workflow 计划

## Plan 任务映射
- SVRMG0：归档旧 plan.md
- SVRMG1：发布 `management/v0.1.2` tag
- SVRMG2：升级 Server 依赖到 `management v0.1.2`

## 关键设计决策与权衡
- 仅做 patch 升级：不改 wire schema，只通过发布新 patch 版本让依赖方获得 children-only 语义修复。
- `list_nodes` 语义收敛为 children-only：若未来需要调试/展示 upstream(parent) 拓扑，应新增独立 upstream 查询 action，而不是复用 `list_nodes`。

## 测试与验证
- 模块可解析：
  - `go list -m github.com/yttydcs/myflowhub-subproto/management@v0.1.2`
- Server 编译与测试：
  - `GOWORK=off go test ./... -count=1 -p 1`

## 潜在影响
- 任何依赖 `list_nodes` 读取 upstream(parent) 的调试用法将不可用（不再返回 parent）。
- 如需保留该能力，应以“新增 upstream 查询 action”的方式提供（避免 children-only 语义被污染）。

## 回滚方案
- Server：
  - 回滚提交，或执行：
    - `GOWORK=off go get github.com/yttydcs/myflowhub-subproto/management@v0.1.1`
    - `GOWORK=off go mod tidy`
- Tag（高风险）：
  - 原则上不删除已推送 tag；若必须删除，需先确认未被任何下游消费，并同步通知相关仓库/CI。

