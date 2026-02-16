# Plan - Auth 直返 *_resp 统一为 MajorOKResp（PR10-Auth-DirectOKResp）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`fix/auth-okresp-direct`
- Worktree：`d:\project\MyFlowHub3\worktrees\auth-okresp-direct`
- Base：`main`
- 参考：
  - `d:\project\MyFlowHub3\target.md`
  - `d:\project\MyFlowHub3\repos.md`
  - `d:\project\MyFlowHub3\guide.md`（commit 信息中文）
- 约束（本 workflow 明确边界，避免范围外扩）：
  - 只改 `subproto/auth` 的“直返客户端的响应头 Major”，不改 wire（action/JSON/SubProto 不变）。
  - 仅覆盖：`register_resp` / `login_resp` / `revoke_resp` 的直返场景。
  - 失败响应仍使用 `MajorOKResp`（错误由 `data.code` 表达），本轮不引入 `MajorErrResp` 语义标准化。
  - 不改 `assist_*` / `up_*` 等逐跳内部链路（保持现状）。
  - 不改 `MyFlowHub-SDK`（不增加兼容开关）。
  - 所有实现性改动只在本 worktree 内完成；`repo/` 仅用于合并/推送/集成验证。

## 当前状态（事实，可审计）
- SDK v1 Awaiter 仅拦截 `MajorOKResp/MajorErrResp` 的响应帧：`MyFlowHub-SDK/await/client.go`。
- Auth 的 `sendResp` 在 `reqHdr!=nil` 时会 `Clone()` 请求头，从而导致部分直返 `*_resp` 继承为 `MajorCmd`：`subproto/auth/transport.go`。
- 结果：客户端使用 Awaiter 等待 `login_resp/register_resp/revoke_resp` 时可能超时（响应帧 Major 不匹配被当作 unmatched）。

---

## 1) 需求分析

### 目标
- 让 Auth 的“直返客户端”响应（`register_resp/login_resp/revoke_resp`）在 header 上标记为 `MajorOKResp`，从而与 SDK Awaiter 的拦截规则一致。
- 保持 `MsgID/TraceID` 等字段继承（不引入额外生成规则），确保请求-响应匹配语义稳定。

### 范围（必须 / 可选 / 不做）
#### 必须（本 PR）
- `subproto/auth`：
  - 新增/调整一个“直返响应头”构建路径：在保留原请求头字段的前提下，将 `Major` 强制改为 `MajorOKResp`。
  - 在以下直返路径调用该构建路径：
    - `register` → `register_resp`（包含参数非法等失败分支）
    - `login` → `login_resp`（包含 authority 不存在等失败分支）
    - `revoke` → `revoke_resp`（permission denied 分支与成功分支）
- 单测覆盖：断言上述直返 `*_resp` 的 `hdr.Major()==MajorOKResp`。
- 回归：`go test ./... -count=1 -p 1`（Windows）。

#### 可选（本 PR 如无额外风险）
- 在归档文档中补充“Major 规范化”的后续路线：按子协议逐个审计哪些 action 属于直返响应、哪些属于逐跳内部链路（避免一刀切）。

#### 不做（本 PR）
- 不改 `assist_*_resp` / `assist_query_credential_resp` / `up_*_resp` 等内部链路 Major（保持 `MajorCmd` 逐跳语义）。
- 不统一其它子协议的 Major（后续另起 workflow）。
- 不调整 `SourceID/TargetID` 等其它字段语义（本 PR 只处理 `Major`；避免引入不可预期的路由/审计变化）。
- 不改 SDK Awaiter 规则、不加兼容开关。

### 使用场景
- Win/CLI/脚本等客户端通过 SDK Awaiter 发起 `register/login/revoke` 并等待 `*_resp`。

### 验收标准
- 对上述直返响应：响应帧 header 的 `Major==MajorOKResp`。
- 单测覆盖并通过。
- `go test ./...` 通过（Windows）。

### 风险
- 若存在历史客户端“只处理 MajorCmd 响应帧”的非规范实现，可能需要同步升级客户端；本 PR 通过“仅 Auth 且仅直返 *_resp”将风险降到最小。

---

## 2) 架构设计（分析）

### 总体方案（采用）
- 在 Auth 内新增一个最小的“直返响应发送函数”：
  - 输入：原请求头 `reqHdr`（用于继承 `MsgID/TraceID/...`）、payload。
  - 输出：在 `reqHdr.Clone()` 基础上仅覆盖 `Major=MajorOKResp` 的响应头（其余字段保持不动）。
- 在 `register/login/revoke` 的“直返 *_resp”路径改用该函数；内部链路保持原 `sendResp`（继承 `MajorCmd`）。

### 备选对比（为什么不选）
- 备选 A：修改 `buildHeader/sendResp` 在 `reqHdr!=nil` 时也强制 `MajorOKResp`
  - 问题：会影响 `assist_*_resp/up_*_resp` 等内部逐跳链路，范围外扩且风险高。
- 备选 B：SDK 兼容接收 `MajorCmd` 的 `*_resp`（开关）
  - 已明确不做（你已确认“不用保持兼容开关”）。

### 模块职责
- `subproto/auth/transport.go`：提供“直返响应头 Major 修正”的集中实现点（避免散落在各 action 文件）。
- `subproto/auth/actions_*.go`：仅负责在合适的分支选择 `sendDirectResp` 或 `sendResp`。

### 数据 / 调用流（关键链路）
- 直返链路（本 PR 覆盖）：
  - client → `register/login/revoke`（MajorCmd）→ auth handler → `*_resp`（MajorOKResp）→ client Awaiter 匹配成功。
- 内部逐跳链路（本 PR 不改）：
  - child hub → `assist_*`（MajorCmd）→ authority → `assist_*_resp`（MajorCmd）→ child hub handler 消费 → 回落 `*_resp` 给 device（已是 MajorOKResp）。

### 接口草案（内部函数）
- `sendDirectResp(ctx, conn, reqHdr, action, data)`：发送“直返客户端”的 `*_resp`，保证 `MajorOKResp`。
- `sendResp(...)`：保持现有，供内部链路/逐跳语义使用。

### 错误与安全
- 仅更改响应头 Major，不改变权限判断/白名单/签名验证等逻辑。
- `revoke` 的 permission denied 仍返回 `revoke_resp`，只是 Major 从 Cmd 变为 OK（业务码仍为 4403）。

### 性能与测试策略
- 性能：仅多一次 `Clone()+WithMajor()`，不增加额外网络 I/O 与序列化次数（payload 不变）。
- 测试：
  - 单测断言直返 `register_resp/login_resp/revoke_resp` 的 `hdr.Major()==MajorOKResp`；
  - 保留现有 assist/pending 用例，确保未引入回归。

### 可扩展性设计点（后续方向，不在本 PR）
- 建议形成全局约定：对“客户端直返的 *_resp”统一用 `MajorOKResp`；对需要逐跳消费/翻译的内部动作保持 `MajorCmd`；对通知类统一用 `MajorMsg`。

---

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（范围、失败语义、是否做 SDK 兼容开关均已确认）。

### AOR1 - Auth：新增直返响应发送函数（MajorOKResp）
- 目标：集中、可复用地构造直返响应头（仅覆盖 Major）。
- 涉及模块 / 文件：
  - `subproto/auth/transport.go`
- 验收条件：
  - `sendDirectResp`（或等价函数）存在且只做 `MajorOKResp` 覆盖；
  - 不影响现有 `sendResp` 行为。
- 测试点：见 AOR3。
- 回滚点：revert 本提交。

### AOR2 - Auth：register/login/revoke 直返 *_resp 改用 sendDirectResp
- 目标：让直返 `register_resp/login_resp/revoke_resp` 全部为 `MajorOKResp`。
- 涉及模块 / 文件：
  - `subproto/auth/actions_register.go`
  - `subproto/auth/actions_login.go`
  - `subproto/auth/actions_revoke.go`
- 验收条件：
  - self-authority（无 parent）下的直返 register/login/revoke 响应：MajorOKResp。
  - authority 存在的 assist 链路不受影响（仍按原逻辑 pending 回落）。
- 回滚点：revert 本提交。

### AOR3 - 单测：断言直返 *_resp 的 Major=MajorOKResp
- 目标：锁定行为，避免未来回归。
- 涉及文件：
  - `tests/auth_handler_test.go`
- 验收条件：
  - 至少覆盖：self-authority 的 `register_resp`、authority nil 的 `login_resp`（直接失败回包）、`revoke_resp`（permission denied）。
- 回滚点：revert 本提交。

### AOR4 - 回归测试（Windows）
- 命令（建议统一，避免临时目录权限/并发问题）：
  - `$env:GOTMPDIR='d:\project\MyFlowHub3\.tmp\gotmp'`
  - `New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null`
  - `go test ./... -count=1 -p 1`
- 验收条件：测试通过。

### AOR5 - Code Review（阶段 3.3）+ 归档变更（阶段 4）
- 归档文件：
  - `docs/change/2026-02-16_auth-direct-resp-major-ok.md`
- 验收条件：
  - Review 覆盖：需求/架构/性能/安全/测试；
  - 归档包含：任务映射、关键决策、测试命令与回滚方案。
