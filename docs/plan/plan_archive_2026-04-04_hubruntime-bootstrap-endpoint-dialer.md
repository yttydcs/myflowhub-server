# Plan - Hubruntime Bootstrap Reuse Parent Endpoint Dialer

## Workflow Information
- Repo: `MyFlowHub-Server`
- Branch: `feat/bootstrap-endpoint-dialer`
- Base: `origin/master`
- Worktree: `D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/MyFlowHub-Server`
- Current Stage: `completed`

## Stage Records

### Initialization
- guide.md: `D:/project/MyFlowHub3/guide.md` reviewed; workflow uses dedicated worktree and later Chinese commit wording.
- base/worktree confirmation: dedicated worktree under `D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/`; no implementation in main repo path.

### Stage 1 - Requirements Analysis
#### Goal
- Make pre-start bootstrap reuse the same generic parent endpoint dial path already used for persistent parent links, without changing the two-stage bootstrap/rebind semantics.

#### Scope
- Must:
  - remove the current pre-start TCP-only gate in `hubruntime.Runtime.Start`
  - route pre-start bootstrap through the generic parent endpoint dialer path
  - keep post-start `sendRegisterOnConn(...)` behavior unchanged
- Optional:
  - add small test seams if required for deterministic unit tests
- Not in scope:
  - serial transport support itself
  - removing startup bootstrap
  - auth protocol contract changes
  - parent watcher/reconnect redesign

#### Use Cases
- `ParentEndpoint=quic://...` should work for startup bootstrap the same way persistent parent connection dialing already works.
- future transports such as RFCOMM or serial-backed dialers should only need runtime endpoint support once, not separate bootstrap-specific code.

#### Functional Requirements
- `Runtime.Start` must still self-register when `parent.enable + self_id + parent target` are present.
- `selfRegisterNodeID(...)` must obtain/confirm `node_id` through the generic bootstrap helper path.
- existing `ParentAddr` and bare `host:port` compatibility must stay intact.

#### Non-functional Requirements
- minimum safe change surface in `hubruntime`
- preserve explicit startup failures on invalid parent endpoint or bootstrap failure
- avoid duplicating endpoint parsing rules

#### Inputs / Outputs
- Inputs:
  - runtime options with `ParentEndpoint` or `ParentAddr`, `SelfID`, `ParentJoinPermit`
- Outputs:
  - startup either derives/overrides `NodeID` successfully or fails with explicit error

#### Edge Cases
- unsupported endpoint scheme
- dial failure on non-TCP parent endpoint
- bootstrap returns pending/rejected
- configured `NodeID` differs from assigned `node_id`

#### Acceptance Criteria
- pre-start bootstrap no longer rejects supported non-TCP parent endpoint schemes up front
- startup still reuses assigned `node_id` override behavior
- runtime tests cover the generic endpoint path or its seams

#### Risks
- runtime tests may need a dialer seam because real QUIC/RFCOMM bootstrap is too heavy for unit tests
- changing bootstrap wiring must not affect post-start watcher semantics

#### Issue List
- None.

### Stage 2 - Architecture Design
#### Overall Solution
- Reuse existing runtime endpoint ownership:
  - `parseParentEndpoint(...)` continues to validate supported schemes
  - `dialParentEndpoint(...)` continues to materialize `core.IConnection`
- change `selfRegisterNodeID(...)` to accept a `func(context.Context, string) (core.IConnection, error)` dialer and pass it into `bootstrap.SelfRegisterOptions`
- remove the explicit `parentScheme != "tcp"` guard from `Runtime.Start`

#### Alternatives Considered
- Keep bootstrap TCP-only and defer this to a later transport-specific workflow:
  - rejected because it duplicates parent dial capability and blocks already-supported `quic://` parent endpoints during startup.
- Move endpoint parsing into Core bootstrap:
  - rejected because Server already owns endpoint syntax and transport selection.

#### Module Responsibilities
- `hubruntime/runtime.go`:
  - parse parent target once for validation
  - invoke bootstrap helper with runtime endpoint dialer
  - keep watcher/rebind logic unchanged
- `hubruntime/runtime_test.go`:
  - cover scheme acceptance and dialer wiring

#### Data / Call Flow
- `Runtime.Start(...)`
- validate effective parent target with `parseParentEndpoint(...)`
- if startup bootstrap conditions match:
  - call `selfRegisterNodeID(ctx, parentTarget, selfID, joinPermit, dialParentEndpoint, log)`
  - helper delegates dialing to Core bootstrap via injected dialer
- after `srv.Start(...)`, existing watcher sends register on persistent parent connection as before

#### Interface Drafts
- `selfRegisterNodeID(ctx, parentTarget, selfID, joinPermit string, dialer func(context.Context, string) (core.IConnection, error), log *slog.Logger) (uint32, error)`

#### Error Handling and Safety
- keep endpoint parse failure as startup error
- keep node-id zero and non-approved bootstrap as explicit failures
- no silent fallback from unsupported schemes to TCP

#### Performance and Testing Strategy
- no steady-state performance impact; startup path only
- add/update runtime tests for the new dialer path and the removed TCP-only gate

#### Extensibility Design Points
- once runtime supports a new parent endpoint scheme, startup bootstrap inherits it automatically
- serial transport can later plug into the same runtime dialer path without another bootstrap-specific fork

#### Issue List
- None.

### Stage 3.1 - Planning
#### Project Goal and Current State
- Goal: make startup bootstrap reuse the same endpoint abstraction already used by persistent parent dialing.
- Current state: `Runtime.Start` validates parent endpoint generically, but pre-start bootstrap still rejects non-TCP schemes and calls `bootstrap.SelfRegister` with `ParentAddr` only.

#### Docs Governance Routing Decision
- Using `$m-docs` for routing.
- Requirements impact: `none`
- Specs impact: `none`
- Lessons impact: `none`
- Stable truth remains in:
  - `docs/specs/auth.md`
  - `docs/specs/core.md`
- This worktree root `plan.md` is workflow control only; completed results go to `docs/change/`.

#### Related Requirements / Specs / Lessons
- Related requirements: `none`
- Related specs:
  - `docs/specs/auth.md`
  - `docs/specs/core.md`
- Related lessons: `none`

#### Executable Task List
- [x] `SRV-BOOT-1` wire startup bootstrap through generic endpoint dialer
- [x] `SRV-BOOT-2` update runtime tests for non-TCP-capable bootstrap path
- [x] `SRV-BOOT-3` run Server test suite relevant to hubruntime/bootstrap

#### Task Details
##### `SRV-BOOT-1` - Runtime bootstrap wiring
- Owner: main agent
- Worktree: `D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/MyFlowHub-Server`
- Plan Path: `D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/MyFlowHub-Server/plan.md`
- Goal: remove TCP-only bootstrap gate and reuse `dialParentEndpoint(...)` for pre-start bootstrap.
- Files / Modules:
  - `hubruntime/runtime.go`
- Write Set:
  - `hubruntime/runtime.go`
- Acceptance:
  - supported parent endpoint schemes no longer fail before bootstrap dialing
  - node-id assignment semantics remain unchanged
- Test Points:
  - supported non-TCP endpoint reaches bootstrap helper path
  - invalid endpoint still fails early
- Rollback:
  - revert runtime bootstrap wiring changes in this branch

##### `SRV-BOOT-2` - Runtime tests
- Owner: main agent
- Worktree: `D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/MyFlowHub-Server`
- Plan Path: `D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/MyFlowHub-Server/plan.md`
- Goal: prove pre-start bootstrap now respects the generic endpoint dialer seam.
- Files / Modules:
  - `hubruntime/runtime_test.go`
  - optional nearby test helpers if needed
- Write Set:
  - `hubruntime/runtime_test.go`
- Acceptance:
  - tests cover bootstrap request path beyond raw TCP-only assumption
- Test Points:
  - self register uses supplied dialer with non-TCP endpoint string
  - startup parse gate no longer rejects supported schemes
- Rollback:
  - remove new tests with runtime revert

##### `SRV-BOOT-3` - Validation
- Owner: main agent
- Worktree: `D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/MyFlowHub-Server`
- Plan Path: `D:/project/MyFlowHub3/worktrees/feat-bootstrap-endpoint-dialer/MyFlowHub-Server/plan.md`
- Goal: verify hubruntime/bootstrap regression coverage stays green.
- Files / Modules:
  - `hubruntime/*`
- Write Set:
  - none beyond test execution
- Acceptance:
  - targeted or full `go test ./...` passes in Server worktree
- Test Points:
  - `hubruntime`
  - bootstrap integration users if touched
- Rollback:
  - do not merge failing test state

#### Dependencies
- Depends on Core worktree exposing injected bootstrap dialer support.

#### Risks and Notes
- Keep `sendRegisterOnConn(...)` unchanged; this task is only about the helper path.

#### Parallelism Assessment
- No sub-agent use.
- Work is serialized because Server wiring depends on the Core helper contract.

#### Issue List
- None.
