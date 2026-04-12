# 2026-04-13_server-tag-docker-image

## 变更背景 / 目标
- 当前 `MyFlowHub-Server` 的 `v*` tag workflow 只会产出 Windows/Linux 压缩包并创建 GitHub Release。
- 为了方便直接在服务器上部署，需要让同一条 tag 流水线额外发布可拉取的 Docker 镜像。
- 本次目标是在不改变 `hub_server` 运行时行为的前提下，补齐镜像构建定义和 GHCR 发布链路。

## 具体变更内容
- 新增 `Dockerfile`
  - 使用多阶段构建编译 `./cmd/hub_server`
  - 运行时镜像采用 Alpine，内置 CA 证书
  - 以非 root 用户运行，并把可写工作目录固定到 `/data`
  - 默认暴露 `9000/tcp` 与 `9000/udp`
- 新增 `.dockerignore`
  - 排除 `.git`、`docs`、`tests`、`dist`、`plan.md` 和本地 `config/runtime_config.json`
  - 减少 Docker build context，避免把本地运行态文件打进镜像
- 更新 `.github/workflows/release-hub-server.yml`
  - 保留现有 `v*` tag 触发、二进制构建、压缩包和 GitHub Release 流程
  - 新增 `packages: write` 权限
  - 新增 Docker Buildx、GHCR 登录、metadata 提取和镜像推送步骤
  - 镜像名使用 `ghcr.io/yttydcs/myflowhub-server`
  - 镜像 tag 直接复用 git tag，例如 `v0.0.15`
  - 首版仅发布 `linux/amd64`

## Requirements impact
`none`

## Specs impact
`none`

## Lessons impact
`none`

## Related requirements
- `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\docs\requirements\README.md`

## Related specs
- `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\docs\specs\README.md`

## Related lessons
- none

## 对应 plan.md 任务映射
- `T1` - Add Docker packaging assets
- `T2` - Extend tag workflow for GHCR publication
- `T3` - Validate and review

## 经验 / 教训摘要
- 对 GitHub 托管仓库，GHCR 是最小改动的镜像发布落点，因为可以直接复用 `GITHUB_TOKEN`，不需要额外注入 Docker Hub 凭据。
- 对带有上层 `go.work` 的 MyFlowHub3 worktree，本仓库级验证应显式使用 `GOWORK=off`，否则 `go test ./...` 可能因为 workspace 选择而失败。
- 对 tag 驱动的发布场景，直接复用原始 git tag 作为镜像 tag，排查和部署时最不容易发生版本映射歧义。

## 可复用排查线索
- 症状
  - 推送 `v*` tag 后只有 GitHub Release 压缩包，没有可拉取镜像
  - 本地 `go test ./...` 在 worktree 内报 `directory prefix . does not contain modules listed in go.work`
- 触发条件
  - release workflow 只有 `contents: write`，没有容器 registry 发布步骤
  - 在 `D:\project\MyFlowHub3` 工作区内直接按默认 workspace 模式跑 repo 级测试
- 关键词
  - `release-hub-server`
  - `packages: write`
  - `docker/build-push-action`
  - `ghcr.io/yttydcs/myflowhub-server`
  - `GOWORK=off`
- 快速检查
  - 检查 `.github/workflows/release-hub-server.yml` 是否包含 GHCR login / metadata / build-push
  - 检查镜像 tag 是否来自 `type=ref,event=tag`
  - 本地验证使用 `GOWORK=off go test ./...`

## 关键设计决策与权衡
- 选择 GHCR 而不是 Docker Hub：
  - 原因是 GHCR 可直接使用 GitHub Actions 内置令牌，减少凭据管理和失败面
- 选择 `linux/amd64` 单平台首发：
  - 原因是当前 release workflow 现有产物本来就是 amd64，先做最小安全闭环
- 选择 Alpine 运行时镜像而不是 distroless：
  - 原因是本次更关注可写工作目录、非 root 运行和 CA 证书的直观可控配置
- 不增加 `latest` 镜像别名：
  - 原因是当前需求只要求 tag 驱动部署，先保持版本到镜像 tag 的一一对应

## 测试与验证方式 / 结果
- `go test ./...`
  - 结果：失败
  - 原因：受上层 `go.work` 影响，当前 worktree 不是该 workspace 选中的 module 前缀
- `GOWORK=off go test ./...`
  - 结果：通过
- `git diff --check`
  - 结果：通过
  - 备注：只有 Git 的 CRLF 提示，没有 whitespace 错误
- 本地 Docker build
  - 结果：未执行
  - 原因：当前环境没有 `docker` CLI / daemon，首个镜像发布证明将依赖 GitHub Actions tag run

## 潜在影响
- 首次镜像推送的最终闭环证明要等真实 tag workflow 运行。
- 如果 GHCR 包默认仍是私有可见性，服务器拉取时需要 GHCR token，或者后续手动把包调整为公开。

## 回滚方案
- 若分支尚未合并、tag 尚未推送：
  - 删除 `Dockerfile`、`.dockerignore` 与 workflow 扩展步骤
- 若 tag 已推送且镜像已发布：
  - 不重写既有 tag
  - 通过更高 patch tag 发布修正版本
  - 如有必要，手动删除对应 GHCR 包版本作为运营层回收动作

## 子Agent执行轨迹
- 本轮未使用子Agent。
