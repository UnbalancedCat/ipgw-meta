---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
status: proposed
---

# CAS 公钥来源诊断设计草案

2026-10-30 当前路线：按 [ADR-0021](../architecture/decisions/ADR-0021-cas-key-use-relation.md) 修复共享产品取键，旧匿名真实入口在新构建中停用。下文 D 系决策、旧包与回执保留历史含义，不形成新的运行授权。


本页定义维护者开发工具的设计与验收，不是现有产品 CLI/JSON 契约，也不是网络执行许可。各阶段的完成记录见当前交接；下文保留早期离线阶段的边界，新增真实传输使用独立入口和身份。关联 [ADR-0016](../architecture/decisions/ADR-0016-cas-provenance-before-key-selection.md) 仍为 proposed；现有安全、协议和 evidence 规范优先。

## 2026-10-29 优先级与历史保留说明

当前后续顺序以[认证收敛 roadmap](../upgrade/plan.md#roadmap-20261029)为准。D4–D15 证书/CTL 专项退出默认开发主线，保留全部设计、实现、包和结果；不因下文“待执行/下一步”或历史同意自动重跑、扩展或继续校园窗口。重新安排需要直接阻塞产品认证的新证据及明确的假设、观察和退出条件。

D15 尚未取得 CAS 页面/公钥来源事实，其显式 ROOT 池验证与产品默认 Windows 系统验证不同；历史成员观察不能替代系统政策，也不能证明产品 TLS 成功或失败。下一开发工作包先对照旧版与 v1，不以完成自建 CTL/Windows 信任解释作为认证修复前置。

未来产品等价观测的系统 PKI 路径须由 P2 先同步规范、ADR、实现与回归，明确可能的系统网络/缓存行为后，才进入新的真实授权窗口。本页历史诊断接受范围未因此扩大，ADR-0016 仍为 proposed；历史一次性授权不续期。

## REL-CAS-DIAG-001：独立身份与执行边界

- 使用新开发身份 cas-provenance-dev、独立源码入口和输出目录，不使用 R1/R2 名称，不修改既有安装，不声明为 Candidate 或正式 live-gate 结果。
- 第一实现阶段只包含内建合成场景和 self-test。无 URL、HTML/脚本文件、profile、凭据、Cookie、环境认证参数或历史 evidence 输入；未知参数拒绝。无真实 HTTP transport、DNS resolver 或 SDK.Login 调用。测试夹具以新生成的 RSA 和合成 canary 构造，原始内容仅在内存。
- 未来采集适配器是独立工作包；不能以环境变量、隐藏参数或 self-test 模式切到真实网络。添加可联网代码之前重新审核采集契约与限额，实现并不等于取得执行授权。
- 状态分开表示工具绑定、分析完整性和未知原因。输出中没有 login_succeeded、认证 pass 或 live_verified。完整分析允许报告多个键，但不决定产品如何选键。

## 分析管线与证据等级

输入管线为：封闭合成 source 集合 → 有界 HTML/脚本分类 → 词法上下文 → 受支持的有限赋值形态 → 内存 RSA 解析/等价关系 → 有限静态使用关系 → 固定投影输出。各阶段不得执行 JavaScript 或使用 eval、VM、浏览器、动态 import。

| 层次 | 输出含义 | 不允许推断 |
|---|---|---|
| 原始匹配 | 兼容观察现有正则命中的文本位置 | 它是变量声明或被调用代码 |
| 词法上下文 | script code、comment、string、template 或 unknown | 字符串里的赋值是实际赋值 |
| 候选赋值 | 在明确支持的语法子集中识别变量/属性与字面值 | 首次出现是最终值、属性等于全局变量 |
| 键等价 | 同轮有效 RSA 的模数与指数相等 | 不同 DER/Base64 文本一定是不同键 |
| 使用关系 | 在限定单一词法块中可证明的直接引用 | 跨函数、异步、闭包或运行时调用顺序 |

第一实现阶段先完成词法分类和赋值事实；使用关系默认 unknown，直到有独立规则与测试支持。大小写按 JavaScript 标识符区分；普通 ASCII 标识符、点属性、Unicode/转义标识符、模板插值和正则字面量不得被一个模糊正则等价处理。未支持形态标为 unknown，不能输出“完整且零候选”。内部识别名称可限定 publicKey/publicKeyStr，但输出只含其预定义类别，不回显任意标识符。

页面和各脚本分别分配 S1…S17 的轮内标签；内联 script 和外部 script 不混成一个正则输入。顺序字段仅为 source 集合及 DOM 出现序号。async/defer/module、条件分支或未调用函数只表示结构事实，不推测是否执行。页面存在键时，诊断仍保留集合内其他来源；不能照搬产品的发现短路来声称来源收集完整。

公钥按 PKIX/PKCS#1 解码，遵循 [ADR-0014](../architecture/decisions/ADR-0014-cas-rsa-compatibility-envelope.md) 的算法、范围和边界；诊断不执行凭据加密。有效键按 (N,E) 在内存去重并赋 K1…K64 标签，标签只表示同轮等价，不携带指纹，不支持跨轮关联。无效键只分类 invalid_encoding/non_rsa/out_of_bounds，不输出原值或原始错误。

## 有界输入与输出草案

建议离线硬上限为 source 17 个（页面 1 个及脚本片段/外部脚本合计 16 个）、单 source 1 MiB、总输入 4 MiB、原始匹配/候选合计 64 个、关系 128 条、词法 token 100000 个、嵌套深度 64；任何超限停止，禁止截断后报完整。上述数字是设计限额，不是对真实页面大小的观察。实现前须通过上限/上限加一测试确定资源占用。

拟定报告固定包含以下字段，编码不超过 64 KiB，UTF-8 加一个 LF；消费者拒绝未知/重复/缺失字段和尾随 JSON。错误不能回显输入。字段名属于工具草案，不修改产品 JSON：

| 字段 | 约束 |
|---|---|
| schema_version / tool_kind / mode | 1 / cas-provenance-dev / synthetic |
| bundle_id | 来自已验证的开发 manifest，闭集格式及长度 |
| binding | verified 或 invalid；invalid 时不启动分析器 |
| analysis | complete / partial / rejected；表示所声明分析层次的覆盖，不表示登录 |
| coverage | 固定对象，分别声明 lexical/assignment/use_relation 的 complete/partial/unknown 状态 |
| reason | none / unsupported_syntax / resource_limit / invalid_input / binding_invalid / internal_failure |
| sources | 固定对象数组：局部 id、kind、ordinal、analysis；不含 URL、文件名或正文 hash |
| occurrences | 固定对象数组：source_id、ordinal、context、identifier_class、assignment_class、key_ref 或 null |
| keys | 固定对象数组：局部 id、rsa_bits、exponent_class；无 key bytes/指纹 |
| relations | 固定对象数组：定义与使用 occurrence 索引、kind（direct/unknown）；不能含 selected/correct |
| operations | 固定计数：external_requests/dns/profile_reads/credential_reads/cas_posts/activation/logout 全为 0 |

enum 必须在实现前列全并锁定；不得用任意 map、任意 reason/message 字符串或原始 parser error 扩展输出。binding_invalid 报告由 runner 构造，sources/occurrences/keys/relations 必须为空，analysis=rejected。超限或语法不支持时，部分结果的每个来源都要标明范围；一个 source 的未知状态不可被其他 source 的成功覆盖。

### 第一离线实现的收敛范围

以下只约束新内部原型，不改变产品解析器或 SDK。原型命令仅接受 `--self-test`，无输入文件/URL、标准输入、环境认证开关。分析器自身不验证包来源，所以内层报告固定 `mode=synthetic-prototype`、`binding=not_packaged`、`bundle_id=null`；不得伪造 verified。第二阶段由包外可信启动器在外层回执记录绑定结果，保留内层原始身份；此状态不允许任何真实采集绕过绑定。

词法支持 ASCII 标识符、空白、单/双引号无转义值、行/块注释和不含插值的模板文本。斜杠歧义（正则/除法）、Unicode/转义标识符、模板插值及未覆盖词法形态从不确定点开始标为 unknown。完整字符串的普通转义只用于定位边界，不解码成候选公钥。截断引号/注释/括号和无效 UTF-8 拒绝；所有扫描受 token、深度、大小及 context 取消约束。

赋值分析只确认直线语句中的 `var/let/const identifier = literal`、`identifier = literal` 和 `identifier.identifier = literal`；语句由分号分隔，最后一条可在 EOF 结束。其他语法（包括函数/控制流/自动分号插入）使后续赋值覆盖为 partial；已确认前缀的事实可以保留，但不推测执行。标识符精确区分大小写；只把精确 publicKey/publicKeyStr 的受支持赋值作为候选。原始文本匹配另行保留上下文，注释/字符串/模板匹配不转成候选。

HTML 只用有界 tokenizer 分离页面和 script 内容，不执行 DOM/JavaScript。原型只支持无属性 script，或仅有空/text/javascript/application/javascript type 的 script；其他属性（包括 src、async/defer、language、nomodule、nonce）及未知 type 均使内容标为 unknown；不解析/输出其 URL。该保守子集不声称这些属性在浏览器中均为非 JavaScript。输入集合中的外部脚本由内建场景提供，尚不证明与页面 src 的关联。页面、各内联片段与外部脚本保持独立来源；页面最多一个，脚本最多 16 个，提取后的总字节同样受 4 MiB 限制。

第一阶段不实现使用关系：`relations=[]`、`coverage.use_relation=unknown`。因此正常报告为 `analysis=partial`，`reason=use_relation_unknown`；有不支持形态时 reason 为 unsupported_syntax。拒绝则 reason 为 resource_limit/invalid_input/canceled/deadline/internal_failure，清空所有内容数组，不保留看似完整的前缀。`coverage.lexical` 与 `coverage.assignment` 只取 complete/partial/unknown；source.analysis 只取 complete/partial。

source.kind 闭集为 page/inline_script/external_script/unknown_script；occurrence.context 为 code/comment/string/template/html/unknown；identifier_class 为 public_key/public_key_str/other/unknown；assignment_class 为 bare/declaration/property/none/unknown。occurrence 增加 `key_class`（not_checked/valid_rsa/invalid_encoding/non_rsa/out_of_bounds），只有 valid_rsa 才允许非 null key_ref。keys.exponent_class 为 common（65537）或 other（其他有效奇指数）。occurrence.ordinal 是全报告顺序，不表示运行顺序；报告中不含输入字节位置或真实行号。所有枚举都由报告编码/解码校验，原型拒绝 verified/binding_invalid 等尚未实现的打包状态。

命令输出为固定 self-test envelope，包含 tool_kind、mode、binding、self_test_result、checks 和示例 report；成功仅代表内建断言通过。报告及 envelope 分别限制 64 KiB。报告消费者必须验证语义引用、计数和上面的字段闭集；未知/重复/缺失字段、非规范状态组合或尾随 JSON 均拒绝。self-test envelope 不适用于正式 evidence。

## 来源绑定和自测

开发包闭集建议为 EXE、runner、manifest.json、build-inputs.json 四文件。manifest 记录 EXE、runner 和 build-inputs 的大小与 SHA-256（不自哈希），build-inputs 固定源文件清单/精确字节 hash、Go 版本/平台/构建参数和干净或 dirty 状态。当前开发工作树有未提交改动，HEAD 只能作为背景字段；source snapshot digest 才绑定实际输入。不得把 b4eeb760、PR head 或 Candidate source 混作同一身份。

信任链为：单独审阅的外部归档摘要与 manifest 摘要 → 包文件闭集 → manifest 严格解析 → runner/EXE/build-inputs 校验 → 校验后使用保护 → 整链 self-test → 开发安装回执。外部期望值不得从待验证包自己读取；内置 EXE 哈希不是独立信任根，不引入 runner 自哈希循环。缺少受信任外部期望值时来源保持 unverified，即使所有本地 checksum 自洽也拒绝启动。

链的入口必须是独立受信任的启动校验器，先验证待执行 runner；禁止直接启动未知 runner 后让它自证安全。外部期望值和启动校验器的审阅/固定记录不放在待验证四文件包中。runner 继续校验 EXE 及整组身份，但不能替代入口信任锚点。此规则使“替换 runner 和它宣称的 hash 后仍自洽”成为明确的负例。

检查必须在 profile、网络、拓扑查询和 EXE 启动之前完成。拒绝缺失/多余文件、重名或大小写冲突、目录/reparse/链接替代、重复 JSON 键、未知字段、路径逃逸、身份或 hash 不匹配。读取校验后保持相同已验证文件对象，Windows 需以禁止写/删共享的句柄覆盖创建子进程到子进程退出的窗口；若平台行为不能证明，停止并保留 blocked。二次 hash 不是消除检查/使用竞态的证明。

整链 self-test 从真实 runner 入口进入，必须覆盖 runner 验证及 EXE 内置合成套件；不只运行 EXE --self-test。失败路径用计数器/标记证明 EXE 未启动、profile/网络未触及；不能依赖操作本身失败来证明“未尝试”。任何混配测试只使用新生成的合成包和独立测试目录，不更改历史 R2 文件。

## 离线验收矩阵与实施顺序

### 第二阶段绑定实现的收敛范围

本阶段只实现 Windows/amd64 开发包的验证与运行，不安装到产品或旧诊断目录。四文件固定为 `cas-provenance.exe`、`runner.json`、`manifest.json`、`build-inputs.json`。runner 是声明式调用清单，不是可任意执行的脚本；唯一动作是以固定参数 `--self-test` 启动该 EXE，30 秒超时、stdout/stderr 各不超过 64 KiB。它独立记录 EXE 大小/hash，必须与 manifest 和实际文件同时一致；同目录的自洽错误绑定仍须被拒绝。

包外的独立 Go 启动器先接收维护者已审阅的 manifest SHA-256，再持有目录祖先和四文件句柄、校验闭集及严格 JSON。启动器不从包内或邻近 sidecar 自动读取可信期望值。runner 由可信启动器严格解码并解释；拒绝额外动作、参数、输入或字段。Linux 仅支持纯元数据测试，不能宣称具备 Windows 文件锁保护。

manifest 的闭集字段为 schema_version=1、tool_kind=cas-provenance-dev、bundle_id（dev- 加 build-inputs 的 64 位十六进制 SHA-256）、platform=windows/amd64、files（按名称排序的三个 name/size/sha256 记录，不包括 manifest 本身）。runner 闭集为 schema_version、tool_kind、mode=synthetic-only、executable=cas-provenance.exe、size、sha256、args（仅 --self-test）、timeout_seconds=30、max_output_bytes=65536。build-inputs 闭集为 schema_version、tool_kind、go_version=go1.25.0、platform、cgo_enabled=0、head（背景提交）、worktree=dirty/clean、source_digest、sources（有序相对路径/大小/hash）、build_flags。输入记录不声称远端 attestation；源码快照、模块校验记录与本地构建回执共同支持本地来源核验。

构建器只复制明确列出的源码到本次临时构建快照，从快照离线构建分析器及包外启动器；拒绝覆盖已有输出目录。快照不是第二份权威仓库。使用本地 Go 1.25.0、GOTOOLCHAIN=local、GOENV=off、GOWORK=off、GOPROXY=off、GOSUMDB=off、CGO_ENABLED=0 和固定参数，不自动更新依赖。同一开发目录另保存四成员 ZIP、归档/manifest/启动器摘要和构建回执；目录启动只验证外部 manifest 锚点，不谎称重新验证了 ZIP。ZIP 验证与新目录安装另由下述第三阶段入口负责。

Windows 仅接受规范化本地盘符绝对路径，拒绝 UNC、设备路径、ADS、dot/dotdot、尾随点/空格。祖先目录及四文件逐级以不跟随 reparse 的方式打开并锁住，普通文件须仅有一个硬链接；禁止写/删除共享直到分析器子进程退出。启动后有界捕获输出，失败不回显子进程原文。新增目录项不能成为新的执行输入；四个锁住的文件对象才是本次执行集合。保护不抵抗管理员终止进程、修改内核或同一身份的代码注入。

启动器只在成功验证闭集和 runner 后增加 exe_attempts；子进程创建成功才增加 exe_starts，不能将“尝试启动但失败”计作“未尝试”。外层闭集字段为 schema_version、tool_kind、binding、bundle_id、runner、reason、exe_attempts、exe_starts、operations、result。binding 只取 invalid/verified，runner 只取 not_started/failed/passed；reason 只取 binding_invalid/unsupported_platform/canceled/execution_failed/output_limit/canceled_or_timeout/output_invalid/none。invalid 时 bundle_id/result 为 null，两个计数都为 0，runner=not_started；verified 时 bundle_id 为 dev- 加输入摘要，passed 时两个计数为 1、reason=none、result 为经严格解码的原型自测输出；其他情况下 result=null。operations 与原型使用同一闭集，均为 0，由无网络/profile 能力的代码边界保证，不是系统防火墙证明。

内层分析器仍为 not_packaged，因为它不自行认证来源；外层 verified 仅表示与显式外部 manifest 期望值相符，不等于身份签名、完整 JavaScript 分析或真实登录。所有启动前拒绝用例必须同时证明尝试/启动计数为零，独立合成执行标记不存在；校验后篡改测试通过独立打开/替换尝试证明锁有效，不能只靠二次 hash。

| 工作包 | 必须验证的结果 |
|---|---|
| 词法/来源分析 | 单键、同键重复、不同键顺序互换、页面/脚本差异、前后缀、大小写、注释/字符串/模板、转义标识符、属性、未调用函数、遮蔽及不支持语法；只断言事实与 unknown，不断言正确键 |
| 输入边界 | 无效 UTF-8、截断 token、HTML/JS 边界、超限、深度、取消与超时；每种失败都不回显输入 |
| 等价关系 | PKIX/PKCS#1 同一 RSA 归为同轮同键；不同键保持不同，轮间不关联 |
| 工具绑定 | 正确配对通过；错 EXE/runner/manifest/构建记录、缺失/篡改/额外文件、无信任锚点、路径攻击、检查后替换均在副作用前失败 |
| 安装/自测 | runner 整链运行、同一字节闭集回执、重复安装不覆盖未知目录、失败保留原件；无账号/profile 输入 |
| 脱敏 | 在各输入位置注入合成 canary，断言 stdout/stderr/JSON/错误/报告均无原文、URL、key bytes 或其可恢复编码/指纹；只允许闭集字段 |

### 第三阶段：仅新目录的离线安装

开发安装器属于包外可信启动器，只接受显式 ZIP 路径、包外归档 SHA-256、包外 manifest SHA-256 和一个尚不存在的目标目录。它不是产品安装器，不更新已有版本、不修改 PATH/ACL、不注册服务或创建 active 链接。本轮仅在合成测试目录执行；已有用户工具目录、历史 R2 和 Candidate 均不操作。

安装前以受保护句柄读取本地 ZIP（最大 40 MiB），先验证外部归档摘要，再在内存中解析严格四成员集合。只接受 Store/Deflate 的普通文件，拒绝目录、符号链接、加密、重复/大小写冲突名称、路径逃逸及额外成员。每成员按既有 metadata/EXE 大小上限读取并核对实际长度、CRC、manifest、runner 配对及 EXE 摘要；未完成全部验证前不创建目标。不得采用通用解压工具或把 ZIP 内的名称当作任意路径。

Windows 锁定已有目标父目录及祖先，使用相对目录句柄和 CREATE_NEW 语义创建唯一末级目录及四个固定文件；目标存在即拒绝，包含空目录、失败遗留、junction 和未知安装。每个文件写入并 Flush，随后关闭写句柄，重新由已有保护执行链核对完整字节、锁定并运行整链自测。关闭写句柄到重新校验之间的变化只能导致拒绝或得到相同已验证字节，不能绕过启动前验证。目标目录句柄始终防止目录替换。成功只代表本次验证时刻；以后运行仍须重新验证外部 manifest 摘要。

不提供覆盖、回滚删除或自动修复。创建前失败不留目标；创建后失败保留有限的部分文件供审计，状态为 incomplete，不能声明安装成功。重复调用不覆盖这些遗留文件。此开发流程不声明目录元数据的断电原子性或发布安装器的 durability 门禁。

安装回执闭集为 schema_version=1、tool_kind=cas-provenance-dev、state（rejected/incomplete/installed）、reason（archive_invalid/destination_invalid/copy_failed/self_test_failed/canceled/unsupported_platform/none）、archive_verified、bundle_id、directory_created、files_written（0..4）、operations（全 0）、self_test（既有严格外层回执或 null）。未验证归档时 bundle_id=null、不创建目标；installed 必须归档已验证、创建目录、写入四文件且 self_test.runner=passed。self_test 中非空 bundle_id 必须与安装归档一致；destination_invalid 不允许已写入文件，copy_failed 只允许完成 0..3 个文件。失败只返回固定分类，不输出 ZIP 名称、路径或原始错误。原始 ZIP 永不修改。

第一实施窗口仅完成分析器与封闭报告，使用内建合成来源。第二窗口再实现包生成/绑定 verifier/runner 与整链测试。每个窗口先更新本文相应未决细节，完成后才在交接记录标注已实现；本设计文档本身不使这些测试变为已通过。

## 尚未授权的真实采集

### 2026-10-29 采集适配器设计复核

源码核验发现以下不能默认为等价的边界：

- [auth.go](../../auth.go) 的 Login 在第 103–115 行先匿名发现 ac_id，再构造 service 和 CAS login query；裸 CAS 页面不能直接声明与产品请求上下文相同。此结论来自源码，不来自当前校园网响应。
- [client.go](../../client.go) 的 newHTTPClient 在第 190–195 行创建 Cookie Jar；[auth.go](../../auth.go) 第 114–125 行的路径还含 ticket 捕获与 activation 分支。因此不能复用 Login/newHTTPClient 后仅拦截 POST 来实现独立零凭据采集。
- [internal/cas/page.go](../../internal/cas/page.go) 第 25–75 行同时解析隐藏表单、QR 和脚本；新适配器只需独立提取 script 来源，不应复制这些认证状态。
- SDK 的脚本选择在 [auth.go](../../auth.go) 第 153–179 行以页面是否已有键、路径是否含 login/rsa、首次匹配为边界。这不是完整来源集合规则，不能直接用于来源诊断。
- 现有源码和合成夹具未证明匿名 CAS 页面/脚本必须回传 Cookie，也未证明脚本 query 必需。历史 CAS-submit 的 Cookie 拦截记录涉及另一条诊断路径，不能据此推断新匿名工具必须启用会话。未知保持未知。

维护者已于 2026-10-29 确认下列两个独立分支均采用推荐方案；确认仅覆盖离线实现，不是网络执行许可：

| 决策 | 推荐方案 | 代价及另一选择 |
|---|---|---|
| D1：入口上下文 | 保留与产品相同的动态 service/ac_id 语义，但独立实现匿名发现和 URL 构造，先用假响应验收；不调用 SDK.Login，不读 profile | 多出匿名发现状态及预算。若先做裸 CAS 页面，只能标注独立基线，不能声称复现产品取键上下文 |
| D2：第一版会话与脚本 URL 子集 | 不创建 Jar、不读取既有 Cookie、不回传 Set-Cookie；200 响应的 Set-Cookie 丢弃且不输出值；脚本只接受无 query/fragment 的同源 HTTPS URL，任何需要会话、额外 query 或 CAS/脚本重定向的情况停止，不自动扩大允许范围 | 可能得到 blocked/partial，而非完整来源。另一选择是另行设计仅本轮内存 Cookie 及逐字段 query allowlist；没有证据前不默认引入 |

D2 不禁止 D1 明确构造的初始 CAS service query，也不允许把响应 Location、脚本 query、表单值或任意用户 URL 当作该例外。D1 的匿名 ac_id 发现仍仅按既有协议契约允许一次受限网关 redirect；D2 的零跟随范围为 CAS 页面及脚本。相同 service/ac_id 语义只说明请求上下文对齐，不证明无 Cookie 页面与 SDK 会话页面字节相同。无 Cookie 请求仍可能触发服务端匿名状态创建；工具不能保证服务器零副作用，也不能为此自动 logout。

不依赖这两个分支的实施边界已经确定：第一步只实现假响应驱动的离线状态机，没有 HTTP/DNS/连接默认实现，没有可联网命令参数，不将源文件或 URL 输入接入既有 self-test；现有四文件开发包与安装回执保持独立身份。真实传输适配器、真实工具构建绑定与执行审批是后续步骤，不能由离线状态机通过自动解锁。

拟议离线验收矩阵如下；这是待实施清单，不是通过记录：

| 层次 | 必须证明的结果 |
|---|---|
| 初始上下文 | D1 若采用产品语义，service/ac_id 构造与产品规则相符，但不进入 SDK 的 status/凭据/activation 路径；裸页模式不能伪装等价 |
| 请求边界 | HTTP 仅可能用于既有匿名 ac_id 发现契约；CAS/脚本仅 HTTPS；禁止 userinfo、非允许端口、外源、路径歧义；所有拒绝早于下一次模拟发送 |
| 来源发现 | 独立记录页面与显式 script[src]，页面出现键不短路；不按 login/rsa 名称筛掉其他来源；重复引用保留出现关系，实际获取可去重；不递归 import、不执行 JS |
| 未知形态 | base、module、动态脚本、未支持 URL/语法等不得被当作“完整且没有键”；发现边界外来源时停止后续获取，并明确受限覆盖 |
| Cookie 与 redirect | 按 D2 验证；不解析/输出 Cookie 值；401/403、认证跳转、ticket 或需要登录态时无后续请求；不能只统计成功请求 |
| 预算 | 每次模拟请求、解析/连接尝试分别预扣预算，失败也计数；禁止自动重试、隐式 redirect、IPv4/IPv6 竞速或备用地址重连绕过限额 |
| 响应边界 | 独立限制 header、压缩前/后 body、总字节、source/脚本数、时间和取消；超限立即停止，关闭输入，不能截断后报告完整 |
| 脱敏 | URL、Cookie、HTML、脚本、公钥及原始错误的合成 canary 不得进入输出；局部 source/key 标签只对本轮有意义 |
| 证据等级 | 模拟计数与实际网络计数分开；实际请求/DNS/profile/凭据操作全部为零；采集完整不代表分析完整或选键正确 |

历史 GET/DNS 各 12 只能作为下一轮预算提案。实现时须区分逻辑 resolver 调用与实际 DNS 报文，不能把前者宣称为后者；若真实传输层无法证明物理尝试上限，未来执行仍 blocked。当前不承诺已实现 DNS、地址绑定、TLS 验证或传输隔离。

### 第四阶段：纯内存采集回放

内部 `cascollect` 只消费有界的合成 Scenario：三个 `.invalid` 域名端点（captive、service、CAS）和按顺序预置的响应。不存在 HTTP client、resolver、socket、回调式 transport、文件读取、环境配置或真实输入适配器。固定只读请求描述只有 GET 和目标，不能附加 Cookie、Authorization 或 body。响应内的 Set-Cookie 只作为待丢弃字段，既不解析值也不进入后续请求或输出。新增内部命令 `cas-collect --self-test` 仅运行内建场景，不加入既有四文件包，不改变既有分析器命令与绑定身份。

回放阶段限制为最多 12 次模拟请求尝试、12 次模拟解析尝试、12 次模拟连接尝试、每响应 header 名称和值合计 16 KiB（最多 64 项）、匿名发现 body 256 KiB、CAS/脚本各 1 MiB、所有获取 body 合计 4 MiB。模拟请求与解析在消费响应前计数，连接失败也计数；模拟 DNS 失败不增加连接计数，任何失败停止且不重试。这些计数不代表 DNS 报文或物理连接，更不构成真实限额验收。输入响应为内存 bytes，不模拟网络流关闭或 TLS 握手；这些仍是未来传输适配器必须单独验证的事项。

仅接受 identity 编码，拒绝压缩、冲突/重复的单值控制 header、非十进制或与 body 长度不同的 Content-Length。CAS 仅 200 + text/html；脚本仅 200 + text/javascript 或 application/javascript，参数只允许 charset=utf-8。CAS/脚本出现 Location、3xx、401/403、其他状态或不支持 MIME 均停止。对页面的源发现不读取隐藏 input；拒绝 base、template、noscript、iframe、svg、math、非普通 script、重复属性、未知 script 属性、src 同时带内联内容、自闭合/截断 script；script 仅允许 src 和经典 JavaScript type。普通内联代码不执行，不将动态生成/import 资源纳入采集范围。

URL 上限 2048 bytes，路径限 256 bytes；拒绝 userinfo、fragment、query（除明确构造的初始 service/ac_id）、非默认端口、转义路径、反斜杠、控制字符、空路径、dot/dotdot 和重复斜杠。相对 src 在校验原始引用后才解析，不能用规范化抹去危险成分。发现阶段只允许同网关 http/https、唯一 1–10 位十进制 ac_id；含 ac_id 的 Location 只消费该值，绝不发送其目标，且额外 query 拒绝。无 query 的中间跳转只允许一次。正文中有限候选必须唯一且有数字结束边界；不从超长数字前缀取值。与 SDK 的少数宽松输入差异只代表新诊断子集更窄，不修改产品规范。

读取页面后先核对全部显式 script[src] 及剩余预算，再读取任何脚本；不因已找到键或名称不含 login/rsa 而跳过。重复同一规范 URL 只获取一次，保留各 DOM 引用到轮内 R1…R8 的映射。分析器仍保留其原有 page/script 占位来源，因此总来源限制按 page + 全部 script 标签 + 去重外部资源 <=17 预检；不放宽原型边界。完成后将原页面和外部 bytes 交给现有分析器，并关联资源到其局部 source ID。此关联只证明回放中的显式页面引用，不证明运行、动态资源或使用关系。

闭集报告固定包含 schema_version=1、tool_kind=cas-collection-dev、mode=synthetic-replay、binding=not_packaged、scope=explicit_page_scripts、context_constructed、collection（complete/blocked）、reason、simulated_attempts（requests/dns/connects）、operations（既有全零集合）、references（ordinal/resource）、resources（id/source_id）、analysis（既有严格分析报告或 null）。complete 只表示显式静态集合获取完成，不代表分析完整；analysis 仍可 partial/use_relation_unknown。零脚本的 200 页面不证明无需会话或取得了预期认证页面；本阶段只识别 HTTP 状态/Location 等明确停止信号，不声称理解任意上游挑战页面或动态脚本。blocked 清空 references/resources/analysis，不输出似乎完整的前缀；只保留模拟计数和上下文构造布尔值。

reason 闭集为 none/invalid_scenario/url_rejected/discovery_failed/response_rejected/unsupported_source/budget_exceeded/resource_limit/dns_failed/connect_failed/request_failed/analysis_rejected/canceled/deadline。报告最多 64 KiB，UTF-8 单 LF；拒绝重复/未知/缺失字段、尾随 JSON、非法引用与矛盾状态；错误不回显输入。自测命令直接输出严格报告，只有内建正负断言全部通过才返回 0。Go 合成测试另验证边界、泄漏、取消、去重关系和无联网依赖。该阶段不会创建新的 Candidate 或真实诊断包。

未来若需要真实依据，先另拟 BHKDesktop 原生有线单次窗口，确认拓扑、账号 offline、独占/恢复路径、零凭据能力和脱敏输出。历史 GET 12 / DNS 12 只是待审上限，不能作为隐含许可或无条件重试预算。采集适配器还需具体约束同源 HTTPS、redirect、Cookie、DNS/连接尝试计数、响应上限和停止条件；当前草案不实现该适配器。

真实页面/脚本只在内存处理，不保存原始响应或原始错误。需要登录态、超出预算/来源边界、发现 ticket/认证重定向或无法解释的语法时停止；不进行 POST、activation、logout 或任何自动恢复。当前及未来来源诊断都不替代 [PROTO-LOGIN-001](../architecture/protocol-correctness.md) 与 [REL-LIVEGATE-001](live-validation.md) 的正式成功闭环。

### 第五阶段：维护者单次匿名采集开发入口

维护者要求连续完成到原生可测试包。新身份为 `cas-acquire-dev`，与上述纯合成命令、历史 R2、Candidate 分离。默认及未知参数均不联网；可信启动器接受 `--self-test`、仅本地 `--preflight` 或 `--run-once`。最后一项先验证并锁定新包、执行零网络 self-test，再在原生终端显示具体范围，要求维护者输入固定确认词，确认 TUN 关闭、有线独占窗口、账号已 offline、需要时可手动恢复 TUN。拒绝/EOF 即结束；确认只启动一次采集，没有自动重试。Agent 只运行 self-test/合成测试及只读本地 preflight。直接执行未绑定 payload 的 live 参数不是受支持入口。

固定端点与产品默认值相同：匿名网关发现、service 构造和 CAS 登录页；不提供 URL、Cookie、profile、凭据或任意文件输入。纯内存规划器复用第四阶段策略，以已收集的响应前缀产生唯一下一请求；`.invalid` 回放入口仍只接受合成域。真实结果单独投影为匿名采集报告，不输出内部回放的 synthetic 标签或模拟计数。输出只含预定义状态、实际尝试计数、局部来源/键类别；不包含输入地址、正文、header、错误文本或公钥 hash。

2026-10-29 维护者在首次 topology_rejected 后明确批准改为验证指定有线路径，保留虚拟接口。Windows payload 在网络前只读选择唯一活动有线硬件接口，拒绝活动硬件无线接口；唯一 IPv4 默认路由仍须属于选定接口，两个半默认路由仍拒绝。其他虚拟接口及不匹配目标的私有路由不再单独构成拒绝理由。选择该接口唯一首选 IPv4 地址和首个 IPv4 DNS，并固定接口 index/LUID；不输出它们。

新增约束补足路径保证：每个 DNS/TCP socket 在 connect/send 前设置并读回 IP_UNICAST_IF，同时绑定选定源地址；每个 DNS 地址及解析后的目标地址在本地核对最长匹配路由。若没有选定接口的匹配路由，或其他接口存在同等/更具体的匹配路由，即拒绝，不因 metric 更优而放过歧义。每次 socket 创建前重查接口 index/LUID、运行状态及源地址归属，发送前再次核对路径；失效即停。接口选项以网络字节序设置、主机字节序读回，见 [Microsoft IP_UNICAST_IF](https://learn.microsoft.com/en-us/windows/win32/winsock/ipproto-ip-socket-options)。本地路由查询不发送探测包。这不能证明物理路径、计费路径或抵御其他进程在核验后改变系统状态，仍要求独占窗口且不修改网络配置。

PowerShell 查询显式抑制 Progress 流，仍拒绝任何残留 stderr、非闭集输出或超时。内层仅返回固定原因枚举，避免把模块加载进度误判成拓扑失败。新增原因：topology_query_failed/topology_timeout/topology_output_rejected/wired_interface_ambiguous/wireless_active/default_route_rejected/split_route_rejected/source_address_rejected/dns_configuration_rejected/route_rejected/interface_binding_failed。前九项只发生在网络前；后两项可发生于逐目标检查或 socket 绑定阶段。旧 topology_rejected 保留解码兼容，不改变公共产品 JSON。

修订包新增 `--preflight`：可信启动器先绑定校验/self-test，再启动 payload 的 `--local-preflight`，只读查询本地有线接口、DNS 目标的路由和公共证书库；不解析域名、不创建网络连接、不发送探测包。内层 mode=local-preflight，collection=ready/blocked，ready 的 reason=none；attempts/operations 全零，context_constructed=false，内容数组为空、analysis=null。外层 state=preflight_completed，live_attempted/live_started=false。ready 只说明本地前置条件满足，未验证 DNS 响应或 CAS 目标路径；后者仍须真实采集时逐目标检查。该入口不请求或代替真实运行许可。

每轮最多 12 次 GET 尝试、12 次 UDP DNS 查询发送尝试和 12 次 TCP 拨号尝试；均发送前计数，失败也计数。DNS 为向所选服务器发送单个 A 查询，不调用系统 resolver、不重试、不用 TCP fallback、不查询 AAAA；校验事务 ID/question、拒绝截断/错误响应，只接受有界 CNAME 链到同包 A 记录。选择一个 IPv4，不做备用地址尝试。计数是应用发送/拨号尝试；操作系统 TCP 重传不算新连接，不能据此声称线上的总报文数量受同一上限约束。

每次 GET 使用一个全新的 HTTP/1.1 连接，禁止代理、连接池、自动 redirect、压缩和 Cookie jar；只使用匿名固定 header。HTTPS 从本机只读注册表 ROOT/Disallowed 证书库构造独立 CertPool，使用 Go 进程内离线证书链验证、正确 ServerName、TLS >=1.2；拒绝链中明确不信任证书，失败不降级。不调用 Windows 可补取证书的系统链构造器，不下载 AIA/CRL/OCSP，不声称在线撤销状态或完整 Windows 证书策略等价；服务器未提供完整链或本机缺根时即 blocked。HTTP 仅用于匿名网关发现。单请求 10 秒，整轮 120 秒；取消关闭连接。header wire 上限 16 KiB（含状态行），最多 64 项；拒绝分块/其他 Transfer-Encoding、压缩、非规范/冲突 Content-Length、1xx、非 HTTP/1.1，以及超出响应/总 body 限额。无需读取的敏感 header 值立即丢弃，不进入后续请求。超限/不支持/失败即停。

报告 schema_version=1，tool_kind=cas-acquire-dev，mode 为 self-test/anonymous-once/local-preflight；binding 由包外启动器提供，payload 不自证。collection 为 complete/blocked，仅 local-preflight 另有上述 ready 状态；reason 使用第四阶段闭集，增加 topology_rejected/transport_failed/self_test_failed 及上述路径门禁枚举。attempts 分别记录 requests/dns/connects；operations 保留凭据/profile/POST/activation/logout 的零计数。analysis 只投影 coverage、reason、sources、occurrences、keys、relations，不复制旧工具身份。complete 仍仅为显式静态集合完整，使用关系 unknown，不证明会话等价、选键或登录成功。中途强杀/启动异常由启动器输出 execution_indeterminate，计数未知而非伪造零。

匿名发现失败分类修订：不改变请求规划、匹配规则、两步发现上限或任何接受/拒绝边界，仅细分原 discovery_failed 的停止原因。第四阶段回放与第五阶段采集共用以下闭集；旧 discovery_failed 保留历史解码，不从旧回执推断新的细分类别。

| reason | 所证明的有限事实 |
|---|---|
| discovery_query_rejected | 受支持同主机 Location 中存在 ac_id，但其 query 键集合、重复数量或数字长度未通过既有规则 |
| discovery_body_boundary | 正文匹配到数字前缀，其后的字符不满足既有终止边界 |
| discovery_body_value_rejected | 正文匹配到的数字串不满足既有 1–10 位规则 |
| discovery_body_conflict | 正文存在不同的有效匹配值；不输出值或数量 |
| discovery_redirect_limit | 已消费允许的一次跳转，当前受支持 Location 仍无 ac_id，正文也无有效匹配；不跟随下一跳 |
| discovery_no_match | 当前正文无既有规则认可的匹配，且没有可跟随的受支持 Location；不声称页面中不存在其他形式的标识符 |

按原检查顺序报告首个停止原因；header/URL/资源限制失败仍使用既有类别。上述六类只能用于 blocked、context_constructed=false、1–2 次完整发现响应计数（requests=dns=connects）；redirect_limit 恰为 2 次。local-preflight/self-test 不生成这些真实采集失败类别。闭集检查拒绝任意错误文本及矛盾计数，不新增正文、URL、HTTP 状态、query、Cookie 或具体标识符字段。验收以合成一跳/两跳响应验证分类、停止位置和无第三次请求，并验证原有成功/拒绝矩阵保持不变。

2026-10-29 query 分类补充：discovery_query_rejected 继续解码历史回执；新报告将其三项谓词按以下顺序区分，所有原拒绝仍拒绝，不增加请求或输入输出字段：

1. discovery_query_duplicate：解码后的 ac_id 值不唯一（包括相同值重复）；不输出数量或值，不继续判定其他键。
2. discovery_query_invalid_value：ac_id 唯一但不满足 1–10 位 ASCII 十进制规则；不输出长度、字符或值。
3. discovery_query_extra_fields：ac_id 已唯一且格式合法，但解码后的 query 含其他键；不输出其名称、数量或值，不声明其他键无害或无会话语义。

这三类沿用上述失败计数/阶段约束。合成验收须含编码后的重复键、相同/不同值重复、空/非数字/超长值、同时有额外键和非法或重复 ac_id，以及唯一合法值配额外键；只在最后一种形态报 extra_fields，避免错误推断。此前 query_rejected 不能追溯重标为其中任一类。

源码差异记录：产品 auth.go 的 inspectACIDDiscoveryResponse 校验 ac_id 唯一与格式，不要求 query 仅有该键；诊断 inspectDiscovery 另有键集合约束。该差异不证明当前真实失败仅由额外键导致，也不允许直接删除诊断约束。未来若考虑只取 ac_id 并丢弃其他键，应先明确 D1/D2 与会话/认证跳转停止规则的关系，维护者确认后再改诊断接受语义；本补充仅改善分类。

### D3 已确认：匿名发现只消费 ac_id

状态：accepted for offline implementation，维护者于 2026-10-29 明确回复“可以，请开始”。本节覆盖前述第四阶段及分类历史中额外 query 一律拒绝的规则；其余边界保持。extra_fields 回执未提供其他键的名称或语义，不能预判 D3 一定适配当前响应。确认不授权真实网络重试。

仅对已通过状态、URL/host/端口/path 和 ParseQuery 检查、且含唯一合法 ac_id 的匿名发现 Location 建立以下例外：

1. 仍不请求原 Location。仅复制 ac_id 标量到新的内存值，由固定 service/CAS 端点和新的 url.Values 构造下一请求；禁止复制原 Location 的 host、path、query 或其他键值。其他字段不参与路由、选择、header、body、Cookie、缓存或报告。
2. 其他键一次 URL 解码后须匹配 ASCII `[A-Za-z][A-Za-z0-9_]{0,63}`；空键、空白/控制字符、括号、非 ASCII 或残余百分号等异常形状停止，不进行反复解码。保留键 ac_id 必须精确小写；其他键按小写并移除下划线后若为 acid，则作为保留名碰撞拒绝。
3. 同样的键名规范化后，以下固定认证/会话标记命中即停止，不因值为空而放过：ticket、tgt、lt、execution、eventid、code、state、session、sessionid、sid、jsessionid、token、accesstoken、idtoken、refreshtoken、samlrequest、samlresponse、relaystate、assertion、cookie、authorization、password、passwd、pwd、credential、credentials、otp、smscode。语法拒绝已覆盖以 `_` 开头的标记。使用固定原因 discovery_query_key_rejected / discovery_query_auth_marker 报告，均不输出命中的名称或值。
4. ParseQuery 自身失败仍拒绝整个响应。未知普通键的值只按标准单次 query 解码被动接收，不递归解析、不解释嵌套 URL/编码、不跟随、不回传、不记录；受既有整个 URL 长度上限约束。丢弃这些值不是认可其无害，也不是完整识别全部认证/会话语义。固定标记只代表可检测信号；未命中不证明匿名页面与 SDK 会话等价。
5. 重复/非法 ac_id 仍停止；无 ac_id 的 queryful 中间跳转仍拒绝。至多一次无 query 同网关中间跳转、最多两份发现响应、D2 的 CAS/脚本零重定向与脚本无 query/fragment、不回传 Cookie、不读 profile/凭据、不 POST/activation/logout，以及预算/TLS/脱敏规则均保持。检测到固定认证/会话标记、CAS/脚本 Location、401/403 等明确信号仍停止，不增加会话处理能力。

先离线更新实现、报告闭集和合成测试，生成独立新包；真实执行仍需维护者另行单次确认。验收至少包括：一/两步合法 ac_id 加普通字段仅进入固定重建的 CAS URL；所有标记及大小写/单层编码/下划线变体拒绝；编码重复键、保留名碰撞、非法键和值、复合异常停止；未知普通值含合成秘密/嵌套 URL 时不透传不输出；原有 URL/跳数/会话边界全部保留。只产生开发验证，不改变公共 SDK 或正式 password_core 状态。

确定性失败优先级：既有 ParseQuery、重复 ac_id、非法 ac_id 校验在前；随后检查全部键的语法和保留名碰撞，再检查全部认证标记，避免 map 遍历顺序影响分类。两个新原因沿用 1–2 份发现响应、context_constructed=false 的闭集约束。旧 extra_fields 原因继续解码历史回执，新采集不再仅因普通额外键存在而输出该原因。

未选方案记录：继续保留 extra_fields 阻塞，先采集额外键的有限分类再逐项决定接受范围；不要求维护者提供原始 Location 或参数值。

新包目录闭集为 `cas-acquire.exe` 与 `build-inputs.json`。包外原生可信启动器在构建时固定两者摘要，绑定精确源码快照及工具链，保持 Windows 祖先目录/两文件读锁到子进程退出，并使用 volume GUID 执行路径。外层输出携带 source digest 与 verified/invalid、live_attempted/live_started（分别为尝试执行与 CreateProcess 成功）；内层严格解码后才转发。缺失、额外、链接、篡改或混配在子进程/拓扑/网络之前拒绝。包外启动器自身是需审阅和固定摘要的信任入口；本地摘要不是线上 attestation。新输出目录独占创建，不覆盖历史工具，不提供安装升级行为。

离线验收需覆盖 planner 原有矩阵、实际 loopback UDP DNS/HTTP/TLS、证书失败、超时/取消、Cookie 不回传、redirect 不跟随、压缩/分块/超限拒绝、计数上限、原生绑定/混配/锁/整链 self-test。测试注入只允许 loopback fixture，不从 CLI 或环境开放 TLS/DNS 替换。每个新包的真实网络执行均须维护者显式启动，历史诊断结果不自动继承。

DNS 实现加固：D3 离线 fuzz 发现当前 dnsmessage 全报文 Unpack 会进入无关 SVCB RDATA 的越界路径。独立诊断先验证 4096-byte 报文上限、事务 header、唯一 A/INET question 和最多 32 个 answer，再逐项核对资源 header 与声明长度。仅解释 answer 中的 A/CNAME；其余 answer 及 authority/additional RDATA 作为有界 opaque bytes 跳过，不调用 SVCB/HTTPS 等无关类型解析器，不声明验证其内容。A 长度必须为 4，CNAME 非空且可解析；坏 framing、截断、错误 question 等仍拒绝。该修复不修改依赖版本、产品 DNS 或真实请求预算；失败 fuzz 样本保留为生成的回归输入。

### 传输失败阶段分类

2026-10-29 在 D3 的维护者运行已构造 CAS 上下文、第三次尝试返回 request_failed 后，仅补齐传输层分类，不修改接受规则、超时、证书来源或请求能力。旧 request_failed 保留历史解码，不能追溯推断旧失败阶段。

- TLS：tls_unknown_authority、tls_hostname_mismatch、tls_certificate_time_invalid、tls_certificate_invalid、tls_certificate_denied；按错误类型/显式拒绝标记分类，不匹配错误字符串。unknown_authority 仅表示离线信任链未建立，不能区分缺根、缺中间证书等；time_invalid 不输出日期，也不直接认定系统时间有误。其余为 tls_handshake_timeout / tls_handshake_canceled / tls_handshake_failed，最后一项不声称具体协议根因。
- 写请求、读状态行、读 headers、读 body：分别为 request_write、response_status、response_headers、response_body，加 _timeout / _canceled / _failed。状态行/headers 未完成时的 EOF、连接错误等归 failed；body 正常结束的 EOF 保持既有处理，不计为 I/O 失败。与 context 或 net.Error 对应的超时/取消归各自类别。状态/headers 格式与大小限制仍独立使用既有 response_rejected/resource_limit。
- 本地 socket deadline 设置失败为 request_deadline_setup_failed；HTTP 请求对象构造失败为 request_build_failed。

上述新原因仅允许 anonymous-once 的 blocked、1–12 次且 requests=dns=connects 的回执，内容数组为空、analysis=null；context_constructed=false 对应发现阶段，true 对应 CAS/脚本阶段。计数仍为尝试，TLS 失败不代表 GET 已发送。内层固定类别之外不输出底层错误、证书 Subject/SAN/指纹/有效期、IP、URL 或响应片段。证书失败仍在 GET 写入前停止，不补取证书、不切系统在线验证、不降级 TLS、不换源重试。

离线验收覆盖自造受信任/不受信任、主机名不符、有效期不符及显式拒绝证书；确认 GET 写入前停止。使用 loopback 或内存故障读取器覆盖握手/各 HTTP 读取阶段的超时、取消、EOF/失败和报告闭集，原有正例与响应边界保持通过。

### D4：TLS 失败后的本地中间证书对照

2026-10-29 维护者通过 grilling 明确同意仅做失败后的离线对照，授权实现、合成验证及独立打包，不授权 Agent 真实执行。维护者返回 tls_unknown_authority 后，当前输出仍不足以区分缺根、缺中间证书或其他链构建问题。D4 只增加失败后诊断的证书读取范围，不改变连接接受规则；连接验证仍使用原 ROOT/Disallowed 来源。

1. 仅在原自动 TLS 校验返回 UnknownAuthorityError 时进入；先关闭失败连接，不重连、不发送 HTTP。通过固定 Go 1.25.0 的 CertificateVerificationError 取得本轮未验证证书链，仅保留内存，不提供证书文件/URL/环境输入。不能取得该类型、超限或无法复现原校验结论时，报告对照不可用。
2. 保持原 ROOT 和 Disallowed 快照；只读枚举 CurrentUser/LocalMachine 的注册表 CA 证书库，不枚举私钥，不加载系统链构造器，不访问 AIA/CRL/OCSP。CA 条目只加入单独的 Intermediates 池，绝不能加入 RootCAs 或成为新信任锚。库不存在视作空；读取失败、无效输入或资源超限不得伪装成“缺根”。
3. 用与原验证一致的 DNSName、验证时间及 serverAuth 用途，先复验原 peer 链，再加本地中间证书作一次纯内存 x509.Verify 对照。任一返回的验证链含明确拒绝证书，都不得促成可用链结论，保持既有拒绝语义。证书数量、单证书/累计字节及链构建工作量设硬上限；取消或窗口结束即停止后续工作，不延长网络窗口。Verify 本身没有 context 参数，只能限制输入、检查调用前后期限，不承诺验证内部即时取消。只有同一基线失败、加入缓存后成功且拒绝策略满足，才能报告缓存可补链。
4. 输出仍为 blocked，既有空 references/resources 和 analysis=null 不变。增加固定 reason：tls_chain_local_intermediates_available（缓存可构造通向既有根的合规链）、tls_chain_unresolved（缓存仍未解决）、tls_chain_probe_denied（对照命中明确拒绝）、tls_chain_probe_unavailable（无法完成有界对照）。沿用传输分类的 anonymous-once、1–12 次且 requests=dns=connects 限制。不输出证书内容、名称、日期、指纹、证书数量、IP、URL、系统错误或 verify 原文；旧 tls_unknown_authority 保留兼容。
5. 对照成功也不恢复 TLS 握手，不继续 GET，不代表已经验证 TLS 对端持有私钥，不表示缓存来源已获外部证明；未解决也不证明“本机缺根”或“服务器证书错误”。既有注册表 ROOT 子集与完整 Windows 逻辑信任策略的差异仍是未消除的限制。
6. 离线验收使用生成的根/中间/叶证书覆盖：缺中间可补、完整 peer 链原本可验、缺根仍失败、无关/自签/伪造 CA 不能成为根、错误域名/用途/有效期、Disallowed、空库/读取失败/超限/取消，以及实际失败连接关闭、零 HTTP GET、无第二次拨号。两端测试、Linux race/fuzz、闭集脱敏及启动器绑定审计通过后才交付独立新开发包。
7. 这项选择只授权离线实现、合成验证及打包。真实复测仍须维护者在原生终端当次确认；上限保持 GET/DNS/TCP 各 12 次尝试、120 秒，所有凭据/profile/POST/activation/logout 为零。任何让缓存证书影响连接接受的后续方案另作决策，不能从 D4 推导授权。

D4 资源边界：peer 链最多 16 张、单 DER 64 KiB、合计 512 KiB；两个 CA 库累计最多 512 个枚举条目、单 DER 64 KiB、合计 8 MiB，重复项也计入读取限额。证书先校验长度，再复制/解析；无效 DER 或超限返回 unavailable，不使用部分结果。对照至多两次 Verify，沿用固定 Go 的链构造签名检查限额；任何单一 RawSubject 的 peer/缓存候选总数超过 16 即 unavailable，以限制同名分支。原 ROOT 快照不修改，不因 D4 增加根；这不是完整 Windows 信任策略模拟。逐条枚举及每次 Verify 前后检查原请求 context，超过期限不继续；不启动后台校验任务。

### D4 对照不可用的阶段分类

维护者运行 D4 返回 tls_chain_probe_unavailable，原原因无法区分输入、基线复验、CA 枚举、资源上限或期限。仅细分停止原因，不扩大 D4 来源、接受规则、预算或输出数据。旧 tls_chain_probe_unavailable 保留历史解码，不能追溯重标。

新增固定原因均以 tls_chain_probe_ 开头：canceled / deadline（context 状态）；input_rejected（必要输入或原错误类型不符）；peer_invalid / peer_limit（对端链结构或既定资源上限）；baseline_mismatch（同参数复验未重现 UnknownAuthority）；peer_subject_limit / cache_subject_limit（加入相应来源时触发既定同名候选上限）；store_open_failed / store_read_failed（只读库打开或枚举失败）；cache_invalid / cache_limit（缓存 DER 解析/结构或既定资源上限）；loader_failed（非 Windows 或内部 loader 非闭集返回）。所有原因仍只能 anonymous-once、blocked、完整尝试计数 1–12、空内容数组/analysis=null，不输出库位置、数量、DER、证书元数据或底层错误文本。

检查顺序保持 D4：先 context，再必要输入和 peer 限额，原链复验，CA 读取及缓存限额，同名候选限制，最后补链 Verify；每次操作后的 context 状态优先于该操作结果。缓存只读函数错误改为内部闭集原因，部分读取结果不交给对照器。对照成功、未解决与明确拒绝三类含义不变；不会因分类修订新增 CA 读取时机、证书写入或请求。

### D5 已确认：可解析缓存子集与零网络缓存检查

状态：2026-10-29 维护者明确选择“同意推荐：部分缓存＋零网络检查”。以下规则覆盖 D4 中“任一 DER 失败整批停止”和“仅在 TLS 失败后读 CA”两项约定，其余 D4 边界保持。维护者返回 cache_invalid，可能为结构异常或 Go DER 解析失败，尚不知道实际是哪一种。合成内存库已证明 Windows 可存储/枚举被固定 Go 解析器拒绝的自造证书；这不等于现场原因，也不能作为放宽解析器的理由。

确认的边界：

1. 只跳过非空、长度合法、已安全复制的 DER 在 Go ParseCertificate 中的解析失败。仍使用固定 Go 原解析规则，不匹配错误文本、不启用兼容开关，不输出证书、数量、原始错误或指纹。跳过条目及其字节仍计入原累计预算，重复项也计入。指针/零长度异常、store 错误、超限、取消/期限及同名候选限制继续停止，不使用遇到这些错误前的部分结果。
2. CA 解析成功项仍只作 Intermediates，ROOT/Disallowed 不变；原 TLS 失败后关闭连接、复现基线再对照的顺序不变。仅 DER 解析失败导致的部分覆盖显式区分：tls_chain_partial_cache_available、tls_chain_partial_cache_unresolved、tls_chain_partial_cache_empty；空子集不能伪装完整读取，未解决不能证明原缓存无可用证书。完整缓存沿用 D4 原类别；明确拒绝仍返回 denied。所有结果始终 blocked，无恢复握手、GET 或重试能力。
3. 新增绑定启动器的 --cache-check 本地入口，允许在任何校园 run 前单独只读相同 CA 库；无拓扑探测、DNS/TCP/HTTP、peer 证书、root 链验证或 profile/凭据操作，不需要网络窗口。只报告完整可读取、部分不可解析、部分且无可用项或不可用的闭集类别；本地结果不预测真实链验证，不授权 run-once。
4. 本地检查也执行相同资源与缓存同名候选限制；实际对照仍须重新读取本轮缓存，并在加入 peer 后再检查合并上限。不保存证书或复用上次缓存快照。未知参数不能联网；launcher 必须先验证绑定和自测，local 模式的 live 标志 false、attempts/operations 全零。
5. 验收覆盖有效/不可解析/有效的顺序、坏项仍耗预算、全坏缓存、结构异常停止、部分可补与仍未解决、Disallowed、原完整路径兼容及本地入口零网络。先完成离线测试和本地检查，再决定是否交付单次校园复测步骤；不得为了检查本地缓存再次发起校园请求。

本地报告 mode=local-cache-check；完整可读取（也可能为空）为 ready/none，部分可解析为 blocked/cache_partial，全不可解析为 blocked/cache_partial_empty；不可用仅接受 loader 闭集失败及 tls_chain_probe_cache_subject_limit。这里的 ready 只说明本地检查完成，不能预测任何真实链能否验证。所有本地报告 context_constructed=false，attempts/operations 全零，空 references/resources、analysis=null。启动器成功取得并验证本地报告为 cache_check_completed（包含 blocked），绑定、自测或执行/解码失败保持独立失败状态，live_started/live_attempted 始终 false。

本地检查使用 10 秒 context，逐条和返回前检查；不启动后台任务，解析内部不承诺即时取消。包执行器为该子进程另设 15 秒硬期限；错误输出仍不回传。实际网络对照沿用原窗口期限和预算，重新读取本轮缓存。不允许 Agent 自动执行真实网络，不新增信任锚、不删除/安装证书、不改变产品认证。

### D5 对照失败的类型分类

维护者返回 tls_chain_partial_cache_unresolved：证明非空可解析子集已进入第二次 Verify，但现有类别合并了所有验证错误。只细分该次 Verify 的公开类型，不改变读取来源、信任池、参数、Verify 次数、期限或网络预算，也不重新验证失败链。不追溯重标历史 unresolved；原 tls_chain_unresolved 和 tls_chain_partial_cache_unresolved 保留解码。

新增原因由固定前缀和固定后缀组成。完整缓存前缀为 tls_chain_cache_，部分缓存为 tls_chain_partial_cache_；后缀闭集为 unknown_authority、hostname_mismatch、time_invalid、usage_invalid、constraints_invalid、certificate_invalid、verify_failed。

- unknown_authority：公开 x509.UnknownAuthorityError；不解析其私有 hint 或 Error 文本，不将其等同于缺根或缺中间证书。
- hostname_mismatch：公开 x509.HostnameError。time_invalid：CertificateInvalidError 的 Expired（包含尚未生效）。usage_invalid：IncompatibleUsage 或 CANotAuthorizedForExtKeyUsage。
- constraints_invalid：CANotAuthorizedForThisName、TooManyIntermediates、NameConstraintsWithoutSANs、UnconstrainedName、TooManyConstraints。certificate_invalid：其他 CertificateInvalidError（包括 NoValidChains 与未知枚举）、UnhandledCriticalExtension、InsecureAlgorithmError 或 ConstraintViolationError。
- verify_failed：未覆盖的公开错误类型；不匹配错误消息推断内部签名检查限额或原因。固定 Go 可能把候选失败隐藏在 UnknownAuthorityError 内；类型分类不保证穷尽根因或暴露全部失败候选。

仅使用 errors.As 和枚举，禁止输出证书、名称、时间、算法值、数量、指纹、Detail 或错误原文。context 期限/取消优先，空部分缓存仍为原 empty 类别，成功对照及明确拒绝的含义不变。新增类型只允许 anonymous-once、blocked、requests=dns=connects 且各 1–12，内容数组为空、analysis=null；本地 cache-check 不得输出这些类型。

验收用生成证书覆盖第二次 Verify 的未知权威、失效中间证书、用途/约束失败及原成功/拒绝路径；用包装错误与会在 Error 调用时失败的 canary 覆盖闭集、未知类型和无文本解析。旧报告兼容、状态矛盾拒绝、两端回归与独立包绑定须通过。实现不授权真实重试，亦不改变产品认证能力。

### D6 已确认：同源、纯内存的证书关系观察

状态：2026-10-29 维护者明确选择“同意推荐：同源、有界关系观察”。维护者返回 tls_chain_partial_cache_unknown_authority 后，现有公开错误类型已无法区分无父候选与隐藏的候选拒绝。本阶段固定证书来源，增加有界关系观察，不切换 Windows 信任来源或链引擎。

1. 只在原 TLS 已失败并关闭 socket、D4/D5 基线及缓存检查完成、第二次 Verify 返回 UnknownAuthority 时执行。不增加 Verify，不重复读库，不联网下载，不恢复 TLS/GET。复用本轮 peer、可解析 CA、原 ROOT 与 Disallowed；不使用证书文件、历史回执里的证书、其他本地库或环境输入。
2. 为复用原根，原 ROOT 枚举时仅额外保留已经加入根池的相同证书对象；不从 CertPool 私有字段取值，不另行枚举，不给 RootCAs 加证书。诊断快照最多512个原枚举条目、单DER 64 KiB、总DER 8 MiB；重复项计入。超限只使 D6 不可用，丢弃该诊断快照，原根池构建及 TLS 行为不变。peer/CA继续原限额。D6所用集合必须是本轮输入的完整可解析快照，不能截断后声称全部关系已检查。
3. 从叶证书探索 RawIssuer/RawSubject 相等的候选，调用 CheckSignatureFrom 观察签名及有限 CA 条件；AKID/SKID 失配不得直接排除候选。根成员只按本轮已有根的精确DER识别，不按自签名、名称或相同公钥提升信任。重复DER可在关系计算中去重，原读取预算不变。
4. 独立预算：最多128个访问节点、深度16、每个节点最多16个同名候选、总计100次 CheckSignatureFrom。开始时取原context剩余期限与5秒两者较短者，逐步检查；不启动后台任务。密码学调用无中途取消保证，须先检查公钥工作量：RSA最多8192位，ECDSA只支持标准P-256/P-384/P-521，Ed25519要求标准长度；超限/未知表示无法完成观察，不改变TLS算法政策。达到任何边界即不可用，不发布截断关系结论。
5. 必须探索所有有界可达分支，不能遇到首个死分支即认定无路径。输出仅固定关系维度：是否观察到通向既有ROOT的签名关系、叶级/上游是否缺同名父候选、是否观察到CA条件拒绝/算法拒绝/其他签名检查失败、是否有循环或明确拒绝。多个分支可同时记录不同类别，结果须对输入排列不敏感；不输出证书数量、链长度、证书标签、名称、指纹、DER、日期、算法值或原始错误。
6. 只以公开 ConstraintViolationError 分类有限CA条件拒绝；InsecureAlgorithmError或ErrUnsupportedAlgorithm分类算法拒绝；其他错误只称signature_check_failed，不断言数学签名不匹配。明确拒绝必须显式保留，不能被一条可达根路径覆盖。部分缓存标签保留，“没有候选”只表示本轮可解析子集中未观察到；通向根的签名关系不等于完整路径有效，也不验证对端私钥持有或TLS成功。
7. 报告仍为 anonymous-once、blocked，原attempts/operations预算和空references/resources、analysis=null不变；仅新D6类别可附加闭集关系结果，旧回执保持可解码。未执行/不可用时不得伪造否定关系。实现前锁定具体字段、枚举及状态组合；未知/重复字段、矛盾状态继续拒绝。
8. 离线验收必须覆盖交叉签名、同名不同公钥、AKID失配但关系成立、无父候选、有限CA约束/算法拒绝、循环/混合分支、签名路径可达根但完整Verify失败、Disallowed、根快照超限、节点/深度/候选/签名预算、取消及排列不变性。两端回归、Linux race、闭集脱敏、绑定审计和新开发包完成后再交付；真实复测仍由维护者单次确认。

本阶段已批准新增ROOT诊断快照和额外签名关系计算，不授权扩大信任来源、调用系统链引擎、安装/删除证书或自动执行真实网络；不能借关系图“修复”TLS接受规则。

D6 输出锁定：仅 reason=tls_chain_relation_observed / tls_chain_relation_denied / tls_chain_relation_unavailable 可增加 chain_relation 对象；其他报告保持原字段集合，不输出该字段或null占位。新解码器兼容旧字段集合，但不放宽任何旧闭集。D6仍为schema_version=1的独立开发工具扩展，不是产品JSON升级。

chain_relation固定含 status（complete/unavailable）、cache_coverage（complete/partial）、reason、facts。status=complete时reason=none，facts包含且仅包含八个布尔值：root_path_observed、leaf_no_parent、upstream_no_parent、ca_constraint_rejected、algorithm_rejected、signature_check_failed、cycle_observed、denied_observed。denied_observed=true时外层必须为tls_chain_relation_denied，否则为tls_chain_relation_observed；root_path_observed不覆盖拒绝，也不是信任成功。沿签名关系可达的证书即使命中Disallowed也保留关系观察，但绝不能据此接受连接。

status=unavailable时外层为tls_chain_relation_unavailable，facts必须null；reason闭集为root_snapshot_invalid、root_snapshot_limit、input_invalid、key_unsupported、node_limit、depth_limit、candidate_limit、signature_limit、canceled、deadline。部分缓存标签仍保留。关系结果必须anonymous-once、blocked、attempts各1–12且相等、操作计数对应、空references/resources、analysis=null；不要求CAS上下文已构造（匿名发现也可能到达HTTPS失败），但绝不用于local/self-test报告。原context终止原因仍优先于D6结果，不能回传不完整事实。

关系节点按DER去重，以固定DER摘要顺序广度优先遍历所有可达候选（摘要不输出）；仅将签名检查通过的边计入可达图，以完整图检测循环。到达原根即结束该分支，不要求根自签名。访问深度为从叶到节点的最短签名关系边数；替代路径仍记录为图边，循环不会无限增加访问深度。任何输入排列下，都不得用预算内的首个成功/失败分支代表全部；发生限额时丢弃全部facts。

D6解释边界：cycle_observed包括单节点自环。一个位于peer/CA集合、但不在原根池中的自签名链尾，可以只有cycle_observed=true；多个证书互相签发也可得到完全相同输出。upstream_no_parent=false只表示存在同名候选，自身也可能是该候选。不能由此称服务器发出循环链、认定证书损坏或决定安装根；当前去重后的报告不区分环的peer/CA来源。

### D7 已确认：一次性细分末端关系与来源

状态：2026-10-29维护者明确选择“同意推荐：一次补齐同源关系分类”。D6返回仅cycle_observed=true，合成反例证明自签名非信任链尾与多证书环不可区分。以下三个维度一起收敛，真实执行仍须维护者当次确认。

1. 仅在D6完成、cycle_observed=true且root_path_observed=false时执行；保留D6原八项事实、partial标签和blocked。复用本轮完整可达图、原证书对象与来源位，不额外验签、Verify、读取证书库或获取证书，不恢复TLS/GET。
2. 分解强连通分量，末端定义为没有通向其他分量的成功签名边。分别观察末端单节点自环、多节点循环、无自环单节点死端；另外报告是否存在非末端循环。不输出节点标签、数量、长度或证书数据，不把“末端”称为可信根。
3. 对三种末端类别分别聚合成员来源：是否存在仅peer、仅CA、同时存在于peer与CA的成员。DER去重时保留来源位；这些是成员存在性，不把若干成员的来源拼成“整个环由服务端提供”。
4. 对各末端类别再观察是否存在不同DER、但与原ROOT相同RawSubject及RawSubjectPublicKeyInfo的成员；以及仅RawSubjectPublicKeyInfo相同而RawSubject不同的成员。只作字节关系比较，不声称名称语义等价，不新增根或更换验证参数。两个true可以来自不同成员；没有匹配仅限原ROOT可解析快照。
5. 计算仍受D6同一个5秒context及原网络窗口限制；新增阶段上限为两者剩余时间与1秒的较小者。节点最多128、成功边最多100、原ROOT最多512；新增原始名称/SPKI处理字节累计最多32MiB、成员与根比较最多65536次。达到任何上限、取消或不完整来源映射则D7不可用，丢弃D7全部部分事实，保留已完成D6事实及缓存覆盖标签；不启动后台任务。
6. 新输出独立于旧D6固定字段集合，只允许明确D7类别；旧解码保持兼容。D7结果仅含完成/不可用状态、固定失败原因、非末端循环布尔和三个固定末端类别对象；每个类别包含存在性、三项成员来源和两项原根表示关系布尔。不可用时这些事实必须整体null。实施前锁定字段及矛盾组合，不通过自由字符串输出细节。
7. 验收覆盖自环、双节点环、多环/混合分支、非末端环、死端、重复DER跨来源、排列不变、同键同名不同DER/同键异名/同名异键、预算停止、旧报告兼容、启动器传递及零额外网络/Verify。全部离线验证完成后打独立开发包；若需要真实运行，仍由维护者当次确认。

D7只帮助区分后续调查方向，不能证明其他Windows信任来源的内容，也不能直接恢复取键。若仍需扩大信任来源，另行提出只读来源差异审计；不安装证书、不启用在线链引擎或放宽TLS。

D7输出锁定：仅tls_chain_topology_observed、tls_chain_topology_denied、tls_chain_topology_unavailable三类增加chain_topology；旧D6保留原字段集合。保留完整chain_relation且其status必须complete、facts.cycle_observed=true、facts.root_path_observed=false；父context终止仍优先，清除全部扩展结果。D7完成时外层denied与D6 denied_observed一致；D7不可用统一unavailable，但不能丢失D6拒绝事实。

chain_topology固定含status、reason、facts。status=complete时reason=none，facts包含nonterminal_cycle和self_loop、multi_node_cycle、acyclic_dead_end三个对象。每个对象固定含present、peer_only_member、cache_only_member、shared_member、root_same_subject_spki、root_different_subject_spki六个布尔值。不存在的类别六项全false；存在时至少一个来源true；至少一个末端类别存在；全图有环必须由末端环类别或nonterminal_cycle体现。存在死端还须D6有上游无父候选或候选拒绝事实。原根表示关系为不同DER之间的字节关系，不表示信任可替换。

status=unavailable时facts整体null；reason闭集为input_invalid、source_invalid、node_limit、edge_limit、root_limit、byte_limit、comparison_limit、canceled、deadline。保留D6事实，但不发布任何D7部分结论。32MiB预算还保守计入根DER摘要计算及相等比较的两侧名称/SPKI字节，每个根身份摘要只计算一次。来源位在D6既有DER去重时保留；D7不重复读取或重建签名图。

### D8 已确认：现有注册表来源的零网络覆盖审计

状态：2026-10-29维护者明确选择“同意推荐：先做零网络覆盖审计”。D7可证明当前可解析图的末端自环来自peer，且没有原ROOT的SPKI字节匹配；不代表完整Windows信任视图或产品TLS结论。现有ROOT加载会跳过Go解析失败条目，覆盖状态尚未输出。产品默认Windows链验证与诊断显式Go根池不同，不得用普通SystemCertPool.Verify冒充零网络检查。

建议先停止同源校园复测，独立实现只读 --trust-audit，本阶段仅限以下范围：

1. 复用固定注册表provider，只读CurrentUser/LocalMachine的ROOT、Disallowed、CA。不得打开泛化SYSTEM集合、额外策略/企业/自动根来源、MY/私钥，亦不得调用系统链构造、AIA/CRL/OCSP、更新或网络接口。不改变原工具或产品的信任池。
2. 每种库在两个位置合计最多512个条目、单DER64KiB、累计8MiB；重复及跳过项仍计预算。总context10秒，逐项检查，子进程15秒硬上限；异常指针/结构、库读取错误、超限及取消停止相应审计，不把不完整读取当作空库或完整覆盖。
3. ROOT/CA用固定Go原解析器观察覆盖；Disallowed保留原始DER拒绝语义，不因Go解析失败丢弃拒绝记录。不更改解析兼容开关。只在内存中比较ROOT解析成功项与Disallowed，报告是否存在被排除项、是否有剩余可用项；不把CA项提升为根。
4. 输出固定来源类别及闭集读取/解析/拒绝覆盖状态，不输出证书名称、DER、SPKI、指纹、数量、有效期、错误文本或任意路径。不导出或保存证书。实现前锁定字段与状态组合，失败必须明确unknown/unavailable，不能输出伪否定。
5. 工具没有真实握手或请求入口。先完成内存/生成证书、坏DER、拒绝优先、限额、取消、脱敏和绑定测试，再交付本机原生零网络执行步骤。无需校园独占窗口、账号变更或关闭TUN；不沿用任何网络授权。
6. 该审计只能回答当前固定来源是否被完整读取/解析，不能关联已经释放且未保存的历史peer。全部覆盖通过也不证明Windows完整信任源与工具等价，亦不证明TLS可用。若要引入额外本地来源或关联新peer，先依据本审计结果提出具体新决策，不自动进入新的握手窗口。

D8输出锁定：独立tool_kind=trust-audit-dev、schema_version=1，mode=self-test/local-trust-audit，state=complete/unavailable；complete仅表示三个固定来源均完成有界枚举，不要求解析覆盖完整或信任可用。root/disallowed/ca各含state、reason、current_user、local_machine、presence、parse_coverage。位置状态为complete/absent/unavailable/not_checked；presence=present/empty/unknown。ROOT/CA解析覆盖为complete/partial/none/empty/unknown（none表示非空但全不可解析）；Disallowed为not_applicable/unknown。库不可用时presence及parse_coverage均unknown，丢弃该库部分对象，不影响独立其他库的报告；全局取消优先并阻止后续库读取。

库reason闭集none/open_failed/read_failed/invalid_entry/limit/canceled/deadline/unsupported。拒绝对照comparison含state、parsed_root_denied_overlap（present/none/unknown）、remaining_parsed_roots（present/empty/unknown）。仅ROOT和Disallowed均完整时计算；ROOT解析失败项不当成可用根，Disallowed原始DER仍按精确摘要匹配，不解析或降低拒绝效力。来源未完整或全局取消时comparison为unavailable且两个事实unknown。operations固定network_requests、dns、connects、profile_reads、credential_reads、certificate_writes均为0；不输出读取数量。未知、重复、缺失字段和矛盾组合拒绝。

独立绑定启动器只接受--self-test和--trust-audit，先验证独立payload与构建输入、完成合成self-test，再运行本地入口；15秒子进程上限，错误原文不回传。无--run-once、网络确认或账号操作入口；旧CAS采集包的命令白名单不扩大。新包使用独立名称及source digest，不是Candidate。

### D9 已确认：同轮 peer 与固定额外本地来源的差异观察

状态：2026-10-29维护者明确同意实施D9并完成离线验收、打包；真实执行仍须当次确认。D8返回固定ROOT可见项解析完整、固定Disallowed为空、CA解析partial，不能关联已释放的D7 peer。下一问题是：当轮失败握手中的证书关系，是否在明确的额外本地来源中出现可对应材料？这不等价于Windows完整链验证，也不决定TLS根因或是否应信任某证书。

不选择默认或自定义Windows链引擎作为本阶段实现。[CERT_CHAIN_ENGINE_CONFIG](https://learn.microsoft.com/en-us/windows/win32/api/wincrypt/ns-wincrypt-cert_chain_engine_config) 的exclusive根约定没有保证排除其他系统库读取；仅缓存检索也不是所有副作用的隔离证明。继续使用原显式Go根池、原TLS接受规则，不改变产品。

#### D9 来源和读取边界

原CU/LM ROOT、Disallowed、CA仍为原诊断基线；只新增以下七个固定来源，不接受路径、库名、URL或证书输入：

| 闭集来源 | 固定名称 | 固定位置 |
| --- | --- | --- |
| user_policy_root / machine_policy_root | ROOT | CurrentUserGroupPolicy / LocalMachineGroupPolicy |
| user_policy_ca / machine_policy_ca | CA | CurrentUserGroupPolicy / LocalMachineGroupPolicy |
| user_policy_disallowed / machine_policy_disallowed | Disallowed | CurrentUserGroupPolicy / LocalMachineGroupPolicy |
| machine_authroot_cache | AuthRoot | LocalMachine |

使用SYSTEM_REGISTRY_W物理注册表provider、READONLY及OPEN_EXISTING，不使用SYSTEM逻辑集合、PHYSICAL自动成员、Enterprise、智能卡、服务/其他用户库、MY、远端前缀、UNPROTECTED或更新标志。不读取私钥、证书属性或自行解码注册表blob；GroupPolicy存储交给官方provider，不能猜测序列化格式。只枚举证书，不枚举/校验CTL、CRL或AuthRoot的AutoUpdate材料，不触发同步、AIA、系统链引擎或证书安装。

[System Store Locations](https://learn.microsoft.com/en-us/windows/win32/seccrypto/system-store-locations) 给出了策略库的本地映射；[CertOpenStore](https://learn.microsoft.com/en-us/windows/win32/api/wincrypt/nf-wincrypt-certopenstore) 区分物理库和逻辑集合。provider可能过滤条目，特别是CurrentUser ROOT保护列表；D8/D9的完整均仅指provider返回项，不宣称所有底层blob或全部Windows来源。[Certificates and trust](https://learn.microsoft.com/en-us/windows-server/identity/ad-cs/certificate-trust) 区分CTL与证书材料；AuthRoot缓存命中不能推出当前CTL接受或信任有效。

#### D9 顺序和停止条件

1. 离线实施/合成验收先完成；真实入口仍由维护者在原生终端当次确认。Agent不自动运行真实采集，也不将本次设计同意视为校园执行授权。
2. 当次确认后、首次网络操作前，读取七个新增固定来源为有界内存快照。读取/结构失败、超限或期限停止；若没有可解析候选证书且没有原始拒绝材料，返回无可比较材料、零网络，不为了得到旧结果再握手。快照不保存到文件、不跨运行复用；Windows后台可能更新库，不能称全机原子快照。
3. 保持现有产品上下文发现及有线接口/DNS/TCP绑定，预算仍最多12次GET尝试、12次DNS发送、12次TCP连接和120秒总窗口。新增本地读取和分析消耗原期限，不延长窗口。凭据/profile/CAS POST/activation/logout仍为零。
4. 仅在原TLS失败且原D6关系/D7末端分析完整时进入同轮对照。先关闭失败连接；沿用该轮peer、原ROOT/Disallowed、可解析CA、验证时间和完整可达图，不保存peer，不重连、不恢复握手、不继续GET。若现场已不满足该前提，只返回既有分类与固定未执行原因，不扩大条件来强行产生结论。
5. 新增ROOT/CA/AuthRoot全都仅作为观察对象，绝不加入任何Roots池。对原图可达成员和末端成员分别比较精确DER身份、同RawSubject/SPKI及仅SPKI字节关系；对原图节点检查新增且原集合不存在的同名父候选是否通过CheckSignatureFrom。仅观察一层新增父关系，不递归扩图、不调用额外Verify、不把签名关系称为完整有效链。
6. 策略Disallowed保留原始DER摘要，即使Go解析失败也不丢弃；只报告它与原图可达对象的明确重叠。任何拒绝命中不得被“发现父候选/缓存匹配”解释为可继续连接。无重叠也不代表完整Windows拒绝策略通过。
7. 结果始终blocked。报告各固定来源的读取/解析覆盖及上述有限关系；源不可用/计算不完整必须unknown或null，不能把未检查计作不存在。没有相关差异时停止追加同类校园复测；有差异时再设计如何取得可信验证依据，不能自动安装根或改变信任池。

#### D9 资源和输出约束

- 七来源各最多512个枚举条目、单DER64KiB；七者累计最多2048条/16MiB，重复、解析跳过及拒绝项都计入。新增读取期限为剩余原窗口与10秒的较小值；逐条检查，失败不使用该批部分候选。
- 关系分析最多128个原图节点、65536次身份/名称/公钥比较、累计32MiB比较输入、100次新增签名检查；期限为剩余原窗口与5秒的较小值。签名检查不能在调用内部即时取消，检查前后期限；不创建后台验证任务。预算用尽时丢弃全部新增关系事实，保留完整的原D6/D7结果与明确不可用原因。
- 不输出证书数量、DER、名称、SPKI、指纹、算法值、日期、路径、IP、URL、错误文本或证书来源中的属性。新增source_difference对象与D6/D7分开，避免改变旧字段含义；固定七来源、覆盖状态、关系事实和闭集原因，实施前锁定字段和矛盾组合。
- 每个候选来源仅报告原图DER重叠、末端同名同SPKI、末端异名同SPKI、新增直接验签父候选四项存在性；拒绝来源仅报告原图DER重叠。存在性可来自不同对象，不能拼成一条有效链。解析partial时所有否定只覆盖可解析子集；未执行/失败事实整体null。

#### D9 离线验收和交付

生成证书覆盖额外来源同证书、同键不同DER/名称、同名异键、额外直接父候选、无关来源、重复/排列不变、原集合已存在项不算新增、拒绝重叠、坏DER/全坏/空/缺失库、超限和取消。用内存库、假provider和固定参数断言验证来源白名单及全零网络路径；不调用系统链引擎，不把mock测试宣称为Windows完整信任验证。

集成测试覆盖：新来源不可用/无材料时网络为零，失败连接先关闭，分析后没有HTTP/第二次拨号，所有原绑定/拓扑/确认保护保持，旧输出解码兼容，新事实缺失/矛盾/未知字段被拒绝。两端聚焦测试/vet、Linux race/有界fuzz、绑定负例及脱敏扫描通过后，生成独立新开发包；旧D7/D8包保留。

本范围已由维护者整体确认：允许上述额外本地来源和有界同轮内存观察，先实施离线验证及打包。真实窗口仍须维护者当次确认；不要求为离线实施切换TUN或运行命令。不能据此推导系统链引擎、其他信任源、CTL验证、凭据登录或发布授权。

D9输出锁定：新增source_difference含phase（gate/comparison）、status（complete/unavailable/no_material/not_run）、reason和固定顺序七项sources。各来源含source、state（complete/absent/unavailable/not_checked）、reason、presence（present/empty/unknown）、parse_coverage（complete/partial/none/empty/not_applicable/unknown）、facts。候选facts固定四项布尔：reachable_der_overlap、terminal_same_subject_spki、terminal_different_subject_spki、additional_signing_parent；拒绝来源使用同一形状但仅首项可为true。不可用/尚未对照facts整体null。库reason为none/open_failed/read_failed/invalid_entry/limit/canceled/deadline/unsupported/not_checked；对照reason另含input_invalid/key_unsupported/node_limit/comparison_limit/byte_limit/signature_limit/baseline_not_observed/no_comparable_material。

gate只在首次网络前停止时输出：无材料为no_material/no_comparable_material，读取失败为unavailable；外层分别source_gate_no_material/source_gate_unavailable，全部尝试零。comparison完整时保留原D6/D7 reason和事实；不可用不丢失原完整分析。若原D6/D7未完整，新增对照为not_run/baseline_not_observed。D9启用时首次HTTPS握手若成功，也在写HTTP前停止为source_baseline_tls_verified；不为收集材料继续GET。父context终止时丢弃新增关系事实，保持原取消分类。

实现按每次身份、名称或SPKI比较分别计数；比较字节预算保守计入原图DER、每对候选DER及名称/SPKI、额外验签输入。相同DER也可满足末端同名同SPKI事实，区别于D7专门针对不同DER的原根表示关系；两个字段不能混用。

### D10 已确认：独立零网络 CTL 材料与固定签名层级检查

状态：2026-10-29 维护者明确同意实施 D10、完成离线验收与打包；实际新增来源读取仍由维护者执行独立本地入口。D9 的 AuthRoot 正匹配仅证明材料对应，不证明当前 CTL 接受。D10 先解决本机是否具备可检查的 CTL 和签名材料；没有校园入口，不重复握手，也不能关联已经释放的 D9 peer。它不实现 Windows 完整信任策略，不改变诊断或产品 TLS。

#### D10 来源与独立身份依据

1. 只读机器 64 位注册表视图中 `HKLM\SOFTWARE\Microsoft\SystemCertificates\AuthRoot\AutoUpdate` 的两个固定 REG_BINARY 值：`EncodedCtl`、`DisallowedCertEncodedCtl`。只查询这两个值，不枚举其他值/子键，不创建键，不触发同步。位置由 [Microsoft Certificates and trust](https://learn.microsoft.com/en-us/windows-server/identity/ad-cs/certificate-trust) 明确列出。缺失、类型错误、读取期间大小变化或超限分别停止该来源；不下载替代材料。
2. 证书候选仅来自这两个 CMS 内嵌证书和原固定 CU/LM ROOT；拒绝比较只用原固定 CU/LM Disallowed 的原始 DER。沿用 D8 只读 provider 和限额；不读取 CA、AuthRoot 证书缓存、额外策略库、MY/私钥、CRL 或任意文件。所有本机材料只在内存中存在，不导出，不跨运行复用。ROOT 是候选材料来源，不能整体成为本检查的可信锚集合。
3. 独立固定身份来自微软公开的 [2026 TRP 审计报告 Attachment A](https://www.microsoft.com/pkiops/docs/content/seals/Microsoft%20TRP%20WTCA%20Independent%20Accountant%27s%20Opinion%20Report%20And%20Management%20Assertion_FY26_Final.pdf)：Microsoft Root Certificate Authority 2010 的 DER SHA-256 为 `DF545BF919A2439C36983B54CDFC903DFA4F37D3996D8D84B4C31EEC6F3C163E`；Microsoft Certificate List CA 2024 为 `56625EC7886445706FE74CCD2ABF69B7C9C9CCDE3C7354903AF1C8F0E30E9D1C`。只接受这两个精确身份作为预定层级，不按名称、自签、系统信任或 CMS 自带根自证可信。公示指纹不是当前吊销状态证明；没有匹配材料就停止，不自动扩大白名单。
4. CA 2024 的公开用途包括 Code Signing 和 Windows Update；不能套用已于 2026-03-29 到期的 CA 2011 的 CTL 用途规则。固定层级只是首版受支持子集，不预断言本机当前 CTL 必须采用它。遇到其他层级、用途或结构，明确报告不支持，保留原件和 blocked。

#### D10 内存分析与结论边界

1. 使用 [CertCreateCTLContext](https://learn.microsoft.com/en-us/windows/win32/api/wincrypt/nf-wincrypt-certcreatectlcontext) 解码复制的内存输入，释放对应 context；不将其加入系统库。检查 CTL/CMS 类型、签名者和全部长度/指针后，使用固定签名者证书与 `CryptMsgControl(CMSG_CTRL_VERIFY_SIGNATURE_EX)` 验证消息签名。签名者模式只允许显式 CERT/PUBKEY，不用 CHAIN/NULL，不使用把外部 store 直接当可信的 trusted-signer 标志。不调用 CertVerifyCTLUsage、CertGetCertificateChain、WinVerifyTrust、系统 Verify、certutil 或更新器。
2. 首版每份 CMS 恰好一个 signer，按 signer identifier 唯一匹配内嵌证书；歧义、未支持的属性/算法或额外嵌套签名停止该项。只观察固定 Root 2010 → CA 2024 → signer 的两条证书签名边及 CMS 签名；不搜索泛化信任图，不用 ExtKeyUsageAny 伪装用途已验证。仅支持 RSA 2048–8192 位、SHA-256/384/512 和已明确测试的签名编码；其他表示为 unsupported，不能降级。证书有效期、基本 CA/KeyUsage 条件、未知关键扩展及原始 DER 拒绝命中独立分类；拒绝不能被成功签名覆盖。
3. 对签名覆盖的 CTL 内容分别报告：ThisUpdate/NextUpdate 相对本轮本机时间的窗口、SubjectUsage 是否为已识别类别、条目结构/属性是否在已检查范围。无 NextUpdate 为 unbounded，不算新鲜；单次 SequenceNumber 不能证明无回滚，窗口内不能证明最新。用途出现不等于完整策略授权，固定 `purpose_policy=not_established`、`revocation=not_checked`、`latest=not_established`；不输出完整信任通过的总判定。
4. 条目仅作有界结构检查，不关联 peer 或选择 TLS 根。SubjectAlgorithm 决定成员标识含义，不能把 SPKI 相同当作 CTL 成员；第一版仅识别 SHA-1 的证书 DER 索引形状，不能据此允许 SHA-1 消息/证书签名。条目属性存在、重复或未识别时显式标记未覆盖，不能解释为无约束。微软 [Trusted Root Program 测试说明](https://github.com/TrustedRootProgram/Program-Requirements/blob/main/Testing%20Instructions.md) 的 NotBefore、NotBeforeEKU、Disable 与 Remove 差异说明成员存在本身不足以证明可接受；本阶段不声称实现这些完整语义。
5. 两个值分别读取，不承诺原子快照。未签名/签名不成立的结构观察不得标为已认证内容；缺材料、解析/签名错误、拒绝、期限/资源失败都保持未知或不可用，不能把未检查输出为 false。完整检查也始终保留 authentication blocked，不恢复 TLS，不授权校园运行。

#### D10 预算、输出与离线验收

- 两个值各最多 4 MiB，累计 8 MiB；先查长度再单次读取，变化即停止，不无限重试。每份 CTL 最多 16384 条目、每条 32 属性、每属性 4 值、每值 64 KiB、OID 128 字节；累计被遍历属性字节最多 8 MiB。每份 CMS 最多 16 内嵌证书、单 DER 64 KiB、累计 1 MiB；不处理内嵌 CRL、递归 countersignature 或 timestamp。重复也计预算；异常不能截断后称 complete。
- 两份合计至多 2 次 CMS 验签、4 次证书边验签；固定签名边可去重但不增加工作量上限。10 秒协作期限，启动器对子进程另设 15 秒硬期限；原生解析不能承诺内部即时取消。解析 worker 在读取/解码前纳入专用 Windows Job，限制单进程内存 256 MiB、禁止衍生进程、关闭即终止；设置失败就停止，不无保护继续。句柄仅限本次子进程，不调整整机策略或其他进程。
- 独立 `ctl-check-dev` 包只提供 `--self-test` 与 `--ctl-check`。绑定验证及合成 self-test 在实际来源读取之前；没有 `--run-once` 或网络代码路径。报告区分完成观察与不可用，外层绑定状态独立；每个来源输出读取、结构、固定签名层级、CMS 签名、时间、用途覆盖、属性覆盖及拒绝状态的闭集枚举。实施时先锁定具体 JSON 字段及矛盾组合，再写实现；禁止任意字符串/错误原文、证书/CTL 内容、名称、指纹、数量、日期或环境路径。固定 operations 的网络、DNS、连接、profile、凭据、证书/注册表写入均为零。
- 合成验收使用新生成的证书和 CTL/CMS，测试专用 pin 仅经内部依赖注入，真实 CLI 不接受 pin/文件/URL 覆盖。覆盖正确固定链、伪根/同名异键、缺材料、篡改内容、错误签名者/多 signer、算法/用途不支持、有效期、拒绝命中、未知关键扩展、缺/过期更新窗口、条目属性与资源边界、取消和 Job 设置失败/强制超时。平台 API 测试只用内存合成材料，不读取真实注册表。
- 完成 Windows 聚焦测试/vet、Linux 纯分析回归/race、有界解码 fuzz、闭集与脱敏检查、绑定负例和独立开发包后，交付维护者一次原生 `--ctl-check` 步骤。它无需关闭 TUN、改变账号或校园独占窗口。实际材料不足时交付明确缺口，不现场下载根/CTL、不加入信任池、不把重新生成材料冒充原件。

上述新增固定 CTL 来源、固定签名层级检查及独立零网络包范围已整体确认，先实施离线验证与打包；Agent 不自动读取真实新增来源。后续若需要更完整用途/约束语义、额外材料或同轮 peer 对照，依据本次本地结果形成新的具体范围，不把本决定当作真实网络或扩大信任授权。

2026-10-29 维护者随后明确授权 Agent 自动执行当前零网络本地检查。本授权覆盖已审阅 D10 包的 `--ctl-check`，替代前述“必须维护者原生执行/Agent不自动读取”的执行安排；固定来源、只读、零网络与信任边界不变。执行须使用受支持工具权限、记录账户及工具上下文，不把工具环境与维护者原生会话默认视为同一观测面。不需要以浏览器刷新或校园账号状态作为本地检查前提；此授权不包含校园请求、登录/注销或扩大来源。

D10 输出锁定：schema_version=1、tool_kind=ctl-check-dev、mode=self-test/local-ctl-check，state=complete/unavailable，authentication=blocked；purpose_policy/latest 固定 not_established，revocation 固定 not_checked。固定顺序两个 sources（authroot_ctl/disallowed_ctl），各含 source、state、reason、read、structure、hierarchy、signature、certificate_time、conditions、rejection、window、usage、attributes。完整观察仅指固定层级、签名和已支持结构检查完成，不是可信判定。

编码子集锁定：legacy PKCS#7 SignedData.version=1、直接 CTL SEQUENCE；CMS SignedData.version=3、OCTET STRING 包装 CTL DER。两者 signer 均为版本1的 issuer/serial 唯一匹配；不支持 SKI signer、未签名属性、嵌套签名或 CTL extensions。按 [RFC 5652 §5.2.1](https://www.rfc-editor.org/rfc/rfc5652.html#section-5.2.1) 与 [RFC 2315 §9.3](https://www.rfc-editor.org/rfc/rfc2315.html#section-9.3)，前者 messageDigest 摘要 CTL SEQUENCE 的 value octets，后者摘要 OCTET STRING 内完整 CTL DER；封装确定唯一规则，不尝试替代摘要。signedAttrs 的 RSA 输入始终是 DER SET 编码，消息摘要、contentType 和属性唯一性均先检查；原生验签失败不能被纯 Go 合成结果覆盖。用途类别中 Root List 为 .311.10.3.9，Disallowed List 为 .311.10.3.30，其他值仅为 other，不等同于 signer EKU。

读取为 complete/absent/unavailable/not_checked；结构 valid/unknown；层级 observed/unknown；签名 verified/unknown；证书时间 valid/invalid/unknown；有限条件 satisfied/unknown；拒绝 none/present/unknown；窗口 within/expired/future/unbounded/invalid/unknown；用途 root_list/disallowed_list/other/unknown；属性 absent/present_uninterpreted/unknown。未完成的检查保留 unknown，不能因其他阶段成功填充。reason 闭集 none/absent/read_failed/type_invalid/size_changed/limit/parse_failed/unsupported/material_missing/signer_ambiguous/signature_failed/certificate_time_invalid/conditions_rejected/denied/store_unavailable/canceled/deadline/protection_failed；失败按已完成阶段保留有限事实，state=unavailable，完整两项才有顶层 complete。operations 固定七项全零：network_requests、dns、connects、profile_reads、credential_reads、certificate_writes、registry_writes。拒绝重复/缺失/未知字段、越界枚举、非零操作及矛盾阶段；不回传底层错误。

#### D10 解析停止的阶段分类

维护者原生回执显示两个固定值均读到但 parse_failed；当前不能定位到包封装、CTL字段、证书DER或签名属性，不能据此诊断材料损坏。本修订仅在既有 reason 中细分原解析拒绝分支；来源、预算、接受条件、签名次数、输出字段和零网络边界不变。旧 parse_failed 保留解码，不追溯重标旧回执。新增原因均以 parse_failed_ 开头，后缀闭集为 envelope、signed_data、content_info、content、ctl_header、ctl_usage、ctl_identifier、ctl_sequence、ctl_time、ctl_algorithm、ctl_entries、ctl_entry、ctl_attributes、ctl_attribute、ctl_values、ctl_tail、certificate、signed_attribute_order、signed_attribute、signed_attribute_identifier、content_type、signing_time、required_attributes。

该类别只说明当前解析器在哪一步拒绝，并不区分所有不支持格式、损坏或资源边界；不输出原始字节、OID、长度、计数、日期、错误文本或证书标识。上述停止仍必须 structure/hierarchy/signature 等分析事实均 unknown，state=unavailable；取消/期限优先并清空未完成事实。用合成拒绝分支和闭集/取消回归验证后生成独立修订包；不以现场输入试探放宽解析，不新增缓存读取时机。

进一步区分两个既有停止分支：parse_failed_ctl_algorithm_unsupported 表示 AlgorithmIdentifier 在现有 NULL/省略参数子集中解析成功但不是唯一支持的 SHA-1 成员索引；不回显它的算法值。parse_failed_ctl_extensions_unsupported 表示其余 CTL 字段通过后只有一个尾部 [0] 字段，尚未解析/验证其扩展内容。两类仍停止，不将“不支持”当作已认证或允许接受，也不证明扩展有效。

### D11 已确认：本地 CTL 结构兼容与签名观察

2026-10-29 维护者确认完整实施本工作包。仅扩展独立诊断的结构支持；Windows 本地读取/原生验签与纯 Go 分析保持隔离。产品 SDK/CLI 的依赖闭包不得包含 ctlcheck、casacquire、trustaudit 或诊断启动器；Linux/macOS 不读取或依赖微软 CTL。交叉构建不是原生信任验收。

1. SubjectAlgorithm 仅允许 SHA-1、SHA-256、SHA-384、SHA-512，成员标识分别精确为20/32/48/64字节；参数仅省略或NULL。未知算法/参数停止；不增加成员查询，不放宽CMS/证书签名算法。
2. 支持单一尾部 `[0] EXPLICIT Extensions`：内部为SEQUENCE，每扩展为OID、可选DER BOOLEAN critical、OCTET STRING value。critical省略为false；显式BOOLEAN仅接受DER true（FF），拒绝显式默认false及非规范编码。值作为opaque bytes，不递归解释。每CTL最多32项，OID最多128字节、单值最多64KiB；扩展完整输入计入原8MiB累计遍历预算，不增加总限额。重复OID、畸形、尾随数据、超限和取消停止本来源；不发布部分结果。空扩展序列按无扩展观察。
3. 非关键/关键CTL扩展均可继续固定签名观察，关键语义明确未解释，不能用于信任接受。证书自身未知关键扩展仍拒绝。条目属性省略形式只有经Windows原生合成解码/验签证实后纳入；无独立依据则保留拒绝，记录缺口。原条目属性检查保持独立。
4. 新报告schema_version=2，新增每source必填member_algorithm（unknown/sha1/sha256/sha384/sha512）与extensions（unknown/absent/present_uninterpreted/critical_present_uninterpreted）；attributes保持条目属性原义。新增观察值仅在固定层级及CMS验签完成后发布，否则unknown。complete只表示有限观察完成，不解释扩展政策；authentication/purpose_policy/revocation/latest的D10限制保持。新解析原因仅增加parse_failed_ctl_extensions、parse_failed_ctl_extension、parse_failed_ctl_extension_duplicate、parse_failed_ctl_extension_limit；旧原因不追溯修改。
5. 解码器按版本严格验证v1/v2完整字段集合；v1不带新增字段、不补造观察值，原封装及历史原因保持。v2必有新字段；拒绝混配、缺失、重复/未知字段、非法枚举和阶段矛盾。外层开发包绑定回执schema维持1；其result携带独立v2报告，产品JSON不变。
6. 保持原来源、固定pins、读取和签名预算、10秒协作/15秒硬期限及256MiB单进程Job。实施顺序为合成结构/签名交叉验证、Windows/Linux回归及race/fuzz、闭集和绑定审计、六平台产品构建/依赖隔离检查、定向扫描，随后Agent通过受支持工具自动本机零网络复核。无需维护者复制命令、切换TUN或刷新门户；记录账户/上下文，不默认工具与桌面会话等价。
7. 同包同结果不重复运行；范围内缺陷自动修复，只有代码或相关输入变化且有信息增益才复核。约定外编码、签名属性/身份、材料缺失、原生/Go不一致或保护失败集中报告并保持未完成，不自动扩源、改信任或联网。旧包和历史快照保留，password_core仍blocked/live_unverified。

D11本机复核后的分类细化：首轮AuthRoot仅返回unsupported，不能定位CMS停止阶段；Disallowed仍为四种成员算法之外。仅将既有CMS unsupported分支细分为v2闭集原因 unsupported_envelope、unsupported_signed_data、unsupported_digest_set、unsupported_digest_algorithm、unsupported_content_encoding、unsupported_signer_info、unsupported_signer_identifier、unsupported_signer_digest、unsupported_signed_attributes、unsupported_signed_attribute_values、unsupported_signed_attribute_unknown、unsupported_signature_algorithm、unsupported_signature_size。保留原unsupported解码及非CMS用途；v1不接受这些新原因。不改变算法/属性/编码接受集合，不输出OID或原始错误，不增加来源、验签或网络操作。

### D12 已确认：独立零网络 CTL 标识辨识

状态：2026-10-29维护者明确选择“同意推荐：有界OID辨识并自动完成本机复核”。D11已将两个来源定位为未知CMS签名属性/成员算法；现有报告刻意不输出其OID，所以不能据此决定增加哪种支持。不继续试加白名单，也不重跑相同包。

官方依据：[CertFindSubjectInCTL](https://learn.microsoft.com/en-us/windows/win32/api/wincrypt/nf-wincrypt-certfindsubjectinctl)区分成员索引与签名，历史支持SHA-1/MD5；[微软SDK wincrypt.h](https://github.com/microsoft/win32metadata/blob/main/generation/WinSDK/RecompiledIdlHeaders/um/wincrypt.h)另定义szOID_DISALLOWED_HASH为证书签名哈希属性OID。这些只证明四种SHA算法不穷尽可能标识，不能认定本机实际值。[RFC5652](https://www.rfc-editor.org/rfc/rfc5652.html#section-5.3)定义签名属性集合及必需属性，允许额外属性；出现未知属性不自动等于坏签名，也不授权忽略其语义。

1. 新增独立开发工具 `ctl-identifiers-dev`，仅 `--self-test` 与 `--ctl-identifiers`。同两固定注册表值只读、先Job保护后读取；不读ROOT/Disallowed证书库，不解析证书，不调用CryptoAPI验签/链引擎，不执行散列成员查询。CMS嵌入证书字节可能随整个值读入内存，但不解析或输出，不能称完全没有读取证书字节。现有ctl-check接受规则、pins和报告保持。
2. 只提取每份CTL的SubjectAlgorithm OID，以及唯一SignerInfo中全部signedAttrs属性OID；不取属性值、条目标识、证书、签名、公钥、日期、名称或原始响应。OID允许未在本地字典中识别的规范数字点分串，明确属于未经签名认证的输入观察；未知OID不保证是公开注册标识，不将其当作指令、URL或自动查询地址。
3. 复用已审阅DER游标/长度检查，定位已支持的legacy PKCS#7 v1直接CTL或CMS v3 OCTET STRING封装、唯一v1 issuer/serial SignerInfo。定位SubjectAlgorithm时不按D11算法白名单提前退出；不把跳过的CTL条目、属性值或证书内容标为已验证。签名属性最多16项，检查DER顺序、OID唯一和属性外层OID/非空SET结构；每SET最多4个有界不透明值、每值及SET各最多64KiB，不递归解释。未知封装、多个signer、结构异常、重复OID、超限或取消使本来源全部标识清空，不发布部分列表。
4. 两个值各最多4MiB，读取方式仍一次长度查询和一次精确读取，无重试；每OID DER内容最多128字节、规范文本最多256 ASCII字符，必须可按ASN.1 OID重编码一致。每份最多一个成员OID及16个属性OID，固定两个来源，总输出最多16KiB。10秒协作期限、15秒硬终止、256MiB单进程Job和禁止衍生进程保持。值变动/类型错误/保护失败停止，不调整整机配置。
5. 独立闭集报告schema_version=1、tool_kind=ctl-identifiers-dev、mode=self-test/local-ctl-identifiers、state=complete/unavailable；authentication=blocked、signature=not_checked、input_authenticity=unverified固定。sources固定顺序authroot_ctl/disallowed_ctl，每项source/state/reason/member_algorithm_oid/signed_attribute_oids；完整项reason=none，成员OID非空，属性列表完整并按数字OID弧顺序排序；不可用项成员OID为空字符串、属性列表为空数组。reason只允许none/absent/read_failed/type_invalid/size_changed/limit/parse_failed/unsupported/canceled/deadline/protection_failed。operations沿用固定七项全零。complete只表示限定标识提取完成，不代表完整结构/签名/信任成立。
6. 报告是原“禁止回显OID”的明确、有限例外：只允许上述两个角色的数字标识及列表长度所隐含的属性数量，禁止其他任意字符串/原始错误/值。不写入原始CTL或证书；单独开发包和回执保留在开发目录，旧包不覆盖。
7. 确认后连续完成纯内存合成输入、两封装/算法与属性跨白名单、重复/畸形/超限/取消/防泄漏、Windows保护执行与整链绑定、Linux race/fuzz、文档和定向扫描，随后自动执行一次已绑定新包的本地入口。再依据实际OID查官方定义，集中交付两项身份与后续支持建议；不自动接受新算法/属性或切换到验签、更不执行校园请求。权限/来源/信任边界不另作推定。

### D13 已确认：固定两种表示的有界签名观察兼容

状态：2026-10-29维护者明确选择“同意推荐：实施D13并自动完成离线验收和本机复核”。D12已取得成员/属性OID，但未认证输入、未解释值或重新验签。按以下完整范围连续实施、合成验收、独立打包和一次自动本机零网络复核。目标是继续有限签名观察，允许结果仍为unavailable，不以获得成功为验收条件。

事实依据：[微软SDK wincrypt.h](https://github.com/microsoft/win32metadata/blob/main/generation/WinSDK/RecompiledIdlHeaders/um/wincrypt.h)将 `1.3.6.1.4.1.311.10.11.15` 定义为signature-hash属性并别名为disallowed hash；[CertGetCertificateContextProperty](https://learn.microsoft.com/en-us/windows/win32/api/wincrypt/nf-wincrypt-certgetcertificatecontextproperty)及[CryptHashToBeSigned](https://learn.microsoft.com/en-us/windows/win32/api/wincrypt/nf-wincrypt-crypthashtobesigned)不能作为把所有实际成员统一当整证书DER哈希、或仅凭OID硬编码一种长度的依据。[微软SpcSpOpusInfo说明](https://learn.microsoft.com/en-us/openspecs/office_file_formats/ms-oshared/91755632-4b0d-44ca-89a9-9699afbbd268)说明SEQUENCE及可选描述/链接字段，但这是Office格式上下文，不能直接外推为本机CTL所有字段语义。[RFC5652 §5.4–5.6](https://www.rfc-editor.org/rfc/rfc5652.html#section-5.4)要求签名覆盖完整signedAttrs，支持把字节覆盖检查与属性用途解释分开；不授权忽略任何安全策略。

1. **来源和用途保持。** 仍为ctl-check-dev独立Windows辅助诊断，原两固定CTL值、ROOT/Disallowed证书材料、固定签名层级、原生内存验签及拒绝检查不变。产品SDK/CLI不接入诊断包；Linux仅纯Go合成验证，macOS交叉构建不视为原生信任验收。禁止新增来源、联网补证书、Windows链引擎、信任变更、成员查询、校园请求或认证。
2. **固定SPC属性的有限容纳。** 仅增加signedAttrs OID `1.3.6.1.4.1.311.2.1.12`，仍恰好一项属性、恰好一个SET值；值必须是单个完整、确定长度的SEQUENCE，完整编码最多64KiB、无外部尾随字节。SEQUENCE内容保持opaque：不递归解码、不提取/打印描述、链接或其他字段、不访问任何引用，不声称内部ASN.1或用途语义已验证。其他未知属性仍停止；必需content-type/message-digest、DER集合排序、重复拒绝与既有signing-time规则保持。验签继续使用原始完整signedAttrs字节，禁止删除SPC属性后重算、重新编码后尝试或只验已知属性。
3. **固定Disallowed属性索引的有限容纳。** 仅在disallowed_ctl来源且CTL SubjectUsage为唯一disallowed_list时增加成员AlgorithmIdentifier OID `1.3.6.1.4.1.311.10.11.15`；参数仍只允许省略/NULL。每个条目标识仍须OCTET STRING，非空且最多64字节，作为不透明字节通过有界遍历；1–64是项目资源上限，不是官方哈希长度集合、不证明这些字节是合法signature hash。禁止尝试算法、计算或匹配真实成员；未知OID及其他用途保持拒绝。原四种SHA算法保留各自精确长度检查，不能被该分支放宽。
4. **预算与保护。** 保留每值4MiB、两值8MiB、16384条目、32条目属性、16签名属性、每值64KiB及累计8MiB遍历预算；新增成员字节/SPC值计入累计预算，不额外建立不受限副本。保留两份最多2次CMS及4次证书边验签、10秒协作期限、15秒硬终止、256MiB单进程Job。外层畸形、重复、超限、取消、原生拒绝均停止相应来源；明确不检查的opaque内部不能称结构已验证。
5. **报告v3消除过强结论。** 新ctl-check报告schema_version=3；v3的structure仅unknown/bounded_checked，后者只表示本规范列出的结构子集与预算检查通过，不沿用可能误读为完整校验的valid。member_algorithm新增signature_hash_property枚举；新增member_identifier_coverage（unknown/exact_length_checked/bounded_uninterpreted）与signed_attribute_coverage（unknown/known_subset_checked/spc_opus_outer_only）。新覆盖字段仅固定层级和CMS验签完成后发布；之前unknown，不把OID身份当认证结果。四SHA算法对应exact_length_checked，属性索引对应bounded_uninterpreted；SPC存在时必须spc_opus_outer_only。attributes仍仅指条目属性，extensions仍仅指CTL扩展。v1/v2按各自原字段/枚举严格解码和编码，旧回执不补造覆盖结论；拒绝版本混配、未知/缺失/重复字段及矛盾状态。失败继续使用既有闭集阶段原因，不增加任意文本、OID或属性值输出。外层包绑定schema不变，D12独立报告不变。
6. **验收与自动执行。** 先以新生成证书和两种封装的合成CTL建立正确/篡改/重复/外层畸形/参数/用途/空标识/边界/取消矩阵。SPC内部只验证保持opaque及不泄漏，不将合成测试误称内部语义验收。Windows原生内存解码/签名与Go分别检查新增表示；任一路拒绝不能被另一路成功覆盖。覆盖两新增表示单独及组合、签名篡改SPC值/成员值、原四SHA回归、v1/v2/v3兼容、Job/硬期限和绑定拒绝；Linux执行race/vet/有界fuzz，六平台执行产品依赖隔离与开发构建。完成文档/定向扫描和新快照绑定后，Agent自动运行一次新包--ctl-check，记录账户、上下文、包身份和脱敏结果，不要求维护者操作终端、切换TUN或刷新门户。
7. **停止与交付。** complete仅表示以上有限观察完成；authentication=blocked、purpose_policy/latest=not_established、revocation=not_checked始终保持。固定材料/签名者不符、原生/Go不一致、保护无法成立或约定外表示集中报告；不现场扩范围或重跑同包同结果。保留D9–D12历史包、Candidate及历史handoff/status，不提交、推送、发布。下一步如需解释SPC内部、实现属性索引/CTL成员或信任策略，必须独立决策，不能由D13成功隐含授权。

### D14：自动推进纯合成成员与策略分析

2026-10-29维护者要求自动继续，不再逐开发步骤确认。执行约定据此更新：Agent自行完成下一离线工作包的事实核对、最小实现、合成回归及开发交付，不重复询问PowerShell或测试权限。grilling用于内部审查假设，常规实施选择不再逐项交给维护者；原校园网络、凭据、信任和发布边界不变。

本工作包只处理新生成证书/CTL，先用Windows受限子进程的内存API确定signature-hash属性的字节规则，再实现可跨平台测试的有限成员关系观察。不能根据文档简称猜哈希对象或根据实际缓存尝试算法。原生对照不读证书库、不读真实属性、不调用链引擎；不重跑D13实际来源。事实不成立则保留该分支unsupported并继续其余合成工作。

分析必须以同一输入的签名检查为前提；未认证、错用途、重复/歧义条目、超限、取消或不支持表示不能给出肯定成员结论。命中/未命中均不等于信任接受或拒绝，未知条目约束、CTL扩展、无更新上界与未建立最新性必须保留。实际来源读入、同轮peer接入、信任池变更与校园握手不在D14执行范围。产品SDK/CLI及D13报告契约保持。

D14实施子集：固定authroot_ctl/root_list/SHA1使用完整证书DER的SHA1；固定disallowed_ctl/disallowed_list/signature-hash属性仅支持生成对照已验证的RSA/SHA256、384、512，使用对应TBS编码散列。两分支证书DER均最多64KiB；Disallowed分支还要求内外签名AlgorithmIdentifier一致、RSA签名参数仅省略/NULL，其他表示拒绝。成员关系不验证目标证书签名、用途或有效期，因此即使命中也不输出可信结论。整表最多16384条，重复标识（包括未命中条目）使关系unknown；不截断、不按SPKI或同名匹配。CTL仍按D13验证原始签名及固定层级，复用已检查内容定位条目；每CTL4MiB、10秒上下文，取消清空关系。原型仅保留命中项条目属性的覆盖分类与CTL扩展/窗口分类，不解释任何策略。

独立`ctl-member-synthetic`命令只接受`--self-test`，不接收文件/URL/证书/标准输入、真实来源或网络模式。只生成内存证书和CTL；Windows先建立既有256MiB单进程Job，开发runner对子进程设15秒硬期限，Linux仅纯Go。输出固定schema_version=1、tool_kind=ctl-member-synthetic、mode=self-test、state=complete、authentication=blocked、trust=not_established及固定合成检查名称，七项operations全零；失败无部分成功JSON并非零退出。新开发目录保存源码快照、单一合成EXE及摘要；没有本地材料读取入口，无需仿造真实诊断启动器或安装流程。

### D15：同轮peer成员观察的离线实现与待执行包

维护者要求继续自动开发。Agent连续完成本节实现、假来源/loopback验证及独立新包，不逐实现步骤询问。此授权不执行新校园握手；新包仍由原单次确认入口启动真实模式。D9–D14旧包及产品认证行为保留。

1. 在既有拓扑检查之后、任何DNS/TCP之前，只读D13同两个CTL值及同ROOT/Disallowed材料，固定签名层级和签名算法保持。建立256MiB单进程Job后读取；CTL准备总计10秒协作、15秒进程硬停止，每值4MiB、两值8MiB、最多两次CMS及四次证书边验签。任一来源不可用、索引不支持、重复成员或取消则零网络停止。不新增来源、系统链引擎或信任。
2. 两来源均通过后仅保留有界成员表和覆盖分类；不保留原始CTL、签名证书或原始错误。成员表各最多16384项，重复标识整表拒绝。AuthRoot仅root_list/SHA1；Disallowed仅disallowed_list/已确认属性索引。认证后的内容用于建表，不接收外部“已可信”声明。未知属性及扩展保持未解释，window不转换为信任。
3. 原匿名网络上限与有线绑定保持。只在同轮unknown-authority失败连接显式关闭后比较该错误携带的peer链；最多16张、每张64KiB、合计512KiB，10秒协作期限。按D14有限规则计算成员键，两个来源分别汇总any_match/no_match；不回显位置、计数、名称、DER或哈希。不支持证书表示、结构错误、超限、取消使比较不可用并清空关系。输入只来自本轮内存，不接受历史peer或文件。即使命中也不继续TLS/GET，D9原基线TLS成功仍在HTTP写入前停止。
4. 新采集回执仅增加可选闭集`ctl_membership`，含phase=gate/comparison、state=complete/unavailable/not_run、reason、authentication=blocked、trust=not_established及固定顺序两来源。每源只输出source、state、reason、relation、entry_attributes、extensions、window；未完成项观察均unknown，完整比较仅any_match/no_match及对应覆盖。未到失败peer比较的窗口明确not_run，不能填no_match。旧D9及更早回执保持原形并严格解码，显式null扩展/缺失字段/未知字段/矛盾状态拒绝。
5. 纯合成测试覆盖签名后建表、两封装、正确/缺失/重复成员、错用途、Unsupported目标、预算/取消、原生拒绝、部分结果清除及闭集解码。loopback验证准备失败零DNS、失败连接先关闭再比较、无HTTP写入，成功TLS不继续请求。Windows/Linux回归、race/vet/fuzz、启动器绑定和产品隔离完成后交付独立包。离线阶段不调用真实准备器或真实RunOnce；真实结果仍不能替代password_core。
