auth 协议（SubProto=2，基于 P256 公钥签名）
==========================================

范围与格式
----------
- 仅描述当前 `internal/handler/auth` 的实现；不包含 login_server 旧 credential 流程。
- 消息统一格式：`{"action":"<name>","data":{...}}`，响应 action = `<req>_resp`，状态码在 data.code。
- 签名算法：ES256（P256 + SHA256），公钥/私钥 DER 以 base64 编码。
- SubProto 固定 2；未认证连接 `SourceID=0` 仅放行子协议 2，其余丢弃。

头部与路由规则
--------------
- TargetID=0 仅表示“向子节点广播，不回父”，**不**表示上送；上送父/权威需写明 NodeID。
- Major：命令/状态类用 `MajorCmd`；响应可用 `MajorOKResp`。
- 权威选择：优先配置 `authority.node_id`，否则父链接；无父则本地即权威。

密钥与持久化
------------
- 节点密钥：启动时从 `config/node_keys.json` 读取/生成（字段 `privkey`、`pubkey`，base64 DER），并写入配置键 `auth.node_priv_key`、`auth.node_pub_key`。
- 信任/白名单：`config/trusted_nodes.json`
  - `bindings`: device_id -> `{node_id,pubkey,role,perms}`，注入 whitelist。
  - `meta`: 预留。
 读取时注入 whitelist 与 trusted 节点公钥；持久化时同步写回缺失的 trusted 公钥。

动作与数据字段
-------------
- register / assist_register  
  - req: `{"device_id","pubkey,omitempty","node_pub,omitempty","ts,omitempty","nonce,omitempty"}`（缺省 pubkey 会填本节点公钥）。  
  - resp: `{"code","msg,omitempty","device_id","node_id","hub_id","role,omitempty","perms,omitempty","pubkey,omitempty","node_pub,omitempty","ts,omitempty","nonce,omitempty"}`
- login / assist_login  
  - req: `{"device_id","node_id,omitempty","ts","nonce","sig","alg"}`，需 ES256 签名。  
  - resp: 同 register_resp，失败 code=4001。
- assist_query_credential / _resp  
  - req: `{"device_id","node_id,omitempty"}`  
  - resp: `{"code","msg,omitempty","device_id","node_id","role,omitempty","perms,omitempty","pubkey,omitempty","node_pub,omitempty"}`
- up_login / up_login_resp  
  - req: `upLoginData` 字段：`node_id,device_id,hub_id,pubkey,ts,nonce,device_ts,device_nonce,device_sig,device_alg,sender_id,sender_ts,sender_nonce,sender_sig,sender_alg,sender_pub,alg`。  
  - 校验：设备签名有效；发送节点签名有效且为信任节点或携带合法公钥；路由公钥冲突则拒绝。
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

核心流程
--------
- 注册：本地权威或 assist_register 上送权威分配 node_id；保存 whitelist/路由/信任公钥，返回 register_resp/assist_register_resp。
- 登录：本地查 whitelist，缺公钥时先 assist_query 补齐；命中即验签并回 login_resp；未命中则 assist_login。成功后向父发送 up_login（逐跳报路由与公钥）。
- 权限：角色/权限来自配置与白名单；perms_invalidate 清缓存并可刷新；perms_snapshot 应用后广播下行。
- 撤销：校验权限→删除绑定→回 resp（仅命中）→广播 revoke 上下行。
- 下线：删除绑定与索引；向父 assist_offline；无响应。

错误码（data.code）
-------------------
- 1：成功
- 400：参数非法
- 4001：未找到 / 签名不匹配 / 未注册
- 4403：权限不足（revoke 等）
- 4500：内部错误（预留）

配置键
------
- 权威/持久化：`authority.node_id`，`auth.disable_persist`（true 不读写 trusted_nodes），`auth.node_priv_key`，`auth.node_pub_key`，`auth.trusted_nodes`（JSON map，由文件填充）。
- 角色/权限：`auth.default_role`，`auth.default_perms`（逗号分隔），`auth.node_roles`（例 `1:admin;2:node`），`auth.role_perms`（例 `admin:p1,p2;node:p3`）。

示例
----
- 注册请求：`{"action":"register","data":{"device_id":"mac-001122334455","pubkey":"<base64 DER EC 公钥>"}}`
- 注册响应：`{"action":"register_resp","data":{"code":1,"msg":"ok","device_id":"mac-001122334455","node_id":5,"hub_id":2,"pubkey":"<...>","node_pub":"<...>"}}`
- 登录请求：`{"action":"login","data":{"device_id":"mac-001122334455","ts":1700000000,"nonce":"n1","sig":"<ES256>","alg":"ES256"}}`
- 撤销请求：`{"action":"revoke","data":{"device_id":"mac-001122334455","node_id":5}}`
- 权限失效广播：`{"action":"perms_invalidate","data":{"node_ids":[5,6],"refresh":true}}`

集成提示
--------
- Target=0 只向子节点广播，不会上送父；上送权威必须写明目标 NodeID。
- 登录/注册均需 P256 DER 公钥 + ES256 签名；缺公钥先用 assist_query_credential 获取。
- 节点密钥与 trusted_nodes 启动时自动生成/读取，请妥善保护 `config/node_keys.json`、`config/trusted_nodes.json`。**
