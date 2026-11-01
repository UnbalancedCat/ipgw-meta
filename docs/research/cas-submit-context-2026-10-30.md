---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# P3 CAS 提交上下文离线核对

本记录承接 [P1 行为矩阵](auth-baseline-2026-10-29.md) 和 [当前状态](../upgrade/status.md) 的 P3 私有观察窗口；路线仍以 [roadmap](../upgrade/plan.md#roadmap-20261029) 为准，不另建路线。真实回执证明 CAS POST 返回 500，未提供本轮 form、Cookie 或键关系。以下是代码审阅和合成证据，不是对真实响应的分析。

## 核对结果

| 环节 | 代码及独立检查 | 结论与限制 |
|---|---|---|
| form 来源与 action | [ParsePage](../../internal/cas/page.go) 优先 loginForm，使用响应页 URL 解析 action；[新增测试](../../cas_submit_context_test.go) 在前置诱饵 form 和未知 input 存在时，检查目标表单状态、相对/绝对/缺省/query-only action、重定向后相对解析。 | 合成目标和 service 保持预期；未证明真实 action 的完整目标等于初始/最终页面。历史同值关系不能代替本轮观察。 |
| CAS 会话 | [passwordLogin](../../auth.go) 显式将原 client.Jar 赋给 postClient；脚本仍用原 client。新增测试由假服务设置、轮换、删除 Cookie，并检查路径限制、错误 Domain 拒绝、默认 Path 及 CAS 重定向 Cookie。 | 当前实现没有在这些场景丢失 CAS GET→脚本→POST 会话。不能据此确认真实 Cookie 的属性、轮换或服务端会话有效。 |
| 每轮隔离 | 同一个 SDK Client 连续进行两轮合成 Login；每轮初始 CAS GET 无前轮 Cookie，动态字段每轮改变。 | 本测试覆盖失败后的新会话和动态状态重新发现；没有执行真实重试。 |
| 凭据与六字段 | 假服务独立核对精确六键、唯一值、同轮 lt/execution、UTF-16 长度，并用本轮生成的私钥解密 RSA，核对 UTF-8 拼接输入。 | 满足 [ADR-0014](../architecture/decisions/ADR-0014-cas-rsa-compatibility-envelope.md)/[ADR-0015](../architecture/decisions/ADR-0015-cas-submit-closed-fields.md) 的合成实例。正确密文、表单和 Cookie 也可收到假服务 500，说明 HTTP 状态本身不足以诊断选键。 |
| 提交前拒绝 | 缺失 execution、空 lt、重复 execution、HTTP action、跨源 action，在真实 SDK 假传输中均在 provider 前停止。已有 sdk_test 保留重复 lt 覆盖。 | 所列错误未进入凭据获取或 POST；没有放宽目标或输入规则。 |
| POST 500 后停止 | 每轮假服务校验提交后返回 500；检查只 POST 一次、稳定 network、无 activation/额外请求，错误不包含合成 canary。 | 与真实失败分类相容，但不模拟真实服务端内部失败原因。 |
| 选键来源 | [page.go](../../internal/cas/page.go) 在整页/脚本文本上取首个正则匹配，[auth.go](../../auth.go) 页面命中或外部脚本命中后短路；既有 cas-observe 和来源测试已有多键/顺序/子串/注释观察。 | 局限已知，本轮不重跑无新输入的观察，也不将某个候选定义为正确键。 |
| 来源分析能力 | [casprovenance](../../internal/casprovenance/types.go) 可报告词法上下文、受限赋值和本轮 RSA N/E 相等标签；[报告校验](../../internal/casprovenance/report.go) 强制使用关系 unknown、Relations 为空。 | 当前不具备浏览器实际加密使用关系分析。仅接通此分析器可能再次得到“两键且未知”，不得声称已解释 JS 执行。 |

本轮未复现需要修改产品实现的缺陷。测试通过只能缩小合成缺口，不能排除真实上下文差异。未修改公共接口、产品代码、选键或凭据处理。

## 验收

- 新增 `TestCASSubmitContextSessionAndForm` 的 9 个场景，各执行两个全新合成会话；`TestCASSubmitContextRejectsBeforeProvider` 的 5 个拒绝场景。所有请求均终止于 fake RoundTripper，未创建真实 transport，密钥运行时生成。
- Windows CAS/根 SDK 普通测试通过：`build/dev-runs/windows/20261030T105227Z-7260082f-focus/result.json`。
- Linux CAS/根 SDK 普通、race、vet 通过：`build/auth-submit-context-20261030/linux-result.json`。
- 这是聚焦检查，未重跑安装器、六平台构建或完整质量门禁，不改变其历史结果及缺口。

## 下一步的判别目标

最小有信息增益的观察应同时回答：当前解析规则命中的文本是注释/字符串/标识符子串，还是受支持的直接赋值；动态 action 完整目标是否等于初始/最终页面，service 是否相同。只观察 SDK 实际获取的材料，不再复制另一套请求流程。

具体拟议边界见 [P3 匿名上下文提案](../operations/development-validation.md#p3-anonymous-context-proposal)。其中“当前解析规则命中”仅是算法匹配投影，不是浏览器实际使用或服务端正确键；复杂语法和未获取来源仍为 unknown。若结果只支持普通直接赋值及相同目标，明确保留真实使用关系缺口，不凭两个键继续猜测或自动扩展采集。

本记录不授权实现提案、新的匿名/密码窗口或扩大材料输出。正式 password_core 仍 blocked / live_unverified，M3 仍 in_progress。

## 后续：匿名窗口后的语法覆盖核对

本节是 2026-10-30 后续离线研究，不覆盖上文原交付时状态。已授权匿名窗口的闭集回执见 `build/auth-context-live-20261030/REPORT.md`：S17 为外部脚本，parser_match 已映射，但该 occurrence 的 context/identifier/assignment 均 unknown，key_class 为 not_checked。原始材料没有保存，不能重放或推断其具体文本。

### 可证明的分类边界

| 路径 | 实现依据 | 对本轮结果的含义 |
|---|---|---|
| 原始匹配与语法独立 | [parser_match.go](../../internal/casprovenance/parser_match.go) 在有界输入上对照当前 SDK 文本匹配，再按来源与偏移映射 occurrence。 | mapped 只说明同一文本命中可定位；语法未知并不阻止映射。 |
| 词法遇阻 | [lexer.go](../../internal/casprovenance/lexer.go) 的 unknown 分支从不确定位置到当前来源末尾标记 unknown 并停止词法扫描。触发包括非注释斜杠、模板插值、Unicode/转义标识符、部分控制字符和行注释中的 Unicode 换行。 | S17 的 context unknown 表明命中落在其自身词法未知后缀内。未知点可以早于命中；不能推定键赋值本身复杂或非法。 |
| 赋值语法遇阻 | 同文件 assignments 仅接受直线字面量赋值；函数、数字表达式等使赋值覆盖 partial，但不会自行把词法 context 改为 unknown。 | 单凭函数包裹并不能解释本轮 context unknown；不能把所有 unsupported_syntax 都归为同一原因。 |
| 来源之间隔离 | [analyze.go](../../internal/casprovenance/analyze.go) 为每个来源重建词法/赋值状态。全报告 coverage 是汇总，occurrence 使用该来源自己的 spans。 | 页面或前一脚本 partial 不会直接把后一外部脚本命中降成 unknown。页面中的 unknown_script 与 S17 的词法缺口分开。 |
| 已确认前缀 | 词法在未知点前的 spans 保留；完整结束的受支持赋值可保留，未结束的表达式通过未知尾标记避免被当作 EOF。 | 并非整个来源无条件全部清空；不能简单删去触发字符再继续识别，以免改变真实上下文。 |

### 新的独立合成证据

[parser_boundary_test.go](../../internal/casprovenance/parser_boundary_test.go) 使用每次运行新生成的 RSA 数据，不模仿或保存真实脚本：

- 11 种前后位置及语法场景联合验证 parser 映射、词法上下文、赋值覆盖和使用关系未知；包括未知后缀内的真实代码、注释和字符串。
- 多来源场景证明页面的未知 script 属性及前一脚本的词法缺口不污染后一脚本。
- 直接比较完整闭集报告，证明除法、正则、Unicode 前缀以及未知后缀里的注释/字符串可以得到完全相同的报告与 parser_match。现有回执不能区分这些原因，重复同包不补足该信息。

Windows 普通测试/vet/合成自检通过：`build/dev-runs/windows/20261030T115629Z-f693b6f3-cas-provenance/result.json`。Linux 普通/race/vet、两项有界 fuzz 与合成自检通过：`build/dev-runs/linux/20261030T115631Z-4c5a869f-cas-provenance/result.json`。本轮只新增测试，未改产品、分析实现、输出 schema 或现有包；没有新的校园请求。

### 结论与下一决策

当前 unknown 与已批准的保守子集一致，本轮没有发现可以直接修复的语义偏差。它暴露的是诊断覆盖不足，而不是已证明的产品选键缺陷。只增加“在哪类字符停止”可以缩小原因，但仍不能分类后面的真实匹配，更不能证明浏览器实际用键；因此不应将逐项增加原因、反复校园采集作为默认路线。

后续应集中决定是否扩展有界、纯静态的语法覆盖及验收范围，先离线评估，再决定是否需要新的真实窗口。原始材料缺失意味着离线工作只能证明新增支持的合成能力，不能预言真实页面必被覆盖；不能悄悄加入 JS 执行器、实际使用关系或扩大采集来源。正式 password_core 和 M3 状态保持。

## 后续：已批准的完整静态语法工作包

维护者随后批准 [ADR-0020](../architecture/decisions/ADR-0020-bounded-static-cas-syntax.md)，本节记录离线实现结果，不重解释先前真实窗口。选择 Goja `v0.0.0-20261008201629-63c541af38f8` 的 parser/AST；保留 MIT 许可，以可重复脚本派生七个 parser 文件，只增加原始位置、预算、取消和禁用 source map 自动读取。没有导入 VM 运行时、eval 或执行脚本；产品依赖闭包不包含该 parser。来源和逐文件哈希见 [upstream.json](../../internal/jsparser/upstream.json)。

评估依据是语法覆盖、AST 原始位置及可审核的资源挂钩。另一候选 tdewolff parse 的 AST 缺少本用途所需的原始位置，并有模块/Inline 与共享限制设置差异，因此未采用。Goja 的未修改 parser 自身没有本契约要求的逐 token/长 token 取消保证，不能仅在外层加 context 即视为有界；新增派生补丁需随固定版本一起维护。官方来源：[固定 parser](https://github.com/dop251/goja/blob/63c541af38f8/parser/parser.go)、[另一候选 parser](https://github.com/tdewolff/parse/blob/v2.8.16/js/parse.go)。

### 合成判别与限制

- 同一生成密钥分别放入代码、注释、字符串、函数、控制结构、逻辑短路、类字段、Unicode/转义标识符及模板场景；正则与除法、CRLF 模板及正则后继 token 的位置分别检查，避免规范化长度污染分类。
- 直接赋值只接受可证明的字符串字面量；PKIX/PKCS1 的同键和不同键顺序互换用本轮 RSA N/E 比较。不求值转义/拼接，不将函数内赋值推断为执行或正确用键。
- 与未修改上游 AST 做合成差分；测试输入/累计字节、token、深度、取消、source map 禁用、报告严格解码和脱敏。原生子进程实际验证 256 MiB Job 内存限制和超时硬终止，不仅检查配置结构。
- 复用同一 SDK 假响应路径，旧 context 与请求轨迹保持；语法失败清空新 syntax 的内容数组，旧 context 的已观察事实仍保留。旧前置拒绝规则可能比新 parser 更保守，不绕过采集边界。
- 正则分类只表示词法外壳，不验证完整 pattern/flags 或浏览器特定语义；不支持/超限仍拒绝。真实材料已释放且不曾保存，没有离线重放依据，也不保证未来校园脚本可完整解析。

Windows 最终专项回执 `build/dev-runs/windows/20261030T124136Z-538b1b67-auth-static-context/result.json`，Linux 最终专项回执 `build/dev-runs/linux/20261030T124139Z-62928d30-auth-static-context/result.json`；六平台三个普通产品交叉构建及诊断依赖隔离通过，详见 `build/auth-static-context-p3-20261030/REPORT.md`。静态工具已独立交付，本轮没有新校园运行或产品认证修复。

## 后续提案：浏览器取键调用链对照（待确认）

后续单次匿名结果见 `build/auth-static-context-live-20261030/REPORT.md`：产品文本匹配位于脚本注释。维护者提出浏览器抓取对照；本节只形成下一工作包，不授权导航、登录或产品选键修改。

代码事实：[page.go](../../internal/cas/page.go) 对整页或脚本文本使用无词法上下文的正则；[auth.go](../../auth.go) 优先页面候选，否则按页面 script 顺序筛选同源 HTTPS、路径含 login/rsa 的引用，首个非空候选即停止，之后才调用凭据 provider。新事实支持审阅此短路规则，但不能指定一个尚未证明的替代键。

浏览器对照的判别目标为：页面实际加载的脚本和顺序、公钥活动赋值或动态来源、提交处理器到加密函数的调用链及其参数来源。只看 DOM 变量、刷新后仍为登录页或下载了某个脚本，均不能证明真正使用关系。第一步可审阅公开脚本调用链；只有实际观察到相应调用才升级为运行时证据，不能以静态可达替代已执行。

当前受支持的 Chrome 扩展接口提供页面/DOM 操作及有限日志/资源能力，未提供本任务所需的请求拦截、隔离浏览器上下文和调试断点控制。因此不能把普通导航描述为可证明零 POST 的受控窗口，也不使用页面 evaluate 冒充调试器或读取隐藏会话。此次能力核对没有打开、刷新或读取校园页面。

建议另行明确采用专用 Playwright/CDP 浏览器测试控制后，先离线验证可行性：独立临时测试会话，不继承个人浏览器会话/密码；导航前建立请求阻断及资源/时间边界；合成页面验证自动表单、fetch/XHR、重定向、worker 与弹窗等旁路的停止规则。仅在能够建立所承诺边界时交付具体真实测试方案；不能成立则报告限制，不临时放宽。

拟议真实对照不输入账号/密码、不点击提交，不导出完整 HAR、Cookie、存储或整页认证响应；只保存最小脱敏事实。公开脚本的读取范围、浏览器执行站点 JS 与加载资源的影响，以及可能的本地缓存副作用须在实际方案中明确，不能沿用旧工具“只分析 SDK 已消费来源、不执行 JS”的授权。允许的来源、预算和停止条件需在可行性验证后固定，不先承诺与旧诊断相同的请求数。

本次决策只针对离线准备方法及浏览器对照方向，不授权新的校园窗口、凭据操作或改变产品实现。产品选键修复仍需对应规范/ADR与合成回归，不能把诊断 parser 自动接入 SDK。
