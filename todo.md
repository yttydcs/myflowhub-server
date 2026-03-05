# Plan - MyFlowHub-Server：升级 subproto/file 到 v0.1.2（mkdir）

## Workflow 信息
- 仓库：`MyFlowHub-Server`
- 分支：`chore/server-bump-file-v0.1.2`
- Worktree：`d:\project\MyFlowHub3\worktrees\MyFlowHub-Server-file-v012`
- Base：`main`
- 当前状态：已完成（待你确认是否结束 workflow）

## 项目目标与当前状态
- 目标：
  - 将 Server 依赖 `github.com/yttydcs/myflowhub-subproto/file` 升级到 `v0.1.2`；
  - 同步 Server 协议文档 `docs/5-file.md`，补充 `op=mkdir` 说明；
  - 执行回归验证，确保构建与测试不回退。
- 现状：
  - `go.mod` 依赖仍为 `v0.1.1`；
  - `docs/5-file.md` 仅覆盖 `pull/list/read_text/offer`。

## 可执行任务清单（Checklist）

- [x] `SRV-FILE-1` 升级依赖版本
  - 目标：`go.mod/go.sum` 切换到 `myflowhub-subproto/file v0.1.2`。
  - 涉及文件：
    - `go.mod`
    - `go.sum`
  - 验收条件：
    - `go list -m github.com/yttydcs/myflowhub-subproto/file` 显示 `v0.1.2`。
  - 测试点：
    - `go mod tidy` / `go test` 可通过（在当前 workspace 条件下）。
  - 回滚点：
    - 版本回退到 `v0.1.1`。

- [x] `SRV-FILE-2` 同步协议文档
  - 目标：在 `docs/5-file.md` 新增 `op=mkdir` 语义、请求/响应与权限说明。
  - 涉及文件：
    - `docs/5-file.md`
  - 验收条件：
    - 文档与 `v0.1.2` 行为一致；
    - 不修改既有 action/SubProto 编号语义。
  - 回滚点：
    - 回滚文档改动。

- [x] `SRV-FILE-3` 回归验证 + 归档
  - 目标：执行测试并生成变更归档。
  - 涉及文件：
    - `docs/change/2026-03-05_bump-subproto-file-v0.1.2.md`
  - 验收条件：
    - 测试结果明确记录；
    - 归档文档包含任务映射、影响与回滚。
  - 回滚点：
    - 文档层可独立回滚。

## 依赖与风险
- 依赖：SubProto 仓库 `file/v0.1.2` tag 已可解析。
- 风险：网络/代理不可达会导致 `go get` 失败；若出现，可改为本地 workspace 验证并在文档标注。
