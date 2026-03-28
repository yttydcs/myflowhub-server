# 2026-03-28 Server 远程 Authority 管理规范对齐

## 变更背景 / 目标

- 远程 authority 审批与 permit 管理的真实 runtime 能力已经下沉到 `myflowhub-subproto/auth`。
- `MyFlowHub-Server` 这边本轮主要任务是更新稳定 spec，废弃“approval / permit 仍建议从 authority 所在节点操作”的旧口径。
- 同时需要明确一个现实约束：当前 repo-local `go.mod` 仍停留在 `github.com/yttydcs/myflowhub-subproto/auth v0.1.4`，依赖版本升级需要等新 tag 发布后再补。

## 具体变更内容

- 更新 `docs/specs/auth.md`
  - semi-central authority 段落新增 remote authority admin 语义
  - 受控准入动作段落明确：
    - approval / permit 管理既可 authority 本机执行，也可由具备权限的远程节点发起
    - authority 权限判断始终基于真实 `SourceID`
  - 错误处理段落新增 remote authority admin 的 `authority unavailable`
  - 删除“approval / permit 仍建议从 authority 所在节点操作”的旧说明

## Requirements impact

- `none`

## Specs impact

- `updated`

## Lessons impact

- `none`

## Related requirements

- `none`

## Related specs

- `docs/specs/auth.md`
- `D:\project\MyFlowHub3\worktrees\feat-win-remote-authority-admin\docs\specs\authority-admin-console.md`

## Related lessons

- `D:\project\MyFlowHub3\docs\lessons\authority-local-admin-actions.md`

## 对应 plan.md 任务映射

- `AUTH-REMOTE-REQ-1`
- `AUTH-REMOTE-SERVER-1`

## 经验 / 教训摘要

- Server spec 不应继续把历史 workaround 记录成长期真相。
- 但 stable spec 对齐并不等于 repo-local 依赖已经升级；release chain 仍要单独完成。

## 可复用排查线索

- 症状
  - Server docs 写着 remote authority admin 可用，但 repo-local 运行仍表现为旧行为
- 触发条件
  - `go.mod` 仍依赖旧版 `myflowhub-subproto/auth`
- 关键词
  - `v0.1.4`
  - `auth/v0.1.5`
  - `authority-local`
- 快速检查
  - 检查 `go.mod` 中 `github.com/yttydcs/myflowhub-subproto/auth` 的实际版本
  - 对照 subproto change / tag 发布状态

## 关键设计决策与权衡

- 先对齐 spec，再等待依赖发布
  - 优点：稳定文档先回到正确方向，避免继续固化旧限制
  - 代价：repo-local build 仍需后续 tag 发布后才能真正消费该能力

## 测试与验证方式 / 结果

- `MyFlowHub-Server`
  - 本轮未升级 `go.mod`，因此没有执行“消费新 auth 版本”的集成验证
  - 当前只完成 spec 对齐；依赖升级仍是后续发布链路任务

## 潜在影响与回滚方案

- 潜在影响
  - 文档先于 repo-local 依赖升级完成对齐
- 回滚方案
  - 回退 `docs/specs/auth.md`
  - 回退本归档

## 子Agent执行轨迹

- 未使用子Agent
