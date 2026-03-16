# 变更背景 / 目标

当前 QUIC listener 启动依赖手工提供 `cert_file` 与 `key_file`。  
为便于本地联调与测试，新增开发开关 `quic-dev-cert-auto`：在未提供证书时自动生成自签名证书并注入运行时配置。

# 具体变更内容

## 新增
- `hubruntime/quic_dev_cert.go`
  - `ensureQUICDevCertIfNeeded`：按条件自动生成并注入 QUIC 证书路径；
  - `generateSelfSignedQUICDevCert`：生成本地开发自签名证书（TLS1.3 可用，含 `localhost/127.0.0.1/::1` SAN）。
- `hubruntime/quic_dev_cert_test.go`
  - 覆盖开关关闭、已有证书保留、半配置报错、自动生成成功。

## 修改
- `hubruntime/options.go`
  - 新增 `QUICDevCertAuto`；
  - 新增环境变量：`HUB_QUIC_DEV_CERT_AUTO`。
- `cmd/hub_server/main.go`
  - 新增参数：`-quic-dev-cert-auto`。
- `hubruntime/runtime.go`
  - `Start()` 中在 QUIC listener 装配前调用自动证书注入逻辑。

## 删除
- 无。

# 对应计划任务映射

- `DEV-CERT-1`：配置面接入
- `DEV-CERT-2`：自动证书生成与注入
- `DEV-CERT-3`：测试与回归
- `DEV-CERT-4`：Code Review 与归档

# 关键设计决策与权衡

- 默认关闭（`false`），避免影响生产部署策略；
- 仅在 `QUICEnable && QUICDevCertAuto && cert/key 均为空` 时生效，显式证书优先级更高；
- 自动证书落盘在 `WorkDir`（为空时落系统临时目录下 `myflowhub`），便于定位与复用；
- 保持 Core `quic_listener` 的证书必选约束不变，降低跨仓改动范围。

# 测试与验证方式 / 结果

- `GOWORK=off go test ./hubruntime ./...`
- 结果：通过。

# 潜在影响与回滚方案

- 潜在影响：
  - 开启 `quic-dev-cert-auto` 后会写入本地证书文件（开发用途）；
  - 若仅配置 cert 或 key 其一，会提前返回明确配置错误。
- 回滚方案：
  - 回退本次提交，移除 `QUICDevCertAuto` 相关逻辑与参数，恢复手工证书模式。
