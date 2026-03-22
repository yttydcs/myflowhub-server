# 2026-03-22 Server Docs Governance

## Goal
- Rebuild `MyFlowHub-Server` docs into the governed taxonomy `requirements/specs/plan/change/lessons`.
- Move stable protocol and framework documents out of the top-level `docs/` root into `docs/specs/`.
- Replace `docs/plan_archive/` with `docs/plan/` and rebuild server-local indexes.

## Current Status
- Server docs now expose a governed root `docs/README.md` and category indexes.
- Stable specs are now under `docs/specs/`; historical workflow plans are under `docs/plan/`.
- This workflow uses the pure-migration strategy: no legacy redirect files are kept at the old paths.

## Workflow Info
- Repository: `D:\project\MyFlowHub3\repo\MyFlowHub-Server`
- Branch: `chore/server-docs-governance`
- Base branch: `main`
- Base commit: `ac05581`
- Worktree: `D:\project\MyFlowHub3\worktrees\server-docs-governance`
- Plan path: `D:\project\MyFlowHub3\worktrees\server-docs-governance\todo.md`
- Current stage: `4`

## Docs Governor Routing Check
- Category focus: docs tree bootstrap, spec routing, plan routing, index / entry maintenance
- Related requirements: none existing for server product behavior
- Related specs:
  - current protocol and framework spec docs under top-level `docs/`
  - generated `protocol_map.md`
- Requirements impact: `none`
- Specs impact: `clarify`
- Notes:
  - No protocol semantics change is planned.
  - Stable technical docs will move to governed spec locations and receive updated indexes.

## Task Checklist
- [x] SRV-DOC-001 Create the governed server docs tree and category indexes
- [x] SRV-DOC-002 Migrate legacy stable specs into `docs/specs/` with stable names
- [x] SRV-DOC-003 Migrate `docs/plan_archive/` to `docs/plan/` and rebuild the archive index
- [x] SRV-DOC-004 Rebuild `docs/change/README.md` for the governed taxonomy
- [x] SRV-DOC-005 Update in-repo links from old doc paths to the new spec/plan paths
- [x] SRV-DOC-006 Validate protected generated content and final navigation

## Executable Tasks
### SRV-DOC-001
- Owner: `MainAgent`
- Worktree: `D:\project\MyFlowHub3\worktrees\server-docs-governance`
- Write set:
  - `docs/README.md`
  - `docs/requirements/**`
  - `docs/specs/**`
  - `docs/plan/**`
  - `docs/change/**`
  - `docs/lessons/**`
- Goal: create the server docs taxonomy and root entry points.
- Acceptance:
  - `docs/` contains governed categories and category indexes.
  - `docs/README.md` explains the reading order and category ownership.
- Tests:
  - `Test-Path docs\\README.md`
  - `Test-Path` for all five categories and their `README.md`
- Rollback:
  - Revert the category creation and root index changes.

### SRV-DOC-002
- Owner: `MainAgent`
- Worktree: `D:\project\MyFlowHub3\worktrees\server-docs-governance`
- Write set:
  - `docs/specs/core.md`
  - `docs/specs/auth.md`
  - `docs/specs/varstore.md`
  - `docs/specs/topicbus.md`
  - `docs/specs/file.md`
  - `docs/specs/flow.md`
  - `docs/specs/exec.md`
  - `docs/specs/permission.md`
  - `docs/specs/protocol_map.md`
  - all former top-level spec files
- Goal: move stable technical documents into `docs/specs/` and normalize their names.
- Acceptance:
  - No stable spec remains only at top-level `docs/`.
  - The new names are stable and not date-prefixed.
  - `protocol_map.md` keeps its generated block markers.
- Tests:
  - `rg -n "BEGIN GENERATED|END GENERATED" docs/specs/protocol_map.md`
  - `Get-ChildItem docs -File` no longer lists the moved spec files
- Rollback:
  - Move the spec files back to their previous top-level locations.

### SRV-DOC-003
- Owner: `MainAgent`
- Worktree: `D:\project\MyFlowHub3\worktrees\server-docs-governance`
- Write set:
  - `docs/plan/**`
  - all former `docs/plan_archive/**`
- Goal: move historical workflow plans under the governed `plan` category.
- Acceptance:
  - `docs/plan/README.md` exists and indexes migrated plan files.
  - Server entry docs use `docs/plan/` as the plan archive location.
- Tests:
  - `rg -n "docs/plan_archive|plan_archive/" docs todo.md -S`
- Rollback:
  - Move the plan files back to `docs/plan_archive/`.

### SRV-DOC-004
- Owner: `MainAgent`
- Worktree: `D:\project\MyFlowHub3\worktrees\server-docs-governance`
- Write set:
  - `docs/change/README.md`
  - optionally link references inside change docs when canonical paths moved
- Goal: keep completed workflow results accessible after the migration.
- Acceptance:
  - `docs/change/README.md` routes users through the governed taxonomy.
  - Change index links remain valid after file moves.
- Tests:
  - link spot-checks from `docs/change/README.md`
  - `rg -n "docs/plan_archive|docs/[2-7]-|docs/core\\.md|docs/权限\\.md|docs/protocol_map\\.md" docs/change/README.md -S`
- Rollback:
  - Restore the previous change index and related references.

### SRV-DOC-005
- Owner: `MainAgent`
- Worktree: `D:\project\MyFlowHub3\worktrees\server-docs-governance`
- Write set:
  - server markdown files referencing old paths
- Goal: update in-repo canonical links to the new spec and plan paths.
- Acceptance:
  - High-value server docs refer to `docs/specs/*` and `docs/plan/*`.
  - No server README/index document points to removed old canonical paths.
- Tests:
  - targeted `rg` checks for old doc paths
  - spot-check key changed files
- Rollback:
  - Revert the path rewrites from this workflow.

### SRV-DOC-006
- Owner: `MainAgent`
- Worktree: `D:\project\MyFlowHub3\worktrees\server-docs-governance`
- Write set:
  - `todo.md`
  - review notes generated during 3.3
- Goal: capture verification evidence and residual risks for review and archive.
- Acceptance:
  - Protected generated content status is recorded.
  - Residual historical-path risk is documented.
- Tests:
  - `git diff --stat`
  - final targeted `rg` checks recorded in review notes
- Rollback:
  - N/A for planning notes; revert if incorrect.

## Dependencies and Order
- `SRV-DOC-001` before all server moves
- `SRV-DOC-002` before `SRV-DOC-005`
- `SRV-DOC-003` before `SRV-DOC-005`
- Root worktree can only finalize workspace navigation after server final paths are known

## Risks and Attention Points
- Pure migration means old links in historical docs may need broad updates.
- `docs/specs/protocol_map.md` contains a generated block and must not be damaged during edits.
- Numeric legacy filenames (`2-auth.md`, `6-flow.md`) are widely referenced and require careful rewrite coverage.

## Out of Scope
- No server code changes
- No protocol semantics changes
- No release automation changes
