# 2026-02-16 - Auth 迁移到 subproto/auth

## 变更背景 / 目标
MyFlowHub-Server 的部分子协议实现位于 `internal/handler/*`，只能在仓库内部引用，不利于后续“可裁切组装 / 拆库复用”的目标架构推进。

本次变更聚焦 Auth（SubProto=2）：
- 将 Auth handler 从 `internal/handler/auth` 迁移到 `subproto/auth`
- 装配层与测试切换到新 import path
- `subproto/auth` 直接依赖 `MyFlowHub-Proto` 的协议包（减少对 Server 兼容壳耦合）
- 保持行为与 wire 不变（不改动作语义 / 不改签名与持久化 / 不改路由与权限）

## 具体变更内容（新增 / 修改 / 删除）
### 新增
- `subproto/auth/*`：承载 Auth 子协议实现（由原目录迁移而来）
- `docs/plan_archive/plan_archive_2026-02-16_file-subproto.md`：归档上一轮 workflow 的计划文档，确保可审计

### 修改
- `modules/hub.go`：默认模块集合改用 `subproto/auth`
- `subproto/auth/types.go`：协议常量/类型改为直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/auth`
- `tests/*`：import 路径切换到 `subproto/auth`
- `docs/2-auth.md`：实现路径描述更新为 `subproto/auth`（语义不变）
- `plan.md`：更新为本次 workflow 的计划、验收标准与执行记录

### 删除
- `internal/handler/auth/*`：旧实现目录（已迁移后移除）

## 对应 plan.md 任务映射
- AU0：归档旧 plan 并更新本 workflow 文档
- AU1：迁移 auth 到 `subproto/auth`
- AU2：`subproto/auth` 直连 `MyFlowHub-Proto` 协议包
- AU3：modules + tests 引用切换到新路径
- AU4：文档路径同步（`docs/2-auth.md`）
- AU5：全量回归
- AU6：Code Review + 归档变更（本文档）

## 关键设计决策与权衡
1) **采用“最小迁移”而非 wrapper**
   - 直接 `git mv` 迁移目录，避免额外间接层，降低维护成本与行为漂移风险。

2) **subproto 直接依赖 MyFlowHub-Proto**
   - `protocol/auth` 在 Server 内仍保留为兼容壳，但 `subproto/auth` 直接引用 Proto 协议包，便于后续将 `subproto/auth` 抽成独立库（解耦方向更明确）。

3) **同步切换 tests 与文档**
   - Auth 相关测试与文档明确依赖旧路径；本次一并切换，确保编译/可读性与可接手性。

## 测试与验证方式 / 结果
- Windows：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`
  - 结果：通过（包含 `tests/auth_handler_test.go` 与 integration 测试）。
- Linux：按当前约定暂不验收。

## 潜在影响与回滚方案
### 潜在影响
- 对外功能与 wire 协议不应有变化；本次仅涉及包路径迁移与依赖收敛。
- 历史归档 `docs/change/*` 可能仍包含 `internal/handler/auth` 的文字描述，这是当时版本的事实，不作为本次变更验收对象。

### 回滚方案
- 直接 `git revert` 本次合并的提交即可回滚（目录迁移与引用改动均在同一变更集内）。

## Code Review（结论）
- 需求覆盖：通过（仅路径迁移与依赖收敛；wire/行为不变；装配/测试/文档已切换）。
- 架构合理性：通过（`subproto/auth` 作为可装配子协议落点；依赖方向更清晰：subproto → proto）。
- 性能风险：通过（无新增热路径逻辑；仅包路径/import 调整；无额外 I/O/锁/循环）。
- 可读性与一致性：通过（保留原实现与命名；差异集中在目录与 import）。
- 可扩展性与配置化：通过（迁移到 `subproto/*` 后更易拆库/裁切；`modules` 装配保持集中）。
- 稳定性与安全：通过（不改签名/持久化/权限/路由语义；安全默认不变）。
- 测试覆盖：通过（`go test ./... -count=1 -p 1` 在 Windows 通过；含单测与 integration）。

