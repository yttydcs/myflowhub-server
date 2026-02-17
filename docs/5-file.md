file 协议（SubProto=5）规范（草案）
=============================

范围
----
- 仅描述新增 `file` 子协议的设计目标与消息格式（用于节点间文件传输）。
- 依赖核心框架的路由能力（`TargetID` 自动路由）与连接角色（`role=parent/child`）。
- **注意**：Major 在框架层代表“是否需要逐跳进入 handler / 或可由 Core 快速转发”的路由语义；本协议不将 Major 作为 CTRL/DATA/ACK 判定依据（CTRL/DATA/ACK 仍由 payload[0] 决定）。

总览
----
- 传输模型：提供方 **A** → 接收方 **B**，Hub 仅中继转发数据，不做文件落盘缓存。
- 交互模式：
  - **pull**：B 读取 A 的文件（控制帧 action=`read`，权限 `file.read`）。
  - **offer**：A 向 B 提供文件（控制帧 action=`write`，权限 `file.write`）。
- 存储：接收方按 `dir/name` 落盘到配置目录（默认 `./file`），**默认覆盖**同名文件。
- 断点续传：接收方写入临时文件（建议 `*.part`），中断后可继续；已完成文件不清理；未完成且“最后写入时间超过阈值”（默认 1h，可配置）自动删除临时文件。
- 完整性：仅校验完整性（建议 `size + sha256`），不做签名/加密（后续可扩展）。

权限
----
- 权限节点格式：`协议.action`
- 本协议固定：
  - `file.read`：读取类操作（pull、list/stat 等）
  - `file.write`：写入类操作（offer、覆盖/删除等）

帧分类（由 payload[0] 决定）
--------------------------
本协议在 payload 的第 1 字节定义 `kind`：
- `0x01`：CTRL（控制帧，后续为 JSON）
- `0x02`：DATA（数据帧，二进制）
- `0x03`：ACK（确认帧，二进制）

**建议 Major**（统一框架规则；不作为 CTRL/DATA/ACK 判定依据）：
- CTRL 请求（`read/write`）：`MajorCmd`（逐跳进入 handler，用于判权/转交）
- CTRL 响应（`read_resp/write_resp`）：`MajorOKResp`（按 `TargetID` 由 Core 快速转发）
- DATA/ACK：`MajorMsg`（端到端数据面帧，按 `TargetID` 由 Core 快速转发）
- 失败响应仍使用 `MajorOKResp`，错误通过 payload 的 `code/msg` 表达

HeaderTcp 与路由约定
--------------------
- SubProto 固定为 `5`（预留给 `file`）。
- `TargetID=0` 在核心中表示“向子节点广播（不回父）”，**不要**用 0 表示“上送父节点”。
- **CTRL 请求逐级判权/转交**：`read/write` 控制请求帧使用 `MajorCmd`，Core 不做自动转发而是进入 handler 由协议逻辑决定向上/向下转交，避免子节点通过“直接填最终目标 TargetID”绕过判权链路。
- **CTRL 响应端到端返回**：`read_resp/write_resp` 响应帧使用 `MajorOKResp`，按 `TargetID` 由 Core 快速转发返回请求方（中间节点无需进入 file handler）。
- SourceID/TargetID 建议（端到端）：
  - CTRL：
    - 请求阶段：`SourceID=请求方`，`TargetID=请求方的直接父Hub`（使其进入父Hub 的 handler 做判权与上送）。
    - 授权后转交目标：`SourceID=请求方`，`TargetID=目标节点`（使请求到达目标节点执行）。
    - 响应：`SourceID=执行方（通常为 A 或 B）`，`TargetID=请求方`（端到端返回）。
  - DATA：`SourceID=A`，`TargetID=B`（A→B 端到端传输）
  - ACK：`SourceID=B`，`TargetID=A`（B→A 端到端确认）

来源校验（核心约定）
------------------
为支持端到端 `SourceID` 穿越多跳，核心层需采用“父免检 + 子树后代放行”的一致性校验（全局生效）：
- 来自父连接（`role=parent`）：放行（子节点无条件信任父节点）。
- 来自子连接（`role=child`）：仅当 `SourceID` 为该连接自身或其后代（可由路由表/索引证明）才放行。
- 该机制依赖登录协议的路由索引（如 `up_login`）将后代 nodeID 逐级上报到祖先节点的路由表中。

控制帧（CTRL）格式
-----------------
- 载荷编码：`payload = [0x01] + JSON(UTF-8)`
- JSON envelope：`{"action":"read|write","data":{...}}`
- 响应 action：`read_resp` / `write_resp`
- `data.code`：状态码（1=成功，其他为失败）

### action=read（pull/list 等，权限：file.read）

#### op=pull（B 从 A 读取文件）
- 请求方：B（接收方）
- 目标（data.target）：A（提供方）
- 请求 data（示例字段）：
  - `op`：固定 `"pull"`
  - `target`：提供方 nodeID（A）
  - `dir`：相对目录（可为空），用于定位与落盘
  - `name`：文件名（禁止包含 `/`）
  - `overwrite`：可选，默认 `true`
  - `resume_from`：可选，B 期望从该偏移继续（默认 0）
  - `want_hash`：可选，是否需要 `sha256`（默认 true）

- 响应 read_resp.data（示例字段）：
  - `code` / `msg`
  - `op`：`"pull"`
  - `session_id`：UUID 字符串（由提供方 A 生成）
  - `provider`：A
  - `consumer`：B
  - `dir` / `name`
  - `size`：文件大小（bytes）
  - `sha256`：可选（hex）
  - `start_from`：本次实际起始偏移（A 可拒绝不合法的 resume_from 并返回 0）
  - `chunk_bytes`：可选，建议数据帧单块大小

#### op=list（列出目录文件）
> 该能力用于“展示 file 目录下可见文件列表”；仍归属于 `file.read` 权限。

- 请求 data（示例字段）：
  - `op`：固定 `"list"`
  - `dir`：相对目录（可为空）
  - `recursive`：可选，默认 false
- 响应 read_resp.data（示例字段）：
  - `code` / `msg`
  - `op`：`"list"`
  - `dir`
  - `dirs`：目录名数组（可选，兼容扩展）
  - `files`：文件名数组（建议仅返回 `name`）

#### op=read_text（文本预览）
> 用于 UI 预览文本文件内容；仍归属于 `file.read` 权限；不走 DATA/ACK 会话。

- 请求 data（示例字段）：
  - `op`：固定 `"read_text"`
  - `target`：提供方 nodeID
  - `dir` / `name`
  - `max_bytes`：可选，最大读取字节数（默认 64KB，建议上限 256KB）
- 响应 read_resp.data（示例字段）：
  - `code` / `msg`
  - `op`：`"read_text"`
  - `provider` / `consumer`
  - `dir` / `name` / `size`
  - `text`：UTF-8 文本内容
  - `truncated`：是否被截断

### action=write（offer 等，权限：file.write）

#### op=offer（A 向 B 提供文件）
- 请求方：A（提供方）
- 目标（data.target）：B（接收方）
- 关键点：`session_id` 由提供方 A 生成并在 offer 中携带（B 在响应中回传该 session_id）。

- 请求 data（示例字段）：
  - `op`：固定 `"offer"`
  - `target`：接收方 nodeID（B）
  - `session_id`：UUID 字符串（A 生成）
  - `dir`：相对目录（可为空），用于定位与落盘
  - `name`：文件名（禁止包含 `/`）
  - `size`：文件大小（bytes）
  - `sha256`：可选（hex，建议提供以便完整性校验与断点续传判定）
  - `overwrite`：可选，默认 `true`

- 响应 write_resp.data（示例字段）：
  - `code` / `msg`
  - `op`：`"offer"`
  - `session_id`：UUID 字符串（与请求一致）
  - `provider` / `consumer`：可选（回显，便于无状态实现）
  - `dir` / `name` / `size` / `sha256`：可选（回显，便于无状态实现）
  - `accept`：是否接受（默认 true）
  - `resume_from`：接收方 B 允许从该偏移开始接收（默认 0；若存在匹配的 `.part` 可返回其大小）

二进制传输帧（DATA/ACK）
----------------------
二进制帧只用于承载文件内容与接收确认；所有复杂控制（取消/错误/重协商）放在 CTRL 中完成。

### 二进制小头（V1）
payload 结构：
- DATA：`[0x02] + FileBinHeaderV1 + <bytes...>`
- ACK ：`[0x03] + FileBinHeaderV1`

`FileBinHeaderV1`（网络序，大端）：
- `ver`：`uint8`（固定 1）
- `flags`：`uint8`（bit0=FIN：发送方已到达 EOF）
- `reserved`：`uint16`（固定 0）
- `session_id`：`[16]byte`（UUID 原始 16 字节）
- `offset`：`uint64`
  - DATA：本块数据起始偏移
  - ACK：接收方已持久化的“下一期望偏移”（即 `[0, offset)` 已完成）

### 接收侧校验与防攻击
会话必须绑定来源与方向，任何不符合者直接忽略（不回包）：
- DATA（A→B）：必须满足 `hdr.SourceID==A && hdr.TargetID==B`，且 `session_id` 存在且属于该方向。
- ACK（B→A）：必须满足 `hdr.SourceID==B && hdr.TargetID==A`，且 `session_id` 存在且属于该方向。
- offset 必须单调不回退（允许重复 ACK）；DATA 的 offset 必须等于当前期望写入位置或在允许的窗口内（窗口策略实现可选）。

落盘规则（接收方）
---------------
- 配置：
  - `file.base_dir`：默认 `./file`
  - `file.incomplete_ttl_sec`：默认 `3600`（仅清理未完成 `.part`）
- 文件路径：
  - 最终文件：`<base_dir>/<dir>/<name>`
  - 临时文件：`<base_dir>/<dir>/<name>.part`
  - 元数据：建议 `<name>.part.meta`（记录 size/hash/session/最后更新时间等）
- 完成：
  - DATA 接收完毕后校验 `size`（以及可选 `sha256`）
  - 校验通过：原子替换/重命名为最终文件（默认覆盖）
  - 校验失败：保留 `.part`（便于重试）或转为失败态（实现可选）
- 清理：
  - `.part` 若在 `file.incomplete_ttl_sec` 内无写入更新，自动删除（不影响已完成文件）

名称与目录安全
-------------
- `name`：
  - 禁止包含 `/`、`\\`、NUL
  - 禁止为 `.` 或 `..`
  - 为空直接失败（`code=400`）
- `dir`：
  - 可为空；非空时必须为相对路径
  - 禁止出现 `..` 段（例如 `../x`、`a/../b`）
  - 禁止绝对路径/盘符（例如 `/root`、`C:\`、`C:`）
  - 不符合直接失败（`code=400`）

限制与配置（建议）
---------------
- `file.max_size_bytes`：单文件最大字节数（超出返回 `code=413`）
- `file.max_concurrent`：最大并发会话数（超出返回 `code=429`）
- `file.chunk_bytes`：建议单 DATA 帧数据大小（影响内存与吞吐）

典型流程
--------
### pull（B read/pull A）
1) B → 父Hub：CTRL `read(op=pull,target=A,dir,name,...)`（判权点沿父链上送至能覆盖 A/B 的 Hub）
2) 授权 Hub → A：转交请求（`TargetID=A`）
3) A → B：`read_resp` 返回 `session_id/size/hash/start_from`
4) A → B：DATA（端到端）逐块发送；B 持久化写入 `.part`
5) B → A：ACK（端到端）可按阈值回传进度
6) B 校验完整性，落盘为最终文件；会话结束

### offer（A write/offer B）
1) A 生成 `session_id`，A → 父Hub：CTRL `write(op=offer,target=B,session_id,dir,name,size,...)`
2) 授权 Hub → B：转交 offer
3) B → A：`write_resp(accept,resume_from)`；若拒绝则结束
4) A 从 `resume_from` 起发送 DATA；B 回 ACK；完成后校验落盘

错误码（建议）
------------
- `1`：成功
- `400`：参数非法（name/dir/session 等）
- `403`：权限不足
- `404`：文件不存在 / 目标不可达
- `409`：冲突（overwrite=false 且已存在）
- `413`：文件过大
- `429`：并发过多
- `500`：内部错误

集成提示（实现侧）
----------------
- Dispatcher 注册：`dispatcher.RegisterHandler(file.NewHandlerWithConfig(cfg, log))`
- 子协议编号：`SubProto=5`（建议）。
- 若启用端到端 SourceID，多跳可达性依赖核心的“父免检 + 子树后代放行”一致性校验与路由索引同步。***
