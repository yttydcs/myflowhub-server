# 2026-02-18 - assist_* / up_* / notify_* 代码层收敛：action 模板与分类（wire 不改）（PR2-3）

## 变更背景 / 目标
在 Server 的各子协议实现中，`assist_* / up_* / notify_*` 等 action 的注册方式存在多种写法（结构体、包装类型、`assisted bool` 分支等）。这会带来：
- 新增 action 时容易漏注册或风格不一致，维护成本高；
- 子协议之间难以形成统一的工程承载模板，影响后续拆库/裁切与跨端一致性。

本次变更的目标是：在 **不改变 wire 与运行语义** 的前提下，引入统一的 action 模板，并优先在 `auth` 与 `varstore` 落地收敛。

## 具体变更内容

### 新增
- `subproto/kit/action.go`
  - 新增 `ActionKind`：`Local/Assist/Up/Notify`（仅用于工程组织/可观测，不参与路由/转发语义）。
  - 新增 `KindFromName(name)`：基于前缀（`assist_`/`up_`/`notify_`）推导 kind，支持显式覆盖。
  - 新增 `FuncAction` + `NewAction(...)`：函数式 action 模板，减少 `Name/RequireAuth/Handle` 样板代码。
- `subproto/kit/action_test.go`
  - 覆盖 `KindFromName` 的基本推导规则。

### 修改
- `subproto/auth/*`
  - 将 register/login/offline/perms/query/revoke/up_login 等 action 注册方式统一为 `kit.NewAction(...)`（保留原业务逻辑与行为）。
  - `RequireAuth` 规则保持与原实现一致（仅按原先 action 的语义设置）。
- `subproto/varstore/actions.go`
  - 移除 `varAction` 包装类型，改为统一 `kit.NewAction(...)` 注册列表。
  - 对 `var_changed / var_deleted` 显式标注 `ActionKindNotify`（避免仅靠前缀推导）。

### 删除
- `subproto/auth/actions_up_login_register.go`
  - `registerUpLoginActions` 合并到 `actions_up_login.go` 内，减少分散文件与重复样板。

## plan.md 任务映射
- DOC1 - 更新全局文档（先做） ✅（在 `d:\project\MyFlowHub3\target.md`、`d:\project\MyFlowHub3\repos.md` 同步当前真实状态与 PR2-3 细节）
- KIT1 - kit：新增 action 模板与 kind ✅
- AUTH1 - auth：迁移 assist/up 注册到 kit 模板 ✅（并将 auth 全部 action 注册风格统一）
- VSTORE1 - varstore：迁移 assist/up/notify 注册到 kit 模板 ✅
- TEST1 - 回归测试（GOWORK=off） ✅
- SMOKE1 - 冒烟步骤（hub_server + node_echo） ✅（见下）

## 关键设计决策与权衡
- **策略 A：wire 不改**：不调整 action 名称、JSON schema、SubProto、send/forward/header 语义，避免行为漂移导致跨端兼容风险。
- **模板保持“薄”**：`kit.NewAction` 只解决样板与分类元信息，不引入统一转发/发送策略（该类语义变更应另起 PR）。
- **性能**：kind 推导在注册期完成；运行期不增加额外 marshal/unmarshal 或字符串判断。
- **可扩展性**：允许 `WithKind(...)` 覆盖推导结果，为后续“无前缀但语义为 notify”的 action 预留扩展点。

## 测试与验证方式 / 结果

### 单测/集成测试（必选）
```powershell
$env:GOTMPDIR='d:\\project\\MyFlowHub3\\.tmp\\gotmp'
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
$env:GOWORK='off'
go test ./... -count=1 -p 1
```

结果：通过（包含 `tests/TestRootHubPing`、varstore 集成测试等）。

### 冒烟（启动并连到 server）
推荐直接执行内置集成用例（覆盖“启动 server 并发包收包”的最小链路）：
```powershell
$env:GOWORK='off'
go test ./tests -run TestRootHubPing -count=1
```

预期：收到 `node_echo_resp`，且 `code=1`、`echo=ping`。

## 潜在影响与回滚方案

### 潜在影响
- 若 action 注册遗漏，运行期会出现 `unknown <subproto> action` 并丢弃；已通过现有单测/集成测试降低风险。
- action 由“结构体”改为“函数式模板”后，未来调试时堆栈展示会不同，但不影响语义。

### 回滚方案
- 回滚本 PR 的提交（按提交顺序 revert 即可）：
  - `chore: 更新 PR2-3 计划`
  - `refactor: 引入 action 模板与分类`
  - `refactor: 收敛 auth/varstore action 注册`

