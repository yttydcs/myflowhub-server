# 2026-03-25 Pluggable State Backend

## 变更背景 / 目标

为 `flow` 和 `varstore` 增加可插拔持久化后端，同时把 capability registry / permission config 的共享对象改为显式注入，避免继续依赖 `cfg` 指针充当隐式 service locator。

本次目标是：

- 无数据库时保持当前默认行为
  - `flow` 继续使用本地 JSON
  - `varstore` 继续使用纯内存
- 配置 `pg` 时由 `Server` 注入 PG backend
- `varstore` 固定 owner 写序：先持久化，后 cache / event / notify / up_* / success resp

## 具体变更内容

### MyFlowHub-SubProto

- `exec/runtimedeps`
  - 新增显式共享 runtime deps：`CapRegistry`、`PermConfig`
- `exec/file/flow/topicbus/varstore/management`
  - 新增 `WithDeps` / `WithOptions` 构造入口
  - 默认构造函数保持兼容
- `flow`
  - 新增 `Persistence` 接口与 `NewJSONPersistence()`
  - `Init()/set/delete` 改为通过 persistence 读写定义
  - 默认仍由 `flow.base_dir` 驱动 JSON backend
- `varstore`
  - 新增 `Persistence` 接口与 `NewMemoryPersistence()`
  - `Init()` 启动时从 persistence 预热 records cache
  - owner 本地 `set/revoke` 改为先持久化，再更新 cache / trigger / notify / up_* / resp
  - `up_set/notify_set/get_resp/subscribe_resp` 等链路仍只刷新内存 cache，不写持久层
- 新增单测：
  - 显式 deps 构造测试
  - `flow` injected persistence 初始化 / 失败语义测试
  - `varstore` preload 与 persist-failure 顺序测试

### MyFlowHub-Server

- `modules/defaultset`
  - 默认装配阶段统一创建共享 `runtimedeps.Deps`
  - `management/exec/file/topicbus/flow/varstore` 改为显式接收共享 deps
- backend 选择：
  - `flow.backend=json|pg`
  - `varstore.backend=memory|pg`
  - `state.pg.dsn`
  - `state.pg.flow_table`
  - `state.pg.varstore_table`
- 新增 PG backend（Server 仓内实现）：
  - `flow`：直接把完整 flow 定义存为 `jsonb`
  - `varstore`：只存 `(owner, name, value, value_type, visibility)`
  - schema 使用 `CREATE TABLE IF NOT EXISTS`
  - 每次操作按需连接 PG；本轮未引入长生命周期连接池
- 默认 backend 未改：
  - `flow.backend` 缺省时仍走 SubProto JSON backend
  - `varstore.backend` 缺省时仍走 SubProto memory backend
- 新增测试：
  - backend 非法值错误
  - 配置 `pg` 但未提供 DSN 错误
  - PG backend 连接失败错误

## 关键设计决策与权衡

- 将显式共享对象收敛到 `exec/runtimedeps`
  - 避免在多 module 仓根部再引入新的共享 module，减少循环依赖风险
- 默认 backend 仍留在 SubProto
  - `Server` 只在配置 `pg` 时注入外部 backend，避免复制默认 JSON/memory 语义
- PG backend 采用“按操作连接”
  - 本次优先减小接入面和生命周期复杂度
  - 代价是写路径会多一次连接开销；在当前数据量和频率下可接受
- `varstore` 只持久化 owner 权威记录
  - 远端回包 / notify / up_* 继续只刷新 cache，避免把逐跳缓存误写成权威真相

## plan.md 任务映射

- `SUB1` 显式化 runtime deps 与 capability registry 共享
- `SUB2` 为 `flow` 抽取 persistence 接口并保留 JSON backend
- `SUB3` 为 `varstore` 抽取 persistence 接口并固定 owner 写序
- `SRV1` 在 `Server` 引入 backend 选择、PG wiring 与 adapter
- `SRV2` 更新 `modules/defaultset` 构造路径与测试

## 需求 / 规范影响检查

- requirements impact：`none`
- specs impact：`clarify`
- 已更新长期 specs：
  - `docs/specs/flow.md`
  - `docs/specs/varstore.md`
- lessons：本轮未新增
- 需要更新 `docs/change/README.md` 索引

## 测试与验证方式 / 结果

### SubProto 联测

在 `D:\project\MyFlowHub3\worktrees\MyFlowHub-SubProto-refactor-pluggable-state` 临时生成 `go.work`，绑定本地 `Core/Proto` 与相关 subproto modules 后执行：

```powershell
go test ./... -count=1 -p 1
```

分别在以下 module 通过：

- `broker`
- `exec`
- `file`
- `flow`
- `management`
- `topicbus`
- `varstore`

### Server 联测

在 `D:\project\MyFlowHub3\worktrees\MyFlowHub-Server-refactor-pluggable-state` 临时生成 `go.work`，绑定本地 `Core/Proto` 与本次 SubProto worktree 后执行：

```powershell
go test ./... -count=1 -p 1
```

结果：通过。

## 潜在影响与回滚方案

### 潜在影响

- `flow.backend=pg` 或 `varstore.backend=pg` 时，启动 / 首次预热将依赖 PG 可达
- PG table 名目前只允许简单标识符，不支持 schema-qualified 名称
- `Server` 目前使用按操作连接 PG，而不是连接池

### 回滚

1. 回退 `MyFlowHub-SubProto` 中的 `exec/runtimedeps`、`flow/varstore` persistence 抽象与对应测试
2. 回退 `MyFlowHub-Server` 中 `modules/defaultset` 的显式 deps / backend 选择 / PG adapter 改动
3. 回退 `docs/specs/flow.md`、`docs/specs/varstore.md` 与本归档文档
4. 重新执行 SubProto / Server 对应 `go test`
