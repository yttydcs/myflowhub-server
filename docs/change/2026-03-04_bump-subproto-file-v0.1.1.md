# MyFlowHub-Server：升级 `myflowhub-subproto/file` 至 `v0.1.1`（修复 Hub File Console not found）

## 变更背景 / 目标

### 背景
- Win 的 File Console 访问 Hub（node1）时提示：`not found`。
- 根因位于 `myflowhub-subproto/file v0.1.0`：
  - root list 时默认 `file.base_dir=./file` 目录不存在会被映射为 `404 not found`；
  - 相对 `file.base_dir` 以 CWD 为基准导致目录落点不稳定。

### 目标
- 将 `MyFlowHub-Server` 升级到 `myflowhub-subproto/file v0.1.1`，获取修复（详见 SubProto 仓的变更文档与 tag）。

---

## 具体变更内容
- 文件：`go.mod`
  - `github.com/yttydcs/myflowhub-subproto/file v0.1.0` → `v0.1.1`
- 文件：`go.sum`
  - 更新依赖校验和

---

## 对应 plan.md 任务映射
- SRV0：归档旧 plan 并建立本次计划
- SRV1：升级依赖 `myflowhub-subproto/file v0.1.1`
- SRV2：验证 tag 可解析与编译通过
- SRV3：Code Review + 归档变更（本文）

---

## 测试与验证
- 依赖可解析：
  - `go list -m github.com/yttydcs/myflowhub-subproto/file@v0.1.1`
- 单测：
  - `cd MyFlowHub-Server; $env:GOWORK='off'; go test ./... -count=1 -p 1`
  - 结果：通过

---

## 潜在影响与回滚方案

### 潜在影响
- 该升级会让 Hub 的默认 `file/` 目录在未配置 `file.base_dir` 时创建于 `hub_server.exe` 同目录下（相对路径以 exeDir 为基准）。
- 使用该修复需要重新构建并重启 `hub_server`（仅升级依赖不会自动生效于已运行的旧二进制）。

### 回滚方案
- 回滚该提交，将依赖固定回 `github.com/yttydcs/myflowhub-subproto/file v0.1.0`。

