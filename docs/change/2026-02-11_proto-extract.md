# 2026-02-11 协议仓库拆分（Proto）+ exec/flow 解耦地基（Server）

## 背景 / 目标
- 将协议定义从 `MyFlowHub-Server/protocol/*` 抽离到独立仓库 `MyFlowHub-Proto`（wire 不变）。
- Server 保留 `protocol/*` 作为**兼容壳**（策略 A）：不立刻打断外部依赖。
- 消除子协议“实现层耦合”：至少移除 `flow -> internal/handler/exec` 的直接 import，为后续子协议可裁切/可组装铺路。

## 具体变更内容（新增 / 修改 / 删除）
### 修改
- `go.mod`
  - 新增 `github.com/yttydcs/myflowhub-proto v0.0.0`
  - `replace github.com/yttydcs/myflowhub-proto => ../MyFlowHub-Proto`（本地联调）
- `protocol/{auth,varstore,topicbus,file,flow,management}/types.go`
  - 改为兼容壳：类型/常量 alias 到 `github.com/yttydcs/myflowhub-proto/protocol/*`
- `internal/handler/exec/types.go`
  - 改为引用 `protocol/exec` 的类型/常量（wire 不变）
- `internal/handler/flow/handler.go`
  - `exec call/call_resp` 相关逻辑改用 `protocol/exec` + 共享 broker；不再 import `internal/handler/exec`

### 新增
- `protocol/exec/types.go`：exec 子协议兼容壳（指向 Proto）
- `internal/broker/*`
  - 通用 `Broker[T]`（进程内 reqID->resp 投递器）
  - `SharedExecCallBroker`：exec `call_resp` 的共享投递器

### 删除
- `internal/handler/exec/broker.go` 与对应测试：逻辑迁移到 `internal/broker`

## plan.md 任务映射
- S1：接入 Proto 依赖（go.mod）✅
- S2：protocol/* 兼容壳改造（指向 Proto）✅
- S3：解耦 flow 与 exec 实现包（移除跨 handler import）✅
- S4：全量回归 ✅

## 关键设计决策与权衡（性能 / 扩展性）
- 兼容壳（策略 A）：
  - 优点：外部依赖短期不破坏，可渐进迁移到 `myflowhub-proto`。
  - 代价：import 链多一层（可控）。
- `internal/broker` 抽取为泛型：
  - 避免 handler 之间互相引用实现包，降低耦合。
  - 为未来其它“请求-响应配对”的子协议复用提供扩展点。
- 性能关键点：
  - `Broker` 为 `map + mutex` 的 O(1) 投递；`chan` 缓冲 1，避免 Deliver 侧阻塞并保证完成信号（close）。

## 测试与验证方式 / 结果
- `go test ./... -count=1 -p 1`：通过（包含 `tests/`）。
- 说明：当前环境并行编译可能触发 OOM；如遇到临时目录权限问题可设置 `GOTMPDIR` 指向项目内目录后重试。

## 潜在影响与回滚方案
### 潜在影响
- 若外部仍 import `myflowhub-server/protocol/*`：不受影响（兼容壳保留）。
- 若需要仓库独立构建/CI：需在 Proto 发布可拉取版本后移除 `replace`。

### 回滚方案
- `git revert` 回退本 PR 提交（会恢复旧 protocol 定义与旧 broker 位置）；如依赖方已切到 Proto 需同步回滚。

