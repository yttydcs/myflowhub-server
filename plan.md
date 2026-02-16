# Plan - modules/defaultset 引入 build tags（裁切默认子协议集合）（PR6-BuildTags）

## Workflow 信息
- Repo：`MyFlowHub-Server`
- 分支：`refactor/server-defaultset-buildtags`
- Worktree：`d:\project\MyFlowHub3\worktrees\pr6-server-buildtags\MyFlowHub-Server`
- 参考总目标：`d:\project\MyFlowHub3\target.md`
- 约束：commit 信息使用中文（前缀如 `refactor:` 可英文）

## 当前状态
- 目前 hub_server 默认启用集合由 `modules.DefaultHub(cfg, log)` 提供。
- 默认集合的“具体模块列表”已解耦到 `modules/defaultset`（避免 `modules` 直接 import 所有 `subproto/*`）。
- `subproto/*` 已对齐为公开可装配子协议模块（不再使用 `internal/handler/*`）。
- 下一步（见 `target.md`）需要支持“子协议可裁切/可组装”，本 PR 先在 **默认集合** 层面引入 build tags，做到“编译期裁切默认集合”，同时保持默认行为不变。

> 环境备注（不进 git）：本仓库 `go.mod` 使用 `replace ../MyFlowHub-Core`、`../MyFlowHub-Proto`。  
> 在本 worktree 布局下，需要在 `d:\project\MyFlowHub3\worktrees\pr6-server-buildtags\` 下存在同名目录。  
> 当前通过 Junction 指向 `d:\project\MyFlowHub3\repo\MyFlowHub-Core` / `d:\project\MyFlowHub3\repo\MyFlowHub-Proto` 满足。

## 1) 需求分析
### 目标
1) 为 `modules/defaultset` 引入 build tags，使 hub_server 的默认启用集合支持**编译期裁切**（source-level cut/assemble 的第一步）。
2) 保持默认行为不变：不指定 tags 时，默认集合与当前一致（management/auth/varstore/topicbus/exec/flow/file + forward）。
3) 不改变 wire/业务语义：仅调整装配期的“默认集合构造方式”，不改各子协议实现。

### 范围
#### 必须（本 PR）
- `modules/defaultset`：
  - 默认构造函数保持不变（仍为 `DefaultHub(cfg, log)`，签名不变）。
  - 为以下子协议提供“禁用 tag”：
    - `noauth`：默认集合不包含 `subproto/auth`
    - `novarstore`：默认集合不包含 `subproto/varstore`
    - `notopicbus`：默认集合不包含 `subproto/topicbus`
    - `noexec`：默认集合不包含 `subproto/exec`
    - `noflow`：默认集合不包含 `subproto/flow`
    - `nofile`：默认集合不包含 `subproto/file`
  - `management` 与 default forward **保持始终启用**（保证至少一个 handler + default 存在，避免空集合）。
- 回归：
  - `go test ./... -count=1 -p 1`（Windows，默认构建）
  - `go test ./... -count=1 -p 1 -tags "nofile noflow noexec notopicbus novarstore noauth"`（Windows，验证“最大裁切”编译通过）

#### 可选（本 PR，如不增加风险）
- 在 `docs/change` 中补充一段“如何裁切构建”的示例命令（便于接手者复用）。

#### 不做（本 PR）
- 引入 Module registry/Deps（更复杂的模块依赖与运行期选择）。
- 修改 `cmd/hub_server` CLI/配置键以做运行期模块选择。
- Linux 构建验收。

### 使用场景
- 默认构建：行为与当前一致（全量默认集合）。
- 裁切构建：例如希望部署一个不包含 `file/flow/exec` 的轻量 hub_server，可通过 `-tags "nofile noflow noexec"` 编译。

### 功能需求
- 默认集合构造顺序与当前一致（当模块启用时保持原有顺序），避免因顺序变化引入潜在差异。
- `modules.DefaultHub` 对外行为不变；仍负责 `validateSet` 与返回 `modules.Set`。

### 非功能需求
- 性能：仅装配期构造；不引入运行期热路径开销。
- 可维护性：tag 命名清晰；每个模块“启用/禁用”实现成对存在，避免缺符号导致构建失败。

### 输入输出
- 输入：`DefaultHub(cfg, log)` + build tags（编译参数）。
- 输出：默认启用 handler 集合 + default fallback（forward）。

### 边界异常
- 裁切到仅 management：仍应满足 `validateSet`（handlers 非空、default 非 nil、subproto 不重复）。

### 验收标准
- 无 tags 时：默认集合不变（以回归测试通过 + 代码对比确认）。
- 开启 tags（示例组合）时：`go test` 通过，且默认集合中不包含被禁用模块的构造（代码层面由 build tag 保证）。
- `modules` 包本身不因本 PR 引入新 import cycle。

### 风险
- build tag 文件组织不当导致缺符号或重复定义（通过成对文件 + tag 约束避免）。
- tag 命名未来需要调整（通过 docs/change 记录并保持一致性，后续可兼容别名）。

## 2) 架构设计（分析）
### 总体方案（含选型理由 / 备选对比）
- 方案 A（采用）：在 `modules/defaultset` 内为每个可选子协议提供 `newXxxHandler()` 工厂函数，并用 build tags 提供“启用/禁用”的成对实现。
  - 优点：默认行为不变；裁切粒度可控；无需改动 `modules` 的对外 API；无运行期开销。
  - 缺点：文件数量增加（每个模块 2 个文件），但结构清晰、可审计。
- 方案 B（不选）：仅提供 `minimal` 单一 tag（`!minimal` / `minimal` 两份 DefaultHub 实现）。
  - 缺点：裁切粒度粗，难以满足“自由裁切/组装”；后续仍需重构。

### 模块职责
- `modules/defaultset`：承载默认集合构造策略，并在本 PR 引入 build tags 支持编译期裁切。
- `modules`：保持抽象与校验不变（Set/validateSet/RegisterAll/BindServerHooks）。

### 数据 / 调用流
1) `cmd/hub_server` 调用 `modules.DefaultHub(cfg, log)`
2) `modules.DefaultHub` 委托 `defaultset.DefaultHub(cfg, log)`
3) `defaultset.DefaultHub`：
   - 始终加入 management
   - 条件性加入（由 build tags 决定）auth/varstore/topicbus/exec/flow/file
   - 始终设置 default forward
4) `modules` 校验并返回 Set

### 错误与安全
- 不改变权限/路由/协议语义；仅变更装配集合构造。

### 性能与测试策略
- 性能：无运行期开销；仅装配期构造。
- 测试：
  - 默认构建全量回归
  - 最大裁切 tags 组合回归（确保 build tags 组织正确）

### 可扩展性设计点
- 后续可在此基础上增加：
  - 运行期模块选择（config/flags）
  - Module registry/Deps
  - 更细粒度的 build tag 命名规范（如按可执行文件前缀）

## 3.1) 计划拆分（Checklist）

## 问题清单（阻塞：否）
- 无（本 PR 不改 wire/业务语义；tag 命名与验收标准已在本文定义）。

### BT0 - 归档旧 plan.md 并准备本 workflow 文档
- 目标：保留历史 plan 的可审计性，不影响当前 workflow。
- 涉及文件：
  - `plan_archive_2026-02-16_modules-defaultset.md`
  - `plan.md`
- 验收条件：新 `plan.md` 仅描述本次 build tags 引入。
- 回滚点：revert 文档提交。

### BT1 - defaultset 拆分为 build-tag 工厂函数
- 目标：将 `modules/defaultset/DefaultHub` 改为调用 `newXxxHandler()`，并为每个模块提供启用/禁用成对文件。
- 涉及模块/文件（预期）：
  - `modules/defaultset/hub.go`（重构为条件性 append）
  - `modules/defaultset/*_enabled.go` + `*_disabled.go`（按 tag 分组）
- 验收条件：
  - 无 tags 时默认集合不变；
  - tags 组合构建通过；
  - 不引入 import cycle。
- 测试点：见 BT2。
- 回滚点：revert。

### BT2 - 回归测试（含 tags 组合）
- 目标：确保默认构建与裁切构建均通过。
- 验收条件：
  - `go test ./... -count=1 -p 1` 通过（Windows）
  - `go test ./... -count=1 -p 1 -tags "nofile noflow noexec notopicbus novarstore noauth"` 通过（Windows）

### BT3 - Code Review + 归档变更
- 目标：按模板完成审查与归档。
- 涉及文件：
  - `docs/change/YYYY-MM-DD_defaultset-buildtags.md`
- 验收条件：归档包含任务映射、关键决策、测试命令与回滚方案。

## 注意事项
- 禁止计划外改动：若需要引入 Module registry/Deps 或运行期选择，必须另起 workflow。

## 执行记录
- 2026-02-16：创建本 workflow worktree 与计划文档（待确认后进入 3.2）。

