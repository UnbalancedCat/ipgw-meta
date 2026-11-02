---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
status: approved
target_release: v1.0.0
release_policy: gate_based
---

# IPGW-Meta v1 分阶段执行与远程验收计划

本文是 `IPGW-META-V1` 的权威执行计划。发布以门禁为准，不以日期为准。2026-10-28 维护者明确取消本机离线开发的 30–45 分钟强制停顿：在已授权范围内持续实施、验证和记录，只有完成范围、实际阻塞或需要新的授权时停止。真实网络及高风险操作仍遵循独立审批边界。

<a id="roadmap-20261029"></a>

## 0. 2026-10-29 当前认证收敛 Roadmap

本节是当前开发优先级与后续工作包的唯一 roadmap。主线为：**以旧版可用流程作行为对照，修复现有 v1 共享认证实现，完成真实登录闭环，再完成正式发布验收。** 不回退到另一条产品线，不预设旧版只能依靠不安全行为工作。

权威源码目录为 BHKDesktop `D:\project\Go\IPGW-Meta`；Windows 原生与同机 Ubuntu WSL 使用已经建立的开发环境，不依赖 BHKLAPTOP，不重复迁移或初始化。当前事实与证据等级见[实施状态](status.md#status-20261029)，执行索引见[工作包索引](../../agent/plans/stabilization-v1.md#auth-roadmap-work-packages)。

### 0.1 当前有效规则、已确认目标与切换条件

2026-11-02 统一源码准备入口见[检查点记录](../research/unified-candidate-preparation-2026-11-02.md)。现行目标以 ADR-0026 为准：认证/公开摘要/promotion v3、辅助 v2；有线 password、有线 QR、NEU password、NEU-2.4G password 四组认证，以及安装、NEU 网络、NEU-2.4G 网络三份辅助。下表及 0.3 的三组/v2 描述保留为 10 月 29 日至 P4 初版历史，不是新 Candidate 的验收矩阵。

| 层次 | 当前解释 |
|---|---|
| 当前有效约束 | 现有安全、协议、正式 Candidate/live-gate/evidence/promotion 规范及工具继续生效；正式验收迁移以 [P4 v2 增量契约](../operations/bhk-acceptance.md) 为准，下文四组规则保留用于 v1 历史解释。历史执行步骤不构成重新迁移、重建或真实操作授权。 |
| 已确认目标 | 开发认证验证与正式验收分开；未来诊断采用产品默认系统 PKI；BHK 独立完成三组正式认证证据，NAS 不再是目标路线的强制依赖。 |
| 切换条件 | P2/P4 的对应规范、ADR、实现与回归同步完成并验收后，分别切换其规则。此 roadmap 的批准不表示现有工具已经支持目标，也不自动接受尚为 proposed 的 ADR。 |

2026-10-29 后续执行：P1 行为对照及 P2 实现/专项离线验收已交付。独立开发入口现按 [ADR-0017](../architecture/decisions/ADR-0017-development-auth-validation.md) 和[开发验证契约](../operations/development-validation.md)生效；这不授予真实运行权限。Windows 安装/junction 的历史缺口及 P4 后续复核见[当前状态](status.md#status-20261031-p4)。2026-10-31 [P4 v2 契约](../operations/bhk-acceptance.md) 的实现与离线验收已交付，报告及未覆盖项见 [P4 记录](../research/p4-bhk-gates-2026-10-31.md)；旧四组仅用于 v1 历史审计，当前发布输出禁止混入 v1。下一阶段为单独授权的 P5，新 Candidate 及正式证据尚未取得。

D4–D15 证书与 CTL 专项诊断退出默认开发主线，代码、包、历史决策和结果全部保留。仅当新证据表明其直接阻塞产品认证，且明确了待区分的认证假设、所需观察与退出条件时，才重新安排相应工作。诊断安全与解析覆盖测试通过不等于服务端流程正确；受限诊断 TLS 失败也不能直接登记为产品默认系统 TLS 缺陷。不得默认继续扩展自建 Windows 信任引擎或通用 JavaScript 执行器。

2026-11-01 执行更新：P4 残余 symlink 覆盖、新源码合入及正式 Candidate 构建/来源核验已完成，覆盖上段 10 月 31 日“新 Candidate 尚未取得”的历史状态。P5 当前进入同一 Candidate 的 BHK 安装辅助验收，之后才是另行授权的三组认证和 Wi-Fi 辅助窗口；具体身份与未完成项见 [P5 记录](../research/p5-readiness-2026-11-01.md#正式-candidate-与本机验收)。最终签名 tag、promotion 发布及公开 release 不在本次授权范围。

### 0.2 工作包、依赖与退出条件

2026-11-02 最新下一步：QR 开发五步闭环已通过，退出码 0、最终 offline、无门禁失败，见[回执核验](../research/qr-continuation-progress-2026-11-02.md#后续窗口qr-开发闭环通过)。结合既有双 Wi-Fi 开发通过，转入统一源码差异、测试覆盖与剩余缺口审阅，准备新正式 Candidate 检查点；不再重复同包开发扫码。提交/推送/合并、正式 CI 构建、四组认证及三份辅助的真实验收仍需对应授权；旧 Candidate 和开发证据不能移作新 Candidate 证据。旧门禁失败具体原因仍未知，正式 QR 与 M3 不提升。以下同日“当前/下一步”保留为本次通过之前的历史执行记录。

2026-11-02 当前下一步：门禁首因定位新包已完成离线验收和零网络 ready 预检，见[交付](../research/qr-continuation-progress-2026-11-02.md#当前交付门禁首因定位-v2)。只有重新建立目标账号 offline 及测试路径、取得新的单次授权后才进行真实 QR 验证；不自动重试或注销。旧窗口登录成功和注销未闭合事实保留，不提升正式状态。

2026-11-02 后续门禁修复：断开接口残留默认路由误拒绝已修正，新 QR 开发包零网络 preflight ready，见[当前入口](../research/qr-continuation-progress-2026-11-02.md#后续更新残留默认路由修复及当前可测试包)。后续单次窗口已结束：QR 命令和独立目标 online 核对通过，注销阶段门禁 blocked，最终会话未知。下一步离线核对门禁停止原因，不自动重试；详见[窗口结果](../research/qr-continuation-progress-2026-11-02.md#最新窗口登录通过注销阶段门禁停止)。下段旧包 topology_unavailable 不再作为当前阻塞，不因此提升真实能力或正式验收。

2026-11-02 当前 QR 执行入口：[ADR-0028](../architecture/decisions/ADR-0028-qr-service-continuation.md) 的无 ticket service 接续与独立普通产品开发包已完成离线实现。两端回归、Linux race/fuzz、六平台构建和包绑定通过；本机零网络 preflight 为 topology_unavailable，未进行真实请求。完整结果及未覆盖项见[实施记录](../research/qr-continuation-progress-2026-11-02.md)。下一步是重新建立并确认有线测试路径、另行授权一次私有终端 QR 开发窗口；不得沿用浏览器代理路径或历史地址。真实 QR 结果明确后才统一新 Candidate，重新取得全部正式证据。以下按日期保留此前调查与顺序记录，不代表仍需重复采集。

2026-11-01 QR 主线恢复：双 Wi-Fi 开发验证后，维护者批准 [ADR-0027](../architecture/decisions/ADR-0027-browser-qr-contract.md)。先核对网页契约与固定版本 QR 实现，再决定最小修复；旧失败摘要可能合并 QR 后的 status 错误，不能定位为轮询失败。离线准备、缺失材料及剩余授权见[QR 研究记录](../research/qr-browser-contract-2026-11-01.md)。后续无痕桌面扫码对照须单次授权；产品普通状态查询确认目标身份后才注销，不网页注销或手机抓包。当前无已证实产品根因，不生成猜测修复包，不沿用历史真实授权。

2026-11-01 后续需求（仅记录，暂不实现）：维护者提出校园网有线接入自有无线路由器时，其下游 SSID 可以自定义。未来应支持显式指定此类测试路径，不把学校公布的 SSID 集合等同于所有可用接入方式。后续工作须区分桥接/NAT、源地址与校园侧会话身份、状态/注销影响范围及证据场景，不能只增加任意名称白名单就宣称等价。该需求不改变本轮 NEU、NEU-2.4G 验收范围或当前门禁；公共产品枚举/扫描本身不按校园 SSID 白名单隐藏接口。

2026-11-01 双 Wi-Fi 后续：维护者批准 [ADR-0026](../architecture/decisions/ADR-0026-bound-scan-dual-wifi.md)。定向扫描、SSID、认证 v3/辅助 v2/promotion v3 迁移及独立开发包已离线交付，见[实施记录](../research/wifi-bound-progress-2026-11-01.md)。NEU、NEU-2.4G 两个单独授权的开发窗口均已回传辅助及 password_core 全 pass，明确非发布证据；双 Wi-Fi 开发验证已闭合，既有离线检查限制继续单列。NEU-Mobile 与自定义 SSID 暂缓；下一发布依赖是 QR 收敛，之后再统一新 Candidate。当前目标为四组认证和三份辅助，覆盖下方旧三组目标；正式验收尚未完成，不继承开发通过记录。旧四组/三组记录按日期保留。

2026-11-01 最新执行顺序：维护者选择暂缓 QR，先验收同一 Candidate 的 Wi-Fi 辅助及 password_core。先备齐本地文件与断网说明；断开有线后，一次辅助完整通过才进入一次密码套件，失败停止，恢复管理连接后再核验回执。QR 保持未完成的发布必需项，不以调整顺序免除门禁；下段“先处理 QR”的记录保留为调整前安排。

2026-11-01 QR 后续：第一次正式有线 terminal_qr 在扫码后的登录步骤返回 protocol_changed，尚未建立最终会话状态；先处理 [QR 协议证据缺口](../research/p5-readiness-2026-11-01.md#有线正式-terminal_qr-失败回执)，不重复同包或自动推进 Wi-Fi。下段是 QR 窗口之前的通过项与顺序记录；有线密码及安装通过保留，后续修复若改变 Candidate 则不能跨 Candidate 复用证据。

2026-11-01 验收进展：同一正式 Candidate 的 BHK 安装辅助及有线 password_core 已通过。下一待授权项为有线 terminal QR，之后完成 Wi-Fi 辅助与 password_core；三组认证、两份辅助摘要及 promotion 人工审阅尚未全部闭合。依据见 [P5 回执](../research/p5-readiness-2026-11-01.md#有线正式-password_core-回执)，不重复已通过的有线密码窗口。

| 阶段 / WP | 依赖 | 工作与交付 | 退出条件 |
|---|---|---|---|
| P0 / `WP-AUTH-ROADMAP-DOCS` | 已批准本次文档方案 | 更新本计划、日期化状态摘要、执行索引和最小导航；保留历史。 | 当前规则与目标规则区分明确；doccheck、链接及定向脱敏检查通过；本次到此交付，不执行 P1–P5。 |
| P1 / `WP-AUTH-BASELINE` | P0 | 本地 v0.1.3 为主要对照，归档桌面版本为辅助；核对匿名入口、ac_id/service、TLS、Cookie 会话、页面/脚本取键、加密、表单、重定向、activation、最终状态。 | 形成逐项“相同 / 差异 / 未知”矩阵，关联源码与可区分假设的验证方法；未证明的根因保持未知。 |
| P2 / `WP-AUTH-VALIDATION-PATH` | P1 | 先更新对应协议、安全、开发验证/evidence 规范及 ADR，再实现产品等价的最小观测与源码/哈希绑定的开发验证路径。 | 合成回归证明观测不改变请求语义、凭据处理和成功判据；开发回执与正式 Candidate evidence 分开；当前有效规则的切换有完整验收记录。 |
| P3 / `WP-AUTH-MINIMAL-FIX` | P1、P2 | 对独立复现和规范共同支持的缺陷增加回归并最小修复共享认证；取键语义变化先形成规范决策。 | 正常与负面合成场景通过；另行授权的开发认证验证完成目标账号状态闭环。失败按环境、实现或协议证据不足分类，不放宽期望或安全限制求通过。 |
| P4 / `WP-AUTH-BHK-GATES` | P0；可与 P1–P3 离线工作并行，协调共享文档 | 同步迁移发布/实验室/真实验证/认证能力/evidence 规范、相关 ADR、runner/schema、promotion 及回归，落实下述三组目标。 | 三组 v2 认证及安装/Wi-Fi 两项辅助门禁被各层一致接受；helper 自绑定、测试路径监测、历史 v1 只读审计与负面回归闭合；Windows 安装/junction 离线缺口解决。详见 [P4 契约](../operations/bhk-acceptance.md)，不把历史 Linux/NAS 结果改写为 Windows。 |
| P5 / `WP-AUTH-CANDIDATE-ACCEPTANCE` | P3、P4；现行其他发布 hard gate | 完成全仓质量门禁、六平台构建及原生安装验证；经后续授权流程进入受保护 main 和正式 CI，产生并验收新不可变 Candidate。 | Candidate 身份、正式 evidence、promotion 与签名发布条件全部闭合后才更新对应里程碑；开发登录成功不能替代正式验证。 |

P1 先做只读行为对照，不默认运行旧版登录或读取旧凭据。历史研究引用与当前 tag 的提交映射另行核对；不能唯一识别最早 Gemini 版本时保留不确定性，不因名称不确定停止可进行的源码比较。P3 不预设首键、第二键正确，也不预设所有多键页面均应拒绝；必须核对实际使用关系、会话和提交上下文，不能仅凭 HTTP 500 断定公钥错误。

2026-10-31 执行补记：动态取键修复及 [ADR-0024](../architecture/decisions/ADR-0024-logout-explicit-callback.md) callback修复后，BHK有线开发password_core七步已在单次私有窗口通过，见[开发闭环回执](status.md#status-20261031-password-dev-pass)。P4正式验收规则迁移及离线验收也已交付，接下来的主线为按P5形成新的不可变Candidate并验收；不把开发包或离线门禁通过当作P5完成，不默认重复浏览器或CTL诊断。本补记不授权提交、发布或新的真实网络窗口。

### 0.3 已锁定的验证目标（P2/P4 入口已交付，P5 正式证据待取得）

**开发验证与正式验证分开。** 开发包用于快速检验修复，绑定精确源码快照与产物哈希，回执明确标为非发布证据，不提升 Candidate 能力状态。凭据由维护者在私有 TTY 输入，Agent 不处理凭据。成功仍以 [PROTO-LOGIN-001](../architecture/protocol-correctness.md#proto-login-001登录成功不变量) 为准，并验证相应幂等与注销后 offline 状态；不得用公网连通性替代账号状态。正式验收继续使用不可变 Candidate，不复用开发字节冒充既有 Candidate。P2 先建立相应规范与工具支持，再单独申请真实验证授权。

**未来诊断采用产品默认系统 PKI。** 应用层保留请求数量、时间和停止预算，不禁用 TLS 验证、不新增信任。Windows 系统验证可能获取证书或更新缓存，应用层计数不能宣称覆盖全部操作系统 DNS/TCP 流量；这些影响必须写入未来真实窗口的范围与确认。本次文档更新不授权调用系统链引擎或网络。Windows 专用 CTL 辅助诊断不成为 Linux/macOS 产品认证的依赖。

**BHK 发布目标为以下精确三组正式认证证据：**

| 平台 / testbed | 网络 | auth_method | suite |
|---|---|---|---|
| windows-amd64 / bhk_windows | campus_wired | password | password_core |
| windows-amd64 / bhk_windows | campus_wifi | password | password_core |
| windows-amd64 / bhk_windows | campus_wired | terminal_qr | terminal_qr |

表中为 evidence 枚举；runner 参数拼写仍以实际生效接口为准。Wi-Fi QR 仅为补充，不替代有线 QR。NAS 不再是目标路线强制依赖，v1 四组规则仅在显式历史审计中使用，不能直接用 Windows QR 填充旧 NAS tuple；P4 未全部验收时不得进入 P5。

Linux 保留共享实现、合成测试、race 和原生安装门禁；无真实校园验收时保持 `live_unverified`。macOS 构建与安装通过不等于原生认证通过，Windows evidence 不能替代其他平台。六平台构建、原生安装、Wi-Fi 绑定与 network list/scan 等其他门禁不因减少正式认证 tuple 而被删除。

### 0.4 验收纪律与当前交付边界

- P0 只改文档并做文档检查、链接检查、定向脱敏扫描；稳定 ID 索引仅在需要时由既有生成入口更新，不手工编辑。不改产品 SDK/CLI/JSON/退出码或认证能力声明，不启动产品测试、校园诊断、Candidate、提交、推送或发布。
- 历史 handoff、旧包及日期明确的状态正文保留；status 新摘要是当前事实入口。D15 单次真实授权已经消耗，任何历史授权均不自动续期。
- 已批准范围内离线工作连续推进，不以固定工作时长、测试数量或诊断版本数量为目标。只有新的实质性语义决策、材料缺口或真实操作授权需要维护者参与。
- 每项诊断说明它区分哪一个认证假设；没有输入变化或预期信息增益时不重复同包同结果。不通过删除测试、改变期望或放宽安全边界关闭缺陷。
- P0 交付窗口的首个后续工作包为 P1；随后 P1/P2 已按独立指令执行，P3 开发闭环与 P4 离线交付已完成，当前下一主线为经授权的 P5。开发工具或开发成功不表示正式 `password_core` 或 M3 已完成。

## 1. 总体目标、安全边界与职责

### 1.1 目标与实施顺序

将当前“能够实际登录”的初步版本收敛为：

- 协议正确、默认安全的单一 Go SDK。
- 稳定的现代 CLI、JSON 和退出码契约。
- legacy、meta、dispatcher 三入口兼容产品。
- 可恢复的配置迁移与离线安装。
- 测试字节与发布字节完全相同的不可变候选链。
- 可在 BHK 与双网口 NAS 上执行、不会记录凭据的真实校园网验收。
- 经过 SSH 签名 tag、GitHub provenance attestation 和逐资产哈希验证的 `v1.0.0`。

实施顺序为：先完成文档和紧急安全处理，再收敛协议与 SDK，随后完成三入口、配置迁移与离线安装，最后构建一次候选并完成真实网络验收和原样晋升。

### 1.2 已锁定决策

- 第一执行窗口只做只读冻结前检，不创建提交或修改 refs。
- 高风险动作分别确认：凭据、网络/会话、历史重写、强制推送、仓库改名、规则保护、正式发布、敏感备份删除。
- GitHub 仓库最终为 `UnbalancedCat/ipgw-meta`。2026-10-28 迁移后，权威源码目录为 BHKDesktop `D:\project\Go\IPGW-Meta`；开发入口见[本机开发与合成测试](../operations/development.md)。
- 迁移前的 `D:\project\Go\ipgw-meta-clean` 约定属于笔记本历史；台式机旧仓库已归档为 `D:\project\Go\IPGW-Meta.archive-20261028-165538`。不得按历史路径重新迁移或归档当前工作区。
- `v1.0.0` 使用专用 SSH 签名 annotated tag，并为候选和发布资产生成 GitHub artifact attestation。
- 历史清理和小写改名后严格保护 `main`：要求 PR、严格 CI、签名提交，阻止删除和 force-push，并约束 `v*` tag 更新/删除。
- 只有一个授权校园账号。真实环境不测试异账号 conflict/switch；它保持 `synthetic_covered + live_unverified`，不阻塞 v1。
- 泄露会话已由维护者于 2026-08-28 通过官方门户确认失效；只记录该事实，不保存截图、Cookie 或其他认证材料。
- 自更新继续禁用。
- 手机验证码保持 `observed_anonymous + detected_only`，不实现发送、提交或保存。
- Docker 只用于无凭据合成测试；真实认证必须使用 VM 或原生 Windows，不使用 Docker host network。

### 1.3 职责

维护者负责 GitHub、ZOS、物理线缆、网络切换和外部系统操作；在私有终端输入密码、扫描二维码；管理 SSH 签名私钥并创建最终签名 tag；对每项高风险动作逐项批准；审阅脱敏 evidence 后决定入库摘要。

执行 Agent 负责仓库审计、规范、代码、测试、候选构建和哈希验证；提供 ZOS/VM/远端任务清单；协调本地、NAS 和 BHK 任务；只解析固定白名单结果，不接触密码、验证码、QR payload 或原始认证响应；在检查点记录结果、阻塞和下一步，本机离线开发不因检查点到达而强制停止。

外部 CI 可在窗口结束后继续运行，但本地只记录 workflow/run ID，不持续等待；下一窗口再读取结果。

## 2. 产品、SDK、CLI 与配置契约

### 2.1 产品结构

Go module 为 `github.com/UnbalancedCat/ipgw-meta`，根包名为 `ipgw`。三个入口固定为：

| 入口 | 行为 |
|---|---|
| `ipgw-legacy` | 保留 1.x 旧工作流和无参数登录；最早 2.0 移除 |
| `ipgw-meta` | 现代命令；无参数只显示只读状态或帮助 |
| `ipgw` | 纯模式分发器，不执行网络或凭据逻辑 |

模式优先级为 `--mode > IPGW_MODE > launcher 配置 > 安装批次默认`。旧安装升级后保持 legacy；从 `v1.0.0` 起全新安装默认 meta。三入口、launcher 元数据和安装器必须作为同一原子 bundle 更新。NEU/CAS/Srun/Dashboard 是内部边界；v1 不建设通用插件、daemon、GUI 或多语言 RPC。新 SDK 完全替代后移除 `neugo`。

### 2.2 Go SDK

公开 façade：

```go
func NewClient(opts ...Option) (*Client, error)

func WithBindIP(netip.Addr) Option
func WithRoundTripper(http.RoundTripper) Option
func WithObserver(Observer) Option
func WithProtocolStateStore(ProtocolStateStore) Option

func (c *Client) Status(context.Context) (Status, error)
func (c *Client) Login(context.Context, LoginRequest) (LoginResult, error)
func (c *Client) Logout(context.Context) (LogoutResult, error)
func (c *Client) ListInterfaces(context.Context) ([]Interface, error)

type LoginRequest struct {
    Method           AuthMethod
    ExpectedUsername string
    Credentials      CredentialProvider
    Switch           SwitchPolicy
    Interactions     InteractionHandler
}
```

固定行为：

- SDK 不读取 profile、配置路径或 keyring，不使用全局状态；`NewClient` 不访问网络。
- 同账号在线和账号冲突检查发生在读取 `CredentialProvider` 之前。
- 登录成功同时要求网关业务明确成功、最终在线、最终用户名精确等于 `ExpectedUsername`。
- 同账号在线返回 `already_online`。异账号默认返回 `session_conflict`；显式 switch 必须成功注销并验证 offline 后才能继续。
- Logout 幂等；离线返回 `already_offline`。
- `ListInterfaces` 只枚举本地 UP、非 loopback IPv4；联网扫描由 CLI 组合 `Status`。
- Client 的只读调用并发安全；单 Client 内 Login/Logout 串行；不得修改调用者 Transport 或共享认证 Cookie Jar。
- 所有网络方法接受 context、限制响应大小并使用分阶段超时；保留 `context.Canceled` 和 `context.DeadlineExceeded`。
- v1 每次动态发现协议；持久协议缓存仍禁用，不能把 bind IP 当作网络身份。

稳定错误码为：`invalid_argument`、`config`、`network`、`authentication`、`session_conflict`、`protocol_changed`、`interaction_required`、`unsupported`、`internal`。

### 2.3 协议和认证

- 动态解析 CAS form action、`lt`、`execution`、登录脚本和 RSA 公钥。
- CAS 已注册的 HTTP service 字符串可以保留，但 redirect policy 必须在发送下一跳前验证 host/路径并返回 `http.ErrUseLastResponse`；带 ticket 的 HTTP 请求不得离开进程。
- ticket 只提交给系统 PKI 验证的 HTTPS activation；status、activation、logout 全部 HTTPS-only。
- 禁止 `InsecureSkipVerify`、自动 HTTP 降级和不安全开关。HTTP 只允许无账号、无 Cookie、无 ticket 的 captive portal/`ac_id` 匿名发现，并仅作为不可信提示。
- 公网可访问不能作为登录成功证据；未知页面不得归类为成功或密码错误。
- Observer 只接受固定的脱敏结构化事件，禁止原始响应、完整 URL、查询参数、Cookie、ticket、LT、验证码、账号秘密和任意 map。

Terminal QR 使用 `Login(... MethodTerminalQR ...)`：

- `ExpectedUsername` 必填；最终身份仍须精确核对。
- QR payload 只进入内存中的 `InteractionHandler`；SDK 负责轮询、期限、取消和最终页面验证。
- SDK 不启动浏览器、不持久化二维码会话，也不自动回退密码或外部浏览器。
- TTY/SSH TTY 在 `stderr` 显示二维码和步骤；非 TTY、JSON 模式或无法安全呈现时立即返回 `interaction_required`/7，不显示、不轮询。

任意登录方式都可能进入短信或其他人工挑战。OTP 不得误报为密码错误，只输出固定 challenge kind、origin method、能力状态、固定 resume mode、TTY 要求和 help ID；不得输出手机号、投递提示、验证码、URL、Cookie、ticket、页面或自由文本上游错误。OTP 不阻塞 v1，但被 OTP 阻断的运行不能替代 password/QR 成功门禁。

### 2.4 CLI、JSON 和退出码

现代命令树：

```text
status
login [--method password|qr] [--switch]
logout
network list
network scan
profile list
profile show
profile add
profile remove
profile migrate
```

`session` 与 `account` 留到 1.x。JSON stdout 恰好输出一个换行结尾 envelope；`data` 与 `error` 必须且只能出现一个，stderr 只允许脱敏诊断。参数解析前错误的规范命令名为 `cli`，自动化只能依赖 `schema_version`、`code` 和固定 `details`，并忽略新增字段。

| 退出码 | 含义 |
|---:|---|
| 0 | 成功，包括 offline status、already-online、already-offline |
| 1 | internal／未分类 |
| 2 | usage、config、invalid argument、unsupported |
| 3 | network／deadline |
| 4 | authentication |
| 5 | session conflict |
| 6 | protocol changed |
| 7 | interaction required |
| 130 | 用户取消 |

时间使用 UTC RFC3339，时长为整数秒，流量为整数 bytes，金额为 `{currency, minor_units}`，缺失值为 `null`。QR payload、手机号、ticket、Cookie 和原始响应不得进入 JSON。

### 2.5 配置和迁移

配置保存在 OS config dir，profiles、launcher 和未来 protocol cache 使用独立文件。只允许 `keyring`、`env`、`file`、`prompt` 四种 credential provider；command provider 推迟到 v1 后。密码不得写入 YAML。Unix credential file 强制 `0600`；Windows 只允许当前用户、SYSTEM 和 Administrators ACL。损坏配置必须停止，不得回写默认配置。Unix credential 与迁移私密状态默认逐组件拒绝 symlink；macOS 只按 [`ADR-0010`](../architecture/decisions/ADR-0010-macos-trusted-system-path-alias.md) 接受固定 `/var` → `/private/var` 系统 alias 例外，锚点后的用户组件与最终文件规则不放宽。

迁移支持旧 `neucn/ipgw` JSON 和当前 Meta YAML：

- discovery/preview 零副作用；所有旧秘密先进入 `pending`。
- TTY 逐 profile 选择 keyring/env/file/prompt；非 TTY 必须使用 `--yes`，且只允许显式 env/file 引用。
- 使用无秘密 journal、随机 transaction ID、恢复快照、受限 backup 和 marker。
- pending journal 恢复必须先于 already-migrated 快路径；apply 使用同一次读取的来源 bytes，不能再次读取形成两代数据。
- keyring 只接受全新不可猜引用，已有项不得覆盖。
- marker 必须绑定来源 stamp、transaction ID 和目标 config digest。
- 无法确定当前状态时保留恢复材料并 fail closed。

## 3. 文档治理与分阶段执行

### 3.1 权威文档与稳定 ID

`docs/` 是唯一人类规范源；`agent/` 只保存工作包索引和交接。r2 目标结构为：

```text
docs/
  upgrade/{plan,status,migration-matrix}.md
  architecture/{overview,protocol-correctness,security}.md
  architecture/decisions/
    ADR-0007-immutable-candidate-promotion.md
    ADR-0008-offline-transactional-installer.md
    ADR-0009-separated-live-test-plane.md
    ADR-0010-macos-trusted-system-path-alias.md
  compatibility/auth-capabilities.md
  operations/{config-migration,release,offline-install,live-validation}.md
  runbooks/{headless-auth,campus-lab}.md
  reference/{cli,go-sdk,json-cli}.md
  evidence/
agent/
  plans/stabilization-v1.md
  handoff.md
AGENTS.md
```

新增稳定 ID 为 `ADR-0010`、`REL-WINDOW-001`、`REL-APPROVAL-001`、`REL-PROMOTION-001`、`REL-ATTEST-001`、`REL-INSTALL-001/002/003`、`REL-LIVEGATE-001/002/003`、`REL-LAB-001/002/003`、`REL-LIVE-MATRIX-001`、`REL-TRANSFER-001`、`EVID-BUNDLE-001`、`EVID-CAPTURE-001`、`EVID-REVIEW-001`、`AUTH-CONFLICT-001`。

所有 docs/agent revision 统一为 `2026-08-28-r2`；`doccheck` 必须同步校验 revision、required paths、稳定锚点、交叉引用和派生 agent 引用。`agent/handoff.md` 只保留当前执行者、当前/已完成 WP ID、阻塞和下一步，不复制里程碑状态或产品规范。

### 3.2 窗口规则

1. 开始时声明唯一 WP、允许修改范围、停止条件。
2. 本机离线开发按已授权范围持续推进，随工作包完成验证和交接，不再以 35 分钟为强制停点；执行约定见 [`REL-WINDOW-001`](../operations/release.md#rel-window-001执行范围与检查点)。
3. 不在窗口末开启新的高风险动作。
4. 超时测试停止或转为外部 CI run ID，不持续等待。
5. 窗口结束时不得存在后台登录/QR 轮询、未知活动校园网会话、半完成 ref 更新、未处理安装事务或未说明敏感临时文件。
6. 高风险动作必须在对应窗口再次取得批准。

### 3.3 工作包顺序

以下保留 r2 原始工作包编排，不是当前待执行队列。已完成迁移、环境和 RC-BUILD 不重复执行；当前优先级及新增依赖以[本页日期化 roadmap](#roadmap-20261029)为准，正式门禁的变更仍须完成 P4。

| 顺序 | WP | 内容与停点 |
|---:|---|---|
| 1 | `WP-M0-PREFLIGHT` | 只读检查 GitHub auth、会话确认、工具 hash、工作树、EOL、全部 refs/tags 和远端 SHA；不创建提交 |
| 2 | `LAB-DISCOVER` | 独立远程/用户窗口，只读检查 ZOS 网络能力；不创建 VM |
| 3 | `WP-R2-SPEC-A` | 落盘完整 plan、ADR 和 status；不改产品代码 |
| 4 | `WP-R2-SPEC-B` | 落盘 release/offline/live/evidence/auth 规范 |
| 5 | `WP-R2-SPEC-C` | 更新 doccheck、stable IDs、AGENTS 和派生 agent 索引 |
| 6 | `WP-M0-FREEZE-AUDIT` | 暂停其他编辑，增加 EOL 策略，stage 计划内变更并审查；不提交 |
| 7 | `WP-M0-FREEZE-COMMIT` | 二次核对 index，创建单一冻结提交、安全 tree archive 和受限完整历史备份 |
| 8 | `WP-M0-REWRITE-LOCAL` | 在隔离 mirror 中重写，删除 mirror 内非 commit `refs/codex/**`，复扫并验证 tree/tag |
| 9 | `WP-M0-REWRITE-REMOTE` | 单独批准；逐 ref lease + atomic push，禁止 `--mirror` |
| 10 | `WP-M0-VERIFY-REHOME` | fresh clone 到 `D:\project\Go\ipgw-meta-clean`，复扫、测试并切换权威工作区 |
| 11 | `WP-M0-RENAME-GOVERN` | 小写仓库名、更新 module/origin，再 fresh clone；启用严格 main/tag 规则 |
| 12 | `WP-BASELINE-VERIFY` | 在 clean repo 重跑 M1/M2 现有门禁；失败则重开对应里程碑 |
| 13 | `WP-M2-CONFIG-CLOSE` | 完成迁移失败注入、Unix 实机权限和 keyring backend 验证 |
| 14 | `WP-M2-INSTALL-UNIX` | 离线 Unix installer、路径安全、权限和事务 failpoint |
| 15 | `WP-M2-INSTALL-WINDOWS` | Windows 当前用户 installer、ACL/reparse 和事务 failpoint |
| 16 | `WP-M2-INSTALL-NATIVE` | 六个原生平台 smoke；三种代表实现完整故障矩阵 |
| 17 | `WP-M3-LIVEGATE-SCHEMA` | 固定 runner 接口、状态机、退出码和 evidence schema |
| 18 | `WP-M3-LIVEGATE-RUNNER` | 实现 maintainer-only runner 和无泄漏测试 |
| 19 | `WP-M3-CANDIDATE` | 实现一次构建、candidate-set、manifest、artifact digest 和 attestation |
| 20 | `WP-M3-PROMOTION` | 实现 promotion lock、签名 tag 验证、draft release 和禁止重建 |
| 21 | `WP-M0-NONBLOCKING-GATES` | 按 ADR-0011 解耦 M0 状态，保留其余 hard gate；签名 PR 合入后停止 |
| 22 | `LAB-PROVISION` | 仅在 ZOS 隔离能力通过时建立管理 VM/测试 VM；只做匿名 topology/status 预检 |
| 23 | `RC-BUILD` | 从受保护 main 的精确 SHA 生成不可变 v1.0.0 candidate |
| 24 | `LAB-TRANSFER` | 本地下载、验 attestation/hash，经批准的测试路径发送；远端不重建 |
| 25 | `LAB-PASSWORD-NAS/BHK` | 每个网络/平台独立窗口执行 password suite |
| 26 | `LAB-QR-NAS` | 私有 TTY 完成 QR suite |
| 27 | `LAB-EVIDENCE` | 导出、校验、人工复核证据；提交脱敏摘要和 promotion lock |
| 28 | `WP-M3-RELEASE` | 用户创建 SSH 签名 tag；workflow 原样晋升候选并发布 |
| 29 | `WP-SECURE-DISPOSAL` | 单独批准后删除秘密备份和旧工作区 |

### 3.4 M0 历史清理

冻结前检硬前置：`gh auth status` 和布尔 API 请求成功；固定使用仓库忽略目录内 `git-filter-repo 2.47.0`，其 SHA-256 必须为 `67447413E273FC76809289111748870B6F6072F08B17EFE94863A92D810B7D94`；枚举工作树、全部本地/远端 refs、tag 类型、remote SHA 和 `refs/codex/**`。任一项失败立即停止。

冻结阶段增加明确 `.gitattributes`，源码、文档、YAML、JSON、shell 和 PowerShell 使用仓库规范 LF；通过 staged raw diff、`--ignore-space-at-eol`、`git ls-files --eol` 和 numstat 排除机械行尾污染。冻结分支为 `codex/v1-freeze`，单一冻结提交包含当前全部计划内变更。另以 `git archive` 生成只含安全 tree 的快照并复扫。

敏感历史备份保存于 `C:\Users\Unbal\AppData\Local\IPGW-Meta\secure-backups\2026-08-28\`，ACL 只允许当前用户、SYSTEM、Administrators，标记 `DO-NOT-TRANSFER`；保存精确 refs、工具版本、备份 hash 和恢复说明，不得传到 NAS、BHK、GitHub、云盘或协作者。

隔离 mirror 使用固定无秘密 replace rules 重写全部相关 commit refs，并删除 mirror 中 `refs/codex/**` 等非 commit tree refs；执行 `git fsck --full` 和全 refs/full-history/reflog secret scan。重写前后冻结 tip tree SHA 必须一致；`v0.1.0` 不变，`v0.1.1`–`v0.1.3` 依输入重新映射。未知 ref、秘密残留或 tree 差异都必须丢弃 mirror，不接触远端。

远端 push 前重新 fetch；使用逐 ref `--force-with-lease=<ref>:<old-sha>` 和一次 `--atomic` 更新，禁止 `git push --mirror`，不得推送 `refs/codex`、remote-tracking 或备份 refs。成功后只允许在干净历史上 forward-fix。随后 fresh clone 到新权威路径，复核 refs、tags、secret scan、tests 和 tree；旧工作区只读隔离，不再运行 Git。删除旧工作区或备份必须另行批准。

GitHub Support 对已失去 ref 可达性的旧对象所做的外部缓存处置继续属于 M0/`SEC-HISTORY-001`，但按 [`ADR-0011`](../architecture/decisions/ADR-0011-nonblocking-m0-governance.md) 不再阻塞 Candidate、Promotion、release 或真实验收。M0 仍保持 `in_progress` 直至外部事项实际完成；不得用非阻塞语义伪造完成状态。

## 4. 离线安装、不可变候选与真实网络实验室

### 4.1 离线安装器

Unix 接口：

```text
install.sh --bundle ABS_ARCHIVE
           --bundle-sha256 HEX64
           [--version EXPECTED]
           [--install-root ABS_PATH]
           [--bin-dir ABS_PATH]
```

Windows 接口：

```text
install.ps1 -BundlePath ABS_ARCHIVE
            -BundleSha256 HEX64
            [-Version EXPECTED]
            [-InstallRoot ABS_PATH]
            [-BinDir ABS_PATH]
```

bundle 与 SHA 必须成对出现。离线模式不得初始化或调用网络。bundle 必须是绝对本地普通文件、非 symlink/reparse/UNC，大小 `1..100 MiB`；Unix 拒绝 group/world writable，Windows 拒绝 Users/Everyone 可写。先复制到安装器创建的私有临时目录，再验证外部 SHA。

在线与离线 acquisition 后共用 `outer SHA → exact members/types/sizes → bounded extraction → inner SHA256SUMS → canonical manifest → transactional activation`。逐祖先拒绝 symlink、junction 和 reparse point；install root、bin dir、config dir 必须绝对、本地、不重叠、非根路径；原子 rename 两端同卷。Unix 正式目录 `0755`、binary `0755`、公开 metadata `0644`、私有 stage/backup `0700/0600`；Windows v1 仅当前用户安装。

隐藏测试变量只有在离线模式、私有 test root、匹配 token 且所有路径严格位于 test root 内时才有效：`IPGW_INSTALL_TEST_ROOT`、`IPGW_INSTALL_TEST_TOKEN`、`IPGW_INSTALL_TEST_FAILPOINT`、`IPGW_INSTALL_TEST_ROLLBACK_FAILPOINT`。固定前向点为 `after_verified_stage`、`after_version_publish`、`after_old_active_detach`、`after_active_switch`、`after_entry_1`、`after_entry_2`、`after_launcher_publish`、`after_path_update`、`before_commit`；固定回滚点为 `before_restore_active`、`before_restore_entry_1`、`before_remove_new_version`。禁止 eval、任意命令、任意路径和 sleep hook。

### 4.2 不可变 candidate-set

candidate workflow 只接受 `release_version: v1.0.0` 和完整 40 位 `source_commit`。`source_commit` 必须等于受保护 `main` 当前 tip，M1–M2 自动化门禁全部通过，最终 tag 尚不存在；M0 状态不参与判定。候选构建后进入代码冻结；除 evidence/status/auth/release-note 白名单外的变化都会使候选失效。

Candidate ID 为 `v1.0.0-<SOURCE_SHA12>-<RUN_ID>.<RUN_ATTEMPT>`。集合结构：

```text
candidate-set/
  candidate-manifest.json
  SHA256SUMS
  release/
    ipgw-meta-linux-amd64.tar.gz
    ipgw-meta-linux-arm64.tar.gz
    ipgw-meta-windows-amd64.zip
    ipgw-meta-windows-arm64.zip
    ipgw-meta-darwin-amd64.tar.gz
    ipgw-meta-darwin-arm64.tar.gz
    install.sh
    install.ps1
    release-manifest.json
    SHA256SUMS
  test-tools/
    ipgw-live-gate-linux-amd64
    ipgw-live-gate-windows-amd64.exe
```

`test-tools` 是私有候选测试工具，不是公开 release asset。根 manifest 分列 `release_assets` 与 `test_tools`，并包含 plan/revision、version、candidate ID、source commit/tree、Go/toolchain、build-input digest、每个文件的 hash/size/platform；full schema、canonical bytes、build-input 算法、归档元数据、两级 checksums 与 attestation subjects 以 [`REL-ATTEST-001`](../operations/release.md#rel-attest-001candidate-manifest-与-provenance) 为唯一规范。六个平台、三入口和两个 helper 各只构建一次；使用确定性归档和独立 SHA-256。GitHub 上传单一不可变 artifact，记录 artifact ID、artifact digest、run ID/attempt，并为 candidate-set 与每个公开资产生成 provenance attestation。本地 Windows 可保存只读副本，但不能替代 GitHub artifact；artifact 不可用、attestation 失败、hash 不符或构建输入变化时必须废弃候选并重做真实测试。

### 4.3 live-gate

源码位于 `internal/livegate/` 和 `internal/cmd/ipgw-live-gate/`。它不是产品 CLI，不安装到 PATH，也不进入公开 release。固定接口：

```text
ipgw-live-gate run
  --candidate ABS_PATH
  --candidate-sha256 HEX64
  --candidate-manifest ABS_PATH
  --suite password-core|terminal-qr
  --testbed nas_vm|bhk_windows
  --network campus_wired|campus_wifi
  --profile NAME
  --output-dir ABS_PATH
```

- runner 执行冻结的 `ipgw-meta`，不重新实现 SDK；candidate 必须为私有目录普通文件，运行前后 hash 必须一致。
- profile/config 只用于运行，不进入 evidence。password suite 强制 TTY prompt，不接受命令行、env、file 或 keyring 密码；QR payload 直接进入用户私有 TTY，不被 runner 捕获。
- 初始 status 必须明确 offline；online/unknown 一律 blocked，不自动 logout。runner 仅在本轮确认创建会话后有清理权。
- 失败/取消最多执行一次有界 logout 再验证 offline；清理失败报告 fail/blocked，由维护者通过官方门户处理。

`password-core` 固定顺序为：initial offline、login logged_in、status online、second login already_online 且不读取凭据、logout logged_out、status offline、second logout already_offline。`terminal-qr` 固定顺序为 initial offline、TTY QR login、status online、logout、status offline。

runner 退出码：0 suite pass；10 gate fail；11 blocked；12 security reject；13 evidence durability failure；14 internal；130 cancel。

### 4.4 NAS/BHK 拓扑

以下保留原 NAS 实验室拓扑及迁移前职责描述。BHKDesktop 当前已是本机开发主机，WSL 可执行离线 Linux 开发/测试，见[本机开发入口](../operations/development.md)；“只协调/不放源码/WSL 只做准备”不适用于当前权威开发目录。正式测试面隔离规则未因此取消；BHK 三组发布目标须待 P4 生效。

```mermaid
flowchart LR
    L["本地 Windows<br/>协调、下载、验证、SCP"] --> C["NAS 管理 Ubuntu VM<br/>Codex + SSH"]
    M["独立管理网络"] --> A["NAS 物理口 A"]
    A --> C
    C -->|"固定 hash 候选、无凭据 SSH"| T["一次性 Ubuntu 测试 VM<br/>不安装 Codex"]
    T -->|"唯一默认路由/DNS"| B["NAS 物理口 B"]
    B --> G["校园网测试上游"]
    U["用户私有 TTY"] --> T
```

管理 VM 建议 Ubuntu 24.04、2 vCPU、4 GiB、20 GiB，使用管理网口并安装 Codex；测试 VM 建议 Ubuntu 24.04、2 vCPU、2–4 GiB、20 GiB，不安装 Codex。测试 VM 管理接口只有私网静态路由，无默认路由/DNS；测试接口使用固定 MAC，是唯一默认路由和 DNS。NAS 宿主测试物理口不得配置 IP、DHCP 或默认路由。

只使用 ZOS 官方桥接/直通能力，不 root hack、不使用 Docker host network。ZOS 无法实现隔离时 fail closed：NAS 只做离线 Linux 测试，真实认证转到 BHK。VM 每轮恢复干净快照，物理换线和网络切换由维护者在 VM 关机后完成。

NAS 管理 VM 与 BHK WSL 分别保存为独立 Codex 远程项目；本地只协调 candidate ID/hash 和脱敏结果。远端不放 GitHub token，不传源码、`.git`、秘密历史或备份，不重新构建候选。BHK WSL 只做准备和 hash；Windows 原生认证由维护者在私有 PowerShell 执行。

## 5. 测试、证据、晋升和发布

### 5.1 自动化门禁

基础门禁：

```text
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
go run ./cmd/doccheck --check
secret scan
```

协议测试必须覆盖动态 CAS form/public key、JSON/JSONP、错误 HTML、响应截断、ticket HTTP 下一跳从未发送、redirect allowlist/恶意 redirect、系统 TLS/无降级、业务失败、最终用户名不匹配、同账号不读取凭据、异账号拒绝、logout failure 中止 switch、offline status、幂等 logout、QR pending/success/expired/cancel/恶意 URL/final OTP、SMS/unknown challenge、context cancel、response limit、并发和 Observer 泄漏 canary。

配置测试覆盖两类来源、格式歧义、超限、未知字段、preview/apply 变化、keyring 不可用/已有引用、每个 journal phase/rollback failure、pending journal 阻止普通写入、marker/config digest 不一致、Windows ACL、Unix owner/mode，以及 stdout/stderr/JSON/journal/marker/backup name 泄漏 canary。

安装器需在 `ubuntu-24.04`、`ubuntu-24.04-arm`、`windows-2022`、`windows-11-arm`、`macos-15-intel`、`macos-15` 六个原生 runner 上执行 fresh install、upgrade、三入口 `--version`、launcher 默认行为和基础 rollback；Linux amd64、Windows amd64、macOS arm64 完成全部 failpoint、rollback-failure、路径攻击和权限矩阵。native runner 不可用时门禁为 blocked，不得以交叉编译替代。离线测试以立即失败的网络命令替身证明未联网。

Promotion 测试必须证明 tag job 不运行 setup-go/build/package；artifact ID/digest/hash 或构建输入变化立即失败；不得覆盖已有单资产；draft 上传后重新下载逐项核验；attestation 失败即停止。

### 5.2 真实网络矩阵

本节为 P4 迁移前仍有效的矩阵；已确认的三组 BHK 目标见[当前 roadmap](#roadmap-20261029)，不能据目标提前放行现有 runner 或 promotion。

| 环境 | 必须执行 |
|---|---|
| NAS Ubuntu / campus wired | password-core、same-account、logout 幂等、Terminal QR |
| BHK Windows / campus wired | password-core、Windows 原生离线安装 |
| BHK Windows / campus Wi-Fi | password-core、bind IP、network list/scan |
| GitHub macOS runners | 安装和 CLI smoke；不声明校园网认证 |

QR 只要求 NAS 完成一次。异账号 conflict/switch 只做合成测试，不借用第二账号。OTP/unknown challenge 返回 blocked/interaction_required 且不泄漏 details。被 OTP 阻断的必需矩阵不算通过，应由维护者通过官方门户处理后重跑。

### 5.3 Evidence

本地私有目录为 `build/live-evidence/<candidate-id>/<evidence-id>/`，权限为 Unix `0700/0600` 或 Windows 当前用户私有 ACL。每次运行只产生 `evidence.json`、`summary.md`、`SHA256SUMS`。

`evidence.json` 只允许 schema version、plan/revision、evidence/candidate ID、candidate-set hash、source commit、platform、testbed、network type、auth method、suite、capability before/after、result、UTC 起止时间和固定 step 的 name/result/exit code/error code/duration。fail/blocked 不得提升能力状态。所有枚举固定；禁止自由 notes、reviewer handle、命令文本、profile、username、IP、interface、URL、details 或原始输出。

禁止 pcap、截图、HTML、headers、Cookie、ticket、TGT、LT、QR payload、验证码和任何可恢复认证材料。runner 在源端直接构造 allowlist，不得先保存原始日志再脱敏。文件通过临时文件、flush、原子替换和目录 durability barrier 写入；无法确认持久化时退出 13，不宣布通过。`summary.md` 使用固定模板。维护者人工审阅后，只有选定固定字段摘要和 bundle hash 可进入 `docs/evidence/`。

### 5.4 Promotion lock

`docs/evidence/releases/v1.0.0/promotion-lock.json` 的 closed-world 字段、canonical bytes、workflow/artifact、11 个 attestation subjects、四份 public evidence summary 与 `release-notes.md` 哈希契约以 [`REL-PROMOTION-001`](../operations/release.md#rel-promotion-001promotion-lock-与原样发布) 为唯一规范；不得使用“至少包含”的开放 schema。

candidate source 与最终 tag 之间只允许修改 `docs/evidence/releases/v1.0.0/**`、`docs/upgrade/status.md`、`docs/compatibility/auth-capabilities.md`。任何 Go、module、installer、workflow、Makefile、release script、doccheck 构建输入或其他文件变化都会废弃候选；promotion commit 必须重算并匹配 build-input digest。

### 5.5 SSH 签名与最终发布

使用专用 Ed25519 release signing key。私钥只存于本地 Windows 用户环境，不进入仓库、CI、NAS、BHK 或 Agent 输出；公钥登记为 GitHub signing key。clean repo 使用 `gpg.format=ssh`。最终 `v1.0.0` 由维护者在 clean repo 手动创建指向 promotion commit 的 SSH 签名 annotated tag。

Promotion workflow 必须按 [`REL-PROMOTION-001`](../operations/release.md#rel-promotion-001promotion-lock-与原样发布) 验证 annotated tag/main tip、lock/source/build-input、精确 artifact、full Candidate 与 attestation；全程禁止 setup-go/build/package/repack。远端同 tag release 必须不存在；workflow 只创建无 clobber 的不可见 draft，上传十个原始公开资产并重新下载逐项核对，最后一次复核通过后才公开。失败时保留不可见 draft，不得自动删除、重试 mutation、公开半包或替换单资产。

Release notes 必须声明：新安装默认 meta、旧安装保持 legacy；配置迁移方式；password/QR 真实证据范围；conflict 只有合成覆盖；OTP detected-only；自更新禁用；macOS 只有安装/CLI 验证，没有校园认证声明。

### 5.6 v1 完成条件

只有以下全部满足，M3 才能标记 complete：

- Candidate source 的 required CI、完整可达历史/refs/reflog 与工作树 secret scan 通过，clean clone 仍是唯一权威工作区；GitHub 不可达旧对象的外部缓存处置按 ADR-0011 独立跟踪，不是 M3 完成条件。
- main/tag 严格保护已启用。
- M1 SDK/协议全部门禁通过；M2 配置迁移、三入口和离线安装器完成。
- 六平台原生安装门禁通过。
- 不可变 candidate、attestation 和 promotion workflow 通过。
- NAS/BHK 真实网络矩阵通过；Terminal QR 至少一次真实闭环。
- evidence 人工复核完成。
- 签名 `v1.0.0` tag 和公开资产逐字节验证成功。

## 6. v1 后范围

功能迁移顺序固定为：

1. 在线设备、精确 session disconnect、批量下线。
2. 套餐与当前用量。
3. 历史用量、账单和充值。

每项先进入 SDK domain model，再进入 meta CLI，最后补 legacy 映射；不得把 Dashboard HTML/API 细节暴露为公共 SDK 类型。

## 7. 参考

- [GitHub：清除仓库中的敏感数据](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository)
- [GitHub：签名验证](https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification)
- [GitHub：Artifact attestations](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations)
- [GitHub：Ruleset 可用规则](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets)
- [GitHub-hosted runners](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)
