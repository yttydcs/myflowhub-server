# 2026-02-16 - Flow 迁移到 subproto/flow

## 背景 / 目标
MyFlowHub-Server 的部分子协议实现位于 `internal/handler/*`，只能在仓库内部引用，不利于后续“可裁切组装 / 拆库复用”的目标架构推进。

本次变更聚焦 Flow（SubProto=6）：
- 将 Flow handler 从 `internal/handler/flow` 迁移到 `subproto/flow`
- 装配层切换到新 import path
- `subproto/flow` 直接依赖 `MyFlowHub-Proto` 的协议包（减少对 Server 兼容壳耦合）
- 保持行为与 wire 不变（不改调度/落盘/权限语义）

## 具体变更内容
### 新增
- `subproto/flow/*`：承载 Flow 子协议实现（由原目录迁移而来，包含 `graph_test.go`）
- `plan_archive_2026-02-16_exec-subproto.md`：归档上一轮 workflow 的计划文档，确保可审计

### 修改
- `modules/hub.go`：默认模块集合改用 `subproto/flow`
- `subproto/flow/types.go`：协议常量/类型改为直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/flow`
- `plan.md`：更新为本次 workflow 的计划、验收标准与执行记录

### 删除
- `internal/handler/flow/*`：旧实现目录（已迁移后移除）

## 对应计划任务映射（plan.md）
- FL0：归档旧 plan 并更新本 workflow 文档
- FL1：迁移 flow 到 `subproto/flow`
- FL2：`subproto/flow` 直连 `MyFlowHub-Proto` 协议包
- FL3：`modules` 装配切换到新路径
- FL4：清理旧目录与引用
- FL5：全量回归
- FL6：Code Review + 归档变更（本文档）

## 关键设计决策与权衡
1) **采用“最小迁移”而非 wrapper**
   - 直接 `git mv` 迁移目录，避免额外间接层，降低维护成本与行为漂移风险。

2) **subproto 直接依赖 MyFlowHub-Proto**
   - `protocol/flow` 在 Server 内仍保留为兼容壳，但 `subproto/flow` 直接引用 Proto 协议包，便于后续将 `subproto/flow` 抽成独立库（解耦方向更明确）。

3) **不改 BindServer/调度/落盘等行为**
   - Flow handler 依赖启动期 `BindServer(core.IServer)` 绑定，本次不调整绑定机制与启动顺序，避免引入隐蔽的初始化问题。

4) **不顺手迁移到 subproto/kit**
   - 本次仅做路径迁移与依赖收敛，避免对 header/source/target/转发细节产生潜在行为差异；统一收敛应另起专项 PR。

## 测试与验证
- Windows：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`
  - 结果：通过（包含 `subproto/flow/graph_test.go`）。
- Linux：按当前约定暂不验收。

## 潜在影响
- 对外功能与 wire 协议不应有变化；本次仅涉及包路径迁移与依赖收敛。
- 历史归档 `docs/change/*` 可能仍包含 `internal/handler/flow` 的文字描述，这是当时版本的事实，不作为本次变更验收对象。

## 回滚方案
- 直接 `git revert` 本次合并的提交即可回滚（目录迁移与引用改动均在同一变更集内）。

## Code Review（结论）
- 需求覆盖：通过（仅路径迁移与依赖收敛；wire/行为不变；装配层已切换）。
- 架构合理性：通过（`subproto/flow` 作为可装配子协议落点；依赖方向更清晰：subproto → proto）。
- 性能风险：通过（无新增热路径逻辑；仅包路径/import 调整；无额外 I/O/锁/循环）。
- 可读性与一致性：通过（保留原实现与命名；差异集中在目录与 import）。
- 可扩展性与配置化：通过（迁移到 `subproto/*` 后更易拆库/裁切；`modules` 装配保持集中）。
- 稳定性与安全：通过（不改权限/调度/落盘语义；启动期 `BindServer` 机制保持不变）。
- 测试覆盖：通过（`go test ./... -count=1 -p 1` 在 Windows 通过，包含 `subproto/flow/graph_test.go`）。
