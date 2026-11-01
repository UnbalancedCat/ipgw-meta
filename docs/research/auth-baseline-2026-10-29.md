---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
snapshot_date: 2026-10-29
---

# P1：旧版与 v1 认证行为对照

本页交付 [roadmap P1](../upgrade/plan.md#roadmap-20261029) 的本地只读研究：固定来源、行为差异、现有合成覆盖与下一阶段判别假设。它不是新认证规范，不证明今日服务端接受某一种行为，也不授权执行旧程序、系统链探测或真实登录。本轮未运行认证/诊断或产品测试，未读取配置、凭据、真实页面、证书或私有 evidence。

## 结论与证据边界

旧版与 v1 的默认传输都采用系统 TLS 验证、直接 IPv4 连接和内存 Cookie Jar；所读旧代码没有关闭 TLS 验证。二者也都构造同类 CAS service、使用 RSA PKCS#1 v1.5 封装 UTF-8 的用户名与密码拼接结果，并提交六个字段。因此，不能将“旧版能用”简单解释为“旧版跳过 TLS”，也不应把产品认证修复变成完整重写 Windows 信任系统。

源码已确认的兼容差异集中在：门户预访问与会话边界、User-Agent、页面和外部脚本的取键优先级、form/action 与 service 的关联，以及后续 ticket/activation 路径。它们是待验证的解释，不是已证明的 HTTP 500 根因。历史 R3 已到 CAS POST 并收到 HTTP 500；后来的 D4–D15 TLS 阻塞发生在另一条受限诊断路径，不能将两者合并成一个根因。

旧版确有硬编码公钥 fallback、允许 ticket HTTP 跳转、敏感 debug 输出及较弱成功判据等风险。这些代码分支的存在，不证明维护者当时的成功运行依赖了这些分支；也不能据此否定其实际可用经验。后续应保留有用行为线索，同时遵守当前安全和成功判据。

## 固定来源及可复核身份

| 标记 | 来源 | 本次确认 |
|---|---|---|
| L | 当前仓库本地 `v0.1.3`，commit `5ad8c1fa05102fbcb249e7195f4801563b3d44a5`，tree `ea4e4e589deea80600c56b6bc57685bd3bdbe03c` | 2026-03-24 的本地 tag 解析结果；用 `git show` 读对象，没有 checkout 或执行。 |
| A | 指定归档 `D:\project\Go\IPGW-Meta.archive-20261028-165538`，commit `819203b3d81d5960cde7e2777fa92991658ed04a`，tree `817904b2494c07791ed4aeb9dd41c94144c477da` | 2026-03-21 的本地 HEAD，工作区干净；只读辅助对照，没有改动归档。 |
| V | 当前工作树，分支 `codex/cas-submit-field-closure`，HEAD `b4eeb760945f0568ca163d957ddd55df2afa86c4` | 工作区有既有未提交开发变更。本研究八份产品实现文件与 HEAD 按 LF 规范化后相同；这些 HEAD blob 与 Candidate source `dac245947b9bd11b5489e2fd87ca2823c2d2dac5` 同文件 blob 内容相同。此限定比较不是整个工作树/构建身份相同。 |
| H | 旧研究引用 `3721fd60dbdc8e6a35cab9569ec295b6fcc7bdea` | 两处仓库都无法读取该对象，均没有本地 filter-repo commit-map 可用；未建立与 L 的独立 tree 映射。保留历史引用，不能用版本名推定相等。 |

L 与 A 的 `pkg/utils/http.go`、`pkg/utils/encrypt.go`、`pkg/cmd/login.go` 字节相同。主要 handler 差异是 L 增加动态 ac_id 发现、按网卡猜测的 fallback，并将其接入 service、门户预访问、activation 与 logout Referer；A 固定使用旧 ac_id。L/A 均有动态脚本取键及硬编码公钥 fallback。两者都只是本地比较基线，尚不能唯一认定哪一个就是维护者所称的最早 Gemini 可用二进制。

源码身份、逐文件 SHA-256、行数及限定比较结果保存于 `build/auth-baseline-p1-20261029/source-identities.json`，不保存旧源码副本、硬编码公钥内容、认证 URL 或原始响应。旧版 `go.mod` 声明 Go 1.25.0，但无法据此确认当时已安装 EXE 的构建工具链、字节、环境或实际走过的分支。

### 源码定位约定

L 的路径属于固定 Git 对象，不是当前 checkout 中的文件：`git show 5ad8c1fa05102fbcb249e7195f4801563b3d44a5:<path>` 可只读定位。下表行号均来自上述固定对象或本次 V 文件；后续源码变化后须按符号重新定位。

| 缩写 | 路径与关键符号 |
|---|---|
| L-H | `pkg/handler/ipgw.go`：FetchAcID L77、DoLogin L130、TestInternet L269、DoLogout L299、GetNetworkStatus L322、extractDynamicPublicKey L494 |
| L-T | `pkg/utils/http.go`：userAgentTransport.RoundTrip L17、NewNetworkClient L26、NewBoundNetworkClient L63 |
| L-E | `pkg/utils/encrypt.go`：RSAEncrypt L14 |
| L-C | `pkg/cmd/login.go`：loginCmd；预检查及登录调用 L62 起 |
| V-A | [auth.go](../../auth.go)：Login L50、passwordLogin L140、finishActivation L378、Logout L433、discoverACID L490、casRedirectPolicy L664、doCASRequest L704 |
| V-C | [client.go](../../client.go)：NewClient L100、defaultEndpoints L161、newHTTPClient L190、request L225、encryptCredential L252 |
| V-P | [internal/cas/page.go](../../internal/cas/page.go)：ParsePage L24、extractPublicKey L87 |
| V-T | [internal/transport/transport.go](../../internal/transport/transport.go)：UserAgent L14、Default L51、ReadAll L71 |
| V-S | [status.go](../../status.go)：status L26、verifyStatus L58；[状态解析](../../internal/srun/status.go)；[业务响应解析](../../internal/srun/business.go) |

## 行为矩阵：相同、差异与未知

“相同”只表示所述源码语义相同，不表示真实请求逐字节相同；“未知”指服务端要求或旧运行实际分支尚无证据。

| 环节 | L / A 的行为 | V 的行为 | 判断、依据与剩余问题 |
|---|---|---|---|
| 传输、TLS、源地址 | Proxy=nil、强制 tcp4、可绑定源 IP；未覆盖 RootCAs 或关闭验证；总超时 15 秒。 | 默认同样直接 tcp4、系统根验证、可绑定源 IP；20 秒 client 超时及更细的阶段/体积上限。 | **基本相同，资源边界不同**（L-T；V-C、V-T）。不等同于旧 EXE 与当前系统 trust store/链获取行为相同，也不保证当前产品 TLS 成功。 |
| User-Agent | 缺省填入浏览器样式 Chrome UA。 | 缺省使用 IPGW-Meta 产品 UA；不覆盖显式请求 UA。 | **差异**（L-T:17；V-T:14、26）。网关/CAS 是否按 UA 分流未知；没有证据就不先修改 UA。 |
| 凭据与初始状态 | CLI 先解析/读取密码，再查 GatewayInfo；同账号跳过，异账号自动 logout，即使 logout 失败仍可继续 DoLogin。 | Login 在读取 provider 前检查 HTTPS status；同账号无需凭据，异账号默认拒绝，显式 switch 必须确认注销成功。 | **安全差异**（L-C；V-A:50、433）。不得恢复旧自动切换或提前读取密码；旧“成功”可能走 already-online，实际历史分支未知。 |
| ac_id | L 用同一带 Jar client 跟随 redirect，发现 query 即截取；正文首个正则命中，失败按网卡 fallback。A 固定旧值。 | 匿名 client 无 Jar；至多两个响应，校验同网关目标/参数及正文唯一候选，无网卡 fallback。 | **差异**（L-H:50–128；V-A:490–662）。需要核对发现路径，但 R3 历史已到 POST，不应重做已经通过的旧 ac_id 修复。 |
| service 构造 | L 将动态 ac_id 放入 HTTP service 字符串，再作为 CAS HTTPS 登录 query；A 固定旧值。 | 同样以 HTTP service 注册上下文构造 CAS HTTPS URL，使用 url.Values 编码。 | **L/V 基本相同**（L-H:141；V-A:103–115）。并不允许 ticket 经 HTTP 发出；action 跳转后是否仍保留同一 service 另见下行。 |
| 门户预访问 | CAS 前额外 GET 门户页；忽略返回错误，与后续请求共享 Jar。A 也有此步。 | 匿名发现后直接 GET CAS；没有同等门户预访问。 | **差异、必要性未知**（L-H:143–147；V-A:103–126）。可能影响网关 Cookie/会话，但不能假定跨主机 Cookie 一定参与 CAS，也不能未经证据复制旧 HTTP Cookie 路径。 |
| CAS Cookie 会话 | 单一 handler client/Jar 贯穿预检、发现、门户、CAS、脚本、POST、redirect、activation；handler 生命周期内复用。 | status、匿名发现、CAS 分开；CAS GET/脚本共享 Jar，POST 显式继承该 Jar，activation 再继承；每次 Login 新 CAS 会话。 | **会话边界不同，CAS GET→POST 连续性代码存在**（L-T:28；V-C:190；V-A:118、209–210、395–396）。不能说 v1 完全丢 Cookie；实际 Cookie 属性、轮换和接受情况未知。 |
| form 与动态状态 | 全文查找第一个 lt/execution，POST 固定发往最初 CAS URL，不使用 form action。 | 优先 loginForm，否则首个带 action 的 form；仅取该 form 控件，lt/execution 各要求唯一非空，解析动态 action 并校验 HTTPS 同源。 | **差异**（L-H:158–193；V-P；V-A:144–151、211）。相对/绝对/缺省 action、是否保留 service、是否选对 form 都需分开验证，不能先恢复固定 POST URL。 |
| 页面与外部脚本取键 | 不先提取 inline key；只遍历 src 文本含 login/rsa 的外部脚本，首个 publicKeyStr 且特定 Base64 前缀匹配获胜。 | 先在完整页面字节中正则取 publicKeyStr/publicKey；有匹配就不读外部脚本，否则最多八个同源候选，按路径含 login/rsa 和顺序查找首匹配。 | **明显差异**（L-H:494；V-P:12、62、87；V-A:153–185）。两个不同 publicKeyStr 的实际用途未知；不能据“首匹配”确定最终加密键。 |
| 词法与脚本解析 | 正则不是 JS 解析器，未验证声明作用域/执行；相对 src 用固定目录拼接，允许绝对 src，失败继续，最终可 fallback。 | 正则同样不是 JS 解析器，大小写不敏感且无左侧标识符边界；URL 相对响应页解析，限制同源 HTTPS；脚本失败可继续，取消/超时立即停止。 | **部分共同局限、来源策略不同**（L-H:494；V-P；V-A:153–185）。注释、字符串、同名变量等合成观察已有，但尚不证明真实页面触发；不默认扩大为通用 JS 执行。 |
| RSA 与 fallback | 动态键缺失使用固定旧键；解析 PKIX，RSA PKCS#1 v1.5，加密 UTF-8 username+password，Base64 输出。 | 只接受本轮动态键，支持 PKIX/PKCS#1、512–8192 位及容量约束；同样 PKCS#1 v1.5 和拼接编码，512 位有专门兼容实现。 | **主封装相同，接受范围不同**（L-H:167–182；L-E；V-C:252–334）。不复制硬编码键；旧轮究竟用了动态键还是 fallback 未知。 |
| 六字段与长度 | 新建 form，字段为 rsa、ul、pl、lt、execution、_eventId；长度用 Go 字节数。 | 同样新建六字段 form；ul/pl 按 JS UTF-16 code unit 数，动态状态唯一性严格。 | **字段集合相同，非 ASCII 长度不同**（L-H:185–193；V-A:33、201–215）。ASCII 时长度相同；不询问或读取真实密码来判别，不把字段集合已相同当作状态值/会话正确的证明。 |
| CAS POST 与 HTTP 500 | POST 后自动跟随，记录 ticket 但不停止；没有 v1 同等 HTTP 状态分类，以最终 URL/正文决定分支。 | doCASRequest 在未捕获 ticket 且非 2xx 时返回 network；未自动重试 POST；识别挑战与认证错误。 | **差异**（L-H:201–268；V-A:218–240、704）。R3 明确 500 只定位到该次请求响应阶段，不说明是凭据、键、状态、header 或服务端异常。 |
| ticket 重定向 | hook 捕获 ticket 后返回 nil，允许继续访问，包括 HTTP service；可能已消耗 ticket，再额外调用 HTTPS activation。 | 校验精确 service 目标与 ticket/ac_id，返回 ErrUseLastResponse 阻断 HTTP，再显式 HTTPS activation。 | **必须保留的安全差异**（L-H:202–246；V-A:664–687、793；[PROTO-REDIRECT-001](../architecture/protocol-correctness.md)）。旧自动跟随是否是可用关键未知，但不能用带票 HTTP 验证该假设。此差异发生在 ticket 产生后，不能单独解释历史 CAS POST 500。 |
| activation 与 Referer | 设置 X-Requested-With 和最终 URL Referer；不严格检查业务返回，等待 3 秒后查状态/探测公网。 | 同一 HTTPS activation 路径及 X-Requested-With；无显式 Referer；业务 code/message 一致成功后最多三次状态检查，间隔 500ms。 | **差异**（L-H:243–264；V-A:378–430；V-S）。Referer 或等待时机可能影响后续兼容，须等执行抵达该阶段才验证，不把它当作 POST 500 根因。 |
| 登录成功 | TestInternet 先查网关“有账号在线”，再依赖外部站点 200/204；没有在该函数核对目标 username，也没有强制 activation 明确成功。 | activation 明确业务成功、最终 online、username 精确等于期望才 logged_in；同账号幂等在凭据前处理。 | **判据差异**（L-H:269–296；V-A:378；V-S:58）。旧码正常返回不能代替 [PROTO-LOGIN-001](../architecture/protocol-correctness.md)，但也不是旧运行必为假成功的证明。 |
| status 与 logout | HTTP status；网络错误可解释成不在校园，CSV/JSON 宽松识别；HTTPS logout 仅检查请求错误，输出正文后成功，不复核 offline。 | HTTPS status、严格身份与结构；HTTPS logout 要求业务成功并复核 offline，支持 already_offline；不设置旧 Referer。 | **安全与兼容差异**（L-H:299–380；V-S；V-A:433–481）。必须保留严格身份与注销闭环；失败分类不应降级成 offline。 |

## 已有合成覆盖与未证明部分

以下仅为测试源码审阅，不是本轮重新执行或对真实服务端的证明。

| 已有检查 | 覆盖范围 | 尚不能推出 |
|---|---|---|
| [sdk_test.go](../../sdk_test.go) 的 TestDefaultTransportUsesSystemTLSVerification | 默认客户端拒绝 loopback 自签名 TLS。 | 当前校园证书在产品 Windows 系统 PKI 下能否通过。 |
| 同文件 TestPasswordLoginCapturesHTTPServiceTicketWithoutSendingIt | 动态 action、service、六字段、HTTP ticket 不发送、明确 activation 与最终身份的 fake transport 路径。 | 真正服务端的 form/session/脚本契约。该场景未设置需要保持/轮换的 CAS Cookie，也没有证明所有 action/service 形状。 |
| 同文件 RSA/UTF16、DuplicateDynamicState、SessionDecision、SwitchLogout 测试 | 密钥结构/容量、长度、动态状态唯一性、凭据前判断及失败中止。 | 真实两个键的用途或当前账户凭据为何收到 500。 |
| 同文件匿名发现测试与 [page_test.go](../../internal/cas/page_test.go) | 无 Cookie 匿名跳转、发现边界、页面元数据与重复控件。 | 旧门户预访问的实际作用、真实脚本运行关系。 |
| [cas-observe](../../internal/cmd/cas-observe/main.go) 及独立来源分析器 | 已有生成 RSA 的单键/重复/双键顺序、页面与脚本差异、子串/注释/字符串等观察入口。 | 这些观察不是批准选键契约，也不能代替真实匿名来源证据；不需为了 P1 重做工具。 |

## P2 输入：按信息增益排序的判别工作

P1 已完成静态范围，不以查明真实根因为退出条件。下面是待 P2 收敛规范并实现的输入；不在 P1 编写新诊断或执行真实窗口。

| 优先级 / 假设 | 已知与未知 | 最小判别方法与停止条件 |
|---|---|---|
| 先确认产品路径可观察 | 受限显式 ROOT 诊断阻塞不等于产品默认系统 PKI 阻塞。 | P2 采用产品默认验证路径，先合成证明不降级 TLS、不记录材料。未来另行授权匿名窗口只需区分 TLS 阶段失败与取得受支持页面；将可能的系统取证书/缓存副作用纳入范围。若仍 TLS 失败则停在此阶段，不先扩展 CTL 语义。 |
| form/action/service 与会话关联 | v1 继承 CAS Jar 的代码存在，但真实 action 是否保持 service、动态状态是否与同轮 Cookie 对应未知。 | 先生成同源相对/绝对/缺省 action、service 保持/变化、Cookie 轮换/Path/Secure/删除的 fake 响应，验证同轮 GET→脚本→POST 连续性；只有规范支持的差异可成为修复。未来实际观测仅输出关系分类，不保存值或 Cookie。 |
| 外部脚本与页面键的实际用途 | L 从外部脚本取键，V 可被页面首匹配短路；历史有两个不同 publicKeyStr。 | 复用已有合成观察，先按固定源码比较候选、请求顺序与加密输入；未来同轮页面/脚本只作有界内存来源与使用关系分类。无法建立关系时输出 unknown，不改选第二键，不硬编码。 |
| UA 或门户预访问改变入口上下文 | 两个明确差异存在，服务端是否依赖未知。 | 先让 fake 服务对单一变化给出不同状态，证明观测可分辨且不会混用会话。只有前述观测提示此方向时才提出受控匿名对照；一次只改一个变量，不恢复共享 HTTP Cookie 或新增任意 endpoint。 |
| 凭据封装与字段值 | 算法/六字段集合大致一致，长度及动态值存在差异；真实密码内容不属于研究输入。 | 合成 ASCII、BMP/非 BMP 字符和容量边界，以解密结果/长度关系验证，不要求不同随机密文逐字节相等。与同轮 form/state、键身份分别核对；仅为有独立证据的缺陷新增回归。 |
| ticket/activation/status 下游兼容 | 旧版自动 HTTP 跟随、Referer 和等待时机不同；这些不发生在正常首次 CAS POST 之前。 | 先合成 ticket 截获、明确业务成功、延迟状态、错账号及失败注销；仅在未来运行抵达相应阶段后扩大该阶段观察。保持 HTTPS、身份与注销判据，不用旧弱成功分支获取“通过”。 |

每个未来真实窗口独立批准、绑定精确包身份，结果不确定时保持 blocked/unknown。P2 首先交付可审阅的产品等价观测/开发验证契约，再进入实现；本研究不批准公共 API、JSON、六字段 allowlist、取键或信任语义变更。

## 本次交付与剩余缺口

- 已完成：两处本地来源身份、限定 blob 对照、认证链行为矩阵、测试覆盖边界和 P2 判别输入。源码/配置文件哈希前后检查及文档验收回执位于 `build/auth-baseline-p1-20261029/`。
- 未完成：最早实际使用 EXE 与源码的精确绑定、历史 H→L tree 映射、当前产品系统 TLS 实测、真实 JS 密钥使用关系、CAS POST 500 根因、目标账号登录闭环。前两项不阻止基于固定 L/A 的独立开发。
- 正式 password_core 仍 `blocked / live_unverified`，M3 仍 `in_progress`。下一工作包为 P2；D4–D15 不恢复为默认前置，不新建 Candidate、不提交或发布。
