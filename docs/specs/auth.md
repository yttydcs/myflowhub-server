auth 协议（SubProto=2，基于 P256 公钥签名）
==========================================

范围与格式
----------
- 仅描述当前 `subproto/auth` 的实现；已移除 login_server 旧 credential 流程。
- 消息统一格式：`{"action":"<name>","data":{...}}`，响应 action = `<req>_resp`，状态码在 data.code。
- 签名算法：ES256（P256 + SHA256），公钥/私钥 DER 以 base64 编码。
- SubProto 固定 2；未认证连接 `SourceID=0` 仅放行子协议 2，其余丢弃。

头部与路由规则
--------------
- TargetID=0 仅表示“向子节点广播，不回父”，**不**表示上送；上送父/权威需写明 NodeID。
- Major：命令/状态类用 `MajorCmd`；响应可用 `MajorOKResp`。
- 权威选择：
  - 默认 legacy 模式：
    - 若显式配置 `authority.node_id`，则该 `node_id` 是唯一 authority；只有当前连接表里存在该节点连接时才可上送，**不得**回退为父节点或本地 authority。
    - 若未配置 `authority.node_id` 且已配置父链，则直接父节点是唯一 authority；父链不可达时，**不得**回退为本地 authority。
    - 仅在既未配置 `authority.node_id` 也未配置父链时，本地才是 authority。
  - `auth.authority_mode=semi-central`：
    - root 通过 `authority_policy_sync` 下发运行时 authority lease，声明当前 `effective_authority_id`。
    - 非 root 节点**不会**把直接父节点视为最终 authority；当 lease 有效时：
      - `assist_register / assist_login / assist_query_credential` 使用 `SourceID=发起 edge hub`、`TargetID=effective_authority_id` 上送，途中节点只按 `TargetID` 转发，不建立本地 pending/binding。
      - `list_pending_registers / approve_register / reject_register / list_register_permits / issue_register_permit / revoke_register_permit` 保持 `SourceID=真实操作者`、`TargetID=effective_authority_id` 上送，authority 侧只在当前入站连接确实拥有该 `SourceID` 路由归属时才接受执行。
    - 若 lease 尚未收到或已过期，但父链仍在线，则 admission 相关 assist 请求允许按父链逐级上送，以保留 parent bootstrap / 初始 register 时序；一旦父链断开，新准入冻结。
    - 半中心退化期只允许“本地已知身份”登录；需要上游 authority 的 login / register / assist_query_credential 都返回 `code=4500,msg=\"authority unavailable\"`。

密钥与持久化
------------
- 节点密钥：启动时从 `config/node_keys.json` 读取/生成（字段 `privkey`、`pubkey`，base64 DER），并写入配置键 `auth.node_privkey`、`auth.node_pubkey`。
- 信任/白名单：`config/trusted_nodes.json`
  - `bindings`: device_id -> `{node_id,pubkey,role,perms}`，注入 whitelist。
  - `meta.pending_registers`: 待审批注册请求列表。
  - `meta.approved_registers`: 已批准但尚未完成最终 register 的预留身份。
- `meta.register_permits`: 当前活动的一次性角色 permit 列表；成功消费、显式撤销或过期后移除。
  - `meta.first_register_bootstrap`: 首个注册 bootstrap 的消费状态（`consumed_epoch` 等）。
- `auth.disable_persist=true` 时，不读写 `trusted_nodes.json`，因此 pending / approved / permit 也不会落盘。

动作与数据字段
-------------
- register / assist_register  
  - req: `{"device_id","requested_role,omitempty","join_permit,omitempty","pubkey,omitempty","node_pub,omitempty","display_name,omitempty","ts,omitempty","nonce,omitempty"}`。  
  - `requested_role` 仅作为“申请的目标角色”提示；是否生效取决于 approve / permit。  
  - `join_permit` 是 authority 保存的一次性 opaque token，绑定 `device_id + role + expiry`。  
  - resp: `{"code","msg,omitempty","device_id","node_id,omitempty","hub_id,omitempty","role,omitempty","perms,omitempty","pubkey,omitempty","node_pub,omitempty","display_name,omitempty","status,omitempty","request_id,omitempty","reason,omitempty","ts,omitempty","nonce,omitempty"}`  
  - `status` 语义：
    - `approved`: 注册已完成（`code=1`，必须带 `node_id`）
    - `pending`: 已进入待审批（`code=202`，不带正式 `node_id`）
    - `rejected`: 被拒绝或 permit 无效（常见 `code=4001`）
- login / assist_login  
  - req: `{"device_id","node_id,omitempty","display_name,omitempty","ts","nonce","sig","alg"}`，需 ES256 签名；`display_name` 为可选提示字段，不进入签名摘要。  
  - resp: 同 register_resp，失败常见 `code=4001`；authority 不可达时为 `code=4500`。
- assist_query_credential / _resp  
  - req: `{"device_id","node_id,omitempty"}`  
  - resp: `{"code","msg,omitempty","device_id","node_id","role,omitempty","perms,omitempty","pubkey,omitempty","node_pub,omitempty"}`
- up_login / up_login_resp  
  - req: `upLoginData` 字段：`node_id,device_id,hub_id,pubkey,ts,nonce,device_ts,device_nonce,device_sig,device_alg,sender_id,sender_ts,sender_nonce,sender_sig,sender_alg,sender_pub,alg`。  
  - 校验：
    - 设备签名有效；
    - sender 签名优先用 trusted sender 公钥验签；
    - 若验签失败且请求携带 `sender_pub`，则仅在满足约束 `sender_id == hdr.SourceID == conn.meta(nodeID)` 下允许用 `sender_pub` 二次验签；
      - 二次验签成功：自愈 trusted/binding 公钥并继续写入路由索引；
      - 否则：拒绝处理；
    - 路由公钥冲突则拒绝。
- revoke  
  - req: `{"device_id","node_id,omitempty"}`；需权限 `auth.revoke`。  
  - resp: 仅删除命中时回 `{"code":1,"device_id","node_id"}`；否则静默。向上下行广播同一动作（除来源）。
- offline / assist_offline  
  - req: `{"device_id","node_id,omitempty","reason,omitempty"}`；无响应。移除绑定与路由索引，向父转发 assist_offline。
- 权限与角色  
  - get_perms / _resp: `{"node_id"}` → `{"code","msg,omitempty","node_id","role","perms"}`。  
  - list_roles / _resp: `{"offset,omitempty","limit,omitempty","role,omitempty","node_ids,omitempty"}` → `{"code","msg,omitempty","total","roles":[{node_id,role,perms}]}`。  
  - perms_invalidate: `{"node_ids,omitempty","reason,omitempty","refresh,omitempty"}`；清缓存，可选触发上行刷新；向子节点广播（target=0）。  
  - perms_snapshot: 下发/广播权限快照（结构见 core/permission.Snapshot）。
- authority_policy_sync  
  - data: `{"mode","effective_authority_id","epoch","ttl_sec"}`  
  - 仅 root 负责起源；非 root 仅接受来自父连接的同步并继续下发给子节点。  
  - `epoch` 更小的同步会被忽略；相同 `epoch` 仅用于刷新 TTL，不写入持久化。
- 受控准入动作（都要求调用方已登录并显式命中 `auth.*` 权限）  
  - 这些动作既可在 authority 本机执行，也可在 remote authority 场景由任意已登录且具备对应权限的节点发起；authority 侧权限判断始终基于请求头中的真实 `SourceID`。  
  - list_pending_registers / _resp  
    - req: `{"offset,omitempty","limit,omitempty","device_id,omitempty"}`  
    - resp: `{"code","msg,omitempty","total","items":[{"request_id","device_id","requested_role,omitempty","display_name,omitempty","created_at","expires_at"}]}`  
    - 权限：`auth.pending.list`
  - list_register_permits / _resp  
    - req: `{"offset,omitempty","limit,omitempty","device_id,omitempty"}`  
    - resp: `{"code","msg,omitempty","total","items":[{"permit","device_id","role","issued_by,omitempty","issued_at","expires_at"}]}`  
    - 语义：只返回当前仍有效的 permit；已消费、已撤销、已过期的不返回。  
    - 权限：`auth.permit.issue` 或 `auth.permit.revoke`
  - approve_register / _resp  
    - req: `{"request_id","role,omitempty"}`  
    - resp: `{"code","msg,omitempty","request_id","device_id","node_id","role,omitempty","status":"approved"}`  
    - 语义：分配并预留 `node_id`，但**不会**立即创建 whitelist/binding；申请方必须重试 `register` 才完成最终入网。  
    - 权限：`auth.register.approve`
  - reject_register / _resp  
    - req: `{"request_id","reason,omitempty"}`  
    - resp: `{"code","msg,omitempty","request_id","device_id","status":"rejected","reason,omitempty"}`  
    - 权限：`auth.register.reject`
  - issue_register_permit / _resp  
    - req: `{"device_id","role","expires_at,omitempty"}`  
    - resp: `{"code","msg,omitempty","permit","device_id","role","expires_at"}`  
    - 权限：`auth.permit.issue`
  - revoke_register_permit / _resp  
    - req: `{"permit"}`  
    - resp: `{"code","msg,omitempty","permit","device_id,omitempty","role,omitempty"}`  
    - 权限：`auth.permit.revoke`

核心流程
--------
- register（统一入口）：
  1. 若 `device_id` 已存在 whitelist，则视为幂等 rebind：直接返回 `status=approved`，并把当前连接重新绑定到已有 `node_id`。  
  2. 否则若 `join_permit` 存在，则按 permit 路径校验并消费：成功后立即入网为 permit 指定角色。  
  3. 否则若存在 `approved_registers[device_id]`，说明该申请已被 approve：这次 retry register 会消费该预留记录并真正创建 whitelist/binding。  
  4. 否则若启用了 `auth.bootstrap.first_register.enabled`，且当前是 local authority、`device_id` 命中配置、`epoch` 尚未消费：
     - 若配置了 `auth.bootstrap.first_register.pubkey`，则请求必须携带且匹配该公钥；不匹配时显式返回 `status=rejected`，**不会**降级到 pending。
     - 匹配成功后，节点立即以 `auth.bootstrap.first_register.role` 完成首次准入，并把消费状态持久化到 `trusted_nodes.json.meta.first_register_bootstrap`。
  5. 否则若 `auth.register.require_approval=true`，创建/刷新 pending 记录并返回 `status=pending`；这一阶段**不创建** whitelist、trusted、route index。  
  6. 否则走兼容的开放注册路径，立即分配 `node_id` 并创建 binding。  
- authority 不可达：
  - 当显式 authority 或已配置父链不可达时，普通 `register` 返回 `code=4500,msg="authority unavailable"`，不会回退为本地 authority。
  - `login` 中依赖上游 authority 的路径（例如本地缺 credential / 本地未命中需 assist）同样返回 `code=4500,msg="authority unavailable"`。
  - remote authority admin 动作在无法找到 authority 路由时，同样返回 `code=4500,msg="authority unavailable"`，而不是伪装成本地权限不足。
  - 半中心模式下，lease 过期本身**不会**阻塞一个仍然在线的父链 bootstrap 路径；真正的冻结条件是“需要上游 authority，但当前父链不可用”。
- approve 流程：`approve_register` 只完成“批准并预留身份”，不会直接把申请方接入网络；申请方必须再次发起 `register`。
  - `approve_register / reject_register / list_pending_registers / issue_register_permit / revoke_register_permit / list_register_permits` 在 remote authority 场景下仍只修改 authority 本地状态；中间节点只负责转发与回包，不复制 pending / permit 数据。
- permit 流程：permit 一次性、绑定 `device_id`，成功消费后立即失效；`device_id` 不匹配不会消费 permit。
- permit list 流程：`list_register_permits` 只返回当前活动 permit；消费、撤销、过期后的 permit 不保留在列表中。
- 登录：本地查 whitelist，缺公钥时先 assist_query 补齐；命中即验签并回 login_resp；未命中则 assist_login。成功后向父发送 up_login（逐跳报路由与公钥）。
- 直连名称缓存：当 direct child 在 register/login 或对应 assist 回包中携带 `display_name` 时，直接父节点可以把该值缓存到连接 metadata，供 management `list_nodes` 低成本返回；缺失时保持回退 `node_id`。
- 权限：角色/权限来自配置与白名单；perms_invalidate 清缓存并可刷新；perms_snapshot 应用后广播下行。
- 撤销：校验权限→删除绑定→回 resp（仅命中）→广播 revoke 上下行。
- 下线：删除绑定与索引；向父 assist_offline；无响应。

错误码（data.code）
-------------------
- 1：成功
- 202：register 已进入 `pending`
- 400：参数非法
- 4001：未找到 / 签名不匹配 / 未注册 / permit 无效
- 4403：权限不足（revoke、approve、issue permit 等）
- 4500：内部错误 / authority unavailable

配置键
------
- 权威/持久化：`authority.node_id`，`auth.disable_persist`（true 不读写 trusted_nodes），`auth.node_privkey`，`auth.node_pubkey`，`auth.trusted_nodes`（JSON map，由文件填充）。
- 半中心 authority lease：`auth.authority_mode=semi-central`，`auth.authority_policy_ttl_sec`
- 角色/权限：`auth.default_role`，`auth.default_perms`（逗号分隔），`auth.node_roles`（例 `1:superadmin;2:admin;3:node`），`auth.role_perms`（例 `superadmin:*;admin:p1,p2;node:p3`）。
- 开箱默认角色层级：
  - `superadmin:*`
  - `admin:file.read,file.write,flow.set,flow.delete,exec.call,exec.cap.query,exec.cap.sync,var.private_set,var.revoke,var.subscribe,auth.revoke,auth.pending.list,auth.register.approve,auth.register.reject,auth.permit.issue,auth.permit.revoke`
  - `node:file.read,file.write,flow.set,exec.call,exec.cap.query,exec.cap.sync`
  - `auth.default_role=node`
  - `auth.default_perms=""`
- 受控准入：  
  - `auth.register.require_approval`：`true` 时普通 register 先进入 pending  
  - `auth.register.pending_ttl_sec`：pending 与 approved 预留记录 TTL  
  - `auth.register.permit_ttl_sec`：permit 默认 TTL  
  - `auth.bootstrap.first_register.enabled`：启用首个注册 bootstrap 槽位  
  - `auth.bootstrap.first_register.role`：bootstrap 命中后授予的角色，默认 `superadmin`；必须是已定义角色  
  - `auth.bootstrap.first_register.device_id`：允许走 bootstrap 的唯一设备标识  
  - `auth.bootstrap.first_register.pubkey`：可选；配置后要求请求公钥必须匹配  
  - `auth.bootstrap.first_register.epoch`：正整数；同一 epoch 只消费一次，手工提升后才可重新开启  
- hubruntime / 父链：  
  - `parent.join_permit`：parent bootstrap 用的一次性 permit；会透传到 pre-start `SelfRegister` 和持久 parent 连接上的 register rebind。

示例
----
- 开放/已批准后的注册请求：`{"action":"register","data":{"device_id":"mac-001122334455","pubkey":"<base64 DER EC 公钥>"}}`
- 待审批响应：`{"action":"register_resp","data":{"code":202,"msg":"pending approval","device_id":"mac-001122334455","status":"pending","request_id":"req_xxx","reason":"approval required"}}`
- permit 注册请求：`{"action":"register","data":{"device_id":"mac-001122334455","join_permit":"permit_xxx"}}`
- permit/批准后注册成功响应：`{"action":"register_resp","data":{"code":1,"msg":"ok","device_id":"mac-001122334455","node_id":5,"hub_id":2,"role":"admin","status":"approved"}}`
- 登录请求：`{"action":"login","data":{"device_id":"mac-001122334455","ts":1700000000,"nonce":"n1","sig":"<ES256>","alg":"ES256"}}`
- 撤销请求：`{"action":"revoke","data":{"device_id":"mac-001122334455","node_id":5}}`
- 权限失效广播：`{"action":"perms_invalidate","data":{"node_ids":[5,6],"refresh":true}}`

集成提示
--------
- Target=0 只向子节点广播，不会上送父；上送权威必须写明目标 NodeID。
- 登录/注册均需 P256 DER 公钥 + ES256 签名；缺公钥先用 assist_query_credential 获取。
- 节点密钥与 trusted_nodes 启动时自动生成/读取，请妥善保护 `config/node_keys.json`、`config/trusted_nodes.json`。
- first-register bootstrap 仅适用于 local authority 冷启动；若配置了 `authority.node_id` 或 `parent.addr`，启用它会被视为非法配置并导致 auth 初始化失败。
- first-register bootstrap 依赖持久化消费状态；`auth.disable_persist=true` 时不得启用。
- 当前默认 bootstrap 角色是 `superadmin`；若部署侧显式覆盖为其他角色，仍要求该角色在当前 `auth.role_perms` 中真实存在。
- parent hub 在受控准入网络中应配置 `parent.join_permit`；否则 pre-start bootstrap 会收到 `status=pending/rejected` 并显式启动失败。
- 初次 permit / approve 成功后的 parent bootstrap，后续在持久 parent 连接上的 register 属于“已有身份的幂等 rebind”，不再要求新的 permit。
- 若配置了 `authority.node_id` 或 `parent.addr`，但对应 authority / 父链连接不可达，auth 不会再回退为本地 authority；部署侧应接受显式失败语义。
- 半中心模式是“root authority + 断链只读登录”的运行时约束，不会把 `effective_authority_id` 落盘到配置；重启后需要 root 重新下发 lease。
- 半中心模式下，approval / permit 管理与 assist admission 一样支持多跳 remote authority 转发；若远程管理仍失败，优先检查消费者依赖版本和 authority 路由归属是否已建立。
