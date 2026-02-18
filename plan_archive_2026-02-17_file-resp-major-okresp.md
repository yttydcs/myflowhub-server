# Plan - File 控制响应帧统一为 MajorOKResp（PR12-File-Resp-MajorOKResp）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`fix/file-resp-major-okresp`
- Worktree：`d:\project\MyFlowHub3\worktrees\file-resp-major-okresp`
- Base：`main`
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）
- 约束（本 workflow 明确边界，避免范围外扩）：
  - 只改 `file.read_resp` / `file.write_resp` 的 HeaderTcp `Major`：统一为 `MajorOKResp`（成功/失败均如此）。
  - 不改 wire（action/JSON/SubProto 不变）。
  - big-bang：不提供兼容开关（你已确认：无需兼容旧客户端）。
  - 不改其它子协议（Auth/VarStore/TopicBus/Exec/Flow/Management 等不在本轮范围）。
  - 不改 `MyFlowHub-Core`（Core 的快速转发规则已具备，本轮只对齐 File 响应帧语义）。
  - 不改 `MyFlowHub-SDK` / `MyFlowHub-Win` / `MyFlowHub-Proto`。
  - 所有实现性改动只在本 worktree 内完成；`repo/` 仅用于合并/推送/集成验证。

## 当前状态（事实，可审计）
- Core 路由统一规则（已落地）：
  - `MajorCmd`：必须进入 handler（逐跳可见），Core 不做自动转发/广播。
  - `MajorMsg/MajorOKResp/MajorErrResp`：按 `TargetID` 由 Core 快速转发/广播（并在发生转发时递减 `hop_limit`）。
  - 代码位置：`MyFlowHub-Core/process/prerouting.go`、`MyFlowHub-Core/process/dispatcher.go`。
- 当前 Server 侧仍有一类“返回帧”使用 `MajorCmd`：
  - `file.read_resp` / `file.write_resp`：`subproto/file/handler.go` 的 `sendCtrlToNode` 构造头部使用 `MajorCmd`。
- 影响（结构性问题）：
  - `file.*_resp` 在跨节点链路中需要依赖 File handler 逐跳转发，而不是走 Core 快速转发；与统一框架规则不一致，耦合增大、后续重构成本变高。

---

## 1) 需求分析

### 目标
- 将 `file.read_resp` / `file.write_resp` 统一标记为 `MajorOKResp`，使其在跨节点链路中由 Core 按 `TargetID` 快速转发（中间节点不需要进 File handler 转发）。
- 请求帧保持 `MajorCmd`（仍逐跳可见并进入 handler），不影响逐级权限裁决/转交链路。
- 失败响应仍使用 `MajorOKResp`，错误通过 payload 的 `code/msg` 表达（你已确认：无需兼容旧客户端）。

### 范围（必须 / 可选 / 不做）
#### 必须（本 PR）
- `subproto/file`：
  - `read_resp` / `write_resp` 的出站 HeaderTcp：`Major=MajorOKResp`（成功/失败均如此）。
- 单测覆盖：断言 `read_resp` / `write_resp` 的 header `Major==MajorOKResp`。
- 回归：`go test ./... -count=1 -p 1`（Windows）。

#### 可选（本 PR 如无额外风险）
- 同步更新文档：
  - `docs/5-file.md`：明确 `read_resp/write_resp` 使用 `MajorOKResp`，依赖 Core 快速转发。

#### 不做（本 PR）
- 不改 `read/write` 请求帧（仍使用 `MajorCmd`）。
- 不改 `DATA/ACK` 的发送（仍使用 `MajorMsg`，端到端传输）。
- 不引入 `MajorErrResp` 标准化：失败仍以 `MajorOKResp` + payload `code` 表达。
- 不统一其它子协议的 Major（后续另起 workflow）。
- 不新增“兼容开关”（big-bang 已确认）。

### 使用场景
- 多级 hub 网络中：
  - `file.read_resp/write_resp` 从裁决/执行节点返回到请求方节点；
  - 中间节点应“只转发，不理解”。

### 验收标准
- `file.read_resp/write_resp`：发送时 header 的 `Major==MajorOKResp`。
- 单测覆盖关键路径并通过。
- `go test ./... -count=1 -p 1` 通过（Windows）。

### 风险
- mixed-version 风险：旧节点若仍产出 `MajorCmd` 的 `file.*_resp`，行为会不一致；本 PR 按 big-bang 执行（你已确认）。

---

## 2) 架构设计（分析）

### 总体方案（采用）
- 在 Server 的 File 协议实现处修正响应帧 Major：
  - `subproto/file/handler.go`：`sendCtrlToNode` 改为 `MajorOKResp`（覆盖 `read_resp/write_resp`）。
- 依赖 Core `PreRoutingProcess` 对 `MajorOKResp` 在 `target!=local` 时的快速转发与 `hop_limit` 递减（既有能力，本轮复用）。
- 不删除 File handler 现有“逐跳转发”逻辑（降低风险，不扩大改动面；mixed-version 时仍可能作为兜底）。

### 备选对比（为什么不选）
- 备选 A：保持 `MajorCmd`，继续依赖 File handler 转发 `read_resp/write_resp`
  - 问题：与 Core “响应帧走快速转发” 的统一框架规则冲突，协议实现需要重复维护转发逻辑；未来拆库/裁切时耦合更重。
- 备选 B：引入 `MajorErrResp` 标准化（失败响应用 Err）
  - 本轮不做：避免跨协议/跨端同步成本；保持现有 payload `code` 语义。

### 模块职责
- `MyFlowHub-Core/process/prerouting.go`：数据面/响应帧（Msg/OK/Err）统一快速转发与 hop_limit 递减（既有能力，本轮复用）。
- `subproto/file/handler.go`：`read_resp/write_resp` 的出站头部 Major 纠正为 `MajorOKResp`。
- `tests/*`：锁定行为，避免未来回归。
- `docs/*`（可选）：明确 Major 与转发边界说明。

### 数据 / 调用流（关键链路）
- File：
  - requester → `file.read/write`（MajorCmd, kind=CTRL）→ …（逐级判权/转交）→ provider/consumer → `file.read_resp/write_resp`（MajorOKResp, kind=CTRL, TargetID=requester）→ Core 快速转发 → requester 消费。
  - DATA/ACK：`MajorMsg`（kind=DATA/ACK）端到端按 `SourceID/TargetID` 由 Core 快速转发。

### 错误与安全
- 仅更改 header.Major，不改变权限判定/逐级裁决/请求参数校验。
- 仍依赖 Core 的 `sourceMismatch` 机制防止非父连接伪造 SourceID（本轮不改）。
- `hop_limit` 仍由 Core 在转发时递减，避免环路/风暴（本轮不改）。

### 性能与测试策略
- 性能收益：减少中间节点 handler 的 JSON 解包 + action 查找（响应帧改为 Core 快转发）。
- 测试策略：
  - 单测断言 `file.read_resp/write_resp` 的 header `Major==MajorOKResp`；
  - Windows 回归 `go test ./...`。

### 可扩展性设计点（后续方向，不在本 PR）
- 可按相同准则逐个子协议审计：
  - “端到端返回帧/响应帧” → `MajorOKResp`；
  - “逐跳可见需裁决/翻译的控制帧” → `MajorCmd`；
  - “广播/通知” → `MajorMsg`。

---

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（你已确认：big-bang；无需兼容旧客户端；包含 docs + tests）。

### FMO1 - File：read_resp/write_resp 改为 MajorOKResp
- 目标：让 `file.read_resp/write_resp` 作为响应帧走 Core 快速转发（中间节点无需进 File handler 转发）。
- 涉及模块 / 文件：
  - `subproto/file/handler.go`
- 验收条件：
  - `sendCtrlToNode` 构造的 header：`Major==MajorOKResp`（覆盖 `read_resp/write_resp`）。
- 测试点：见 FMO2。
- 回滚点：revert 本提交。

### FMO2 - 单测：断言 File 响应帧 MajorOKResp
- 目标：锁定行为，避免未来回归。
- 涉及模块 / 文件：
  - `tests/file_handler_test.go`（新增）
- 验收条件：
  - 至少覆盖：
    - 非法 `file.read` 触发的 `read_resp`：`MajorOKResp`
    - 非法 `file.write` 触发的 `write_resp`：`MajorOKResp`
- 回滚点：revert 本提交。

### FMO3 - 文档：补齐 File Major 与转发边界说明
- 目标：让文档与统一框架规则一致，便于后续接手者理解。
- 涉及文件：
  - `docs/5-file.md`
- 验收条件：
  - 文档明确：
    - 请求帧（CTRL：`read/write`）使用 `MajorCmd`
    - 响应帧（CTRL：`read_resp/write_resp`）使用 `MajorOKResp`
    - DATA/ACK 使用 `MajorMsg`
    - 失败响应仍使用 `MajorOKResp`，错误通过 payload `code/msg`

### FMO4 - 回归测试（Windows）
- 命令（统一临时目录与并发）：
  - `$env:GOTMPDIR='d:\project\MyFlowHub3\.tmp\gotmp'`
  - `New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null`
  - `go test ./... -count=1 -p 1`
- 验收条件：通过。

### FMO5 - Code Review（阶段 3.3）+ 归档变更（阶段 4）
- 归档文件（worktree 内）：
  - `docs/change/2026-02-17_file-resp-major-ok.md`
- 验收条件：
  - Review 覆盖：需求/架构/性能/安全/测试；
  - 归档包含：任务映射、关键决策、测试命令与回滚方案。
