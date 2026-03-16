# 变更背景 / 目标

Server runtime 接入 QUIC listener，支持与 TCP/RFCOMM 并存监听，并允许 parent endpoint 通过 `quic://` 建链。

# 具体变更内容

- `hubruntime/options.go`
  - 新增 QUIC 配置项：`QUICEnable`、`QUICAddr`、`QUICALPN`、证书与客户端证书配置；
  - 新增对应环境变量读取与 `Normalize` 规则。
- `hubruntime/runtime.go`
  - 新增 `quic_listener` 引入与 runtime 构建；
  - `listeners` 组合加入 QUIC 分支；
  - `parseParentEndpoint` / `dialParentEndpoint` 支持 `quic://`。
- `cmd/hub_server/main.go`
  - 新增 QUIC CLI 参数：`-quic-enable`、`-quic-addr`、`-quic-alpn`、证书相关参数等。
- `go.mod` / `go.sum`
  - 对齐 Core 到包含 QUIC transport 的开发版本；
  - 补齐 `quic-go` 依赖校验。

# 对应任务映射

- QUIC-SERVER-1：Server Runtime 挂载 QUIC listener

# 关键设计决策与权衡

- 维持 `multi_listener` 模式，不引入额外调度层；
- QUIC 默认为显式开关，避免影响现有 TCP/RFCOMM 部署路径；
- parent 自注册逻辑仍保持 TCP-only（既有策略不变）。

# 测试与验证

- `GOWORK=off go test ./...`

# 潜在影响与回滚

- 影响：Server 增加 QUIC 配置面和依赖；
- 回滚：删除 QUIC 配置分支并回退 Core 依赖版本。
