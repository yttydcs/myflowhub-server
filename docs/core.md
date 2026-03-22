核心框架详解（对齐 Core v0.4.7）
==============================

范围
----
- 本文描述 `MyFlowHub-Server` 当前依赖的 `MyFlowHub-Core v0.4.7` 实际行为。
- 若文档与代码不一致，以代码为准。

接收链路（RX Pipeline）
----------------------
1) `IListener.Listen` 接入连接并加入 `ConnManager`，触发 `Process.OnListen`。  
2) `IReader.ReadLoop` 从 `conn.Pipe()` 读取字节流，解帧后调用 `conn.DispatchReceive`。  
3) `Server` 为连接注册回调，将收包事件交给 `Process.OnReceive`。  
4) 常见组合：`PreRoutingProcess`（base） + `DispatcherProcess`：  
   - Dispatcher 入队后由 worker 执行 `route`；  
   - `route` 顺序为：`selectHandler` -> `sourceMismatch` -> `preRoute` -> `handler.OnReceive`（按条件）。  

说明：`TCPReader` 只是历史命名，当前已是传输无关 reader（底层统一走 `Pipe`）。

PreRouting 与 HeaderRouter（当前实现）
------------------------------------
- Header-only 决策由 `HeaderRouter.Decide` 给出：
  - `SourceID==0 && SubProto!=2`：丢弃；
  - `SubProto==2`：放行（登录协议）；
  - `MajorCmd`：放行到 handler（逐跳可见）；
  - `TargetID==0`：广播给子节点（不回父）；
  - `TargetID!=local`：快速转发；
  - 其他：本地分发。
- `PreRoutingProcess` 执行具体转发动作：
  - 广播与快速转发都先克隆 header，再递减 `hop_limit`；
  - `hop_limit==0` 时按默认值补齐；`<=1` 视为耗尽并丢弃；
  - `routing.forward_remote=false` 时，跨节点帧直接丢弃；
  - 来自父连接且目标不可达时，不会再回父。

SourceID 一致性校验（Dispatcher.sourceMismatch）
-----------------------------------------------
- 校验发生在进入 handler 之前（在 `preRoute` 之前）。
- 放行规则：
  - `handler.AllowSourceMismatch()==true`：放行；
  - 连接未绑定 `nodeID`（`meta nodeID==0`）：拒绝；
  - `hdr.SourceID == conn.meta(nodeID)`：放行；
  - 连接角色是 `role=parent`：放行；
  - 子连接场景下，仅当 `ConnManager.GetByNode(hdr.SourceID)` 映射到当前连接时放行；
  - 其余情况拒绝。

发送链路（TX Pipeline）
----------------------
1) 入口：`server.Send(ctx, connID, hdr, payload)`。  
2) 安全默认：若 `hop_limit/trace_id` 未设置，自动补默认值。  
3) 调用 `Process.OnSend`（每次 `Send` 一次）。  
4) 默认通过 `SendDispatcher` 入队：  
   - 按连接映射到 per-connection writer，单连接串行写；  
   - 支持 shard+worker、入队超时、连接级缓冲；  
   - writer 最终写 `conn.Pipe()`（`WriteFrame`），不是直接写 `net.Conn`。  
5) 仅在 dispatcher 不可用时才回退到 `conn.SendWithHeader`。

上下文与 Server 注入
-------------------
- `Server.Start` 会将 `srv` 注入 context（`core.WithServerContext`）。
- handler/工具函数可通过 `core.ServerFromContext(ctx)` 获取 `IServer`。
- 手动调用 handler 时若未注入 server context，会失去统一发送/路由能力（通常回退直写连接）。

默认转发与兜底处理
------------------
- `DefaultForwardHandler` 仅在 Dispatcher 找不到该 `SubProto` 的 handler 时触发。
- PreRouting 发生在 handler 之前；若 PreRouting 已完成转发并返回 false，正常不会进入 handler。
- 兼容扩展：当 `preRoute=false` 且 `MajorCmd`，若 handler 声明 `AcceptCmd()==true`，仍可本地处理一次。

连接模型与演进状态
------------------
- `IConnection` 现为 `Pipe` 模型，不再以 `RawConn` 为核心抽象。
- Core v0.4.x 已引入 `ILink/ILinkManager` 与 `HeaderRouter` 收敛路由决策。
- Server 当前仍主要通过 `IConnectionManager` 工作，但与 Link 抽象保持兼容路径。

父链与自动重连
--------------
- `parent.enable=true` 且 `parent.addr` 非空时，Server 维护父连接并自动重连。
- 重连间隔由 `parent.reconnect_sec` 控制。
- 父连接会写入 `meta role=parent`，参与 source 校验与转发判定。

关键默认值/约束
---------------
- SourceID=0 的非登录协议默认丢弃。
- `MajorCmd` 默认逐跳可见（进入 handler）。
- `MajorMsg/OK/Err` 在 `target!=local` 时优先走 Core 快速转发。
- 发送时 `OnSend` 按 `Send` 调用次数执行，不按广播目标数重复执行审计。

快速参考：典型收发路径
---------------------
1) 收：`Listener -> Reader(Pipe) -> DispatchReceive -> Process.OnReceive -> Dispatcher.route -> preRoute -> handler`。  
2) 发：`server.Send -> OnSend -> SendDispatcher -> conn.Pipe() 写帧`。
