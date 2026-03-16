# Plan - Server：新增 QUIC 开发自动证书开关

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`feat/quic-dev-cert-auto`
- Worktree：`d:\project\MyFlowHub3\repo\MyFlowHub-Server\worktrees\feat-quic-dev-cert-auto`
- Base：`main`

## 项目目标与当前状态
- 目标：
  - 新增 `-quic-dev-cert-auto`（及环境变量）用于开发场景自动生成临时自签名证书；
  - 当已显式提供 `-quic-cert-file/-quic-key-file` 时，不改变现有行为；
  - 默认关闭，生产路径不受影响。
- 当前状态：
  - QUIC listener 强制要求 `cert_file` 与 `key_file`；
  - 无证书时 server 启动会失败，需要手工准备证书。

## 范围
- 必须：
  - 新增运行时配置项（Options + env + CLI flag）；
  - 启动前按条件自动生成并注入证书文件路径；
  - 增加单测覆盖关键分支（启用/禁用、已有证书、自动生成成功）。
- 可选：
  - 支持可配置自动证书目录（默认落在 WorkDir 或系统临时目录）。
- 不做：
  - 不更改 Core `quic_listener` 的证书必选约束；
  - 不改变 QUIC 客户端校验策略（`insecure/pin` 语义保持不变）。

## 可执行任务清单（Checklist）

### DEV-CERT-1 - 配置面接入（已完成）
- 目标：新增 `QUICDevCertAuto` 配置与 CLI/ENV 映射。
- 涉及模块 / 文件：
  - `hubruntime/options.go`
  - `cmd/hub_server/main.go`
- 验收条件：
  - `-quic-dev-cert-auto` 和 `HUB_QUIC_DEV_CERT_AUTO` 生效；
  - 默认值为 `false`。
- 测试点：
  - options normalize 不破坏既有字段。
- 回滚点：
  - 回退新增配置字段与参数绑定。

### DEV-CERT-2 - 自动证书生成与注入（已完成）
- 目标：在 `Runtime.Start` 的 QUIC 装配前，按条件生成临时证书并填充 `opts.QUICCertFile/QUICKeyFile`。
- 涉及模块 / 文件：
  - `hubruntime/runtime.go`
  - （可能新增）`hubruntime/quic_dev_cert.go`
- 验收条件：
  - 仅在 `QUICEnable && QUICDevCertAuto && cert/key 均为空` 时生成；
  - 已提供 cert/key 时不覆盖；
  - 生成失败返回明确错误。
- 测试点：
  - 生成后的 cert/key 文件可被读取，路径已注入；
  - 禁用开关时不生成。
- 回滚点：
  - 回退自动生成逻辑并恢复原启动行为。

### DEV-CERT-3 - 测试与回归（已完成）
- 目标：补充单测并验证全仓测试不回退。
- 涉及模块 / 文件：
  - `hubruntime/*_test.go`
- 验收条件：
  - `GOWORK=off go test ./hubruntime ./...` 通过。
- 回滚点：
  - 回退新增测试文件。

### DEV-CERT-4 - Code Review 与归档（进行中）
- 目标：输出强制审查结论并归档到 `docs/change`。
- 涉及模块 / 文件：
  - `docs/change/2026-03-16_server-quic-dev-cert-auto.md`
- 验收条件：
  - 审查项完整（需求/架构/性能/可维护性/测试）；
  - 文档可独立交接。

## 依赖关系
- `DEV-CERT-1` -> `DEV-CERT-2` -> `DEV-CERT-3` -> `DEV-CERT-4`

## 风险与注意事项
- 自动证书仅用于开发环境，必须保持默认关闭；
- 证书文件落盘应使用最小权限（当前平台允许范围内）；
- 需保证退出后不影响已有 `workdir` 行为与其他 listener。
