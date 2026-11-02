---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# 产品等价开发认证验证

<a id="qr-continuation-development"></a>

## 2026-11-02 QR 独立开发验证

按 [ADR-0028](../architecture/decisions/ADR-0028-qr-service-continuation.md) 新增 qr-dev-launch 独立身份，不扩展下方 password observed 模式。开发包绑定源码快照、普通产品、工具和构建输入哈希，运行前后核验；预检仅检查本地包/有线拓扑，无 profile/status 读取。self-test 使用合成执行器。真实模式需新的单次授权和私有 TTY 确认。

复用既有 terminal_qr 状态机：初始 offline、QR、独立目标 status、logout、最终 offline；清理权与拓扑变化停止规则不变。普通产品 QR 输出直通私有 TTY，不捕获或写日志。总期限 10 分钟，QR 子命令 300 秒，其他子命令 120 秒，父上下文更短时优先。开发报告 schema v1、tool_kind=qr-dev、release_evidence=false，独立记录合并前命令枚举/退出码/固定错误分类/结果/耗时及完整性；整体通过仍以状态机身份闭环为准。无原始输出、账号、IP、URL、Cookie、ticket 或二维码；不改变正式回执，不声称观察 QR 内部阶段。

缺失/混配/运行中漂移、TTY 或 profile 不满足、拓扑变化、取消、超时和不完整报告均不得 pass。结束后由维护者恢复管理连接；硬中断意味着账号状态未知。真实失败不自动重试或增加诊断；正式 QR/M3 仍须新 Candidate 的正式证据。

外层固定记录 source_digest、product_sha256、tool_sha256（受绑定 worker）、binding、state/reason、live_attempted、preflight、qr；启动器文件摘要另由包外构建记录固定。qr 对象的 commands 为合并前记录，steps 保持状态机解释，两者须交叉校验；QR 命令 pass 只表示该子进程成功，整体 pass 仍要求目标身份闭环。取消或拓扑中断保留已校验的命令事实并使外层 blocked；文件绑定漂移则拒绝可信报告。preflight 只报告本地 ready/blocked 与零请求/零 profile 计数，不代表账号 offline。最终身份、指令及预检结果见[实施记录](../research/qr-continuation-progress-2026-11-02.md)。

本页落实 P2 和 [ADR-0017](../architecture/decisions/ADR-0017-development-auth-validation.md)。这是独立开发工具的契约，产品 SDK/CLI/JSON、认证能力及正式四组 Candidate 门禁不变。实现与验收进度见[当前状态](../upgrade/status.md#status-20261029)。

### QR 门禁首因定位（2026-11-02 已批准）

新开发启动器外层升级为 schema v2，新增必需 `gate_failure`（无已确认失败为 null）；内层 qr 命令报告保持 v1 原义。v1 外层仅作显式历史解码，不允许新字段，不补造首因；v2 严格拒绝缺失/未知/重复字段、截断和矛盾状态。产品 JSON、正式 runner/evidence 及成功判据不变。

首因对象仅含 command、phase、reason、query_reason、changes。command 是 initial_check、profile_show、initial_status、qr_login、status_online、logout、final_status、cleanup_logout、cleanup_status、final_check；phase 是 before/during/after/final。reason 为 binding_failed/query_unavailable/snapshot_changed；query_reason 闭集为 none/unsupported/launch_failed/canceled/deadline/exit_nonzero/stderr/output_limit/structure_rejected/local_rule_rejected/local_query_failed/query_rejected；changes 精确为 address/interface/dns/routes/adapters/ssid 六个布尔槽。只有 snapshot_changed 可以且必须至少一个 true；只有 query_unavailable 可以有非 none query_reason。命令及位置需与可用轨迹相符，preflight 未建立基线时不伪造差异。

保留第一个确认失败，后续检查不能覆盖；命令正常结束所取消的进行中监测查询不记录为失败。保持现有检查调用次数、接受条件、15 秒单查询及 64 KiB 输出限制，不读取新来源、不保存原始拓扑或可关联指纹、不输出错误文本、地址、接口标识。系统命令读取失败与明确本地规则拒绝分别分类，无法细分时 query_rejected。SSID 路径沿用已有读取规则。

仅独立开发工具使用详细投影；共享检查器保留原 bool 接口及正式行为。首次定位失败意味着整体不能 pass，不据此宣称请求是否已发送；phase 只说明检查位置。离线矩阵须覆盖全部原因、变化槽、首因保留、正常取消、绑定漂移、v1/v2 严格兼容和泄漏 canary。完成新绑定包、自检及零网络预检后，真实测试仍需新单次授权。


## REL-DEV-AUTH-001：身份、授权与输出

开发包绑定精确源码快照、构建输入、观察程序及产品 EXE 的 SHA-256；不接受运行时替换哈希。入口为 `auth-dev-launch.exe`，payload 固定为 `auth-dev.exe`、`ipgw-meta.exe`、`build-inputs.json`。启动前锁定目录和文件，拒绝重解析点、硬链接、缺失/额外文件与哈希不符；结束前复核，失败不发布成功回执。快照与构建记录保留在开发目录，不安装到 PATH。

`--self-test` 仅使用假传输和合成材料；`--preflight` 只核对包和本地拓扑，不读取 profile。`--anonymous-once` 与 `--password-core PROFILE` 是互斥的真实操作，必须先获得新的单次授权并在原生终端确认。没有参数不执行操作。P2 实施只执行合成测试和绑定自检，不执行上述真实模式。

匿名模式可在另行授权后由 Agent 执行。password 模式必须由维护者在私有终端执行，stdin/stderr 为 TTY，凭据仅由已有产品 prompt provider 获取；Agent 不捕获 stderr、密码、账号或 profile 原文。只允许 prompt profile，不支持 QR、switch 或 provider 回退。每次运行前明确 TUN 关闭、指定有线独占、账号 offline；恢复方法是重新启用原 Clash TUN，保留核心、系统代理和节点。门户页面不是本工具的认证成功判据。

开发 JSON 使用独立 `tool_kind=auth-dev` 与 schema 1；固定 `release_evidence=false`。只输出包身份、闭集状态/原因、应用计数和投影后的步骤。`step_coverage=not_started/complete/partial` 区分未开始、状态机完整返回和取消/runner fault 的不完整观察；未观察步骤不补造，partial 不能 pass。不含账号、接口/IP/DNS、URL/query、响应正文、Cookie、ticket、隐藏字段、公钥、证书或自由错误文本。内存中处理材料，不保存响应；日志只接收闭集报告。它不是 evidence bundle，不提供 Candidate ID，不提升正式里程碑或能力。

## REL-DEV-ANON-001：复用 SDK 的凭据前观察

直接调用现有 SDK `Login`，使用非真实的固定占位身份、`SwitchRefuse` 和始终拒绝的 CredentialProvider。SDK 独立完成 status、匿名 ac_id/service 构造、CAS 页面/脚本请求及同轮内存 CookieJar；不复制登录协议，不注入 Cookie，不读取 profile 或凭据。匿名发现仍不回送 Cookie；CAS Cookie 按产品同源 HTTPS 规则接收/回送，只保留在该次调用内。

外层守卫只允许前置 GET：status、无凭据的网关发现、同源 HTTPS CAS。所有 POST、activation、logout、HTTP Cookie、出域和 ticket 请求在发送前停止；初始 online 停止，早期 SSO ticket 触发 `early_ticket` 而不 activation。预算最多 32 次应用 RoundTrip、120 秒协作期限，父进程 145 秒终止。超限、取消或拓扑变化立即拒绝后续请求。此计数不是 DNS/TCP/操作系统流量计数。

provider 被调用只表示 `credential_boundary_reached`；SDK 在 provider 返回后才做 RSA 加密，因此不能把这个结果称为密钥有效、可以登录或登录成功。其他失败只输出 SDK 稳定错误码或守卫闭集原因。观察只记录响应角色、是否接收/回送 CAS Cookie、页面/脚本是否发现候选，以及动态 action 的 service 是否与入口一致；不改变选择顺序、请求参数、重定向、正文或 Cookie 规则。

响应观察使用有界被动读取，在 SDK 消费时收集副本，不预读、不额外 drain、不替换错误；每份最多 2 MiB，完整读取才解析。观察失败不把不完整材料当作完成结果。所有原始字节在调用结束释放，测试证明请求轨迹及凭据边界与无观察 SDK 等价。

## REL-DEV-PASSWORD-001：复用产品与闭环状态机

独立入口执行绑定的 `ipgw-meta.exe`，复用 `livegate.ProcessExecutor` 的有界 JSON 投影与 `StateMachine` 的 password_core 七步和 cleanup 所有权。不调用正式 Candidate 验证器或 evidence publisher。确认后在整个窗口只读锁定现有配置文件及路径，拒绝重解析/硬链接并防止运行中更换；启动器不读取配置正文，由产品 profile show 投影检查 prompt provider 和 refuse switch。账号只在内存中参与精确比较；第一轮 login 连接私有 stdin，第二次 login 不提供 stdin。产品进程只继承 Windows 和用户配置/临时目录必需变量，不继承 IPGW 参数、代理或 TLS 覆盖环境。

初始 offline、登录、目标账号 online、第二次 already_online、注销、offline、第二次 already_offline 必须全部成立才可报告开发 `pass`。cleanup 只按既有状态机确认的所有权进行；拓扑不再匹配则拒绝任何后续网络动作，包括 cleanup，并明确结果 blocked/failed，交由维护者恢复。不能把登录结果 alone 当作 cleanup 所有权。

运行总协作期限 10 分钟；最多七个主步骤及两个 cleanup 步骤，每个产品调用最多 120 秒，cleanup 使用已有每步 30 秒上限。产品内部仍使用自身有界请求/重定向；独立进程入口只能报告 CLI 调用次数，不能宣称测量了其 HTTP/DNS/TCP 总数。每步前检查本地拓扑，执行中监测变化并取消；子进程由 `CommandContext` 终止，stdout 只进入已有有界投影器，stderr 直接到私有 TTY。取消/错误不重试整个窗口。

## SEC-DEV-TRANSPORT-001：产品传输与本地测试路径

匿名模式使用 `transport.Default(selectedIPv4)` 后包装观察器；不能仅在自定义 RoundTripper 外传 `WithBindIP` 冒充实际绑定。password 模式向产品传递现有 `--bind-ip`。均保持产品 UA、无应用代理、IPv4 TCP 源地址绑定及默认系统 TLS/PKI，不设置额外 RootCAs，不禁用验证。

本地检查要求唯一活动物理有线接口/IPv4、没有活动物理 Wi-Fi、默认 IPv4 路由指向所选接口、没有二分覆盖路由；允许无关虚拟接口。基线及每次复核比较所选接口、地址、DNS 和路由快照，仅内存保留。检查是本地快照，不声称原子路由锁定；切换瞬间仍有竞态，所以要求整个授权窗口拓扑不变。

DNS 使用系统默认解析；不再承诺 D 系自定义 DNS 接口绑定、每个目标连接前路由验证或精确系统 DNS/TCP 总量。系统 PKI 可能联网获取材料、更新缓存，应用计数不覆盖这些隐式操作，须纳入每次真实授权。P2 的离线测试不调用这些实际路径。

## REL-DEV-TEST-001：离线验收与切换条件

必须覆盖真实 SDK 假传输轨迹、CAS 同轮 Cookie 与无 Cookie 发现、初始在线、早期 ticket、凭据前停止、出域/POST/HTTP Cookie 拒绝、预算/取消/拓扑变化、被动读取错误、脱敏与严格解码；password 覆盖既有状态机成功/失败/cleanup、第二次 login 无 stdin、绑定 IP 和默认正式调用不变。

包测试覆盖完整自检链、篡改/缺失/额外文件和错误身份在执行前拒绝。Windows 普通测试/vet、Linux race/vet、文档检查和脱敏检查通过；产品依赖闭包不含开发诊断，六平台产品构建继续通过。Windows 原生保护需要 Windows 测试，Linux/macOS 构建不代表原生或真实认证验收。

只有实现、上述回归和绑定开发包齐备，P2 才完成；实际窗口另行授权。P4 之前正式门禁仍为原四组，不接受开发回执替代 Candidate evidence。

<a id="p3-failure-observation-proposal"></a>

## 2026-10-30 已确认：P3 失败阶段观察

本节与 [ADR-0018](../architecture/decisions/ADR-0018-development-auth-failure-observation.md) 已由维护者确认，批准离线实现与验收。上文 P2 schema 1 和旧包保持历史含义；新包验收后才提供下述 schema 2 入口。目的仅为区分本轮宽泛 network 的来源，不预设 HTTP 500、密码错误或选键错误；不恢复 CTL 路线，不授予真实运行权限。

### 接入及输出

推荐一次完成独立 observed CLI：复用现有 CLI/app/SDK、profile 解析及 prompt，不改普通产品入口。internal app 仅增加显式可选构造 seam，将按实际 bind IP 创建的默认 transport 交给被动 decorator；不提供任意 endpoint、根证书或代理覆盖。未选择该构造入口时依赖和行为保持原样。

每个子命令记录固定操作名、SDK 结果类别及有序请求观察；请求观察只含角色、响应状态码和固定传输类别。角色固定为 status、discovery、cas_get、cas_post、activation、logout、other；不把 cas_get 自动宣称为具体页面或脚本，不把最后一次请求自动宣称为失败根因。保留顺序可区分提交前 GET 与提交后 GET，但不输出 URL、header、query、Cookie 或 ticket。

HTTP 状态只接受 100–599；无响应或约定外状态用 unknown，原产品仍按原规则处理。传输类别仅为 none、canceled、timeout、dns、tls_certificate、tls_other、other；按标准 typed error/context 分类，不从自由字符串猜测。TLS 握手失败不等于证书校验失败，类别不确定用 other。请求返回 response/error 的组合必须如实标记，不能把 HTTP 2xx 当作业务成功。

SDK 结果用固定 Code/Message 组合补充 cas_request、cas_http_status、activation_request、activation_http_status、credential_provider、encryption、status_identity、other/unknown；完整列表与映射必须在实现验收前锁定并回归。上下文取消/超时单独保留，不把未知消息作子串匹配。直接返回的原 result/error 原样交给 CLI；不保存错误对象或 Details 中的身份字段。

不增加凭据 hook；无法观察的 provider 返回/RSA 内部步骤保持 unknown。已有密码提示不是诊断输入。请求正文、header、query 和响应正文均不读取/复制；不包装正文 reader，不额外 drain。SDK 在正文读取后返回的固定分类仅作结果补充，无法定位则保留 unknown。

### 预算、身份和子进程

- 复用现有 10 分钟窗口、每命令 120 秒、七个主步骤及最多两个 owned cleanup 的预算；不增加认证或状态请求，不修改 cleanup 所有权。没有新真实授权时完全不运行。
- 每子命令最多 32 条请求观察，全窗口最多 128 条，最终独立报告不超过 64 KiB。超限后清除该观察集合并标记 incomplete，不发布截断的“完整”轨迹；不因观测溢出中止原产品请求、篡改结果或跳过必要 cleanup。诊断完整性与状态机结果分开，不能用验证 pass 掩盖观测缺口。
- 单条只保留闭集标量；固定大小存储，锁覆盖状态更新但不跨网络调用。无历史文件回读、全局缓存、写配置、系统证书查询或网络探测。
- 新包为独立开发身份，同时绑定普通 product 与 observed CLI 的哈希和同一源码快照。旧 P2 包不修改，旧 product_sha256 不重解释。launcher 必须验证完整新增文件集合、拒绝混包/缺失/多余/重解析/硬链接，并在执行前后复核。
- 新开发子进程使用独立严格封装，区分原 CLI 投影和闭集观察；不向公共产品 JSON 随意加字段。父进程仅内存流式处理原产品 JSON，账号仅用于既有精确身份比对。原始产品输出和私有 stderr 不落盘、不回传，不能把敏感字段藏进新的 RawMessage 后序列化。
- 新 schema/工具身份与 v1 回执严格区分；旧回执没有观察数据，不能迁移为已观察。正常产品和正式 runner 继续使用旧协议。第二轮 login 不提供 stdin，私有 TTY、profile 锁及最小环境继承保持。

### 实施顺序与验收

1. 确认 ADR 和本节的整体选择后，锁定闭集 schema、错误映射、协议分帧及源码身份规则；先补相关规范，再实现。不得把 proposed 写成 accepted 或已实现。
2. internal app seam 用合成 fake transport 验证：普通构造行为不变、实际绑定的默认 transport 只被包装、无代理/TLS覆盖；无密码 provider 重实现。
3. 完整假响应轨迹对照：相同 SDK/CLI 在有/无观察器时，请求方法/顺序/参数、Cookie、正文消费、返回值和退出码一致。覆盖 CAS POST 500、CAS传输错误、重定向、activation错误、最终状态失败、成功与取消；全部使用新生成的合成凭据/密钥或明确的 loopback TLS。
4. 验证未知错误、带敏感 canary 的错误/URL/headers/body、预算溢出、并发和中断；未知项不误归类，错误字符串不泄漏，观察失败不放宽判据或取消应有 cleanup。
5. Windows专项测试/vet、自检及整链绑定；Linux专项测试/race/vet；CLI golden、原正式 runner 回归与产品依赖闭包，六平台普通产品构建。macOS构建不代表原生认证验收。
6. 完成新包和脱敏报告后交付一次私有终端步骤；真实运行必须另行授权。任何依赖原始日志、改变选键/TLS/会话/cleanup、增加证书来源或系统网络调用的缺口，集中报告，不能隐式扩大范围。

2026-10-30 确认前的源码审阅和合成接入实验已保存在 `build/auth-phase-design-20261030/`；实验不是上线实现，也不是新包就绪证明。

### 固定封装与字段

新 launcher 外层为 schema_version 2、tool_kind `auth-dev-observed`，分别输出 source_digest、product_sha256、observed_cli_sha256；binding/state/live_attempted 和原 result 含义保持。result 仍严格使用 P2 schema 1，不能新增或重解释旧字段。新 observations 包含 state（not_started/complete/incomplete）、reason（none/overflow/child_invalid/interrupted）、commands。非 password 模式保持 not_started 和空数组。

子进程只接受固定的内部调用形式及四种命令 profile_show/status/login/logout；独立封装为 schema 1 / `auth-observed-cli`，字段只有 command、exit_code、product、observation。product 是内存中的原 CLI JSON，只供父进程沿用原有投影器处理；父进程最终报告绝不转储它。命令、退出码、封装与实际进程退出码必须一致；缺失、未知字段、重复字段、版本混配和尾随数据拒绝。

每个 observation 只有 command、state（complete/incomplete）、sdk_code、sdk_category、requests；sdk_code 使用现有公共错误码或空串，sdk_category 固定为 not_called/success/unknown/cas_request/cas_http_status/activation_request/activation_http_status/credential_provider/encryption/status_identity/canceled/timeout。非白名单 SDK 消息为 unknown，不推导额外阶段。每条请求只有 role、status_code（0 表示 unknown）、response_received、transport；不提供自由文本。中断/缺失的子进程观察标记 incomplete，不编造记录；观察协议故障使该命令按 runner fault 处理，仍由既有状态机决定 owned cleanup。

当原状态机 pass 但 observations incomplete 时，保留真实的原状态机结果和观察缺口，launcher 返回非零（11），不能将该组合发布为完整诊断通过。观察缺口本身不触发额外网络操作。

<a id="p3-anonymous-context-proposal"></a>

## 2026-10-30 已确认：P3 匿名提交上下文

维护者已确认以下完整范围，决策见 [ADR-0019](../architecture/decisions/ADR-0019-anonymous-submit-context.md)。依据为 [提交上下文离线核对](../research/cas-submit-context-2026-10-30.md)：本轮 POST 500 已明确，但现有观察不提供 form/键关系；合成会话和字段检查没有证明产品缺陷。本次批准离线实现、回归和打包，不修改现有 password 被动观察器，不授权真实运行。

### 判别目标与复用边界

一次回答两个有区分力的问题：当前解析规则的首次命中属于哪种词法/赋值上下文；动态 action 与初始/最终 CAS 页面及 service 的关系如何。不能把“两把键存在”当作退出条件，也不能把静态赋值等同浏览器实际执行。

复用现有匿名 SDK 路径、同轮内存 CAS Cookie、被动 body tee 和始终拒绝的 provider；仅分析 SDK 已完整消费的成功响应。禁止额外获取页面、脚本或资源，不接入 cascollect 的独立请求 planner，不改变当前短路、选键或产品行为。没有获取的后续脚本必须标为未覆盖，不报告“没有其他键”。不新增 SDK 原始材料回调，不在公共 Observer 中传递材料。

成功响应、最终页面与脚本的角色须以 HTTP 状态和 SDK 阶段对照建立；中间重定向、非 2xx、未完整消费、无法可靠分类的材料不提升为“已接受来源”。无法确定时输出 unknown。分析在原有消费过程中旁路进行，Read 的 n/error、Close、重定向和网络轨迹保持；不预读、不 drain。

### 拟议闭集输出与限定例外

- 独立工具身份 `auth-context-dev`、schema 1；旧 P2/P3 回执和包不变，不新增 password 模式。结果分别报告 SDK 是否抵达 provider、分析覆盖和缺口；不得把 provider 到达称为登录成功或键正确。
- 表单关系只报告：目标表单识别是否唯一、lt/execution 的 missing/empty/unique/ambiguous 类别、action 是否同源 HTTPS、完整目标与初始/最终页面分别 same/different/unknown、service 为 same/absent/different/unknown。不输出 action、字段名列表或值。
- 来源和 occurrence 使用本轮局部标签，沿用现分析器的注释/字符串/模板/代码/HTML/unknown、标识符类别、受支持赋值与 RSA 相等标签；不输出源码、键值或跨轮指纹。
- **新增限定投影**：允许将当前解析函数首次文本匹配与上述来源/occurrence 关联，名称为 `parser_match`，不能命名为 correct_key 或浏览器 selected_key。映射必须通过同一输入及当前 parser 的合成对照；不成立时为 unknown。它只描述当前算法匹配，不是批准产品选择此键。此项是对旧来源提案“不输出 selected_key”的明确区分，须本次确认，不默认为既有授权。
- `use_relation` 始终 unknown，不新增 JS 执行器、跨函数数据流或实际加密使用关系推断；不新增 Cookie 内容/名称/哈希输出，也不扩大 ADR-0018 的 password 观察范围。

### 预算与失败语义

沿用最多 32 次应用 GET、120 秒协作期限及父进程 145 秒硬终止，产品系统 PKI/DNS 与本地拓扑约束保持。零 profile/凭据读取、POST、activation、logout；初始 online 或早期 ticket 停止。未来真实操作仍需新的单次授权，本次确认只授权离线实施与打包。

分析输入沿用每源 1 MiB、累计 4 MiB（包含规范化副本）、最多 17 来源（包括分离 inline）、64 occurrence、100000 token、深度 64、最终报告 64 KiB。每次调用均受上下文期限；超限或输入/角色不确定只发布完整的拒绝/未知分类，不发布截断列表，不隐式增加请求。观察预算耗尽后释放收集材料，已有 SDK 流程仍按原守卫规则停止，不通过分析结果恢复或扩大请求。使用关系未知是固定能力限制，不能包装成分析全覆盖。

### 连续实施与退出条件

1. 先将本节批准边界落实到 ADR 和严格 schema，再实现独立匿名投影；不修改产品认证实现或旧包。
2. 新生成 RSA 合成场景证明 parser_match 对页面/外部脚本、单键/重复/双键/互换、注释/字符串/子串及不支持语法的区分；证明不宣称实际使用关系。测试完整 action 关系、同源重定向、非 2xx 正文、短路后未取脚本和角色不确定。
3. 对照有/无观察的真实 SDK 假传输轨迹、Cookie、body Read/Close/error、provider 次数；预算/取消/超限和脱敏严格解码均有负面回归。不得让不同请求结果被视为等价。
4. Windows 专项测试/vet/自检/绑定、Linux 普通/race/vet、文档及定向脱敏；独立源码快照与包验证完成后交付，Agent 自动执行范围内离线检查。真实运行另行批准，不重新索取开发/PowerShell许可。
5. 输出只能支持上述有限判别；若真实结果仍未知，集中报告不足，不自动加新来源、放宽语法或切换到密码验证，不以必须修复 500 为本工作包退出条件。

### 固定接口与报告细节

独立入口为 `auth-context-launch.exe`，仅接受 `--self-test`、`--preflight`、`--anonymous-once`。payload 精确闭集为 `auth-context.exe`、`build-inputs.json`、`ipgw-meta.exe`；普通产品仅用于绑定同快照身份，不在此工具中执行认证。启动器固定三项哈希并验证目录/文件句柄、运行前后复核；自检只用生成的合成 RSA 和假响应，预检只做本地拓扑检查。未知参数、password 参数、非 TTY 或不匹配确认均不进入真实模式。

新启动器外层 schema 1 / auth-context-dev 包含 source_digest、product_sha256、observer_sha256、binding、state、live_attempted、result。内层同身份 schema 1 包含 mode、state、reason、scope=sdk_consumed_only、release_evidence=false、sdk、form、successful_sources、non_success_responses、analysis、parser_match。sdk 保留已有匿名摘要；新增来源判别以 form/analysis/parser_match 为准，不把历史摘要中的粗分类当作来源接受证明。

只有 SDK 到达拒绝 provider、成功材料的阶段可确定且分析及映射有效时，内层 state=observed；analysis.state 始终 partial，coverage.use_relation 始终 unknown。successful_sources 只计完整成功响应（page/external），不重复计算分离 inline；non_success_responses 计 CAS 非 2xx 响应，它们不进入分析。候选相等性仅适用于已获取集合，不代表页面全部引用或浏览器实际执行。

form 采用 selection、lt_shape、execution_shape、same_origin_https、initial_target、final_target、service 闭集字段；parser_match 只有 state=mapped/unknown、source_id、occurrence，映射引用 analysis 内 occurrence，词法/赋值类别和本轮 key_ref 从该条目读取。不会输出键值、源码偏移、URL 或跨轮指纹。analysis 只投影 state/reason/coverage/sources/occurrences/keys/relations，不继承旧 prototype 的模式和零操作字段。

SDK 未到达 provider、读取不完整、来源角色不确定、超限、分析拒绝或映射失败时 state=blocked，form 全部 unknown、analysis=null、parser_match=unknown、successful_sources=0，不发布截断部分。保留 sdk 摘要及有界非成功响应计数，便于区分产品停止与观察缺口。非匿名模式所有材料字段均为空/unknown。

4 MiB 累计材料预算为保守预留：页面按原文、遮罩副本、分离 inline 最多三倍预留，外部脚本一次计入；因此部分输入可在实际规范化大小达到上限之前被拒绝。保留每源 1 MiB 等其余上限，原 P2 body tee 的独立既有上限不变。拒绝后清除新观察副本，仍保留原始 Read/Close/error 和 SDK 原有停止逻辑。原始材料只在内存处理，不声称 Go 运行时提供可验证的内存擦除保证。

严格解码拒绝未知/缺失/重复字段、版本混配、无分析的映射、无凭据边界的 observed、请求计数或来源引用矛盾；旧 schema 1/2 的其他 tool_kind 不作为本工具回执接受。

<a id="p3-static-syntax-proposal"></a>

## 已确认：P3 有界静态语法覆盖工作包

维护者已于 2026-10-30 确认本工作包，见 [ADR-0020](../architecture/decisions/ADR-0020-bounded-static-cas-syntax.md)。旧 ADR-0019 和历史工具保留。依据为[语法边界合成核对](../research/cas-submit-context-2026-10-30.md)：当前回执的 unknown 可由多种不同输入产生，重复同包或只增加降级原因标签仍不足以判断 parser 命中的上下文。

已批准以下完整离线工作包，先可行性和规范、后实现与验收，不逐个语法族申请校园复测：

1. 评估成熟的纯静态 JavaScript 解析器，核验官方语法说明、固定版本、许可证、依赖闭包、输入/节点/深度限制、取消与最坏资源表现；不下载或执行校园材料，不运行 JS、浏览器、eval 或动态 import。不能证明有界执行时停止并报告，不通过加大既有预算掩盖问题。
2. 先确定对应规范和 ADR；词法上下文、AST 语法位置、赋值候选和运行时使用必须分开。函数或控制流中的赋值只表示语法出现，不证明语句执行、最终值或服务端正确键。实际使用关系保持 unknown；产品正则选键、SDK、公共 JSON 和认证行为不变。
3. 在已有 SDK 消费的成功来源内，针对函数/控制流、正则与除法、Unicode/转义标识符、注释/字符串及模板插值形成完整合成矩阵，优先解决匹配上下文分类。可证明的字面量候选及本轮相等关系才输出；动态值保持 unknown，不做跨函数执行或数据流推断。
4. 新报告独立版本，旧回执不回填；只输出固定语法类别和本轮局部引用，不输出源码、名称、偏移、URL、键值或新材料指纹。具体枚举和矛盾状态规则先在规范中锁定，不用自由文本扩展。
5. 保持既有来源、每源/累计字节、token/节点遍历、深度、报告大小及时间预算；解析器内部不能受协作取消约束时，必须先证明进程硬终止和资源保护可成立。解析错误、超限、取消及未支持形态保留明确缺口，不能将 AST 成功写成浏览器行为证明。
6. 先做合成独立对照及负面回归，再做 Windows/Linux 测试、race/vet、绑定/自检和定向脱敏。若可行且契约闭合，交付独立开发包；如果仍不足以区分认证假设，交付缺口和停止理由，不为了打包而打包。

本提案的确认只授权上述离线设计、实现和交付。任何新真实匿名窗口仍单独授权；不升级为密码窗口、不增加脚本获取、不恢复 CTL 主线。若需要实际使用关系、新材料来源或更大的预算，停止并集中报告，不能作为实施细节自动加入。

### 静态观察固定边界

新身份为 auth-static-context-dev，schema 2，独立包/命令，不替换旧 auth-context。固定字段为 schema_version/tool_kind/mode/state/reason/release_evidence/context/syntax；context 保留原匿名上下文 schema 1 原义，syntax 是独立静态事实，不把原 lexical unknown 回填为旧版已知。非匿名模式没有真实材料。

syntax 固定字段：state（observed/rejected）、reason（none/invalid_input/syntax_rejected/token_limit/depth_limit/work_limit/input_limit/canceled/parser_failure/mapping_failed/protection_unavailable）、use_relation=unknown、sources、occurrences、keys、parser_match。source 只含本轮 id/kind/ordinal；occurrence 只含 source_id/ordinal/context（code/comment/string/template/regexp/html/unknown）、identifier_class（public_key/public_key_str/other/unknown）、assignment_class（declaration/bare/property/none/unknown）、placement（top_level/function/control/other/unknown）、key_class 与 key_ref（沿用已有 RSA 闭集）。parser_match 沿用本轮引用结构；不回显偏移、名称或材料。rejected 清空所有内容数组，匹配 unknown，不保留截断前缀；未知静态分类不推定运行时语义。

每源 1 MiB、累计材料 4 MiB、最多 17 来源、64 occurrence、累计扫描 token/注释不超过 100000；解析递归和 AST 遍历分别最多 64 层，累计 AST 访问不超过 100000。重读和 lookahead 计入扫描预算，可能比原解析器更早拒绝；累计字符读取步数新增更严格上限 16 Mi，不是输入额度增加。先解析再遍历时，必须在解析过程中实施递归/token/工作检查，不只验收最终 AST。未知点不能通过删改输入恢复。

placement 按保守语法容器划分：函数内为 function，控制语句/条件表达式/逻辑短路及可选链为 control（函数内仍优先 function），类体/字段和未归类块为 other；top_level 不表示该语句实际执行。词法区间以扫描时原始位置为准，不能用规范化模板文本长度或解码标识符长度回推；regexp 的区间须截止真实 closing slash/flags，不包含后继 token。此观察不验证 regexp pattern/flags 的完整合法性，也不证明严格模式或浏览器特定环境接受整个脚本；regexp 分类仅描述词法外壳。未闭合 regexp 仍拒绝。

真实观察仍最多 32 应用 GET、120 秒，父进程最多 145 秒；静态分析在既有总期限内运行。专用分析阶段再限制最多 10 秒，Windows 进程 256 MiB Job 内存限额在材料分析前建立；保护不可建立即停止，不能降级。合成测试可直接调用纯 Go 分析层，但不可将其等同于真实入口资源保护验收。所有原始材料只在内存；不把 Go 内存清除宣称为可验证擦除。

新入口 `auth-static-context-launch.exe` 只接受 --self-test/--preflight/--anonymous-once；payload 精确为 auth-static-context.exe、build-inputs.json、ipgw-meta.exe。外层 schema 2、tool_kind auth-static-context-dev，仍绑定 source_digest/product_sha256/observer_sha256；state/binding/live_attempted/result 含义沿用独立启动器约定，不接受旧工具身份冒充。新单次确认词为 RUN-DEV-STATIC-CONTEXT-ONCE，旧确认词不通用。

内层 context 保留完整旧报告及其原先全部拒绝条件；只有旧观察达到凭据边界且材料完整时才执行新增纯静态分析。syntax 若被拒绝，外层内层结果 state=blocked/reason=static_rejected，同时保留旧 context 的真实观测，不能混成旧观察失败或成功登录。预检和自检的 syntax=null；真实模式 syntax 存在时必须与 context 的来源序列及 parser 来源引用一致。产品命令始终为零。此组合可能在新增 parser 可接受而旧观察拒绝的输入上保守停止，不为追求覆盖解除旧守卫。


<a id="p3-key-use-relation"></a>

## P3 登录使用关系修复后的开发验证

2026-10-30 起新构建按 [ADR-0021](../architecture/decisions/ADR-0021-cas-key-use-relation.md) 动态选键：完整收集允许来源，接受有限静态登录关系及唯一 RSA 身份。不同键、失败来源、未支持语法和预算不足都在凭据提供器前停止；不宣称静态关系等于浏览器实际执行。

当前交付主线为普通产品与独立 observed CLI 的绑定 password_core 开发包。保留 self-test/preflight；新构建中的 auth-dev、auth-context、auth-static-context 和 cas-acquire 匿名真实入口均启动即拒绝。旧包、旧报告与下文日期明确的匿名契约保持历史意义；正则对照仅在诊断包内使用，不再代表当前产品选择。历史分类测试与 SDK 取键分离，不用旧报告解释新算法。

离线入口 `python scripts/dev.py key-relation --context tool` / `python3 scripts/dev.py key-relation --context tool` 执行对应平台专项验证；Linux 增加 race 与十秒有界 fuzz。现有 auth-dev 入口生成独立新快照和绑定包，不能替换或标记为既有 Candidate。

2026-10-30 离线交付时维护者报告 TUN 开启、账号在线，用于远程管理；这是当时状态，不代表当前现场。此后单次测试结果见[2026-10-31 摘要](../upgrade/status.md#status-20261031-password-progress)。下一真实 password_core 仍需新的单次授权与现场条件确认；密码只在维护者私有终端输入。不能自动注销以准备测试。

## P3 注销callback候选修复的开发交付

2026-10-31 维护者批准 [ADR-0024](../architecture/decisions/ADR-0024-logout-explicit-callback.md)。新构建的普通产品与独立observed CLI共同使用仅新增callback的共享注销请求；原结构投影、私有prompt、状态机和cleanup规则不变。两者仍是分别绑定的不同EXE，不能混用旧包回执解释新包。

离线验收使用Windows/Linux `scripts/dev.py auth-dev`、纯解析回归、Linux race/vet及既有有界fuzz，六平台 `cross` 检查产品依赖隔离；Windows入口生成新源码快照并执行整链self-test和构建审计。本工作包只运行零网络自检，不读真实profile；真实password_core随后独立授权。新包交付状态见当前status，不把提案批准写成测试通过。Windows原生安装/junction旧缺口保持。

<a id="p3-logout-structure-proposal"></a>

## 已确认：P3 注销响应结构观察

维护者已于 2026-10-31 批准 [ADR-0022](../architecture/decisions/ADR-0022-development-logout-structure.md) 的完整实施和离线交付。产品仍要求业务成功与最终 offline 同时成立；以下仅为独立观察入口对 ADR-0018 的限定扩展，包交付前不得称已可用。

### 已知事实与待区分问题

首次修复包开发窗口已登录成功，但整套 password_core 因第一次注销响应处理失败而失败。`client.request` 的 256 KiB 上限和 `srun.ParseLogoutSuccess` 的封装/字段检查都可产生 protocol_changed；现有 HTTP 状态投影不能区分二者。清理 Logout 实际只查询状态，不发送第二次 logout。合成回归已经证明体积拒绝和业务解析拒绝均可产生同样轨迹，不能据此猜测真实正文。

建议一次覆盖读取完成性、体积、封装、字段结构及当前 parser 结果，不修改产品响应接受条件，不预设服务端返回某个字段或值。

### 限定材料与语义保持

- 只在独立 observed CLI 的 logout 命令、固定 HTTPS 网关 logout GET 响应上接入；不处理 status、discovery、CAS、activation 或其他响应，不读请求正文、header、query。精确角色无法确认时不采集。
- 仅随 SDK 已发生的 Body.Read 复制返回的字节；不得预读、drain、重试或主动补读。保持每次 Read 的 n/error、Close 返回值、请求参数、Cookie、系统 TLS 和状态机顺序。分析不影响 SDK 的结果、错误或 cleanup 决策。
- 最多保存 256 KiB；看到第 256 KiB + 1 字节时只记录超限并清除材料，不分析截断正文。非 EOF 错误、未完整消费、取消或提前 Close 均不发布字段结论。只读到上限而无法证明 EOF 时保持未完整。
- 每条命令最多两份符合条件的响应、累计复制不超过 512 KiB；超限立即停止该命令的新正文观察，释放材料并报告覆盖缺口，SDK 仍按原流程执行。默认产品仅发一份注销响应，额外额度不授权新请求。
- 分析在 SDK 原响应关闭后同步执行；不得持锁做长分析或遗弃后台任务。每份最多 32768 个 token、深度 32、1000000 扫描步，最长一秒协作期限并服从更短的父 context。预算检查须在解析过程中成立；不能只在解析完成后计数。资源保护不能证明时交付缺口，不扩大预算。
- 原始材料只留内存，完成或拒绝后清除自有缓冲；不声称 Go 运行时提供全部副本的可验证擦除。禁止写正文、字段名列表、字符串值、数值、长度、回调名、URL、摘要或跨轮指纹到报告/日志/仓库。

### 闭集投影与版本

新增独立开发回执 schema 3（auth-dev-observed），观察子进程 schema 2；产品 JSON 不变。旧外层 schema 2 / 子进程 schema 1 仍按原严格解码规则读取，不补造字段。拒绝未知、缺失、重复字段、版本混配和矛盾状态。

观察结果只允许以下固定类别；失败不发布截断的部分字段：

| 项 | 拟议分类及含义 |
|---|---|
| 读取 | not_observed / complete / too_large / read_error / incomplete / budget_exceeded / canceled；无真实 logout 响应时为 not_observed |
| 封装 | unknown / empty / json_object / jsonp_object / other / malformed；完整边界校验，不能仅凭前缀接受 JSONP |
| 结构拒绝 | none / duplicate_key / malformed / depth_limit / token_limit / work_limit / canceled；重复键包含转义后相等及嵌套重复 |
| BOM | unknown / absent / present；仅观察，不能删除 BOM 后冒充产品接受 |
| 固定字段 | 仅 ecode、success、code、message、error、error_msg、res 七个固定槽位；每槽为 unknown / missing / null / boolean / number / string / object / array，不输出实际名称集合或值 |
| ecode 细分 | unknown / missing / integer_zero / integer_nonzero / noninteger_number / integer_out_of_range / string_zero / string_other / other_type；整数规则与当前 parser 的十进制 int64 规则对照，不将字符串零改成数值零 |
| success 细分 | unknown / missing / true / false / other_type；只用于说明当前既有一致性检查，不新建成功判据 |
| 其他字段 | unknown / absent / present；不输出数量或名称，不推断业务含义 |
| 当前 parser | unknown / accepted_success / accepted_rejection / rejected；仅在有界结构完整且调用既有纯解析函数安全时计算，结果不控制 SDK |

七个固定槽位只用于区别表示，非 ecode/success 的字段不匹配字符串字面量、不解释其成功含义。malformed、重复字段、预算不足、非完整读取时各字段均为 unknown；封装 other 只说明不属于支持的对象封装，不猜 HTML 或错误页面内容。已完整分析的合法结构才可以报告字段 missing。BOM 和封装均为观察事实，不修改原文来推定产品接受。

同时新增固定 SDK 分类，用既有 Code/Message 精确区分 logout_response_limit、logout_response_parse、logout_business_rejected；其他保持 unknown。原错误文本及 Cause 不输出。读取类别与 SDK 分类独立，不能以一个推造另一个。只有实际 logout 命令可以携带新增结构投影；旧登录、status 等观察能力保持。

### 连续离线交付与停止条件

批准后先锁定严格 schema，再实现独立旁路：合成矩阵覆盖 JSON/JSONP、空/非对象/BOM、重复键/尾随/截断、所有字段类型与 ecode 表示、success 冲突、大小边界、分块 Read、n 与 EOF/错误并存、取消/提前关闭、节点/深度/时间预算。捕获的材料不用于重新执行 Logout。

用相同假传输比较有/无观察的 SDK 返回、请求序列、参数、Read/Close、错误、cleanup 和私有 provider 次数；保留非 logout 材料一旦被读取便失败的回归。完整状态机覆盖业务接受、明确拒绝、响应解析失败但服务端已 offline、cleanup 仍 online 等不同情形，不放宽测试期望。

完成 Windows 聚焦/普通测试与 vet、Linux 测试/race/vet 和有界 fuzz、启动器严格版本/绑定/self-test、产品依赖隔离及六平台交叉构建、文档及定向脱敏检查。Windows 既有安装/junction 缺口单列。新开发包绑定普通产品和独立 observed CLI 两个哈希，保留旧包及回执，不创建 Candidate。

只执行零网络自检；真实 password_core 仍需新单次授权和私有终端密码。该提案不授权网络操作、不读取当前 profile。未知结构、无法解释的新字段或产品业务规则变化集中报告再决策，不保存原始正文来绕过限制。下一包有信息增益才申请复测，不重复同包同结果。


### 注销结构观察已实现的报告布局与离线入口

外层 schema 3 / auth-dev-observed 的 `observations.commands` 中，仅 logout 命令增加非空 `logout` 对象；其他命令省略该成员。对象包含 state（not_observed / observed / budget_exceeded / incomplete）和 responses；每份响应包含 request_index、read、envelope、structure、bom、fields、ecode、success、other_fields、parser。request_index 只引用本命令已有 requests，按请求发起顺序排序；不代表额外请求。最多两份，超限清空整组。未发生合格响应可为 not_observed；材料缺失不能冒充字段 missing。

fields 精确包含前述七个固定槽位。非完整读取、结构拒绝或预算不足时全部为 unknown；解析预算不足使用 budget_exceeded 和具体闭集 structure 分类。state=incomplete 只能与父命令 incomplete 同时出现且没有响应列表。旧外层 v2 不允许新增成员/类别；新 v3 的 logout 命令必须带此对象，显式 null、版本混配、索引越界和矛盾 parser 结论均拒绝。实际运行 Adapter 只接受新子进程 schema 2，历史 schema 1 仅用于回执读取。

固定 SDK 的体积错误只有在该命令最后已观察请求为 logout 时才归类 logout_response_limit；若初始 status 超限仍为 unknown，不能误标为注销响应。读取与 SDK 分类分别记录，不相互推造。

离线聚焦入口为 `python scripts/dev.py logout-structure --context tool`（Windows）和 `python3 scripts/dev.py logout-structure --context tool`（Linux；包含 race 和十秒有界 fuzz）。`auth-dev` 完整入口继续执行状态机、严格子协议、绑定与自检；Windows 生成独立新包。报告快照与交付状态见[当前状态](../upgrade/status.md#status-20261031-logout-observation)。本次交付没有新的真实测试授权。
