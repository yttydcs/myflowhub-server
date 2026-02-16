# 2026-02-16 - Auth 直返 *_resp 统一为 MajorOKResp

## 变更背景 / 目标
SDK v1 Awaiter 仅拦截 `MajorOKResp/MajorErrResp` 的响应帧；但 Auth 在“直返客户端”的链路中常通过 `reqHdr.Clone()` 复用请求头，导致 `register_resp/login_resp/revoke_resp` 的响应头 `Major` 继承为 `MajorCmd`，从而出现客户端 Awaiter 等不到响应（超时）的情况。

本次变更目标：
- **仅修复 Auth 的直返响应头 Major**：将 `register_resp/login_resp/revoke_resp` 的直返响应统一标记为 `MajorOKResp`；
- **不改 wire**（action/JSON/SubProto 不变）；
- **失败响应仍使用 `MajorOKResp`**，错误由 payload 内 `data.code` 表达；
- **不影响逐跳内部链路**（如 `assist_*` / `up_*` 等仍保持原有路径与语义）。

## 具体变更内容
### 新增
- `subproto/auth/transport.go`
  - 新增 `sendDirectResp(...)` 与 `buildDirectRespHeader(...)`：
    - 在 `reqHdr.Clone()` 基础上 **仅覆盖 `Major=MajorOKResp`**，其余字段保持继承（例如 `MsgID/TraceID`）。

### 修改
- `subproto/auth/actions_register.go`
  - `register -> register_resp` 直返路径改用 `sendDirectResp`（含参数非法等失败分支）。
- `subproto/auth/actions_login.go`
  - `login -> login_resp` 直返路径改用 `sendDirectResp`（含 authority 不存在时的直接失败回包、签名不匹配等失败分支）。
- `subproto/auth/actions_revoke.go`
  - `revoke -> revoke_resp` 改用 `sendDirectResp`（permission denied 与成功分支）。
- `tests/auth_handler_test.go`
  - 增加直返场景断言：`*_resp` 的 `hdr.Major()==MajorOKResp`。

### 删除
- 无。

## plan.md 任务映射
- AOR1 - Auth：新增直返响应发送函数（MajorOKResp）✅
- AOR2 - Auth：register/login/revoke 直返 *_resp 改用 sendDirectResp ✅
- AOR3 - 单测：断言直返 *_resp 的 Major=MajorOKResp ✅
- AOR4 - 回归测试（Windows）✅

## 关键设计决策与权衡
- 只改“直返客户端的 `*_resp`”：
  - 避免把 `assist_* / up_*` 等内部逐跳链路一并改动而扩大风险。
- 失败响应继续使用 `MajorOKResp`：
  - 本次不引入 `MajorErrResp` 的统一语义标准化，避免跨协议/跨端同步成本；
  - 业务失败由 `data.code` 表达（与现有协议文档一致）。
- 性能：
  - 仅新增一次 `Clone()+WithMajor()` 的头部构造；不增加额外网络 I/O、序列化次数或循环。

## 测试与验证方式 / 结果
- Windows：
  - `$env:GOTMPDIR='d:\project\MyFlowHub3\.tmp\gotmp'`
  - `New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null`
  - `go test ./... -count=1 -p 1`
  - 结果：通过。

## 潜在影响
- 若存在历史客户端只在 `MajorCmd` 下解析 `*_resp`（非规范实现），可能需要同步升级客户端侧；本次通过“仅 Auth 且仅直返 `*_resp`”将影响面降到最小。

## 回滚方案
- 按提交粒度 `git revert <sha>` 回退本次变更（不涉及 wire 变更，回滚风险较低）。

