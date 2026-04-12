# Plan - server-tag-docker-image

## Workflow Information
- Repo: `D:\project\MyFlowHub3\repo\MyFlowHub-Server`
- Branch: `feat/server-tag-docker-image`
- Base: `main` @ `c1ec782`
- Worktree: `D:\project\MyFlowHub3\worktrees\server-tag-docker-image`
- Current Stage: `4`

## Stage Records

### Initialization
- guide.md: read `D:\project\MyFlowHub3\guide.md`
- base/worktree confirmation:
  - repo baseline branch is `main`
  - control-plane repo stays untouched except worktree management
  - active execution worktree is `D:\project\MyFlowHub3\worktrees\server-tag-docker-image`

### Stage 1 - Requirements Analysis
#### Goal
- Keep the existing tag-triggered release workflow for `repo/MyFlowHub-Server` and additionally publish a deployable Docker image so tagged releases can be pulled directly on servers.

#### Scope
- Must:
  - keep the current `v*` tag trigger and GitHub Release asset generation working
  - add a repository-managed container build definition for `cmd/hub_server`
  - publish a versioned container image from the same tag workflow without requiring new user-managed secrets
- Optional:
  - expose sensible container runtime defaults for TCP / QUIC and writable runtime config storage
  - attach OCI labels/metadata to the pushed image
- Not in scope:
  - changing hub runtime behavior or protocol defaults
  - adding deployment orchestration files such as `docker-compose.yml` or Kubernetes manifests
  - changing release tag strategy or cross-repo release order

#### Use Cases
- A tagged Server release should still generate downloadable zip assets for Windows/Linux users.
- A server operator should be able to deploy the same tagged version via `docker pull ghcr.io/yttydcs/myflowhub-server:<tag>`.
- The container should persist `config/runtime_config.json` and any auto-generated runtime artifacts under a writable workdir.

#### Functional Requirements
- Extend the existing `release-hub-server` workflow so `push.tags = v*` also builds and pushes a container image.
- Add a `Dockerfile` that builds `./cmd/hub_server` into a runnable Linux container image.
- Provide a `.dockerignore` to keep build context small and avoid shipping repo noise.
- Keep the container default command aligned with the existing binary entrypoint and env-driven configuration model.

#### Non-functional Requirements
- Prefer the smallest safe change surface and reuse GitHub-native auth.
- Avoid introducing environment-specific hard-coded endpoints or secrets.
- Keep the image suitable for production-style deployment: non-root runtime, writable workdir, CA certificates present.
- Keep release workflow understandable and auditable.

#### Inputs / Outputs
- Inputs:
  - git tag matching `v*`
  - repository source at the tagged commit
  - GitHub Actions built-in token
- Outputs:
  - existing GitHub Release zip artifacts and checksums
  - published container image in GHCR under the repository owner namespace

#### Edge Cases
- Docker tooling should not interfere with GitHub Release creation if the binary packaging still succeeds.
- The container must remain writable for `config/runtime_config.json` even when no config file exists at startup.
- Missing Docker on the local workstation is acceptable; validation must still cover static correctness and Go regression checks.

#### Acceptance Criteria
- Repository contains a working container build definition for `hub_server`.
- Tag workflow YAML includes a GHCR login/metadata/build-push path while preserving the existing release asset path.
- Local regression validation for the touched Go code paths passes.
- Docs governance can record `Requirements impact: none` and `Specs impact: none`.

#### Risks
- Image registry/tag naming must be derived from GitHub context correctly or pushes will fail at release time.
- Container runtime defaults could accidentally make the config directory unwritable if the final image user/workdir is misconfigured.
- No local Docker engine is available in this workspace, so image validation is limited to static review plus CI-oriented correctness.

#### Issue List
- none

### Stage 2 - Architecture Design
#### Overall Solution
- Use the existing GitHub Actions release workflow as the single tag-triggered release entry.
- Add a multi-stage `Dockerfile` that compiles `cmd/hub_server` with `CGO_ENABLED=0` and runs it in a minimal non-root Alpine image.
- Extend `.github/workflows/release-hub-server.yml` with a second path that:
  - logs in to `ghcr.io` using `GITHUB_TOKEN`
  - derives image tags from the pushed git tag
  - builds and pushes `ghcr.io/yttydcs/myflowhub-server`
- Keep the first iteration to `linux/amd64` only because the current release workflow already targets amd64 and this is the smallest safe deployment addition.

#### Alternatives Considered
- Docker Hub:
  - rejected for now because it would require extra repository secrets and manual registry credential management
- multi-arch (`linux/amd64,linux/arm64`) image publishing:
  - deferred because the current release baseline is amd64-only and this task does not include arm64 runtime validation
- distroless final image:
  - viable, but Alpine is simpler here because writable workdir ownership and CA certificate provisioning are straightforward with a minimal change

#### Module Responsibilities
- `Dockerfile`:
  - build `hub_server`
  - define runtime image user, workdir, exposed ports, and entrypoint
- `.dockerignore`:
  - exclude git/docs/worktree noise from the Docker build context
- `.github/workflows/release-hub-server.yml`:
  - keep binary release job behavior
  - add GHCR permissions and image publish steps

#### Data / Call Flow
- `git push origin vX.Y.Z`
- GitHub Actions starts `release-hub-server`
- workflow checks out the tagged commit
- workflow builds zip release artifacts as before
- workflow authenticates to GHCR with `github.actor` + `github.token`
- workflow builds the Docker image from repository source and pushes `ghcr.io/yttydcs/myflowhub-server:vX.Y.Z`
- workflow creates/updates the GitHub Release with existing zip assets

#### Interface Drafts
- Container runtime env defaults:
  - `HUB_ADDR=:9000`
  - `HUB_QUIC_ADDR=:9000`
  - `HUB_WORKDIR=/data`
- Exposed ports:
  - `9000/tcp`
  - `9000/udp`
- Image reference:
  - `ghcr.io/yttydcs/myflowhub-server:<git-tag>`

#### Error Handling and Safety
- Keep workflow failure behavior explicit; if image push fails, the job fails instead of silently skipping Docker publication.
- Use only GitHub-provided auth for GHCR to avoid plaintext credentials.
- Run the container as a non-root user and ensure `/data` is writable.

#### Performance and Testing Strategy
- Add `.dockerignore` to reduce context upload size.
- Reuse Go module layer caching in the Dockerfile by copying `go.mod`/`go.sum` first.
- Validate with `go test ./...`.
- Local Docker build is not available; rely on workflow/static review for container publication correctness.

#### Extensibility Design Points
- Future iterations can add `latest` or multi-arch tags without redesigning the workflow structure.
- Future deployment docs can reuse the committed image reference and runtime env model.

#### Issue List
- none

### Stage 3.1 - Planning
#### Project Goal and Current State
- Goal: make `MyFlowHub-Server` tagged releases also publish a deployable Docker image.
- Current state:
  - tag workflow already creates GitHub Release zip artifacts
  - repository has no `Dockerfile` or `.dockerignore`
  - local machine has no Docker CLI/daemon available for image smoke builds

#### Docs Governance Routing Decision
- 使用 `$m-docs` 校验计划文档路由、requirements/specs 影响和 lessons 查询入口。
- Canonical destinations:
  - stable truth: unchanged, existing `docs/requirements` / `docs/specs`
  - workflow control: this worktree root `plan.md`
  - workflow result archive: `docs/change/2026-04-13_server-tag-docker-image.md`
  - reusable lessons: currently not expected unless CI/image publication reveals a non-obvious recurring failure mode
- Requirements impact: `none`
- Specs impact: `none`

#### Related Requirements / Specs / Lessons
- Related requirements:
  - `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\docs\requirements\README.md`
- Related specs:
  - `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\docs\specs\README.md`
- Related lessons:
  - `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\docs\lessons\README.md`

#### Executable Task List
- `T1` add container packaging assets for `hub_server`
- `T2` extend tag release workflow to publish GHCR image
- `T3` run regression validation and static review for the new release path

#### Task Details
##### T1 - Add Docker packaging assets
- Owner: main agent
- Worktree: `D:\project\MyFlowHub3\worktrees\server-tag-docker-image`
- Plan Path: `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\plan.md`
- Goal: create a minimal runtime image definition with non-root execution and writable runtime config storage
- Files / Modules:
  - `Dockerfile`
  - `.dockerignore`
- Write Set:
  - `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\Dockerfile`
  - `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\.dockerignore`
- Acceptance:
  - Docker build definition matches `cmd/hub_server` startup model
  - runtime workdir is writable and env defaults are explicit
- Test Points:
  - static review of workdir/user/entrypoint
  - regression `go test ./...`
- Rollback:
  - delete `Dockerfile` and `.dockerignore`

##### T2 - Extend tag workflow for GHCR publication
- Owner: main agent
- Worktree: `D:\project\MyFlowHub3\worktrees\server-tag-docker-image`
- Plan Path: `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\plan.md`
- Goal: keep release artifacts and add versioned GHCR image publication on `v*` tags
- Files / Modules:
  - `.github/workflows/release-hub-server.yml`
- Write Set:
  - `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\.github\workflows\release-hub-server.yml`
- Acceptance:
  - workflow keeps current zip/checksum/release path
  - workflow adds `packages: write` and a GHCR build/push path using tag-derived metadata
- Test Points:
  - static review of workflow trigger/permissions/order
  - regression `go test ./...`
- Rollback:
  - revert workflow file to the pre-Docker release version

##### T3 - Validate and review
- Owner: main agent
- Worktree: `D:\project\MyFlowHub3\worktrees\server-tag-docker-image`
- Plan Path: `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\plan.md`
- Goal: ensure the change is internally consistent despite missing local Docker tooling
- Files / Modules:
  - touched files only
- Write Set:
  - no new write target beyond T1/T2 unless Stage 3.2 uncovers a defect
- Acceptance:
  - `go test ./...` passes
  - final review can mark the stage checklist as passing
- Test Points:
  - `go test ./...`
  - manual diff review
- Rollback:
  - revert the touched files and do not publish the branch/tag

#### Dependencies
- GitHub Actions hosted runner must support Docker Buildx and GHCR login via `GITHUB_TOKEN`.
- Release tag format remains `v*`.

#### Risks and Notes
- Because local Docker is unavailable, the first real image publication proof will happen in GitHub Actions on a tag run.
- To keep the change small, image tags will initially track the pushed version tag only, without a `latest` alias.

#### Parallelism Assessment
- No safe parallel split is needed.
- Reason:
  - only three touched files
  - Dockerfile and workflow are tightly coupled
  - sub-agent delegation is unnecessary and not used in this workflow

#### Issue List
- none

### Stage 3.3 - Code Review
- 需求覆盖: 通过
  - `T1/T2/T3` 已覆盖镜像构建定义、tag workflow 扩展和回归验证
- 架构合理性: 通过
  - 复用现有 tag workflow，未引入额外 registry 密钥或新的发布入口
- 性能风险（N+1 / 重复计算 / 多余 I/O / 锁竞争）: 通过
  - `.dockerignore` 控制了构建上下文；运行时代码未改动
- 可读性与一致性: 通过
  - 镜像 tag、镜像名和工作目录约定都保持显式
- 可扩展性与配置化: 通过
  - 运行参数仍完全走现有 env/flag 模型；后续可单独扩展多架构或 `latest`
- 稳定性与安全: 通过
  - GHCR 使用 GitHub 内置令牌；容器以非 root 用户运行并保留可写工作目录
- 测试覆盖情况: 通过
  - `GOWORK=off go test ./...` 已通过；本地 Docker 构建因环境缺少 `docker` 未执行，已记录为剩余验证项
- 子Agent治理与审计（任务映射、上下文完整性、文件所有权、结果复核、冲突处理、记录完整性）: 通过
  - 本轮未使用子Agent

### Stage 4 - Change Archive
- Archive path:
  - `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\docs\change\2026-04-13_server-tag-docker-image.md`
- Requirements impact: `none`
- Specs impact: `none`
- Lessons impact: `none`
- Index updates:
  - `D:\project\MyFlowHub3\worktrees\server-tag-docker-image\docs\change\README.md`

阻塞：否
Stage 4 已完成，等待是否结束 workflow
