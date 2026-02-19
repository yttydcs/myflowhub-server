# 2026-02-15 - VarStore 迁移到 subproto/varstore

## 背景 / 目标
当前 MyFlowHub-Server 的子协议实现大多位于 `internal/handler/*`，只能在仓库内部引用，不利于后续“可裁切组装 / 拆库复用”的目标架构推进。

本次变更聚焦 VarStore（SubProto=3）：
- 将 VarStore handler 从 `internal/handler/varstore` 迁移到 `subproto/varstore`
- 装配层与测试改用新 import path
- 保持行为与 wire 不变（不做 assist_* / up_* / notify_* 命名收敛）

## 具体变更内容
### 新增
- `subproto/varstore/*`：承载 VarStore 子协议实现（由原目录迁移而来）
- `docs/plan_archive/plan_archive_2026-02-15_default-forward-subproto-forward.md`：归档旧 workflow 的计划文档，确保可审计

### 修改
- `modules/hub.go`：默认模块集合改用 `subproto/varstore`
- 测试：
  - `tests/varstore_handler_test.go`
  - `tests/integration_varstore_end_to_end_test.go`
  - `tests/integration_root_hub_ping_test.go`
- `subproto/varstore/types.go`：协议常量/类型改为直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/varstore`（减少对 Server 兼容壳的耦合）
- `plan.md`：更新为本次 workflow 的计划与验收标准

### 删除
- `internal/handler/varstore/*`：旧实现目录（已迁移后移除）

## 对应计划任务映射（plan.md）
- V0：归档旧 plan 并更新本 workflow 文档
- V1：迁移 varstore 到 `subproto/varstore`
- V2：`subproto/varstore` 直连 `MyFlowHub-Proto` 协议包
- V3：`modules` 装配切换到新路径
- V4：测试切换到新路径
- V5：清理旧目录与引用
- V6：全量回归
- V7：归档变更（本文档）

## 关键设计决策与权衡
1) **采用“最小迁移”而非引入 wrapper**
   - 直接 `git mv` 迁移目录，避免额外间接层，降低维护成本与行为漂移风险。

2) **subproto 直接依赖 MyFlowHub-Proto**
   - `protocol/varstore` 在 Server 内仍保留为兼容壳，但 `subproto/varstore` 直接引用 Proto 仓库协议包，便于后续将 `subproto/varstore` 抽成独立库（解耦方向更明确）。

3) **避免顺手重构发送/响应逻辑**
   - 本次不将 varstore 的头部构造/发送逻辑迁移到 `subproto/kit`，以避免对 hop/source/target 等细节产生潜在行为差异；待后续统一收敛时再做专项 PR。

## 测试与验证
- Windows：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`
  - 结果：通过（包含 varstore 单测与集成测试）。
- Linux：按当前约定暂不验收。

## 潜在影响
- 对外功能与 wire 协议不应有变化；本次仅涉及包路径迁移与依赖收敛。
- 若外部项目以 `replace` 方式引用 Server 仓库并 import `internal/handler/varstore`（理论上不应发生，因为 `internal` 限制），则会编译失败；建议统一使用 `modules.DefaultHub` 或后续公开的 `subproto/*`。

## 回滚方案
- 直接 `git revert` 本次合并的提交即可回滚（目录迁移与引用改动均在同一变更集内）。

