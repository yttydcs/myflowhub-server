核心框架详解
============

接收链路（RX Pipeline）
----------------------
1) **Listener 接入**：`IListener.Listen` 接受新连接，包装为 `IConnection`，加入 `ConnManager`，触发 OnListen。
2) **Reader 解帧**：`IReader.ReadLoop` 从底层连接读取，使用 `HeaderCodec` 解出 `hdr/payload`，调用 `conn.DispatchReceive`，最终触发 `Process.OnReceive`。
3) **Process 组合**：常用组合为 `PreRoutingProcess`（base）+ `DispatcherProcess`：
   - Dispatcher 将事件放入内部 channel，worker 取出执行。
   - worker 内先调用 `preRoute`（如果 base 实现了 `PreRoute`，如 PreRoutingProcess），返回 false 则终止，不再分发到子协议。
   - 返回 true 时，按 SubProto 选择已注册的 handler 调用 `OnReceive`。
4) **PreRouting 核心规则（现实现）**：
   - SourceID=0 且 SubProto!=2：丢弃（未登录非登录协议）。
   - SubProto=2：直接放行 dispatcher。
   - MajorCmd：**不做自动转发/广播**，直接放行 dispatcher（逐跳可见；是否继续转发由子协议 handler 决定）。
   - MajorMsg / MajorOKResp / MajorErrResp（数据面/响应）：
     - target==0：广播给子节点（不回父），返回 false。不要将 0 作为“上送父节点”。
     - target!=local：按节点索引/父链转发，返回 false（子协议 handler 不会执行）。
     - target==local：返回 true，交给 Dispatcher 调用子协议。
   - hop_limit：仅在发生“转发/广播”（数据面）时递减；耗尽则丢弃（防环/防风暴）。

SourceID 一致性校验（Dispatcher.sourceMismatch）
-----------------------------------------------
- Dispatcher 在进入子协议 handler 前，会进行一次 `SourceID` 一致性校验（可被 handler 覆盖）。
- 目的：避免连接伪造任意 `SourceID`（在树形网络里用于权限、审计、路由一致性等）。
- 例外：登录/注册等阶段可能尚未绑定 nodeID，相关 handler 可通过 `AllowSourceMismatch()=true` 放行。

推荐校验规则（拟全局生效）
------------------------
> 该规则用于支持“端到端 SourceID 穿越多跳”（例如文件传输 DATA/ACK 直达目标、控制帧上送 LCA 判权）。

- 若 `handler.AllowSourceMismatch()==true`：跳过校验（保持现有行为）。
- 否则要求连接已登录：`conn.meta(nodeID)!=0`。
- 若 `hdr.SourceID == conn.meta(nodeID)`：放行（最常见的逐跳发送）。
- 否则根据连接角色：
  - `role=parent`：放行（子节点无条件信任父节点；父节点可代表子树下发控制/缓存）。
  - `role=child`：仅当 `ConnManager.GetByNode(hdr.SourceID)` 映射到该连接时放行（`SourceID` 为该子连接背后的后代节点）。
- 依赖：登录协议需要把“后代 nodeID → 该 child 连接”的索引逐级同步到祖先节点（例如 `up_login`）。

发送链路（TX Pipeline）
----------------------
1) 入口：`server.Send(ctx, connID, hdr, payload)`。
2) 调用 `Process.OnSend` 钩子（仅一次），若返回错误则终止。
3) 若启用 `SendDispatcher`：
   - 事件放入 dispatcher channel，按连接映射到 per-connection writer，串行写出。
   - 支持 shard/worker 并行、入队超时、连接级缓冲。
4) 未启用 SendDispatcher 时，直接调用 `SendWithHeader` 发送（同步，阻塞至写完）。

上下文与 Server 注入
-------------------
- `core.WithServerContext(ctx, srv)` 将 server 放入 context；handler 可用 `core.ServerFromContext(ctx)` 获取 srv/ConnManager/NodeID。
- Server.Start 包装了全局 ctx，Reader/Process 调用时 ctx 中有 server。手动调用 handler 时需自行注入，否则会 fallback 到直接写连接。

路由与默认转发
--------------
- DefaultForwardHandler 仅在 Dispatcher 找不到子协议 handler 时触发，用于按配置将未知子协议转发到父/指定节点。
- PreRouting 的跨节点转发发生在 Dispatcher 之前，返回 false 短路子协议处理。

Major 与子协议处理
------------------
- 框架对 Major 做统一约束（用于“控制面/数据面”分离与减少 Core 协议特例）：
  - `MajorCmd`：控制面，**必须进入 handler（逐跳可见）**；Core 不做自动转发/广播。
  - `MajorMsg` / `MajorOKResp` / `MajorErrResp`：数据面/响应，优先走 Core 快速转发；仅 `target==local` 才进入 handler。
- 协议仍可在 payload 内定义更细的帧类型（例如 `file` 的 CTRL/DATA/ACK），但“是否逐跳可见”应优先通过 Major 表达。
- `AcceptCmd()`：旧扩展点（在 Cmd 被 Core 转发的时代用于“转发后仍本地处理一次”）；在 `MajorCmd` 统一逐跳规则下通常不再需要（保留兼容）。

注册与分发
----------
- 通过 `DispatcherProcess.RegisterHandler(ISubProcess)` 按 SubProto 注册；每个 SubProto 仅可注册一个 handler。
- Default handler 可注册一次，用于兜底未知子协议。
- Dispatcher 使用 worker+channel 解耦收包与处理，避免阻塞 Reader。

连接管理
--------
- `ConnManager` 维护连接表，支持按 connID/nodeID/deviceID 索引；提供 Range/Broadcast/CloseAll 等。
- 登录流程会在连接元数据中写入 `nodeID/deviceID`，用于路由与索引。

父链与自动重连
--------------
- Server 根据配置 `parent.enable/parent.addr` 维护父连接，断线自动重连（`parent.reconnect_sec`）。父连接的 meta 标记为 `role=parent`。

关键默认值/约束
---------------
- SourceID=0 的非登录协议默认被 PreRouting 丢弃。
- `MajorCmd` 帧默认会进入子协议 handler（逐跳）。
- 对 `MajorMsg/OK/Err`：target!=local 的帧默认不会进入子协议 handler（走 Core 快速转发）。
- 发送时 OnSend 仅执行一次（即便广播）。

扩展点提示
----------
- 若需要“数据面（MajorMsg/OK/Err）在 target!=local 时也进 handler 再转发”，需调整 PreRouting/Dispatcher（否则默认走 Core 快速转发）。
- 控制面（MajorCmd）已默认逐跳可见；若某协议希望端到端直达，请不要使用 MajorCmd（或在 handler 内做显式转发并明确 hop_limit 策略）。

快速参考：典型收发路径
---------------------
1) 收：Listener→Reader→Process.OnReceive→Dispatcher.worker→PreRoute（可能转发/丢弃）→子协议 handler。
2) 发：server.Send→Process.OnSend→SendDispatcher（可选）→net.Conn.Write / SendWithHeader。***
