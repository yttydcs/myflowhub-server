# 变更背景 / 目标

在 Server 已接入 QUIC runtime 配置后，将 Core 依赖从开发态 pseudo-version 收敛到正式 `v0.4.7`，保证下游部署与回滚可控。

# 具体变更内容

## 修改
- `go.mod`
  - `github.com/yttydcs/myflowhub-core`：
    - `v0.4.7-0.20260316021423-d992975ec6ad`
    - -> `v0.4.7`
- `go.sum`
  - 同步校验记录，清理 pseudo-version 对应项。

# 对应任务映射

- `QUIC-REL-1`：下游版本对齐与发布

# 关键设计决策与权衡

- 保持 runtime 代码不再变更，仅替换依赖版本：
  - 优点：功能行为稳定，发布风险最小；
  - 代价：需要严格依赖发布顺序（Core 先于 Server）。

# 测试与验证方式 / 结果

- 执行：
  - `GOWORK=off go mod tidy`
  - `GOWORK=off go test ./...`
- 结果：
  - 全量通过。

# 潜在影响与回滚方案

- 潜在影响：
  - 无业务逻辑变更，主要影响为依赖版本标识。
- 回滚方案：
  - 回退本次提交并恢复旧依赖；重新执行全量测试验证。
