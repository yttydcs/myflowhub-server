exec 协议（SubProto=7）规范（草案）
=============================

范围
----
- `exec` 子协议用于在网络中调用“特殊能力”（将来可用于插件系统，例如调用第三方服务商 Web API）。
- **注意**：`exec.call` 不是把 `flow` 交给别的节点执行；`flow` 的执行者永远是接收 `flow.set` 的节点。
- 权限：仅校验 `exec.call`（动作/方法本身不拆分权限）。

总览
----
- 控制帧编码：UTF-8 JSON，envelope 固定为 `{"action":"...","data":{...}}`
- 典型动作：
  - `call`：请求目标节点执行一个已注册的 `namespace::method`
  - `call_resp`：执行结果响应

权限
----
- 权限节点格式：`协议.action`
- 第一版固定：
  - `exec.call`：允许“使用网络中的特殊能力”

HeaderTcp 与路由约定
--------------------
- SubProto 固定为 `7`（预留给 `exec`）。
- `TargetID` 由核心路由自动转发到目标节点。
- 本协议依赖“逐级上送直到可向下转发”的一致性路由语义（见下文）。
- Major 约定（统一框架规则）：
  - 请求帧（`call`）：`MajorCmd`（逐跳可见，需要进入 handler 参与裁决/执行/转发）。
  - 响应帧（`call_resp`）：`MajorOKResp`（按 `TargetID` 由 Core 快速转发；中间节点不需要进 handler 转发）。
  - 失败响应仍使用 `MajorOKResp`，错误通过 payload 的 `code/msg` 表达。

逐级上送与裁决（downstream 判定）
------------------------------
前提：
- 子节点无条件信任父节点；父节点对子节点有绝对控制权。

对 `exec.call` 的规则（你已确认）：
- `exec.call` 本质是“调用网络能力”，请求方（通常为 `flow` 执行者）应先发送给自己的**直接父节点**，再逐级向上。
- 任一节点收到 `call/assist_call` 后：
  1) 判断目标 `target_node` 是否在自己的子树内：
     - 若到 `target_node` 的下一跳连接属于 **downstream（子连接）**，则视为目标在本子树内，当前节点具备对该子树的控制权。
     - 否则（下一跳为 upstream 或不可达），继续向上转发给父节点。
  2) 当“可向下转发”成立时，当前节点即为“裁决/转发点”：
     - 若目标位于当前节点的 **downstream**，则可以直接向下转发（父控子语义）。
     - 权限校验：校验主体为 `executor_node` 是否具备 `exec.call`。
       - 通过：**直接转发**请求到目标（转发即同意，不需要先回 allow）。
       - 拒绝：直接回 `call_resp(code=403)` 给 `executor_node`。

特殊情况（无需权限判断）
----------------------
- 若目标节点位于“请求节点的 downstream”（即 `flow` 执行者对目标是父→子方向可控），则可由 `flow` 执行者直接向目标发送 `exec.call`，不经过逐级授权，也不做 `exec.call` 权限校验。
  - 该规则是“父控子”语义的直接体现：子节点无条件信任父节点。

> 注：上述“免检”仅描述 `exec.call` 这一能力的权限模型；目标节点仍可对 `namespace::method` 做入参校验/限流/失败返回。

控制帧格式
----------
- 载荷编码：JSON(UTF-8)
- JSON envelope：`{"action":"call","data":{...}}` / `{"action":"call_resp","data":{...}}`
- 统一字段建议：
  - `req_id`：UUID（用于匹配一次调用；同时可用于幂等/去重）
  - `executor_node`：发起这次调用的 `flow` 执行者节点（权限主体）

### action=call（执行特殊方法，权限：exec.call）

请求 `data`：
- `req_id`：UUID（必填）
- `executor_node`：uint32（必填）
- `target_node`：uint32（必填）
- `method`：string（必填，形如 `namespace::method`）
- `args`：object（可选）
- `timeout_ms`：int（可选，默认 3000；建议由 `flow` 节点的 timeout 传入）

响应 `action=call_resp`，`data`：
- `req_id`：回显
- `code`：`1` 成功；`400/403/404/408/500` 等失败
- `msg`：可选错误说明
- `executor_node`：回显（可选）
- `target_node`：回显（可选）
- `method`：回显（可选）
- `result`：object（可选，成功时返回）

响应投递建议：
- `TargetID` 设置为 `executor_node`，依赖核心路由将 `call_resp` 直接回到执行者。
- `SourceID` 建议为“执行方法的节点”（最终目标节点）。

方法注册（插件系统接口建议）
--------------------------
- 每个节点维护一个方法注册表：`method(string) -> handler(ctx, args) -> result`
- `method` 使用 `namespace::method` 作为唯一标识。
- 第一版不规定注册机制（静态内置或动态插件均可），但约束：
  - 未注册方法返回 `code=404`（method not found）
  - 入参不合法返回 `code=400`
  - 超时返回 `code=408`

错误码建议
----------
- `1`：ok
- `400`：invalid request / invalid args
- `403`：permission denied（`exec.call`）
- `404`：target or method not found / not reachable
- `408`：timeout
- `429`：too many requests（可选，限流）
- `500`：internal error

