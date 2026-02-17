# Plan - Exec / Flow 响应帧统一为 MajorOKResp（PR11-Resp-MajorOKResp）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`fix/resp-major-okresp`
- Worktree：`d:\project\MyFlowHub3\worktrees\resp-major-okresp`
- Base：`main`
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）
- 约束（本 workflow 明确边界，避免范围外扩）：
  - 只改 `exec.call_resp` 与 `flow.*_resp` 的 HeaderTcp `Major`：统一为 `MajorOKResp`（成功/失败均如此）。
  - 不改 wire（action/JSON/SubProto 不变）。
  - big-bang：不提供兼容开关（你已确认）。
  - 不改其它子协议（Auth/VarStore/TopicBus/File 等不在本轮范围）。
  - 不改 `MyFlowHub-Core`（Core 的快速转发规则已具备，本轮只对齐上层帧语义）。
  - 不改 `MyFlowHub-SDK`（本轮不引入 await 规则变更）。
  - 所有实现性改动只在本 worktree 内完成；`repo/` 仅用于合并/推送/集成验证。

## 当前状态（事实，可审计）
- Core 路由统一规则（已落地）：
  - `MajorCmd`：必须进入 handler（逐跳可见），Core 不做自动转发/广播。
  - `MajorMsg/MajorOKResp/MajorErrResp`：按 `TargetID` 由 Core 快速转发/广播（并在发生转发时递减 `hop_limit`）。
  - 代码位置：`MyFlowHub-Core/process/prerouting.go`、`MyFlowHub-Core/process/dispatcher.go`。
- 当前 Server 侧仍有两类“返回帧”使用 `MajorCmd`：
  - `exec.call_resp`：`subproto/exec/handler.go` 的 `sendCallRespToNode` 构造头部使用 `MajorCmd`。
  - `flow.*_resp`：`subproto/flow/handler.go` 的 `sendCtrlToNode` 构造头部使用 `MajorCmd`。
- 影响（结构性问题）：
  - 这些返回帧在跨节点链路中需要依赖子协议 handler 逐跳转发，而不是走 Core 快速转发；协议实现与 Core 统一框架规则不一致，耦合增大、后续重构成本变高。

---

## 1) 需求分析

### 目标
- 将 `exec.call_resp` 与 `flow.*_resp` 统一标记为 `MajorOKResp`，使其在跨节点链路中由 Core 按 `TargetID` 快速转发（中间节点不需要进 handler 转发）。
- 请求帧保持 `MajorCmd`（仍逐跳可见并进入 handler），不影响逐级授权/裁决路径。
- 失败响应仍使用 `MajorOKResp`，错误通过 payload 的 `code/msg` 表达（你已确认）。

### 范围（必须 / 可选 / 不做）
#### 必须（本 PR）
- `subproto/exec`：
  - `exec.call_resp` 的出站 HeaderTcp：`Major=MajorOKResp`（成功/失败均如此）。
- `subproto/flow`：
  - `set_resp/run_resp/status_resp/list_resp/get_resp` 的出站 HeaderTcp：`Major=MajorOKResp`（成功/失败均如此）。
- 单测覆盖：断言上述响应帧的 `hdr.Major()==MajorOKResp`。
- 回归：`go test ./... -count=1 -p 1`（Windows）。

#### 可选（本 PR 如无额外风险）
- 同步更新文档：
  - `docs/6-flow.md`：明确 `flow.*_resp` 使用 `MajorOKResp`，依赖 Core 快速转发。
  - `docs/7-exec.md`：明确 `exec.call_resp` 使用 `MajorOKResp`，依赖 Core 快速转发。

#### 不做（本 PR）
- 不改 `exec.call` 与 `flow.*` 请求帧（仍使用 `MajorCmd`）。
- 不引入 `MajorErrResp` 标准化：失败仍以 `MajorOKResp` + payload `code` 表达。
- 不统一其它子协议的 Major（后续另起 workflow）。
- 不新增“兼容开关”（big-bang 已确认）。

### 使用场景
- 多级 hub 网络中：
  - `exec.call_resp` 从目标节点返回到 `executor_node`；
  - `flow.*_resp` 从裁决/执行节点返回到请求方节点；
  - 中间节点应“只转发，不理解”。

### 验收标准
- `exec.call_resp` 与 `flow.*_resp`：发送时 header 的 `Major==MajorOKResp`。
- 单测覆盖关键路径并通过。
- `go test ./... -count=1 -p 1` 通过（Windows）。

### 风险
- mixed-version 风险：旧节点若仍产出 `MajorCmd` 的 `*_resp/call_resp`，行为会不一致；本 PR 按 big-bang 执行（你已确认）。

---

## 2) 架构设计（分析）

### 总体方案（采用）
- 在 Server 协议实现处修正响应帧 Major：
  - Exec：`sendCallRespToNode` 改为 `MajorOKResp`。
  - Flow：所有 `*_resp` 的下发发送点改为 `MajorOKResp`。
- 依赖 Core `PreRoutingProcess` 对 `MajorOKResp` 在 `target!=local` 时的快速转发与 `hop_limit` 递减（既有能力，本轮复用）。
- 保留现有 handler 内“逐跳转发兼容逻辑”不删除（降低风险，不扩大改动面）。

### 备选对比（为什么不选）
- 备选 A：保持 `MajorCmd`，继续依赖 handler 转发 `*_resp/call_resp`
  - 问题：与 Core “响应帧走快速转发” 的统一框架规则冲突，协议实现需要重复维护转发逻辑；未来拆库/裁切时耦合更重。
- 备选 B：引入 `MajorErrResp` 标准化（失败响应用 Err）
  - 本轮不做：避免跨协议/跨端同步成本；保持现有 payload `code` 语义。

### 模块职责
- `MyFlowHub-Core/process/prerouting.go`：数据面/响应帧（Msg/OK/Err）统一快速转发与 hop_limit 递减（既有能力，本轮复用）。
- `subproto/exec/handler.go`：`call_resp` 的出站头部 Major 纠正为 `MajorOKResp`。
- `subproto/flow/handler.go`：所有 `*_resp` 的出站头部 Major 纠正为 `MajorOKResp`。
- `tests/*`：锁定行为，避免未来回归。
- `docs/*`（可选）：明确 Major 与转发边界说明。

### 数据 / 调用流（关键链路）
- Exec：
  - executor → `exec.call`（MajorCmd）→ …（逐级裁决/转发）→ target → `exec.call_resp`（MajorOKResp, TargetID=executor）→ Core 快速转发 → executor 消费（broker）。
- Flow：
  - requester → `flow.*`（MajorCmd）→ …（逐级裁决/转发/执行）→ responder → `flow.*_resp`（MajorOKResp, TargetID=requester）→ Core 快速转发 → requester 端消费（通常是客户端/上层）。

### 错误与安全
- 仅更改 header.Major，不改变权限判定/逐级裁决/请求参数校验。
- 仍依赖 Core 的 `sourceMismatch` 机制防止非父连接伪造 SourceID（本轮不改）。
- `hop_limit` 仍由 Core 在转发时递减，避免环路/风暴（本轮不改）。

### 性能与测试策略
- 性能收益：减少中间节点 handler 的 JSON 解包 + action 查找（响应帧改为 Core 快转发）。
- 测试策略：
  - 单测断言 `exec.call_resp` / `flow.*_resp` 的 header `Major==MajorOKResp`；
  - Windows 回归 `go test ./...`。

### 可扩展性设计点（后续方向，不在本 PR）
- 可按相同准则逐个子协议审计：
  - “端到端返回帧/响应帧” → `MajorOKResp`；
  - “逐跳可见需裁决/翻译的控制帧” → `MajorCmd`；
  - “广播/通知” → `MajorMsg`。

---

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（你已确认：big-bang + 顺带调整 flow 的 `*_resp`）。

### RMO1 - Exec：call_resp 改为 MajorOKResp
- 目标：让 `exec.call_resp` 作为响应帧走 Core 快速转发（中间节点无需进 handler 转发）。
- 涉及模块 / 文件：
  - `subproto/exec/handler.go`
- 验收条件：
  - `sendCallRespToNode` 构造的 header：`Major==MajorOKResp`。
- 测试点：见 RMO3。
- 回滚点：revert 本提交。

### RMO2 - Flow：所有 *_resp 改为 MajorOKResp
- 目标：让 `flow.*_resp` 作为响应帧走 Core 快速转发（中间节点无需进 handler 转发）。
- 涉及模块 / 文件：
  - `subproto/flow/handler.go`
- 覆盖 action：
  - `set_resp/run_resp/status_resp/list_resp/get_resp`
- 验收条件：
  - 上述 `*_resp` 出站 header：`Major==MajorOKResp`。
- 测试点：见 RMO3。
- 回滚点：revert 本提交。

### RMO3 - 单测：断言 Exec/Flow 响应帧 MajorOKResp
- 目标：锁定行为，避免未来回归。
- 涉及文件：
  - `tests/test_stubs.go`（增强 stubServer 记录 major）
  - 新增：`tests/exec_handler_test.go`
  - 新增：`tests/flow_handler_test.go`
- 验收条件：
  - 至少覆盖：
    - 本地执行 `exec.call` 触发的 `call_resp`：`MajorOKResp`
    - 非法 `flow.set` 触发的 `set_resp`：`MajorOKResp`
- 回滚点：revert 本提交。

### RMO4 - 文档（可选）：补齐 Major 与转发边界说明
- 涉及文件：
  - `docs/6-flow.md`
  - `docs/7-exec.md`
- 验收条件：
  - 文档明确 `*_resp/call_resp` 为 `MajorOKResp`，依赖 Core 快转发。

### RMO5 - 回归测试（Windows）
- 命令（统一临时目录与并发）：
  - `$env:GOTMPDIR='d:\project\MyFlowHub3\.tmp\gotmp'`
  - `New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null`
  - `go test ./... -count=1 -p 1`
- 验收条件：通过。

### RMO6 - Code Review（阶段 3.3）+ 归档变更（阶段 4）
- 归档文件（worktree 内）：
  - `docs/change/2026-02-16_exec-flow-resp-major-ok.md`
- 验收条件：
  - Review 覆盖：需求/架构/性能/安全/测试；
  - 归档包含：任务映射、关键决策、测试命令与回滚方案。

