# 2026-03-06 VarStore：规范文档对齐逐跳回程（Server）

## 变更规模
- 级别：较大（文档规范与实现策略收敛，影响多仓联动与后续依赖升级）

## 背景 / 目标
VarStore 在多 hop 场景下需要“逐跳可见回程”，以支撑中间节点的：

- 去重扇出
- 沿途缓存
- 并发写入 1:1 匹配（写操作不得按 key 模糊匹配）

但 `docs/3-varstore.md` 中对 `Major`、回程语义、`SourceID/TargetID`、以及若干规则点存在不一致/表述不清。

目标：把 `docs/3-varstore.md` 与已确认的最终语义对齐，作为后续实现与联动升级的单一事实来源。

## 具体变更
更新 `docs/3-varstore.md`，覆盖以下关键点：

- `*_resp/assist_*_resp` 必须使用 `MajorCmd`（逐跳可见），不再建议 `MajorOKResp/MajorErrResp`。
- `SourceID` 为原始 actor，端到端保留（assist/up/notify/resp 均不得改写）。
- `TargetID`：
  - 命令类（非 resp）可指向 owner/Hub；
  - 响应类遵循逐跳回程，仅指向下一跳。
- `set/revoke` 并发写匹配：每次上送分配新的上行 msg_id，并建立映射；回程按上行 msg_id 命中并恢复下游 msg_id/trace_id。
- 规则收敛：
  - `set.value` 禁止空/纯空白；
  - `set_resp code=1` 必带 `value`，仅成功更新缓存；
  - `list` 空集合成功返回（语义 `names=[]`）；
  - `subscribe.data.subscriber` 只允许 0 或等于 `SourceID`（默认不允许代订阅）；
  - `private` 允许权限例外；
  - `notify_set/notify_revoke` 下行链路要求“转发 + 本地处理（缓存/订阅推送）”。

## 关联任务映射（plan.md）
- SRV-1：更新 VarStore 规范文档（已完成）
- SRV-2：依赖联动（待 SubProto 发布可解析版本后执行）

## 测试与验证
- 文档变更：人工审阅 + 与 SubProto/SDK 实现交叉核对。

## Code Review
- 结论：通过

## 潜在影响与回滚
- 影响：文档对实现策略更严格（尤其是 MajorCmd 与 msg_id 映射），需要 SDK/客户端 await 配套支持。
- 回滚：回退 `docs/3-varstore.md` 与本文件即可，不影响运行代码。
