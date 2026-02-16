# 2026-02-16 defaultset build tags（裁切默认子协议集合）

## 背景 / 目标

`hub_server` 的默认启用模块集合由 `modules.DefaultHub(cfg, log)` 提供。随着“子协议可裁切/可组装”的需求推进，需要支持在**编译期**按需裁切默认集合（source-level cut/assemble 的第一步），以便后续实现更自由的组合与发布形态。

本次变更目标：

1. 为 `modules/defaultset` 引入 build tags，使默认集合支持**编译期裁切**。
2. **默认行为不变**：不指定 tags 时，默认集合与当前一致（`management/auth/varstore/topicbus/exec/flow/file + forward`）。
3. **不改变 wire/业务语义**：仅调整“默认集合构造策略”，不改各子协议实现细节。

## 具体变更内容

### 1) 默认集合构造改为工厂函数 + 条件性 append

- `modules/defaultset/hub.go` 将默认集合构造改为：
  - 始终加入 `management.NewHandler(log)`
  - 依次尝试加入（按原有顺序）：`auth/varstore/topicbus/exec/flow/file`
  - 始终设置 default fallback：`forward.NewDefaultForwardHandler(cfg, log)`
- 每个可裁切模块通过 `newXxxHandler(cfg, log)` 工厂函数返回 handler；当对应 tag 禁用时返回 `nil`，从而不会加入默认集合。

### 2) 新增 build tags（负向 tags：指定即禁用）

以下 tags 作用于“默认集合是否包含对应子协议模块”：

- `noauth`：默认集合不包含 `subproto/auth`
- `novarstore`：默认集合不包含 `subproto/varstore`
- `notopicbus`：默认集合不包含 `subproto/topicbus`
- `noexec`：默认集合不包含 `subproto/exec`
- `noflow`：默认集合不包含 `subproto/flow`
- `nofile`：默认集合不包含 `subproto/file`

实现方式为每个模块提供成对文件：

- `*_enabled.go`：`//go:build !noX`
- `*_disabled.go`：`//go:build noX`

从而避免“缺符号 / 重复定义”的构建风险，并确保默认（无 tags）行为不变。

### 3) 永远启用项

- `management` 与 default forward **始终启用**，用于保证：
  - `handlers` 非空
  - `default` 非空
  - 通过上层 `modules` 的 `validateSet` 校验

## 任务映射（plan.md）

- BT1 - defaultset 拆分为 build-tag 工厂函数
  - 对应提交：`b377c1c`
- BT2 - 回归测试（含 tags 组合）
  - 通过：见下方“测试与验证”
- BT3 - Code Review + 归档变更
  - 本文 + Review 结论（见下方）

## 关键设计决策与权衡

1. 采用“负向 tags”（`noauth/...`），默认无 tags 即全量启用：
   - 优点：默认构建不需要额外参数；发布/开发体验更自然。
   - 代价：tag 名称是“禁用语义”，需要在文档中固定并持续维护一致性。
2. 采用“工厂函数 + 成对文件”的组织方式：
   - 优点：模块边界清晰；启用/禁用逻辑可审计；避免 build tags 复杂分支导致的缺符号问题。
   - 代价：文件数量增加（每个模块 2 个文件），但换来更稳定的构建行为。
3. `management` 与 `forward` 固定启用：
   - 目的：保证即使“最大裁切”，也不会出现空集合导致启动期校验失败。

## 测试与验证

Windows（Linux 构建验收暂忽略）：

- 默认构建回归：
  - `go test ./... -count=1 -p 1`
- 最大裁切回归：
  - `go test ./... -count=1 -p 1 -tags "nofile noflow noexec notopicbus novarstore noauth"`

结果：上述两条命令均通过。

## Code Review 结论（3.3）

- 需求覆盖：通过（默认不变；支持 tags 裁切；不改 wire 语义）
- 架构合理性：通过（defaultset 仅负责默认装配；上层校验不变；无 import cycle）
- 性能风险：通过（仅装配期构造，无运行期热路径开销）
- 可读性与一致性：通过（命名清晰；顺序保持；成对文件规则一致）
- 可扩展性与配置化：通过（新增模块仅需新增 `newXxxHandler` 与 tag 文件对）
- 稳定性与安全：通过（不改变权限/路由；management+forward 固定启用避免空集合）
- 测试覆盖情况：通过（默认 + 最大裁切两套构建回归）

## 潜在影响与回滚方案

### 潜在影响

- 未来若引入新的默认子协议模块，需要同步补齐对应的 `newXxxHandler()` 以及（如需可裁切）对应的 build tag 文件对。
- 使用者若在构建参数中指定了 `no*` tags，将导致对应模块在默认集合中不可用（属于预期行为）。

### 回滚方案

- 回滚提交：`git revert b377c1c`（移除 build tags 裁切逻辑）
- 同步删除本文档对应提交（或一并 revert 文档提交）

