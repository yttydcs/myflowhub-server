# Flow Data DAG Requirements

## Background

- 当前 `flow` 主要提供“按 DAG 顺序执行 call 节点”的能力。
- 现有节点写入模型以 `method / target / args` 为主，前端主要依赖手写 JSON 参数。
- 当前运行时不会把节点输出保存为后续节点可消费的显式结果，上游节点成功执行后，下游节点无法直接复用其结果。
- 当前模型也缺少“单次 run 内可显式复用的局部变量”体系；用户若想复用中间态，只能把数据继续包在节点结果里传递，无法声明一个区别于 `varstore` 的 run-local 命名值。
- 当前运行控制面缺少显式 `cancel_run`；若要中断活动 run，只能删除整个 flow 定义。
- 当前 `status` 与 `list` 只能给出最新运行摘要，缺少按 `flow_id` 查看保留窗口内历史 run 的独立入口。
- 当前 `run/status/detail/list_runs/list/get` 仍未建立独立权限边界，调用方无法只授予“运行控制”或“只读观测”能力。
- 当前节点配置虽然有 `retry`，但失败后仍会立即重试，缺少显式 backoff 间隔策略。
- 当前 `flow` 对“活动 run 上限”只有隐式行为：
  - 手动 `run` 默认允许并发重入
  - `interval/event/var_changed` trigger 默认在已有活动 run 时跳过
  - 该差异尚未形成可声明、可审计的稳定契约
- 当前 `event/var_changed` 触发器一旦匹配就会立即尝试启动；若上游重复投递同一事件，或短时间内反复发出同一变量变更通知，运行时缺少显式去重窗口来抑制重复 run。
- 当前 retained run 虽已支持可选本地 archive，但后端仍固定为本地 JSON sidecar；在需要更强持久化或集中化存储时，仍缺少显式 run archive backend。
- 当前 `flow` 仍缺少真正的高阶编排能力：
  - 不能按条件选择分支路径
  - 不能对数组输入做显式 foreach
  - 不能在一个 flow 内同步复用另一个 flow
  - 只能按固定毫秒间隔触发，不能按日历时间调度
- 当前 `flow` 仍缺少正式的纯计算节点：
  - 不能在图内直接完成数值 / 布尔 / 字符串 / object / array 变换
  - 简单计算仍要依赖外部 capability 或调用方预处理

## Goal

- 将 `flow` 升级为“数据流 DAG”：
  - 节点可产生结构化结果。
  - 后续节点可显式消费上游节点结果。
  - flow 可在单次 run 内维护显式的局部变量，供后续节点按名称读取，而不是退回跨 run 的 `varstore`。
  - flow 需要补齐一个纯计算 `transform` 节点，用于在图内完成结构化表达式运算并产出结果。
  - flow 需要补齐基础 run control，支持显式取消指定 run，而不删除工作流定义。
  - flow 需要补齐基础观测面，支持按 `flow_id` 查看保留窗口内的 run 历史摘要。
  - flow 需要把运行控制权限与只读观测权限显式收口，避免继续裸露在无权限动作集合中。
  - flow 需要为节点失败重试补齐最小可控的固定间隔 backoff，避免立即重试。
  - flow 需要把“活动 run 上限 / 重入控制”收敛为显式定义字段，而不是依赖隐式入口差异。
  - flow 需要为 `event/var_changed` trigger 提供默认关闭、显式开启的去重窗口，减少重复通知导致的重复启动。
- flow 需要提供可选 run archive backend，把 retained window 内的终态 run 持久化下来，供重启后继续查询；未配置 PG 时仍必须保持默认可运行。
  - 编辑器默认以表单化绑定提升易用性，而不是要求用户长期手写大段 JSON。
- 在首批能力中新增 `compose` 节点，用于拼装 JSON 结果，作为后续调用节点的输入来源。
- 在本轮表达式扩展中新增 `transform` 节点，用于在无脚本前提下完成显式、可审计的运算。
- 在本轮局部变量扩展中新增 `set_var` 节点，用于写入单次 run 生效的命名局部变量。
- 在本轮高阶编排扩展中新增：
  - `branch` 节点，用于按显式规则选择后续路径
  - `foreach` 节点，用于串行遍历数组并汇总结果
  - `subflow` 节点，用于同步调用同执行器内的另一个 flow
  - `trigger.type=cron`，用于按标准 5-field cron 表达式调度

## Scope

### Must

- `call` 节点执行后必须产生可被后续节点引用的结果。
- 下游节点必须支持显式输入绑定，而不是隐式共享全局变量。
- 首批新增 `compose` 节点，用于将上游结果、触发信息和运行元数据组装为 JSON 结果。
- 必须支持 `transform` 节点，用于执行结构化、白名单式表达式树运算，并把结果暴露给后续节点。
- `transform` 表达式必须至少支持 `literal`、`source`、`object`、`array`、`op + args` 五种变体。
- `transform` 不得引入任意脚本执行、自由表达式字符串求值或用户自定义函数。
- 新增 `set_var` 节点，用于将模板和绑定物化后的值写入单次 run 的局部变量空间。
- 新增 `cancel_run` 动作，用于取消指定 `flow_id + run_id` 的活动 run。
- 新增 `list_runs` 动作，用于查询指定 `flow_id` 当前保留窗口内的 run 摘要。
- `run/cancel_run` 必须要求稳定权限 `flow.run`。
- `status/detail/list_runs/list/get` 必须要求稳定权限 `flow.read`。
- flow 定义必须支持显式 active-run 上限控制，且旧 flow 不被静默破坏。
- `event/var_changed` trigger 必须支持显式 dedup window，且默认关闭。
- flow 必须支持可选 run archive，用于持久化 retained window 内的终态 run。
- 下游节点必须支持显式读取 flow 局部变量，且读取来源要与 DAG 祖先关系保持一致。
- 必须提供独立于 `status` 的 `run detail` 查询能力，用于按 run 查看节点结果详情。
- 编辑器必须提供默认表单化输入绑定模式，并保留高级 JSON 模式作为补充。
- 数据依赖必须与 DAG 保持一致，只允许引用祖先节点结果。
- flow 局部变量必须与 `varstore` 语义明确区分：只在单次 run 内有效，不跨 run 持久化，不参与网络同步。
- 必须支持 `branch` 节点，且第一版只允许显式规则匹配，不支持脚本表达式。
- 必须支持 `foreach` 节点，且第一版按输入顺序串行执行 body graph。
- 必须支持 `subflow` 节点，且第一版仅允许同步调用同执行器内的 flow。
- 必须支持 `cron` trigger，且第一版不要求时区字段。

### Optional

- 利用能力 schema 改善编辑器参数表单和结果提示。

### Not In Scope

- 任意脚本执行节点。
- 任意表达式字符串求值。
- `foreach` 并行 fan-out / fan-in。
- cross-executor `subflow`。
- `subflow` 异步 fire-and-forget 模式。
- `cron` 时区字段与重启补跑。
- 将 flow 局部变量自动映射到 `varstore`，或让它承担跨 run / 跨节点持久化语义。

## Scenarios

1. 先调用节点 A 获取用户信息，再由节点 B 使用 `user_id` 发送通知。
2. 事件触发产生 payload，节点 A 获取设备信息，节点 B 将两者拼装后下发到远端能力。
3. 多个上游节点分别返回部分结果，由 `compose` 节点汇总为一个统一 JSON，再交给后续 `call` 节点。
4. 节点 A 计算出一个中间 JSON，`set_var` 将其写入 `session_payload`，节点 B 和节点 C 都通过局部变量读取它的不同字段，而不需要额外落到 `varstore`。
5. 用户在编辑器中选中某个已运行节点，按 `flow_id/run_id/node_id` 查看该节点结果，必要时只读取结果中的某个子路径，而不是把整个 run 大对象塞回 `status`。
6. 节点 A 根据订单状态走 `approved` 或 `rejected` 分支，只执行被选中的后续路径。
7. `foreach` 读取设备数组，逐项调用远端能力并汇总结果数组。
8. 主 flow 在完成前置准备后，通过 `subflow` 同步调用另一个通知 flow，并读取其结果节点作为当前节点输出。
9. 运维希望在每天工作日早上 9 点触发一个 flow，而不是按固定 `every_ms` 轮询。
10. 节点 A 调用 `varstore::get` 读取一个计数值后，`transform` 将该值 `+1`，再交给后续 `set_var` 或 `call` 节点继续消费。

## Functional Requirements

1. `flow` 运行时必须在单次 run 内维护显式的节点运行上下文，至少包含节点状态、返回码、错误信息和节点结果。
2. `flow` 运行时还必须维护单次 run 的局部变量空间，至少包含变量名、当前值，以及可用于静态校验和调试的写入来源信息。
3. `call` 节点必须支持“模板 + 输入绑定”的输入构建方式，而不是仅依赖原始 `args` 直写。
4. `compose` 节点必须支持基于模板和输入绑定生成 JSON 结果，且不依赖远程能力调用。
5. `set_var` 节点必须支持基于模板和输入绑定生成一个 JSON 值，并把该值写入指定的 flow 局部变量名，同时将该值作为当前节点结果暴露给后续节点。
6. 输入绑定必须显式声明来源、目标位置和是否必填，并支持在保存前做静态校验。
7. 输入绑定来源至少支持：
   - 上游节点结果
   - 触发器上下文
   - flow 运行元数据
   - run 运行元数据
   - flow 局部变量
8. flow 局部变量读取必须显式声明变量名，且可选读取变量值中的子路径。
9. flow 局部变量引用必须限制为当前节点可确定解析到的祖先写入者；若存在缺失写入、歧义写入或非法路径，必须在保存前或运行时给出明确错误。
10. 输入绑定目标必须可映射到 JSON 结构中的明确位置，避免隐式拼接字符串。
11. 运行时遇到缺失必填值、非法引用或不合法目标路径时，必须返回明确错误，不能静默降级。
12. 结果引用必须限制为祖先节点，禁止跨分支或未来节点引用，避免图外隐藏依赖。
13. flow 局部变量的生命周期仅限于当前 run，不参与 `status` 默认摘要返回，也不自动写入持久层或 `varstore`。
14. `status` 仍以摘要为主，不默认携带完整节点结果或完整局部变量值。
15. `run detail` 必须与 `status` 分离，至少支持按 `flow_id + run_id(可选最新) + node_id + path(可选)` 查询节点结果详情。
16. `run detail.path` 为空时表示读取节点根结果；非空时表示按 JSON Pointer 读取结果子路径。
17. `run detail` 命中节点时必须同时返回节点状态摘要与命中的结果值；命中失败、节点不存在或结果路径不存在时必须返回明确错误。
18. `run detail` 第一版只承载节点结果查询，不把局部变量全量调试信息并入默认响应。
19. `cancel_run` 必须要求 `flow_id + run_id`，只允许取消目标 flow 的活动 run，不得删除工作流定义。
20. `cancel_run` 命中已结束 run、未知 run 或不属于该 flow 的 run 时，必须返回明确错误，不能静默成功。
21. 被 `cancel_run` 取消后的 run，`status` 必须返回 `cancelled` 与取消原因；`detail` 查询相关节点时也必须能体现 run 已被取消。
22. `list_runs` 必须支持按 `flow_id` 返回当前保留窗口内的 run 摘要，顺序为最新到最旧。
23. `list_runs` 每条摘要至少包含 `run_id`、`status`、开始时间、结束时间和可选说明信息。
24. `list_runs` 必须支持可选 `limit`，用于只返回最近 N 条保留 run。
25. `list_runs` 不能退化成完整结果调试接口；完整节点结果仍通过 `detail` 查询。
26. `run` 与 `cancel_run` 必须显式要求 `flow.run`，不能继续依赖“无权限动作”默认放行。
27. `status`、`detail`、`list_runs`、`list`、`get` 必须显式要求 `flow.read`，以支持独立观测授权。
28. `graph.nodes[].retry_backoff_ms` 必须支持可选固定间隔毫秒数；当节点失败且仍有剩余重试次数时，执行器必须在下一次尝试前等待该时长。
29. `retry_backoff_ms=0` 或缺失时，节点继续保持当前立即重试行为。
30. `retry_backoff_ms<0` 的新写入 graph 必须被拒绝，不能静默归零。
31. backoff 等待期间若 run 被取消或 flow 被删除，执行器必须立即停止等待和后续重试。
32. flow 定义必须支持可选字段 `max_active_runs`，用于控制同一 `flow_id` 允许同时存在的活动 run 上限。
33. 当 `max_active_runs` 未设置时，执行器必须保持当前兼容行为：
   - 手动 `run` 继续允许并发重入
   - `interval/event/var_changed` trigger 继续在已有活动 run 时跳过本次启动
34. 当 `max_active_runs=0` 时，所有启动来源都视为“不限制活动 run 数”。
35. 当 `max_active_runs>0` 时，手动 `run` 与 trigger 启动都必须遵守统一 active-run 上限；手动超限返回明确冲突，trigger 超限则跳过本次启动。
36. `max_active_runs<0` 的新写入 flow 定义必须被拒绝；读取完整定义时也必须能回显该字段。
37. `event` 与 `var_changed` trigger 必须支持可选字段 `dedup_window_ms`，用于抑制短时间内重复的同类 trigger 启动。
38. 当 `dedup_window_ms` 省略或为 `0` 时，执行器必须保持当前行为，不做额外 trigger 去重。
39. 当 `dedup_window_ms>0` 时，执行器必须在内存中按“同一 flow + 同一规范化 trigger 上下文”做窗口去重；窗口内重复 trigger 不得生成新的 run。
40. dedup 必须是显式 opt-in；执行器重启后 dedup 记忆可清空，不要求跨重启持久化。
41. `flow.run_archive.backend` 必须支持 `off | file | pg`；未设置时保持默认关闭语义。
42. `flow.run_archive_enabled=true` 必须继续兼容，并等价于启用 `file` backend。
43. 当 backend=`file` 时，执行器必须把 retained window 内的终态 run 摘要和节点结果持久化到本地 archive。
44. 当 backend=`pg` 时，执行器必须把 retained window 内的终态 run 摘要和节点结果持久化到 PG；PG 必须是显式 opt-in，未配置 PG 时系统仍可正常运行。
45. run archive 仍复用 `flow.max_retained_runs` 作为 retained window 上限；超出窗口的更旧 archive 可以被清理。
46. `status`、`detail`、`list_runs` 命中 retained window 内的 archived run 时，返回结果必须与内存 retained run 保持一致，即使执行器重启。
47. 删除 flow 定义不应立即删除 retained archive；只要 archived run 仍在 retained window 内，`list_runs/status/detail` 仍可查询这些 run。
48. `branch` 节点必须按声明顺序匹配 case，命中首个匹配项后只激活该 case 对应的后续路径。
49. `branch` 节点未命中任何 case 时，若配置了 `default_case` 则走默认路径；否则节点失败。
50. `branch` 的路由必须通过图中显式标记的 edge case 完成，未选中的路径节点状态必须可观察。
51. `foreach` 节点必须从显式数组来源读取 items；来源缺失或不是数组时必须明确失败。
52. `foreach` 必须为每次迭代提供至少 `loop_item` 和 `loop_index` 两类可绑定来源。
53. `foreach` body graph 的每次迭代必须使用隔离的局部上下文，迭代内 `set_var` 不得污染外层 flow 的局部变量空间。
54. `foreach` 节点必须按输入顺序收集每次迭代的 `result_node_id` 结果，形成结果数组。
55. `subflow` 节点必须同步执行目标 flow，等待目标 flow 终态后再继续当前 flow。
56. `subflow` 节点第一版只允许调用同一执行器内已存在的 flow 定义。
57. `subflow` 节点必须禁止直接自调用和递归调用链。
58. `subflow` 节点必须支持为子 flow 构造结构化输入，并通过子 flow 的 `result_node_id` 读取结果。
59. `cron` trigger 必须支持标准 5-field cron 表达式，并按执行器本地时区计算下一次触发时间。
60. `cron` trigger 第一版不要求时区字段，执行器重启后也不补跑错过的窗口。
61. `transform` 节点必须只产生结果，不直接产生外部副作用。
62. `transform` 节点必须支持显式表达式树，且单个表达式节点只能选择 `literal/source/object/array/op` 其中一种变体。
63. `transform.source` 必须复用现有绑定来源模型，至少支持 `node_result`、`trigger`、`flow_meta`、`run_meta`、`flow_var`、`loop_item`、`loop_index`。
64. `transform.source` 必须支持可选 `required`；默认 `true`，显式 `false` 时来源缺失返回 `null`，以便与 `coalesce/if` 配合。
65. `transform` 第一版至少支持以下白名单运算：`add/sub/mul/div/mod/neg/abs/min/max`、`eq/ne/gt/gte/lt/lte`、`and/or/not/coalesce/if`、`concat/lower/upper/trim`、`len`。
66. `transform` 必须在保存前拒绝未知 op、错误参数个数、非法 source、越界的 `loop_item/loop_index` 引用和多变体混用。
67. `transform` 运行时遇到类型不匹配、除零、必填来源缺失时，必须让当前节点明确失败，不能静默做隐式转换。
68. `transform` 必须可在 `foreach.body` 内使用，并允许显式消费 `loop_item` 和 `loop_index`。

## Non-functional Requirements

- 易用性：
  - 默认使用表单化绑定，而不是要求普通用户长期维护复杂 JSON 模板语言。
- 可审计性：
  - 数据依赖必须可从图和节点配置中直接读出。
- 性能：
  - 避免对大型 JSON 进行无意义的重复序列化、反序列化和深拷贝。
- 可扩展性：
  - `branch / foreach / subflow / set_var` 等节点应尽量复用同一套运行上下文与绑定模型。
- 边界清晰：
  - `flow` 局部变量只解决单次 run 内的局部中间态，不与 `varstore` 的跨 run、跨节点、持久化语义混淆。
- 安全性：
  - 第一版不引入任意脚本求值和动态表达式执行。

## Edge Cases

- 绑定引用的节点不存在。
- 绑定引用的节点不是当前节点祖先。
- 绑定路径存在，但运行时结果中缺少对应字段。
- 触发器没有自然 payload（例如 interval），但下游仍请求读取触发上下文。
- 节点结果过大或包含敏感字段，不能默认通过轻量状态接口暴露。
- 结果对象很大，但用户只需要某个嵌套字段时，应允许通过 detail 子路径读取，而不是重复下载整个结果。
- `flow_var` 引用的变量名不存在。
- 多个祖先 `set_var` 对同一变量名形成歧义写入，当前节点无法唯一确定读取哪一个值。
- 同一路径上后写入的局部变量应覆盖前写入的同名值，但不应破坏图上可审计性。
- `retry_backoff_ms=0`。
- `retry_backoff_ms<0`。
- 节点在 backoff 等待期间被 `cancel_run` 或 `delete` 中断。
- 手动 `run` 被双击或并发请求击中同一 flow，不能穿透 active-run 上限。
- `max_active_runs=0` 与字段未设置的语义不同，必须显式区分。
- 同一 `event` payload 在短窗口内被重复投递。
- 同一变量在极短时间内连续触发相同 `changed/deleted` 通知。
- branch case 与 edge case 不一致。
- branch 未命中任何 case 且没有 `default_case`。
- `foreach` source 为空、不是数组、数组元素是复杂对象。
- `foreach` body graph 合法，但 `result_node_id` 不存在。
- `subflow` 指向不存在的 flow、指向自身或形成递归链。
- `subflow` 子 flow 失败，或指定 `result_node_id` 不存在。
- `cron` 表达式非法，或在跨月/跨周边界计算下一次时间。

## Acceptance Criteria

1. 一个两节点 DAG 中，节点 B 可以将节点 A 的结果字段映射到自身输入后成功执行。
2. `compose` 节点可以将多个上游来源组装成一个 JSON 结果，并被后续 `call` 节点消费。
3. `set_var` 节点可以把一个中间 JSON 值写入局部变量，后续节点可通过显式变量引用读取该值或其子字段。
4. 编辑器默认提供表单化输入绑定配置；高级模式仅作为补充入口。
5. 非祖先引用、缺失必填绑定、局部变量歧义写入和非法目标路径在保存前或运行时会得到明确错误。
6. 既有调度动作与执行顺序语义保持 DAG 拓扑执行，不因数据流增强而退化为隐式共享变量模型。
7. 对某次 run 的某个节点发起 detail 查询时，可以拿到根结果或指定子路径的结果，而 `status` 仍只返回摘要。
8. 对活动 run 发起 `cancel_run` 时，run 会转为 `cancelled` 且 flow 定义仍保留；对已结束 run 发起取消时会得到明确冲突错误。
9. 对一个已有保留 run 的 `flow_id` 发起 `list_runs` 时，可以按最新到最旧看到 run 摘要，并可通过 `limit` 截断结果集。
10. 仅具备 `flow.run` 的调用方不能直接读取 `status/detail/list_runs/list/get`；仅具备 `flow.read` 的调用方不能执行 `run/cancel_run`。
11. 节点配置 `retry_backoff_ms>0` 时，失败重试之间会按固定毫秒数等待；等待期间若 run 被取消，则不会继续发起后续 attempt。
12. `max_active_runs=1` 时，已有活动 run 的 flow 再次手动 `run` 会返回明确冲突，而 trigger 不会生成额外 run。
13. `max_active_runs=0` 时，trigger 来源也允许生成重叠 run；未设置该字段时仍保持当前 trigger 单飞兼容行为。
14. `event/var_changed` trigger 配置 `dedup_window_ms>0` 后，同一规范化 trigger 在窗口内重复出现时不会生成新的 run。
15. `dedup_window_ms` 未设置或为 `0` 时，trigger 行为保持现状；窗口外的同类 trigger 或不同规范化 trigger 仍可正常启动。
16. 开启 run archive 后，执行器重启后仍可对 retained window 内的 run 使用 `status/detail/list_runs`。
17. 删除 flow 定义后，只要 archived run 仍在 retained window 内，`list_runs/status/detail` 仍可继续查询。
18. `branch` 节点命中某个 case 后，未选中的路径节点会被标记为 `skipped`，merge 后的公共节点仍可继续执行。
19. `foreach` 节点对一个三元素数组执行 body graph 后，会得到三个结果组成的数组，且顺序与输入一致。
20. `subflow` 节点可以同步执行另一个 flow，并返回指定结果节点数据；若目标 flow 失败，当前节点也会明确失败。
21. `cron` trigger 配置工作日早 9 点表达式后，会在执行器本地时间命中该窗口；未配置时区字段仍语义明确。
22. `transform(add)` 可以直接实现“读取上游结果中的数字并加一”。
23. `transform` 可以在同一个节点中构造嵌套 object / array 结果，并在 `foreach.body` 内消费 `loop_item/loop_index`。
24. `transform(coalesce)` 可以把 `required=false` 的可选来源与 fallback 值组合起来，来源缺失时仍返回明确结果。
25. `transform` 的未知 op、错误参数个数、非法 source 会在保存前被拒绝；类型不匹配和除零会在运行时明确失败。

## Related Specs

- [../specs/flow.md](../specs/flow.md)

## Related Changes

- 待本次 workflow 完成后补充。
