# Plan - remote-authority-auth-release-server

## Workflow Information
- Repo: `MyFlowHub-Server`
- Branch: `chore/remote-authority-auth-release`
- Base: `main`
- Worktree: `D:\project\MyFlowHub3\worktrees\chore-server-remote-authority-auth-release`
- Current Stage: `3.1`

## Stage Records

### Initialization
- guide.md:
  - 已读取 `D:\project\MyFlowHub3\guide.md`
  - 约束确认：
    - commit 信息使用中文
    - 所有 worktree 必须位于 `D:\project\MyFlowHub3\worktrees`
    - 子协议稳定文档以 `repo\MyFlowHub-Server\docs` 为准
- base/worktree confirmation:
  - 主仓控制面路径仅用于 worktree / merge / release 管理
  - 本轮实现仅在当前 worktree 内完成
  - 关联 worktree：
    - `D:\project\MyFlowHub3\worktrees\chore-subproto-remote-authority-auth-release`

### Stage 1 - Requirements Analysis
#### Goal
- 在不引入计划外功能改动的前提下，将 `MyFlowHub-Server` 升级到新发布的 `myflowhub-subproto/auth v0.1.5`，并用 `GOWORK=off` 验证真实依赖链。

#### Scope
- 必须：
  - 将 `go.mod/go.sum` 中的 auth 依赖升级到 `v0.1.5`
  - 在 `GOWORK=off` 下完成模块解析、构建或测试验证
  - 为本次依赖升级补齐变更归档
- 可选：
  - 如依赖求解联动到其他校验和更新，同步最小必要 `go.sum`
- 不做：
  - 不修改业务逻辑
  - 不变更稳定需求或协议契约
  - 不发布新的 Server tag

#### Use Cases
- Server 需要消费包含 remote authority admin 相关修复与能力的 auth 新版本。
- CI / 单仓 checkout 需要在不依赖本地 `go.work` 的情况下通过真实依赖链验证。

#### Functional Requirements
- `go.mod` 中 `github.com/yttydcs/myflowhub-subproto/auth` 必须从 `v0.1.4` 升级到 `v0.1.5`。
- 升级后 `GOWORK=off` 必须能解析并通过关键验证。
- 归档需要明确这是发布链收口，而不是新的行为设计变更。

#### Non-functional Requirements
- 变更保持最小，只做依赖升级与必要归档。
- 验证必须可复现，不能依赖 sibling worktree 隐式注入。
- 出现版本解析失败时，应显式暴露为发布链问题，而不是静默回退。

#### Inputs / Outputs
- 输入：
  - 稳定需求：`D:\project\MyFlowHub3\docs\requirements\auth-controlled-admission.md`
  - 稳定规格：`D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\specs\auth.md`
  - 经验：`D:\project\MyFlowHub3\docs\lessons\cross-repo-semver-release.md`
  - 上游发布：`github.com/yttydcs/myflowhub-subproto/auth@v0.1.5`
- 输出：
  - `go.mod/go.sum` 升级结果
  - `GOWORK=off` 验证记录
  - 已 push 的分支 `origin/chore/remote-authority-auth-release`

#### Edge Cases
- 若上游 tag 尚不可解析，则必须停止本阶段，先回到 SubProto worktree 完成发布。
- 若 `go mod tidy` 联动更新间接依赖，只接受最小必要变化。
- 若测试暴露的是上游 tag 漏发，而非 Server 代码问题，应按 release-chain 处理。

#### Acceptance Criteria
- `go.mod` 中 auth 版本为 `v0.1.5`
- `GOWORK=off` 解析与测试通过
- 归档记录 requirements/specs/lessons 影响结论

#### Risks
- 上游 tag 未及时可见会导致 `go get` / `go test` 失败。
- 依赖升级可能暴露此前被本地 workspace 掩盖的问题。
- 若误并入其他依赖升级，会扩大回滚面。

#### Issue List
- 无

### Stage 2 - Architecture Design
#### Overall Solution
- 采用最小下游收口方案：
  - 等待 `MyFlowHub-SubProto/auth v0.1.5` tag 就绪
  - 在 Server worktree 内升级 `go.mod/go.sum`
  - 使用 `GOWORK=off` 执行依赖解析与测试
  - 通过后提交、推送，并在 stage 4 做归档

#### Alternatives Considered
- 备选 1：继续用本地 sibling worktree 联调，不做 `GOWORK=off`
  - 否决：无法证明真实发布链可用
- 备选 2：顺手升级更多 subproto 依赖
  - 否决：不符合最小变更原则，也会扩大回归面

#### Module Responsibilities
- `MyFlowHub-Server`
  - 消费已发布的 auth module，保持装配层稳定
- `MyFlowHub-SubProto/auth`
  - 提供本轮 Server 所需的新 auth 行为实现
- `docs/change`
  - 记录此次 release-chain 收口结果与验证证据

#### Data / Call Flow
- `go.mod` -> `go get github.com/yttydcs/myflowhub-subproto/auth@v0.1.5`
- `go mod tidy` -> 更新 `go.sum`
- `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/auth`
- `GOWORK=off go test ./... -count=1 -p 1`

#### Interface Drafts
- 无新增接口；只升级 semver 依赖

#### Error Handling and Safety
- 若 `go get` 解析失败，停止并回查上游 tag 发布状态。
- 若 `GOWORK=off` 测试失败，先判断是否为 release-chain 问题，再决定是否回到 3.1 调整计划。

#### Performance and Testing Strategy
- 不改运行时逻辑，重点验证真实依赖链与全量测试。
- 验证顺序：
  - `go list -m`
  - `go test ./... -count=1 -p 1`

#### Extensibility Design Points
- 保持 Server 作为装配层的职责边界，不把 auth 实现再塞回 Server。
- 归档中沿用 release-chain lesson 的检索路径，便于后续 patch 升级复用。

#### Issue List
- 无

### Stage 3.1 - Planning
#### Project Goal and Current State
- 目标：
  - 将 `MyFlowHub-Server` 的 auth 依赖从 `v0.1.4` 升级到 `v0.1.5`
- 当前状态：
  - 分支与 worktree 已创建
  - `go.mod` 当前为：
    - `myflowhub-core v0.4.9`
    - `myflowhub-proto v0.1.5`
    - `myflowhub-subproto/auth v0.1.4`
  - 当前尚无本 worktree 的 `plan.md`
  - 上游 auth 新 tag 尚未在本阶段确认发布完成

#### Docs Governance Routing Decision
- 使用 `$m-docs` 完成路由与影响判断
- Requirements impact: `none`
- Specs impact: `none`
- Related requirements:
  - `D:\project\MyFlowHub3\docs\requirements\auth-controlled-admission.md`
- Related specs:
  - `D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\specs\auth.md`
- Related lessons:
  - `D:\project\MyFlowHub3\docs\lessons\cross-repo-semver-release.md`
- 结论：
  - 本轮是既有 auth 规格的消费升级，不新增稳定 requirements/specs
  - 归档放入 `docs/change`，lesson 默认复用现有 `cross-repo-semver-release`

#### Related Requirements / Specs / Lessons
- Requirements:
  - `D:\project\MyFlowHub3\docs\requirements\auth-controlled-admission.md`
- Specs:
  - `D:\project\MyFlowHub3\repo\MyFlowHub-Server\docs\specs\auth.md`
- Lessons:
  - `D:\project\MyFlowHub3\docs\lessons\cross-repo-semver-release.md`

#### Executable Task List
- [ ] `SRVAUTH-1` 等待并确认上游 `auth/v0.1.5` 已发布可解析
- [ ] `SRVAUTH-2` 升级 Server 的 auth 依赖到 `v0.1.5`
- [ ] `SRVAUTH-3` 执行 `GOWORK=off` 验证并记录结果
- [ ] `SRVAUTH-4` 提交、推送，并在 stage 4 做归档

#### Task Details
##### SRVAUTH-1 - 确认上游 auth tag 可解析
- Owner: 主代理
- Worktree: `D:\project\MyFlowHub3\worktrees\chore-server-remote-authority-auth-release`
- Plan Path: `D:\project\MyFlowHub3\worktrees\chore-server-remote-authority-auth-release\plan.md`
- Goal:
  - 确认 `github.com/yttydcs/myflowhub-subproto/auth@v0.1.5` 已可被真实依赖解析
- Files / Modules:
  - 无本地代码写入
- Write Set:
  - 无
- Acceptance:
  - `go list -m ...@v0.1.5` 成功
- Test Points:
  - `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/auth@v0.1.5`
- Rollback:
  - 无代码写入，无需回滚

##### SRVAUTH-2 - 升级 Server auth 依赖
- Owner: 主代理
- Worktree: `D:\project\MyFlowHub3\worktrees\chore-server-remote-authority-auth-release`
- Plan Path: `D:\project\MyFlowHub3\worktrees\chore-server-remote-authority-auth-release\plan.md`
- Goal:
  - 将 Server 的 auth 依赖最小升级到新 patch 版本
- Files / Modules:
  - `go.mod`
  - `go.sum`
- Write Set:
  - `go.mod`
  - `go.sum`
- Acceptance:
  - `go.mod` 解析到 `github.com/yttydcs/myflowhub-subproto/auth v0.1.5`
- Test Points:
  - `GOWORK=off go get github.com/yttydcs/myflowhub-subproto/auth@v0.1.5`
  - `GOWORK=off go mod tidy`
- Rollback:
  - 回退依赖升级提交

##### SRVAUTH-3 - 验证真实依赖链
- Owner: 主代理
- Worktree: `D:\project\MyFlowHub3\worktrees\chore-server-remote-authority-auth-release`
- Plan Path: `D:\project\MyFlowHub3\worktrees\chore-server-remote-authority-auth-release\plan.md`
- Goal:
  - 证明 Server 在不依赖本地 `go.work` 的情况下可通过回归
- Files / Modules:
  - 只读验证
- Write Set:
  - 无
- Acceptance:
  - `GOWORK=off go test ./... -count=1 -p 1` 通过
- Test Points:
  - `GOWORK=off go list -m github.com/yttydcs/myflowhub-subproto/auth`
  - `GOWORK=off go test ./... -count=1 -p 1`
- Rollback:
  - 若失败且确认为版本问题，停止并回溯上游发布链

##### SRVAUTH-4 - 提交、推送与归档
- Owner: 主代理
- Worktree: `D:\project\MyFlowHub3\worktrees\chore-server-remote-authority-auth-release`
- Plan Path: `D:\project\MyFlowHub3\worktrees\chore-server-remote-authority-auth-release\plan.md`
- Goal:
  - 固化 Server 消费升级结果并准备 workflow 归档
- Files / Modules:
  - `go.mod`
  - `go.sum`
  - `docs/change/*`
  - `docs/change/README.md`
- Write Set:
  - 以上文件与 git 分支历史
- Acceptance:
  - 分支已 push，归档已记录验证与影响结论
- Test Points:
  - `git push origin chore/remote-authority-auth-release`
- Rollback:
  - 回退当前分支提交；如已 push，追加回退提交

#### Dependencies
- 上游依赖：
  - `D:\project\MyFlowHub3\worktrees\chore-subproto-remote-authority-auth-release` 必须先完成 `AUTHREL-4`
- 下游依赖：
  - 无，本轮不继续扩散到应用层

#### Risks and Notes
- 该任务是典型 release-chain 收口；真实风险在 semver 可解析性，而非业务代码。
- 若 `go get` 时联动其他 indirect 版本变化，需要判断是否为 auth 升级必需。
- 若验证发现 stable spec 与实现仍不一致，需回到阶段 1/2 重新确认，而不是直接归档。

#### Parallelism Assessment
- 当前不派发子 Agent。
- 原因：
  - 必须等待上游 auth tag 发布完成后再实施
  - 当前写集极小，主代理直接执行更可控

#### Issue List
- 无

阻塞：否
进入 3.2
