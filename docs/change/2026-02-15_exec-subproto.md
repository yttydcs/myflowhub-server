# 2026-02-15 - Exec 迁移到 subproto/exec

## 背景 / 目标
当前 MyFlowHub-Server 的部分子协议实现位于 `internal/handler/*`，只能在仓库内部引用，不利于后续“可裁切组装 / 拆库复用”的目标架构推进。

本次变更聚焦 Exec（SubProto=7）：
- 将 Exec handler 从 `internal/handler/exec` 迁移到 `subproto/exec`
- 装配层切换到新 import path
- `subproto/exec` 直接依赖 `MyFlowHub-Proto` 的协议包（减少对 Server 兼容壳耦合）
- 保持行为与 wire 不变

## 具体变更内容
### 新增
- `subproto/exec/*`：承载 Exec 子协议实现（由原目录迁移而来）
- `docs/plan_archive/plan_archive_2026-02-15_topicbus-subproto.md`：归档上一轮 workflow 的计划文档，确保可审计

### 修改
- `modules/hub.go`：默认模块集合改用 `subproto/exec`
- `subproto/exec/types.go`：协议常量/类型改为直接依赖 `github.com/yttydcs/myflowhub-proto/protocol/exec`
- `plan.md`：更新为本次 workflow 的计划、验收标准与执行记录

### 删除
- `internal/handler/exec/*`：旧实现目录（已迁移后移除）

## 对应计划任务映射（plan.md）
- EX0：归档旧 plan 并更新本 workflow 文档
- EX1：迁移 exec 到 `subproto/exec`
- EX2：`subproto/exec` 直连 `MyFlowHub-Proto` 协议包
- EX3：`modules` 装配切换到新路径
- EX4：清理旧目录与引用
- EX5：全量回归
- EX6：Code Review + 归档变更（本文档）

## 关键设计决策与权衡
1) **采用“最小迁移”而非 wrapper**
   - 直接 `git mv` 迁移目录，避免额外间接层，降低维护成本与行为漂移风险。

2) **subproto 直接依赖 MyFlowHub-Proto**
   - `protocol/exec` 在 Server 内仍保留为兼容壳，但 `subproto/exec` 直接引用 Proto 协议包，便于后续将 `subproto/exec` 抽成独立库（解耦方向更明确）。

3) **不顺手改 exec 的转发/权限/响应构造**
   - 本次仅做路径迁移与依赖收敛，避免对逐级裁决、hop_limit、TargetID 路由等细节引入行为差异；后续若要统一收敛到 `subproto/kit`，应另起专项 PR。

## 测试与验证
- Windows：
  - `$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'; go test ./... -count=1 -p 1`
  - 结果：通过。
- Linux：按当前约定暂不验收。

## 潜在影响
- 对外功能与 wire 协议不应有变化；本次仅涉及包路径迁移与依赖收敛。
- 历史归档 `docs/change/*` 可能仍包含 `internal/handler/exec` 的文字描述，这是当时版本的事实，不作为本次变更验收对象。

## 回滚方案
- 直接 `git revert` 本次合并的提交即可回滚（目录迁移与引用改动均在同一变更集内）。

