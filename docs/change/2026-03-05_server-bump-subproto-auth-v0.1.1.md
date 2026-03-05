# 2026-03-05 - Server：升级 subproto/auth 到 v0.1.1

## 变更背景 / 目标
- 背景：上游 `myflowhub-subproto/auth` 发布 `v0.1.1`，修复 `up_login` 传播中 `SenderPub` 字段误用导致的跨级路由传播不稳定问题。
- 目标：Server 对齐该版本，确保运行时引入修复后的 Auth 子模块。

## 具体变更内容
- 修改：
  - `go.mod`
    - `github.com/yttydcs/myflowhub-subproto/auth`：`v0.1.0` -> `v0.1.1`
  - `go.sum`
    - 同步新的模块校验和。
- 无功能代码逻辑改动。

## todo 任务映射
- SRVAUTH-1：升级依赖版本 -> 完成
- SRVAUTH-2：最小回归验证 -> 完成
- SRVAUTH-3：Code Review + 归档 -> 完成

## 关键设计决策与权衡
- 采用“最小变更”策略，仅升级目标依赖，不并入其他功能改动。
- 性能/行为影响：理论上仅影响 Auth 路由传播正确性，不增加 Server 请求路径开销。

## 测试与验证方式 / 结果
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/auth`
  - 结果：`v0.1.1`
- `GOWORK=off go test ./... -count=1 -p 1`
  - 结果：通过。

## Code Review（3.3）结论
- 需求覆盖：通过
- 架构合理性：通过
- 性能风险：通过
- 可读性与一致性：通过
- 可扩展性与配置化：通过
- 稳定性与安全：通过
- 测试覆盖情况：通过（依赖升级回归测试通过）

## 潜在影响与回滚方案
- 潜在影响：Auth 子模块行为升级到 `v0.1.1`；其他子模块版本不变。
- 回滚：将 `go.mod/go.sum` 中 `auth` 回退至 `v0.1.0` 并重新测试。
