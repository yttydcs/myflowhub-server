# 2026-03-08 升级 Auth 以修复多 hop 路由索引缺失（Server）

## 背景 / 目标

多 hop 场景出现回归：Root Hub 因 `sourceMismatch` 丢弃后代节点帧，导致上层能力（典型是 VarStore `list/get`）返回 `not found (code=4)`。

改动规模：较大（升级 Auth 关键依赖并修订规范文档，影响 Root/Hub 的路由索引建立与安全门禁链路）。

本次目标：
- 升级 Server 依赖的 `myflowhub-subproto/auth` 到 `v0.1.2`（修复公钥毒化 + up_login 自愈）；
- 同步更新 `docs/specs/auth.md`，避免继续误导旧行为。

## 变更内容

### 1) 依赖升级
- `github.com/yttydcs/myflowhub-subproto/auth`：`v0.1.1` -> `v0.1.2`

涉及文件：
- `go.mod`
- `go.sum`

### 2) 文档更新
- 更新 `docs/specs/auth.md`：
  - `register`：缺省 `pubkey` 不再自动填本节点公钥；
  - `up_login`：补充 `sender_pub` 自愈策略与约束 `sender_id == hdr.SourceID == conn.meta(nodeID)`；
  - `auth.disable_persist=true`：明确不读写 `config/trusted_nodes.json`。

## 任务映射（plan.md）
- SRV-AUTH-1：完成
- SRV-AUTH-2：完成
- SRV-AUTH-3：未执行（按可选项保留）
- SRV-AUTH-4：完成

## 验证方式 / 结果

依赖解析：

```powershell
$env:GOWORK='off'
go list -m github.com/yttydcs/myflowhub-subproto/auth@v0.1.2
```

测试：

```powershell
$env:GOWORK='off'
go test ./... -count=1 -p 1
```

结果：通过。

## 潜在影响
- Root Hub 可正确接受并建立后代节点路由索引，避免后续 `sourceMismatch` 丢弃。
- Android Release workflow 会在构建时 checkout `myflowhub-server` 默认分支作为 hubmobile replace 源：本分支需合入默认分支后，Android 发布才会消费到该修复。

## 回滚方案
- 回退本仓 `go.mod/go.sum` 中 `myflowhub-subproto/auth` 版本至 `v0.1.1` 并重新发布；
- 或发布后续 Server patch 修正。

