# 2026-02-15 - TopicBus 迁移到 subproto/topicbus

## 背景 / 目标
当前 MyFlowHub-Server 的多个子协议实现位于 `internal/handler/*`，只能在仓库内部引用，不利于后续“可裁切组装 / 拆库复用”的目标架构推进。

本次变更聚焦 TopicBus（SubProto=4）：
- 将 TopicBus handler 从 `internal/handler/topicbus` 迁移到 `subproto/topicbus`
- 装配层、测试与文档改用新 import path
- 保持行为与 wire 不变

## 具体变更内容
### 新增
- `subproto/topicbus/*`：承载 TopicBus 子协议实现（由原目录迁移而来）
- `plan_archive_2026-02-15_varstore-subproto.md`：归档上一轮 workflow 的计划文档，确保可审计

### 修改
- `modules/hub.go`：默认模块集合改用 `subproto/topicbus`
- `tests/topicbus_handler_test.go`：import 切换到 `subproto/topicbus`
- `docs/4-topicbus.md`：实现路径与集成提示更新为 `subproto/topicbus`
- `subproto/topicbus/types.go`：协议常量/类型改为直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/topicbus`（减少对 Server 兼容壳的耦合）
- `plan.md`：更新为本次 workflow 的计划与验收标准

### 删除
- `internal/handler/topicbus/*`：旧实现目录（已迁移后移除）

## 对应计划任务映射（plan.md）
- TB0：归档旧 plan 并更新本 workflow 文档
- TB1：迁移 topicbus 到 `subproto/topicbus`
- TB2：`subproto/topicbus` 直连 `MyFlowHub-Proto` 协议包
- TB3：`modules` 装配切换到新路径
- TB4：测试切换到新路径
- TB5：文档同步更新
- TB6：清理旧目录与引用
- TB7：全量回归
- TB8：Code Review + 归档变更（本文档）

## 关键设计决策与权衡
1) **采用“最小迁移”而非引入 wrapper**
   - 直接 `git mv` 迁移目录，避免额外间接层，降低维护成本与行为漂移风险。

2) **subproto 直接依赖 MyFlowHub-Proto**
   - `protocol/topicbus` 在 Server 内仍保留为兼容壳，但 `subproto/topicbus` 直接引用 Proto 协议包，便于后续将 `subproto/topicbus` 抽成独立库（解耦方向更明确）。

3) **避免顺手重构发送/响应逻辑**
   - 本次不将 topicbus 的发送/响应构造统一迁移到 `subproto/kit`，以避免对 header/source/target 等细节产生潜在行为差异；待后续统一收敛时再做专项 PR。

## 测试与验证
- Windows：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`
  - 结果：通过（包含 `tests/topicbus_handler_test.go`）。
- Linux：按当前约定暂不验收。

## 潜在影响
- 对外功能与 wire 协议不应有变化；本次仅涉及包路径迁移与依赖收敛。
- 若外部项目（理论上不应发生）以 `replace` 方式引用 Server 仓库并 import `internal/handler/topicbus`，会因 `internal` 限制无法编译；建议统一使用 `modules.DefaultHub` 或公开的 `subproto/*`。

## 回滚方案
- 直接 `git revert` 本次合并的提交即可回滚（目录迁移与引用改动均在同一变更集内）。

