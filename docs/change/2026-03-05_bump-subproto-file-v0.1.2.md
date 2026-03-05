# 2026-03-05 - Server：升级 `myflowhub-subproto/file` 至 `v0.1.2`（新增 mkdir）

## 变更背景 / 目标
- 背景：
  - `myflowhub-subproto/file v0.1.2` 新增 `write(op=mkdir)`，用于目录创建。
  - Server 作为装配层，需要升级依赖并同步协议文档，避免实现与文档漂移。
- 目标：
  1) Server 依赖升级到 `v0.1.2`；
  2) `docs/5-file.md` 补充 `op=mkdir` 说明；
  3) 完成回归验证。

## 具体变更内容

### 修改
- `go.mod` / `go.sum`
  - `github.com/yttydcs/myflowhub-subproto/file`：
    - `v0.1.1` -> `v0.1.2`

- `docs/5-file.md`
  - `action=write` 章节扩展为 `offer/mkdir`；
  - 新增 `op=mkdir` 的请求/响应字段说明、语义约束与典型流程。

### 新增
- `todo.md`（本 workflow 任务拆分文档）

### 删除
- 无。

## todo.md 任务映射
- `SRV-FILE-1` 升级依赖版本：✅
- `SRV-FILE-2` 同步协议文档：✅
- `SRV-FILE-3` 回归验证 + 归档：✅

## 关键设计决策与权衡
- Server 保持“装配层”职责：
  - 不在 Server 内重写 file 子协议逻辑；
  - 仅通过 semver 升级依赖并同步文档。
- 文档先行对齐：
  - 避免客户端按旧文档实现导致行为偏差。

## 测试与验证
```powershell
cd d:\project\MyFlowHub3\worktrees\MyFlowHub-Server-file-v012
$env:GOWORK='off'
go list -m github.com/yttydcs/myflowhub-subproto/file
go test ./... -count=1
```
- 结果：
  - `go list -m .../file` 输出 `v0.1.2`
  - `go test ./...` 通过。

## 潜在影响与回滚方案

### 潜在影响
- 节点间 file write 行为新增 `mkdir` 分支；旧客户端若不调用该 op 不受影响。

### 回滚
- 将依赖回退至 `v0.1.1`，并回滚文档对应段落。

