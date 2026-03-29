# 2026-03-28 Server：stream 装配与文档同步

## 变更背景 / 目标

- 在 `MyFlowHub-Server` 中接入新的 `stream` 子协议，但本轮不要求先发布 semver/tag。
- 保持 Server 继续只承担兼容壳、默认装配和文档真相角色，不回退到内置协议实现。

## 具体变更内容

### 新增

- [`protocol/stream/types.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/protocol/stream/types.go)
  - 新增 `stream` compat wrapper，公开类型/常量全部委托到 `myflowhub-proto/protocol/stream`
- [`modules/defaultset/stream_enabled.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/modules/defaultset/stream_enabled.go)
  - 默认构建接入 `myflowhub-subproto/stream`
- [`modules/defaultset/stream_disabled.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/modules/defaultset/stream_disabled.go)
  - 预留 `nostream` build tag 裁切路径
- [`docs/change/2026-03-28_stream-server-integration.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/change/2026-03-28_stream-server-integration.md)
  - 记录本轮 Server 侧装配与文档同步结果

### 修改

- [`modules/defaultset/hub.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/modules/defaultset/hub.go)
  - 默认 handler 集合容量从 `7` 调整为 `8`
  - 默认集合追加 `stream`
- [`modules/hub_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/modules/hub_test.go)
  - 增加默认集合必须包含 `SubProto=8` 的断言
- [`docs/specs/stream.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/specs/stream.md)
  - 锁定内部协调口径：
    - 私有 `delivery_prepare/activate/abort/close`
    - `activate` 前禁止 DATA / ACK 生效
    - owner 撤销后仍需借 coordinator 收敛 route
- [`docs/specs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/specs/protocol_map.md)
  - 同步 Proto canonical 副本，纳入 `SubProto=8 / protocol/stream`

## Requirements impact

- `none`

## Specs impact

- `updated`

## Lessons impact

- `none`

## Related requirements

- [`docs/requirements/stream.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/requirements/stream.md)

## Related specs

- [`docs/specs/stream.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/specs/stream.md)
- [`docs/specs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/specs/protocol_map.md)

## Related lessons

- `none`

## 对应 `plan.md` 任务映射

- `STRM-SRV-1`
- `STRM-SRV-2`
- `STRM-VAL-1`

## 经验 / 教训摘要

- 在未发布 `stream` module/tag 的回合里，不应该把未发布版本硬写进 `Server/go.mod`，否则即使本地有 workspace，也会先命中远端 revision 解析失败。
- 对 Server 来说，`protocol/stream` compat wrapper 和 `docs/specs/protocol_map.md` 同步是一起交付的；只改 defaultset 而不补协议入口，会让 repo 导航失真。

## 可复用排查线索

- 症状
  - `unknown revision stream/v0.1.0`
  - 默认集合缺少 `SubProto=8`
  - `docs/specs/protocol_map.md` 看不到 `protocol/stream`
- 触发条件
  - Server 提前接入本地新 module，但 semver/tag 尚未发布
- 关键词
  - `stream`
  - `defaultset`
  - `protocol/stream`
  - `unknown revision stream/v0.1.0`
  - `go.work`
- 快速检查
  - 检查 [`modules/defaultset/hub.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/modules/defaultset/hub.go) 是否追加 `newStreamHandler`
  - 检查 [`modules/hub_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/modules/hub_test.go) 是否断言 `SubProtoStream`
  - 检查 worktree 根的临时 `go.work` 是否已指向本地 `Proto + stream`

## 关键设计决策与权衡

- 本轮用临时 `go.work` 验证本地联调，不强行伪造未发布 semver。
- `stream` 在 defaultset 中采用和其它模块一致的 `enabled/disabled` build tag 组织，避免未来裁切时结构破例。
- `specs/stream.md` 只补齐长期应稳定的协调口径，不把 handler 细节搬进长期真相。

## 测试与验证方式 / 结果

- 临时 workspace：
  - `D:\project\MyFlowHub3\worktrees\server-stream-subproto-design\go.work`
  - 指向本地 `Core + Proto + SubProto/stream + Server`
- 执行：`go test ./modules/... ./tests/... -count=1 -p 1`
- 结果：通过
- 备注
  - 直接 `GOWORK=off` 仍会落回已发布旧版 Proto / SubProto，不能代表本轮联调结果

## 潜在影响与回滚方案

- 潜在影响
  - 本轮 Server 侧 `stream` 接入依赖本地 workspace；未发布 tag 前不具备 `GOWORK=off` 可复现性
- 回滚方案
  - 回退：
    - [`protocol/stream/types.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/protocol/stream/types.go)
    - [`modules/defaultset/hub.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/modules/defaultset/hub.go)
    - [`modules/defaultset/stream_enabled.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/modules/defaultset/stream_enabled.go)
    - [`modules/defaultset/stream_disabled.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/modules/defaultset/stream_disabled.go)
    - [`modules/hub_test.go`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/modules/hub_test.go)
    - [`docs/specs/stream.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/specs/stream.md)
    - [`docs/specs/protocol_map.md`](D:/project/MyFlowHub3/worktrees/server-stream-subproto-design/docs/specs/protocol_map.md)
  - 删除临时 `go.work` / `go.work.sum`

## 子Agent执行轨迹

- 本轮未使用子Agent
