# 2026-02-16 - modules/defaultset（默认装配集合解耦）

## 变更背景 / 目标
当前 `modules.DefaultHub(cfg, log)` 在 `modules/hub.go` 内直接 import 各 `subproto/*` 并硬编码默认启用集合。

为贴合 `target.md` 的目标架构（默认集合落到 `modules/defaultset`，为后续裁切/组装预留落点），本次变更：
- 新增 `modules/defaultset` 承载 hub_server 默认启用集合的构造策略（handlers + default）
- `modules.DefaultHub` 改为委托 `modules/defaultset`
- 保持默认启用集合与行为不变

## 具体变更内容（新增 / 修改 / 删除）
### 新增
- `modules/defaultset/hub.go`：
  - `DefaultHub(cfg, log)` 返回 hub_server 默认启用的 `handlers` 与 `default` fallback。

### 修改
- `modules/hub.go`：
  - `DefaultHub(cfg, log)` 委托 `defaultset.DefaultHub(cfg, log)` 构造集合；
  - `modules` 包不再直接 import 具体 `subproto/*`（依赖收敛到 defaultset）。

### 删除
- 无。

## 对应 plan.md 任务映射
- DS0：归档旧 plan 并更新本 workflow 文档
- DS1：新增 `modules/defaultset`
- DS2：`modules.DefaultHub` 委托 defaultset
- DS3：全量回归
- DS4：Code Review + 归档变更（本文档）

## 关键设计决策与权衡
1) **defaultset 不反向依赖 modules**
   - `modules` 提供装配抽象与校验；`defaultset` 仅提供“默认集合的构造策略”，避免形成 import cycle。

2) **保持 `modules.DefaultHub` 作为稳定入口**
   - 对外函数签名不变；调用方（`cmd/hub_server`）无需改动，降低迁移风险。

3) **本 PR 不引入 build tags / module registry**
   - 仅先完成“默认集合构造”的解耦，为后续裁切与更复杂装配策略预留落点（小步多 PR）。

## 测试与验证方式 / 结果
- Windows：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`
  - 结果：通过。
- Linux：按当前约定暂不验收。

## 潜在影响与回滚方案
### 潜在影响
- 默认启用集合与行为不应变化；仅装配构造逻辑的包边界调整。

### 回滚方案
- 直接 `git revert` 本次合并的提交即可回滚（新增 defaultset + 装配委托修改均在同一变更集中）。

## Code Review（结论）
- 需求覆盖：通过（默认集合解耦到 `modules/defaultset`；行为不变；回归通过）。
- 架构合理性：通过（职责更清晰：`modules` 抽象/校验；`defaultset` 策略构造；依赖方向更可控）。
- 性能风险：通过（仅装配期构造，不影响运行期热路径）。
- 可读性与一致性：通过（命名直观；变更点集中；保持既有 API）。
- 可扩展性与配置化：通过（为后续 build tags/裁切预留落点；本次不引入额外复杂度）。
- 稳定性与安全：通过（不改权限/路由/协议语义；仅改装配边界）。
- 测试覆盖：通过（全量 `go test ./...` 通过；modules 包与 tests 覆盖注册路径）。

