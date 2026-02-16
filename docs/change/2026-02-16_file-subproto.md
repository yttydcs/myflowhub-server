# 2026-02-16 - File 迁移到 subproto/file

## 变更背景 / 目标
MyFlowHub-Server 的部分子协议实现位于 `internal/handler/*`，只能在仓库内部引用，不利于后续“可裁切组装 / 拆库复用”的目标架构推进。

本次变更聚焦 File（节点间文件传输）：
- 将 File handler 从 `internal/handler/file` 迁移到 `subproto/file`
- 装配层切换到新 import path
- `subproto/file` 直接依赖 `MyFlowHub-Proto` 的协议包（减少对 Server 兼容壳耦合）
- 保持行为与 wire 不变（不改帧格式 / 不改 action/op / 不改权限语义）

## 具体变更内容
### 新增
- `subproto/file/*`：承载 File 子协议实现（由原目录迁移而来）
- `plan_archive_2026-02-16_flow-subproto.md`：归档上一轮 workflow 的计划文档，确保可审计

### 修改
- `modules/hub.go`：默认模块集合改用 `subproto/file`
- `subproto/file/types.go`：协议常量/类型改为直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/file`
- `plan.md`：更新为本次 workflow 的计划、验收标准与执行记录

### 删除
- `internal/handler/file/*`：旧实现目录（已迁移后移除）

## 对应计划任务映射（plan.md）
- FI0：归档旧 plan 并更新本 workflow 文档
- FI1：迁移 file 到 `subproto/file`
- FI2：`subproto/file` 直连 `MyFlowHub-Proto` 协议包
- FI3：`modules` 装配切换到新路径
- FI4：清理旧目录与引用
- FI5：全量回归
- FI6：Code Review + 归档变更（本文档）

## 关键设计决策与权衡
1) **采用“最小迁移”而非 wrapper**
   - 直接 `git mv` 迁移目录，避免额外间接层，降低维护成本与行为漂移风险。

2) **subproto 直接依赖 MyFlowHub-Proto**
   - `protocol/file` 在 Server 内仍保留为兼容壳，但 `subproto/file` 直接引用 Proto 协议包，便于后续将 `subproto/file` 抽成独立库（解耦方向更明确）。

3) **不改帧格式与传输语义**
   - 维持 `payload[0]` 的 Kind 分流与 `binHeaderV1` 帧头格式，避免对存量节点/客户端造成 wire 不兼容。

## 测试与验证方式 / 结果
- Windows：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`
  - 结果：通过（`subproto/file` 包可编译并被全量回归覆盖）。
- Linux：按当前约定暂不验收。

## 潜在影响与回滚方案
### 潜在影响
- 对外功能与 wire 协议不应有变化；本次仅涉及包路径迁移与依赖收敛。
- 历史归档 `docs/change/*` 可能仍包含 `internal/handler/file` 的文字描述，这是当时版本的事实，不作为本次变更验收对象。

### 回滚方案
- 直接 `git revert` 本次合并的提交即可回滚（目录迁移与引用改动均在同一变更集内）。

## Code Review（结论）
- 需求覆盖：通过（仅路径迁移与依赖收敛；wire/行为不变；装配层已切换）。
- 架构合理性：通过（`subproto/file` 作为可装配子协议落点；依赖方向更清晰：subproto → proto）。
- 性能风险：通过（无新增热路径逻辑；仅包路径/import 调整；无额外 I/O/锁/循环）。
- 可读性与一致性：通过（保留原实现与命名；差异集中在目录与 import）。
- 可扩展性与配置化：通过（迁移到 `subproto/*` 后更易拆库/裁切；`modules` 装配保持集中）。
- 稳定性与安全：通过（不改权限/路径清洗/会话管理语义；默认安全策略不变）。
- 测试覆盖：通过（`go test ./... -count=1 -p 1` 在 Windows 通过；当前无专属单测，依赖全量回归覆盖编译与集成）。

