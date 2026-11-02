---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# 协议正确性规范

P2 [产品等价开发验证](../operations/development-validation.md) 复用本页定义的 SDK 流程。匿名 `credential_boundary_reached` 不表示 RSA 有效或登录成功；开发 password 回执仍按本页账号状态闭环判定，但不替代正式 Candidate evidence。P2 不改变本页取键、表单、ticket 或成功语义。

## PROTO-DISCOVERY-001：发现优先

- 2026-10-30 取键契约按 [ADR-0021](decisions/ADR-0021-cas-key-use-relation.md)：完整检查允许来源，以有限局部登录使用关系选择 RSA 公钥；不得用注释、裸赋值、首个文本匹配或脚本次序选键。不同合格键、读取/解析不完整或预算失败在读取凭据前拒绝；相同 RSA 模数与指数可合并。此前仅裸赋值的成功夹具不再代表支持的认证页面。

- 每次认证动态解析 CAS form action、`lt`、`execution`、公开登录脚本和 RSA 公钥。
- CAS RSA 公钥只接受本次认证从官方同源 HTTPS 页面或公开登录脚本动态发现的 PKIX／PKCS#1 RSA，模数范围为 512–8192 位；不得固定、缓存或使用历史 fallback。RSA PKCS#1 v1.5 只编码 UTF-8 `username + password` 协议字段，明文必须不超过 `key_size_bytes - 11`；超长、畸形、非 RSA 或越界密钥均在 CAS POST 前返回 `protocol_changed`，不得截断、改发明文或降低 TLS。该兼容边界见 [`ADR-0014`](decisions/ADR-0014-cas-rsa-compatibility-envelope.md)。
- CAS 密码 POST 必须从新的表单值集合开始，只提交 `rsa`、`ul`、`pl`、`lt`、`execution`、`_eventId=submit` 六个字段；`lt` 与 `execution` 必须分别是动态 form 中唯一且非空的值，其余四项由本次凭据操作计算或固定。不得克隆或透传页面内其他具名 input、未知隐藏字段、按钮、禁用控件或重复字段。未知字段若成为必需协议输入，必须先形成新的匿名观察、规范和显式 allowlist，不得把不可信 DOM 自动扩大为凭据 POST。该闭包见 [`ADR-0015`](decisions/ADR-0015-cas-submit-closed-fields.md)。
- `ac_id` 发现只允许匿名 GET，最多读取两个有界响应：初始 captive 响应，以及至多一个经过严格校验的中间 redirect 响应。中间 Location 必须解析后仍为同一网关 host、`http` 或 `https` 默认端口、无 userinfo、fragment 和 query，路径不得含控制字符、反斜杠或点段；发现客户端不得创建 Cookie Jar，也不得发送账号、Cookie、ticket 或凭据。每个响应先检查同网关 Location 中唯一的 1–10 位十进制 `ac_id`，再检查有界正文中的唯一候选；只有首个响应允许继续一次，第二个响应后无论是否还有 Location 都必须停止。非法 Location、冲突/过多候选、第三跳需求或无法证明的形状统一返回 `protocol_changed`，不得按网卡类型、历史常量或公网可达性猜测 `ac_id`。
- 状态解析必须明确区分 JSON、JSONP、HTML 和未知响应；先判断内容类型/结构，再做业务解析。
- CAS challenge 解析在去除可选 UTF-8 BOM 与首尾空白后，必须先按首个有效字符识别 HTML；首字符为 `<` 的完整页面不得因内部 `<script>` 出现括号或对象字面量而被归类为 JSONP。JSONP 只接受覆盖整个响应的单一安全 callback 调用，其唯一参数必须是严格 JSON object；不完整、重复字段或多余脚本继续返回 `protocol_changed`。
- legacy CSV 仅在去除首尾空白后为单个无内嵌 CR/LF 的 CSV record、字段数至少为 9、位置 0 是合法 username 且位置 8 是全局单播 IPv4 时，才构成 online 身份证据；任一身份条件缺失或无效都返回 `protocol_changed`。其他位置不得覆盖或冲突推断 username/IP，也不得参与 online/offline 判定。
- legacy CSV 的位置 6/7/11 只形成一个 all-or-nothing 的可选摘要候选：位置 6 和 7 必须同时是 `int64` 范围内的非负十进制整数，且位置 11 在存在且非空时必须可精确转换为余额 minor units，才返回完整 `OnlineSummary`；否则整个摘要为 unavailable（`nil`），不得返回部分摘要或因可选摘要不可用而否定已经成立的 online 身份。未解释的非身份列保持 opaque，不得猜测其语义。
- 上述可选摘要降级只适用于 positional legacy CSV。JSON/JSONP 的显式命名字段一旦出现就承诺对应语义；无效、部分或 alias 冲突的显式 summary 仍返回 `protocol_changed`。
- 协议缓存不包含秘密，按可靠的网络上下文隔离，最长 7 天；只有经过业务成功和最终身份验证的值可写入。绑定 IP 不是网络身份，不能单独作为 cache key 或“同一网络”的证据。
- v1 尚无可靠 network fingerprint，因此禁用持久协议缓存的读取、写入与 fallback；每次操作都执行动态发现，发现失败时返回 `protocol_changed`。未来只有在能够可靠判定同一网络后，才可启用已验证缓存 fallback。
- 猜测式恢复只能作为显式诊断，不能默认启用，也不能提升证据等级。

## PROTO-REDIRECT-001：CAS ticket 截获

CAS 注册的 service 可保持 HTTP 字符串以兼容服务端，但客户端必须在真正发送下一跳之前：

1. 解析 Location；
2. 校验 host 精确为允许的 NEU gateway、端口符合策略、路径精确匹配 activation 入口；
3. 确认只含预期 ticket 参数；
4. 返回 `http.ErrUseLastResponse` 阻止 HTTP 请求；
5. 从内存读取 ticket，并仅通过正常系统 PKI 验证的 HTTPS activation 使用。

恶意 host、userinfo、非预期端口、路径穿越、额外敏感参数或不可解析 Location 一律返回 `protocol_changed`。ticket 不得进入日志、Observer、错误、JSON、缓存或 fixture。

## PROTO-QR-CONTINUATION-001：无 ticket 接续

按 [ADR-0028](decisions/ADR-0028-qr-service-continuation.md)，QR 轮询确认后可接受与本轮 service 的 scheme、host、端口、原始路径精确匹配、无 userinfo/fragment、仅含唯一相同 ac_id 的无 ticket 地址。任何 ticket 参数（包括空值/重复）继续按原 ticket 路径校验，不得降格；额外参数、路径变形和错配拒绝。

该地址只作标识，不发送 HTTP 请求。新增分支在 60 秒子上下文内（服从更短父上下文）完成：向既有 HTTPS activation API 仅以本轮 ac_id 发一次 GET、拒绝重定向、共享同轮 Jar；严格解析不超过现有 API 限制的 JSON 对象，必须且只能含 code/message/Redirect/ID，code 为整数 0，其余为字符串，Redirect 非空。message/ID 不形成成功证据、不输出。重复字段、JSONP、未知或矛盾字段、缺失 Redirect 均拒绝，不进行备用接续。

Redirect 必须匹配既有 HTTPS CAS login 的 host/端口/原始路径，无 userinfo/fragment，唯一参数为 service；内嵌 service 必须等于本轮 service 去除 ac_id 后的基址，且无 query/fragment/userinfo。仅在验证后用本地已知端点和本轮 ac_id 重建 CAS URL，同轮 Jar 继续，沿用 ticket 截获、HTTPS activation、业务成功与最终目标身份检查。网关查询一次，不循环；code 0 或无 ticket 不能被解释为登录成功。旧直接 ticket/CAS 确认路径和 180 秒 QR 轮询期限保持；新期限覆盖接续至最终 status，取消/网络/超限沿用既有分类。

## PROTO-TRANSPORT-001：HTTPS-only

status、activation 和 logout 必须使用 HTTPS；禁止 `InsecureSkipVerify`、自动 HTTP 降级和不安全覆盖开关。HTTP 仅可用于完全匿名、不携带 Cookie/ticket/账号/凭据的 captive portal 或 `ac_id` 发现，并只生成不可信提示。HTTPS 不可用时返回 `network` 或 `protocol_changed`，不得伪装为“离线”或“校外”。

## PROTO-LOGIN-001：登录成功不变量

登录成功必须同时满足：

1. 认证/activation 业务响应明确成功；
2. 最终状态为 online；
3. 最终 username 与请求的 `ExpectedUsername` 精确相等。

公网可访问、任意账号已在线、HTTP 2xx 或页面含“成功”字符串都不是充分条件。同账号在线返回 `AlreadyOnline` 且不得读取凭据；异账号默认返回 `session_conflict`。只有显式 switch 才能先注销，注销响应和离线复核任一步失败都必须中止。

## PROTO-LOGOUT-001：幂等注销

Logout 在原本离线时返回成功结果 `AlreadyOffline`；在线时必须检查 HTTPS 响应、网关业务结果并复核离线状态。不得仅因请求发出或返回 2xx 宣告成功。

按 [ADR-0024](decisions/ADR-0024-logout-explicit-callback.md)，在线注销GET在既有action与username之外显式发送固定 `callback=ipgw_logout`，请求可被现有解析器接受的表示格式。只增加该参数，不增加ip/ac_id、浏览器Cookie/header、CAS退出或重试。严格JSON/JSONP对象、整数ecode及可选success一致性检查和最终offline复核保持；callback不执行JavaScript，也不改变已有裸JSON兼容或回调语法。该候选修复是否解决真实封装差异仍以独立开发窗口为准。

## PROTO-IO-001：网络边界

所有网络方法接受 `context.Context`，设置连接/头/总时限，限制响应体大小，并保留 `context.Canceled` 与 `context.DeadlineExceeded`。Observer 只接收固定脱敏事件，不能接收原始 URL、header、body 或任意 map。
