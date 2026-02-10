# Plan - HeaderTcp v2（32B）+ Core 路由统一（Server）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/hdrtcp-v2`
- Worktree：`d:\\project\\MyFlowHub3\\worktrees\\hdrtcp-v2\\MyFlowHub-Server`
- 目标 PR：PR1（跨 3 个 repo 同步提交/合并）

## 项目目标
1) 配合 Core 完成 **HeaderTcp v2（32B）big-bang** 升级，确保 server 收发与测试全部通过。  
2) 配合 Core 路由框架规则统一：`MajorCmd` 逐跳进入 handler；`MajorMsg/OK/Err` 走 Core 快速转发。  
3) 为后续子协议“可裁切/可组装”做准备：server 作为中间层只负责调用子协议库（本 PR 只做地基，不做大拆分）。

## 范围
### 必须（PR1）
- 适配 Core 的 `IHeader` v2 接口变更（编译通过）
- 使用 HeaderTcp v2 编解码（wire 32B）
- 更新/补齐 server 测试与文档（`docs/core.md` 的路由规则描述需同步）

### 不做（本 PR）
- 子协议拆独立 repo/go module（另起 PR2+）
- Linux 构建/验收（用户已允许忽略）

## 已确认的关键决策（来自阶段 2）
- 兼容策略：**S3 / big-bang**；切换后 **v1 不再兼容**。
- HeaderTcp v2：**32B（+8B）**。
- 路由框架规则：**MajorCmd 不由 Core 自动转发，必须进 handler；MajorMsg/OK/Err 走 Core 快速转发**。
- 语义基线：`TargetID=0` 仅表示“下行广播不回父”，不能表示上送父节点。

## 问题清单（阻塞：是）
> 与 Core/Win 共用的 wire 细节确认项；未确认禁止进入阶段 3.2。

1) HeaderTcp v2 `magic` 值（建议 `0x4D48`）是否确认？
2) `hop_limit` 默认值/语义是否确认？（建议默认 `16`，转发递减）
3) `trace_id` 生成策略是否确认？（建议发送侧自动补齐随机 uint32；响应继承；转发不改）
4) `timestamp` 单位是否确认？（建议保持 Unix 秒 `uint32`）

## 任务清单（Checklist）

### S1 - 适配 Core HeaderTcp v2 / IHeader 变更
- 目标：修复 `go test ./...` 期间由 Core 接口变更引起的编译错误；确保 server 所有 header 构造、clone、response helper 与 codec 使用 v2。
- 涉及模块/文件（预期）：
  - `internal/**`（所有使用 `core.IHeader` / `header.HeaderTcp` 的位置）
  - `tests/**`（大量构造 header 的用例）
  - `internal/login_server/*`（登录链路对路由/SourceID 约束敏感）
- 验收条件：
  - `go test ./...` 通过。
  - 与 Win 联调的冒烟链路可跑通（见 Win 侧 smoke 步骤）。
- 测试点：
  - `go test ./... -count=1`
  - 重点回归：auth / varstore / topicbus / file 的集成测试（如存在）。
- 回滚点：
  - 将适配拆为独立提交；可 revert。

### S2 - 路由语义与 Major 使用自检（避免隐式协议特例）
- 目标：确保 server 侧各子协议对 Major 的使用符合框架规则（控制面用 Cmd；数据/响应用 Msg/OK/Err），避免“用 payload[0] 决定路由”的隐式依赖扩散。
- 涉及模块/文件（预期）：
  - `internal/handler/**`（尤其 file：CTRL/DATA/ACK）
  - `protocol/**`（如需补充常量/注释，保持最小化）
- 验收条件：
  - server 端不依赖 Core 的“特定 SubProto 特判”才能正确路由（例如 file CTRL 不再需要 Core 特判）。
- 测试点：
  - file：CTRL 从子节点上送应逐跳进入 handler；DATA 仍能快速转发且吞吐不明显下降。
- 回滚点：
  - 若发现协议 Major 使用不一致，先修正协议侧，避免回退 Core 框架规则。

### S3 - 文档同步（core.md）
- 目标：更新 `docs/core.md` 中关于 PreRouting 与 Major 的描述，使其与新框架规则一致（可审计/可交接）。
- 涉及模块/文件（预期）：
  - `docs/core.md`
- 验收条件：
  - 文档不再描述“file CTRL 特判”之类已移除规则；明确 Major 分流与 `TargetID=0` 语义。
- 回滚点：
  - 文档变更独立提交；可 revert。

### S4 - Code Review（阶段 3.3）与归档（阶段 4）
- 目标：完成 Review 清单并在本 worktree 根目录创建 `docs/change/2026-02-10_hdrtcp-v2.md`。
- 验收条件：
  - Review 逐项“通过/不通过”结论明确；不通过则回到阶段 3.2 修正。
  - 归档文档包含：背景/目标、具体变更、任务映射（S1-S3）、关键决策与权衡、测试结果、影响与回滚方案。

## 依赖关系
- 依赖 Core 的 v2 头部与路由框架落地；同时 Win 需要同步升级，否则无法联调。

## 风险与注意事项
- wire 破坏性变更必须三端同步；建议在本地联调通过后再分别推远端 PR。
- server 测试若依赖固定头长（24B）需要全部更新为 32B。

