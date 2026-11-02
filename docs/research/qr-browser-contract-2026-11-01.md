---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
snapshot_date: 2026-11-01
---

# QR 失败解释与网页契约核对

执行依据为 [ADR-0027](../architecture/decisions/ADR-0027-browser-qr-contract.md)。当前结论：离线核对与窗口方案已完成，真实失败阶段和网页契约仍缺证据，尚无可证明的产品修复。本次未校园导航、扫码、查询账号状态或注销；未运行外部实现、改产品、生成新包或提交。

## 历史失败解释修正

保留 `EVID-20261101-002` 原始 fail 回执及 [P5 记录](p5-readiness-2026-11-01.md#有线正式-terminal_qr-失败回执)。已读既有脱敏核验副本：初始 offline 成功；qr_login_logged_in 返回 exit 6/protocol_changed，166 秒；无 cleanup，最终会话 unverified。维护者报告二维码显示、扫码并确认。

`internal/livegate/runner.go` 的 executeQRPair 将 QR 子进程与紧随的 status 作为一对：QR 非零退出直接结束；QR 成功后 status 失败则将其结果、错误码和耗时合并到 QR 步骤，不保留独立 status 步骤。该规则在失败 Candidate source `1a32aabc2aa3125afe638b0dc3d3444bf7a95c7d` 已存在。

因此历史“没有后续状态”只能解释为没有独立 status 步骤/通过证据，不能证明 status 请求没有运行。166 秒也可能是两次调用总时长，不能推定二维码过期；正常过期使用 interaction_required。无身份在线证明、无清理权和无 cleanup 的结论仍成立。两种失败路径可能产生相同摘要，不能从摘要倒推出唯一路径。

现有 `TestStateMachineStopsAtEveryPrimaryFailure`、`TestStateMachineQRProofPairing`、`TestStateMachineQRStatusRunnerFaultDoesNotSynthesizePass` 已覆盖失败合并及拒绝伪通过。这是定位信息不足，不是当前已证明的成功判据缺陷；不改旧黄金回执来增加历史上未观察的信息。

## 源码—观察—验证矩阵

本地基线：`codex/bhk-v1-acceptance` / HEAD `c21094ab3e7b2a4064b66ae82cf2fbc2173f92cd` 加现有 Wi-Fi 未提交改动；已有普通开发包来源另按窗口约定绑定。

| 阶段 | 当前 SDK/工具行为 | 已有网页材料或真实事实 | 仍需区分的内容 |
|---|---|---|---|
| 创建 | 页面 QRUUID 存在则复用；广告支持但无 UUID 时本地随机生成；二维码携带 UUID 和本轮 service | 本地桌面 login_neu.js 调用 loginQRCode 与 generateLoginQRCode；曾显示可扫码二维码 | 构造器定义、UUID 来源、service 编码和初始化会话 |
| 轮询 | 同轮 CAS 会话 GET checkQRCodeScan，参数 uuid/random；默认间隔 3 秒、生命周期 180 秒 | 维护者曾扫码确认，未提供轮询结构 | 方法/参数键/封装；等待、已扫描、确认、拒绝、过期分别如何表示；不能按 166 秒猜测 |
| 解析 | 精确空体为 pending；严格 JSON/JSONP 对象；redirect 别名可单独确认，不强制 status；已识别消息必须一致，未知/冲突拒绝 | 缺少 loginQRCode 定义与本轮响应投影 | 空白体、空对象、其他字段/类型是否具有官方可解释语义；不能因第三方宽松接受就放宽 |
| 确认 | service ticket 跳转验证后激活；否则解析同源 HTTPS CAS 确认地址，保留同轮 Cookie Jar，再截获 ticket/挑战 | 已提供调用方将 result.redirect_url 赋给 window.location.href | 返回绝对/相对地址、初始 service 的关系、重定向链和本轮会话延续；不回传其值 |
| 激活 | HTTPS 网关 activation，严格业务响应与目标身份验证 | 密码路径已通过，不能移作 QR 路径证明 | QR 是否到达 activation、业务是否成功、SDK 最终 status 是否失败 |
| 工具复核 | QR 人类 TTY 命令后独立 JSON status，但失败可能合并 | 原失败摘要只含合成步骤 | 未来分别记录每个命令的退出与分类，不读取二维码输出来定位 |

本轮只查看维护者已提供的 `login_neu.js` 相关调用片段，不复制脚本或公钥；桌面已提供脚本集合中未见 loginQRCode 的独立定义文件。该缺口不等于官方网页没有该资源。没有打开浏览器或请求缺失资源。

## 固定版本 QR 外部实现

2026-11-01 只读查询并按完整提交读取相关源码；均未运行。以下实现目标不完全相同，不能当作当前 IPGW 协议证明。

| 实现与固定提交 | 确认的行为 | 可比较点与不能直接采纳项 |
|---|---|---|
| [neucn/neust](https://github.com/neucn/neust/blob/3e9f8c0006277a7b3e8a0a89d9fa8d72302c80cd/src/auth/wechat.rs#L81) | 同一 UUID 生成授权与轮询地址；空正文表示未授权，非空后以同一 Session 检查 CAS 状态 | 对照 UUID 与会话衔接；[后续检查](https://github.com/neucn/neust/blob/3e9f8c0006277a7b3e8a0a89d9fa8d72302c80cd/src/session.rs#L135)解析 CAS 页面/Cookie，非空不是网关成功依据 |
| [CreamPig233/neu_wisedu2wakeup](https://github.com/CreamPig233/neu_wisedu2wakeup/blob/8ce265cac63db61ed7b4460ea5f0311493cb5ae5/neu_wisedu2wakeup.py#L91) | 本地 UUID 显示 QR；维护者按回车后轮询，再用同一 session 访问教务 CAS service | 目标为教务/WebVPN；未严格解析轮询，不能移植为成功规则，也不复制其输出授权链接的做法 |
| [RekaYOO/NEU-JWXT-Toolkit](https://github.com/RekaYOO/NEU-JWXT-Toolkit/blob/c46d7856aa146e0c9b59c7120ad51b415bd9f1c6/backend/core/auth/client.py#L711) | 初始化 CAS 会话，生成带 UUID/service 的 QR；空白体 pending；JSON object 缺 redirect_url 也 pending；有字段则检查域并同 session 跟随 callback | WebVPN 目标；对照会话、参数类型、字段和回调，不直接采用缺字段默认 pending、宽松跳转或 Cookie 元数据输出 |

原三个参考项目的已检查入口主要是密码认证：neucn/ipgw、Neboer/ipgw-py-manager、7aaauTools/neu-login，版本见[前次记录](ac-id-comparison-2026-11-01.md)。扩展查询 neucn/neugo、weNEU/neu-eone、DoraTiger/NEU_IPGW 的相关入口也未提供直接 QR 轮询依据；chiyuki0325/NEU-IPGW 读取失败，只记不可核对，不推定无支持。

[官方扫码教程](https://pass.neu.edu.cn/portal/help/smdl.html)的公开搜索索引说明微信企业号绑定与关注前提，但索引并非本轮新登录页面；不能证明当前字段、期限或手机请求。维护者此前已扫码确认，不能据教程倒推其未绑定。桌面 DevTools 仅覆盖该浏览器会话，手机确认链仍未观察，不安排手机抓包。

## 已准备的后续窗口与缺口

完整步骤、材料清单、普通产品身份和停止条件见[QR 浏览器窗口](../operations/browser-comparison.md#browser-qr-comparison)。采用无痕会话、一次网页 QR 登录、三次普通产品 status，目标身份在线证明后才一次产品 logout；不网页注销，不重试。每步由维护者私有终端/浏览器完成，Agent 等待脱敏结果。

| 缺口 | 下一次能够区分它的材料 | 当前处理 |
|---|---|---|
| 缺官方 QR 构造器/轮询源码 | 经授权加载的 Sources 中相关函数和无秘密常量 | blocked，先取网页契约，不改 parser |
| 原错误阶段不可唯一恢复 | 网页对照及后续必要时独立命令结果 | 保留原 fail，不补造历史步骤 |
| 当前 QR 轮询/确认格式未知 | 未扫描至确认过程的脱敏结构与跳转关系 | 不生成猜测格式的修复包 |
| 额外开发观察尚未批准 | 若上述仍不足，集中说明所需阶段/投影与预算 | 不直接复用 password observed CLI |

现有 password 观察器要求 JSON、限制 QR，32 项请求预算与每命令期限也不适配正常 QR 轮询。QR 私有 TTY 不能通过捕获/转录原始终端输出来替代。后续若真有必要，再单独设计有界阶段汇总通道。

新增纯合成回归 `TestStateMachineQRProtocolFailureSummaryIsNotCommandTrace`：分别执行 QR 自身失败，以及 QR 成功后 status 失败，两条路径均产生相同的 protocol_changed / 166 秒步骤摘要。测试同时核对各自实际命令序列，不新增或改变历史报告字段。

2026-11-01 离线检查使用固定 Go 1.25.0：Windows 的 QR 状态机聚焦测试、livegate vet 和文档索引生成通过；Linux 同组测试、vet、race 通过。原始命令与退出码保存在忽略目录 `build/qr-browser-preparation-20261101/windows-20261101T123105Z-a320c464/result.json` 和 `linux-20261101T123109Z-841ed220/result.json`。测试使用假执行器，不运行产品认证或校园请求。

本次没有产品改动，因此不重跑全仓测试或六平台构建、不打包；这些检查在有证据支持的修复实施后执行。此次验收包括文档/链接/定向脱敏检查及既有和新增 QR 合成证明检查。完整 QR 修复仍未完成，Wi-Fi 开发通过、旧 Candidate 与历史回执全部保留，M3 不变。

## 2026-11-02 已授权网页窗口与源码补证

本节追加后续事实，前文未授权/缺少脚本为当时快照。维护者授权一次窗口并确认条件，使用约定普通开发产品（哈希复核一致）及显式有线绑定。初始 status 为 offline、exit 0；维护者确认无痕、完成一次网页扫码，随后普通 status 确认目标账号 online（未回传该次退出码）；普通 logout 为 logged_out、exit 0；最后 status 为 offline、exit 0。未通过网页注销。该结果是网页与普通产品对照，不是 SDK QR 或正式 Candidate 通过证据。窗口完整耗时、浏览器与产品同路径以及是否经过自有路由器未独立证明；不追加请求来补造证明。

既有 Network 记录由维护者查看：确认前轮询在 Response 面板显示空白，不能区分零字节与空白字符；确认后的 GET 返回 HTTP 200，JSON 对象只有 redirect_url 字符串，指向 HTTP 网关 service 接续路径，含 ac_id 而不含 ticket。不保存原始响应、认证 URL、UUID 或地址。现有请求 Initiator 指向 login-qrcode.js 的轮询代码。

维护者提供桌面 login-qrcode.js，SHA-256 为 `345011f32d85302782c989434ed17700410dc93e60eb99a69c514e10ef74d9ba`。仅只读分析，不复制进仓库、不执行。脚本本地生成 UUID，二维码以当前页面 service 参数构造；每 3000ms 发起一次 JSON AJAX 轮询。success 回调停止计时并调用上层回调，error/complete 回调为空；结合此前 login_neu.js，上层直接导航 result.redirect_url。该文件没有校验目标身份或证明网关激活成功的逻辑。第 60 次 tick 清除计时器并显示过期，但该次仍会执行后面的 AJAX；不能把此实现改述为严格 180 秒内结束所有请求，也不照搬宽松错误处理。

当前已证明的兼容差异：SDK ParsePoll 接受仅含非空 redirect_url 的 JSON；随后 finishQRConfirmation 对相同 service 目标调用 ticketFromRedirect/validatedTicket，要求精确的 ticket 与 ac_id，因此上述无 ticket 接续表示会被拒绝。若目标并非所构造 service，也不能通过同源 HTTPS CAS 后备规则。问题不应简单归因为 HTTP scheme：产品本来就能识别匹配的 HTTP service 地址作为 ticket 载体，并通过 HTTPS activation 发送 ticket；没有理由打开 HTTP 凭据/会话传输。

该静态差异尚不能唯一回溯 EVID-20261101-002：原 SDK 轮询正文未保存，runner 仍可能合并后续 status 失败。下一缺口为同轮浏览器接续的实际请求关系（HTTP 是否实际发送或本地升级、是否另经 CAS/ticket、门户如何确认状态），仅查看已有记录。无 ticket 不代表无需激活，也不代表可以仅凭跳转宣称登录成功；在明确接续语义前不修改产品或生成修复包。

后续已有记录补证：无 ticket 的 HTTP service 项显示 307、Non-Authoritative-Reason 为 HttpsUpgrades，目标为 HTTPS；随后 HTTPS document 返回 200、text/html。该记录支持浏览器升级，不能称为服务器收到 HTTP 后重定向。HTTPS 项的 Remote Address 为本机代理，不能把 TUN 关闭解释为浏览器直连，也不宣称与显式绑定的 SDK 同路。请求链可见无 ticket 的门户 API、后续 CAS 导航及带 ticket 的门户接续；仅凭列表尚不能证明中间 API 的字段或会话因果。

维护者提供本地名 reditect.js（浏览器名 redirect.js），SHA-256 `396c9c6faaf780d9facb5fc26677ae8c3cc558f891aaf55890c9d416d7556611`。静态逻辑为设备类型识别和 pc/phone、sms、otp、success 页面配对转换；sso 不在配对表中，不能把它认作当前 CAS 接续实现。已有 Portal.js（SHA-256 `0130f589e9b31e576dc4883f0e38c6d3fb661e3f0e69739a39c778ad4491b277`）才包含 _checkSSO：sso 页面通过 ajax.get 请求固定 /v1/srun_portal_sso 加当前查询；success 分支读取 res.Redirect，处理 service 并拼入本轮 acid 后导航，未提供 Redirect 则导航 success 页面。这里的 res 是 ajax 包装器传给回调的对象，不能未经原始响应结构和包装器核对就认为它是响应顶层对象。下一最小材料为已有无 ticket XHR 的脱敏字段结构及 Redirect 目标关系；不重放请求，不复制脚本、会话材料或响应全文。

该缺口随后由维护者提供的既有 XHR 记录补齐：HTTPS GET /v1/srun_portal_sso，HTTP 200、application/json；顶层 code 为数值 0，message 为字符串，Redirect 为指向 HTTPS CAS login 且含 service 参数的字符串，ID 为空字符串。本条嵌入的 service 为网关 service 路径且尚无 ac_id；已有 _checkSSO 源码随后补入本轮 acid。消息中的 301 字样不是实际 HTTP 状态，code 0 在此处也不等于账号已登录。Initiator 对应 main.js → Portal.info → _checkSSO → _request。包装器直接传入解析后的 res：code 已定义、error 未定义且 code 严格等于数值 0 时调用 success，不拆包或改写 Redirect。因此当前材料支持“网关返回 CAS 接续指示，再由页面导航”，不支持“轮询确认即可完成网关认证”。仍不读取 Cookie 值或声称已证明手机与桌面的全部会话关系。

最后一项既有记录补证：上述后续 HTTPS CAS login 为 GET、HTTP 302；Location 指向 HTTP 网关 service，具有 ac_id 和 ticket 参数。只记录参数存在和目标关系，不保存值或完整 URL。至此浏览器所观察链路为：轮询确认 → 无 ticket service → 浏览器升级 HTTPS → 门户 HTML/JS → 无 ticket HTTPS 网关 API → JSON Redirect → 再次 HTTPS CAS login → 302 service ticket → 后续门户接续；普通产品独立目标 online 核对、注销及最终 offline 的结果另列如上。不能将 HTTP Location 的出现等同于 ticket 已明文发送；产品应沿用截获 ticket、校验 service/ac_id、通过 HTTPS activation 和最终身份核对的边界。带 ticket 那条浏览器升级记录尚未单独检查，不补造其升级原因。

本次采集结束，不要求新的网页扫码或 SDK 重试。材料足以进入离线接续方案和合成复现，尚不足以宣称旧正式失败根因唯一确定或同轮 Cookie 的因果已证明。后续先在协议/ADR 中明确无 ticket service 的精确接受范围、同轮 CAS 接续策略、跳转和请求预算、拒绝/挑战及最终身份判据，再实现最小修复；不执行网页 JS，不任意跟随响应地址，不采用 HTTP 传送凭据或 ticket，不凭 code 0/页面成功判定认证完成。
