---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
status: approved
---

# 实施状态

本页是里程碑状态的正式来源；工作包级执行交接位于 [`agent/handoff.md`](../../agent/handoff.md)。状态仅使用 `not_started`、`in_progress`、`blocked`、`complete`。

| 里程碑 | 状态 | 发布门禁 | 规范入口 |
|---|---|---|---|
| M0 文档与紧急安全 | in_progress | 残余历史对象治理（非发布状态门禁） | [计划](plan.md#34-m0-历史清理)、[安全](../architecture/security.md)、[ADR-0011](../architecture/decisions/ADR-0011-nonblocking-m0-governance.md) |
| M1 协议正确性与 SDK | complete | HTTPS-only、ticket 截获、动态发现、typed errors、最终身份校验 | [协议](../architecture/protocol-correctness.md)、[SDK](../reference/go-sdk.md) |
| M2 三入口、配置与自动化 | complete | 三二进制、JSON/退出码、profiles、迁移和原子安装 | [CLI](../reference/cli.md)、[迁移](../operations/config-migration.md) |
| M3 v1 候选与稳定发布 | in_progress | 自动化门禁、跨平台构建、校园网人工验收 | [发布](../operations/release.md)、[证据](../evidence/README.md) |
| 1.x 后续功能迁移 | not_started | 会话 → 套餐/当前用量 → 历史用量/账单/充值 | [迁移矩阵](migration-matrix.md) |

<a id="status-20261029"></a>

## 2026-10-29 当前摘要：认证收敛路线

本摘要是当前状态入口；下方 8 月至 9 月日期明确的快照及其“当前/下一步”原文保留，只表示当时状态。当前执行交接位于 [BHKDesktop 交接](../../agent/handoff-bhkdesktop-2026-10-28.md)，原 `agent/handoff.md` 保持历史用途。后续顺序以[当前 roadmap](plan.md#roadmap-20261029)为准，不重做迁移、环境初始化或历史 RC-BUILD。

<a id="status-20261031-p4"></a>

### 2026-11-01 P5 有线 terminal QR 首次正式窗口失败

<a id="status-20261102-qr-continuation"></a>

2026-11-02 统一源码准备：Wi-Fi/QR 开发结果闭合后，已完成拟提交范围审阅、干净源码快照的两端集成/vet/doccheck、Linux race 及发布工具合成质量门禁；提交清单和 PR 草稿见[准备记录](../research/unified-candidate-preparation-2026-11-02.md)。尚未创建新分支、提交/推送或生成新 Candidate；下一检查点为精确清单授权。正式 QR/M3 保持，下方开发 pass 不作为正式证据。

2026-11-02 最新结果：维护者私有终端的新 QR 开发窗口五步完整 pass，退出码 0；独立目标身份 online、注销及最终 offline 均通过，coverage=complete、gate_failure=null，无 cleanup。四项包哈希、225 项源码快照及严格 v2 回执解码复核通过，见[开发闭环结果](../research/qr-continuation-progress-2026-11-02.md#后续窗口qr-开发闭环通过)。这是非发布开发证据；普通产品字节与此前门禁停止窗口相同，旧失败原因仍未知。QR 开发验证已闭合，下一步统一 Wi-Fi/QR 源码审阅及新 Candidate 准备，不重复开发扫码；提交、正式构建与真实验收另行授权。正式 QR 仍待新 Candidate 验收，M3 in_progress；以下内容为较早的调查/交付快照，其“下一步”由本段覆盖。

2026-11-02 首因定位已离线交付：独立 QR 启动器外层 v2、旧回执显式历史读取，保留首个门禁失败类别和位置，未放宽停止规则。两端相关回归、Linux race/fuzz、共享消费者、绑定/自检通过，新包零网络 preflight ready。普通产品字节与旧包相同；新身份及剩余真实验证见[当前交付](../research/qr-continuation-progress-2026-11-02.md#当前交付门禁首因定位-v2)。本轮未登录或注销，旧真实失败原因不能由离线结果回填；下一步为另行授权的新包单次验证，正式 QR/M3 保持。

2026-11-02 门禁排查补记：维护者确认账号仍在线。8 次原样本地查询稳定，真实本地门禁配合假认证命令五步通过；正常收尾取消及失败停止边界的两端重复回归/race/vet 通过。历史具体原因仍未知，不修改注销或放宽门禁；下一提案为独立开发报告保留首个门禁失败类别，见[排查记录](../research/qr-continuation-progress-2026-11-02.md#门禁排查未复现历史原因仍缺证据)。未新增校园请求，正式状态不提升。

2026-11-02 拓扑复核更新：维护者已授权一次私有 QR 窗口，但旧包在真实请求前因断开 Wi-Fi 的残留默认路由拒绝。经批准继续修复，仅对适配器与 IPv4 状态均明确断开且唯一匹配的残留默认项排除计数，保留全量变化监测及活动竞争拒绝。两端相关回归和新包绑定通过，新包本地 preflight ready/exit 0，零校园请求/零 profile 读取；普通产品字节与上一包相同。新身份、测试与终端入口见[当前实施记录](../research/qr-continuation-progress-2026-11-02.md#后续更新残留默认路由修复及当前可测试包)。后续维护者单次窗口中 QR 命令及独立目标身份 online 核对通过；注销阶段门禁停止，整体 blocked/partial、exit 11，最终会话未知。包哈希与 218 项快照核验通过；维护者未主动改动网络，具体失败项尚无证据，不能认定协议失败或完整闭环成功。见[窗口结果](../research/qr-continuation-progress-2026-11-02.md#最新窗口登录通过注销阶段门禁停止)。授权已消耗，正式 QR/M3 不变；下段 blocked 是旧包历史。

2026-11-02 离线修复交付：按 [ADR-0028](../architecture/decisions/ADR-0028-qr-service-continuation.md) 补齐无 ticket service → HTTPS 网关指示 → 同轮 CAS → ticket 截获 → HTTPS activation/目标身份复核，保持既有直接 ticket 与 CAS 确认分支。独立 qr-dev 包使用普通产品，分别记录 QR、后续 status、logout、final status，不改变正式摘要。两端测试/vet、Linux race/有界 fuzz、六平台三入口构建及依赖隔离、快照/绑定/self-test 通过；安装套件本轮不重跑，Windows 五项既有 symlink 测试因权限 skip，Linux 一项错误构建主机测试条件 skip，不计为通过。本机预检 binding=verified、topology_unavailable、exit 11，network_requests/profile_reads 均为 0，真实模式未执行。包身份、全部报告与后续窗口说明见[实施记录](../research/qr-continuation-progress-2026-11-02.md)。正式 QR 仍 blocked/live_unverified，M3 in_progress；旧 Candidate、失败回执和双 Wi-Fi 开发通过保持，下段为修复实施前的历史快照。

2026-11-02 后续：已授权的无痕网页 QR 对照完成，维护者核对目标 online，普通产品 logout 与最终 offline 均 exit 0。已补齐官方 QR 脚本及网关/CAS 接续材料，发现 SDK 不接受本轮网页先返回的无 ticket service 表示；网页随后经门户 API 再访 CAS 才获得 ticket。证据、浏览器本机代理路径限制及旧失败不可唯一回溯见[研究记录](../research/qr-browser-contract-2026-11-01.md)。采集已结束，下一步为离线接续契约和合成复现；尚未修改产品或交付 QR 修复包，正式 QR/M3 不提升。下段为此前离线准备快照。

当前 QR 修复取证已开始：依据 [ADR-0027](../architecture/decisions/ADR-0027-browser-qr-contract.md) 完成固定版本外部 QR 源码/本地调用链核对及浏览器窗口约定。新增合成回归验证历史摘要无法区分 QR 命令失败与其后 status 失败；正式 fail 回执不变。缺官方 QR 构造器/轮询实现及真实脱敏状态关系，产品根因仍未确定，未修改产品或创建修复包。后续仅进入另行授权的无痕 Chrome 对照，详见[研究记录](../research/qr-browser-contract-2026-11-01.md)。

补充手动试用：维护者反馈普通开发产品在有线、NEU、NEU-2.4G 上登录/注销均无问题；浏览器门户 ac_id 依次为 1、15、15。此为维护者观察，无新增逐命令/正式回执，不改变正式认证状态。固定版本外部对照及局限见[研究记录](../research/ac-id-comparison-2026-11-01.md)。

最新后续：[双 SSID 与定向扫描](../architecture/decisions/ADR-0026-bound-scan-dual-wifi.md) 已完成离线实现、版本迁移、两端回归及独立开发包。维护者分别回传同包 NEU、NEU-2.4G 开发窗口 pass：两次均为本地 5 个地址、选中 1 项，八项辅助和 password_core 七步全部通过，窗口结束时 offline；release_evidence=false。NEU-2.4G 明确回传启动器 exit 0，NEU 未回传该数字。此前正式辅助 selected_address_unique blocked、开发预检 topology_unavailable 均保留，见[实施记录](../research/wifi-bound-progress-2026-11-01.md)。两个窗口均关闭 Clash，尚未建立故障因果；浏览器门户跳转只作脱敏观察。双 Wi-Fi 开发验证已闭合，下一发布依赖为 QR 收敛及后续新 Candidate，新正式 Wi-Fi 证据仍未取得。当前目标为四组认证和三份辅助，旧三组/两项描述按日期保留。M3 不变。

后续维护者已明确暂缓 QR 排障，改为同一 Candidate 的 Wi-Fi 验收，并批准断网交接范围：一次 Wi-Fi 辅助检查全部通过后再执行一次 password_core，初始在线/未知或任一步失败均停止，不自动重试或注销已有会话。完整 Candidate 和本地入口已校验，只执行 CheckOnly，尚无 Wi-Fi 结果。此安排覆盖下文“先厘清 QR 再进入 Wi-Fi”的当时顺序，不取消 QR 必需门禁。

同一 Candidate 的正式有线 terminal_qr 回执为 fail：初始 offline 通过，QR 登录步骤 166 秒后返回产品 exit 6 / protocol_changed，runner exit 10。维护者确认二维码出现且已扫码确认；回执没有证明扫码后哪个协议阶段失败。私有 bundle/schema/checksums、产品/manifest/helper 绑定已复核，详见 [QR 失败记录](../research/p5-readiness-2026-11-01.md#有线正式-terminal_qr-失败回执)。

未建立 status_online 清理权，因此没有 cleanup 或最终 status；窗口后的账号状态未确认，不能沿用初始 offline。QR 保持 live_unverified，P5 正式验收集合未闭合，M3 保持 in_progress；此前有线 password_core 和安装辅助的通过记录仍有效，但不能代替 QR。停止重复该包 QR 和继续 Wi-Fi 窗口，先厘清协议证据缺口；不改冻结 Candidate 或猜测放宽响应规则。

### 2026-11-01 P5 有线正式 password_core 通过

当前 Candidate `v1.0.0-1a32aabc2aa3-41275535736.1` 已取得 windows-amd64 / bhk_windows / campus_wired / password_core 的正式 v2 pass 回执。七个主步骤完整通过、无 cleanup；私有 bundle/checksums、Candidate manifest、普通产品与 helper 哈希绑定已只读复核。该 tuple 的能力为 synthetic_covered + live_verified，详情见 [P5 有线回执](../research/p5-readiness-2026-11-01.md#有线正式-password_core-回执)。窗口结束时 offline，不据此推断之后持续状态。

BHK 安装辅助已通过；尚缺有线 terminal_qr、Wi-Fi 辅助和 Wi-Fi password_core。后续真实窗口仍逐次授权；公开摘要待人工审阅选入，尚无完整 promotion lock。M3 保持 in_progress，不提升 Linux/macOS 或其他网络的能力。下方“正式 password_core blocked/live_unverified”等描述保留为此次窗口之前的历史状态。

### 2026-11-01 P5 源码准备：symlink 覆盖补齐

最新进展：PR #29 已经授权流程合入受保护 main `1a32aabc2aa3125afe638b0dc3d3444bf7a95c7d`，签名有效、13 项 CI 全部通过，tree 与审阅版本一致。一次正式 [Candidate run 41275535736](https://github.com/UnbalancedCat/ipgw-meta/actions/runs/41275535736) 成功，生成 `v1.0.0-1a32aabc2aa3-41275535736.1`；六平台原生安装、14 文件本地闭集核验及 11-subject 独立 attestation 复核通过。BHK 原生安装辅助第 003 份回执十项全 pass，前两份 blocked 保留；执行详情与限制见 [P5 记录](../research/p5-readiness-2026-11-01.md#正式-candidate-与本机验收)。三组正式认证及 Wi-Fi 辅助证据尚未取得，正式 password_core 仍 blocked/live_unverified，M3 仍 in_progress。以下未合并/未构建描述保留为较早检查点。

最新复核：维护者批准的 8 文件修正已提交/推送为 `0d11f1f`；13 项 CI 中 12 项成功，六平台原生安装全部通过。Windows 全仓测试仍有诊断夹具的路径/超时失败，追加测试修正在本机全仓回归、vet 和独立短路径环境通过，尚待提交及新 CI。未合并、没有新 Candidate；以下首次提交及修正前记录保留。

后续执行：已按批准清单建立 `codex/bhk-v1-acceptance`、提交 `2fc0d1644920a2aefb10a8859e4baebe84cf0501` 并创建 [PR #29](https://github.com/UnbalancedCat/ipgw-meta/pull/29)。首轮 CI 暴露干净检出链接和跨平台测试夹具问题；本地修正及相关回归完成，补充提交和新 CI 尚待完成。首轮失败保持可见，未合并或构建新 Candidate。详情见 [CI 修正记录](../research/p5-readiness-2026-11-01.md#pr-29-与首轮-ci-修正)；下一段保留提交前状态。

五项 Windows symlink 安全测试已通过固定哈希离线测试程序，在原生管理员进程中实际执行并全部 pass；未修改永久权限或真实安装。发布手册中的残留旧矩阵已修正。详情见 [P5 准备记录](../research/p5-readiness-2026-11-01.md)。当前仍在提交/PR 前的审阅检查点，尚无新 Candidate 或 P5 正式证据；正式 password_core blocked/live_unverified、M3 in_progress 保持。下方 P4 的 skip 描述保留为当时记录。

### 2026-10-31 P4 离线交付完成

维护者已批准的 [ADR-0025](../architecture/decisions/ADR-0025-bhk-formal-acceptance.md) 和 [P4 契约](../operations/bhk-acceptance.md) 已完成实现与离线验收。当前源码中的正式认证 evidence/公开摘要/promotion lock 使用 v2：三组 BHK 认证加两项独立辅助摘要；v1 只在显式历史审计入口读取。Candidate/release manifest 仍为 v1，旧 Candidate 不得搭配新 helper。

Windows 全仓测试、vet、原生安装/junction 回归与专项检查，Linux 全仓测试/race/vet、有界 fuzz、质量门禁，以及六平台三入口构建和产品依赖隔离均通过。安装失败定位于测试夹具的所有者写入/继承 ACL 准备；修复限定于私有测试目录，没有绕开 junction 或放宽安装契约。五项 Windows symlink 特权用例仍因环境限制 skip，不能称为已覆盖；普通全仓测试的 Candidate 安装 smoke 未选 Candidate 而 skip，正式辅助入口则强制 Candidate、拒绝 skip。详见[完整交付与缺口](../research/p4-bhk-gates-2026-10-31.md)。

下一工作包为 P5：后续授权的受保护 main/正式 CI、新不可变 Candidate、BHK 原生安装辅助记录、Wi-Fi 辅助窗口及三组正式认证。P4 没有执行校园请求、真实 profile/凭据读取、Candidate 构建、提交、推送或发布。正式 password_core 保持 blocked/live_unverified，M3 保持 in_progress；本节不授予 P5 操作权限，下方日期明确的历史状态保留。

<a id="status-20261031-password-dev-pass"></a>

### 2026-10-31 P3 有线开发password_core完整闭环通过

维护者在私有终端执行新callback包的一次授权窗口，回传schema3结果。严格解码、五项包哈希、211项源码快照及当前源码一致性核验通过；只保存白名单JSON，未保存终端账号提示。证据见 `build/logout-callback-live-20261031/`，source `a396e11e95ec9705b03a2a19527de7070e9735bdca6151199a2cf611c6569669`。这是真实开发验证：执行者为分别绑定的独立observed CLI，不是普通产品同一EXE，也不是Candidate evidence。

七个主步骤全部pass：初始offline、login LoggedIn、独立status online、重复login AlreadyOnline、logout LoggedOut、独立最终status offline、重复logout AlreadyOffline。8个product_commands包含前置profile show；无cleanup命令。最后一次logout只发送status，不重复发出注销请求。回执证明该窗口结束时offline，不代表此后的持续账号状态。

登录轨迹包含CAS POST302、activation200及身份状态复核。首次注销是status200、logout200、status200；正文完整读取，envelope=jsonp_object、ecode=integer_zero、success=missing、parser=accepted_success，随后原有offline判据通过。全部observations共18次应用RoundTrip；旧内层匿名计数为0是password模式未采集，不是零网络，不覆盖操作系统DNS/PKI流量。

与前次失败包相比，产品实现只有固定callback参数变更；本轮支持该候选修复在目标场景有效。没有获得旧other正文，不能把前后两个不同窗口当作全部服务端因果的严格控制实验。未放宽业务判据、导入浏览器会话或新增CAS退出。

本次单次授权已消耗，Agent未重跑。P3的BHK有线开发闭环已取得通过证据；正式password_core仍blocked / live_unverified，M3仍in_progress，旧不可变Candidate不含这些开发修复。下一工作包应收敛P4的BHK正式验收规范、evidence/runner/promotion一致性，随后经P5流程产生并验收新Candidate；不继续重复Chrome/CTL或同包认证。Wi-Fi、terminal QR、Windows原生安装/junction、Linux/macOS原生认证等既有缺口保持。

<a id="status-20261031-logout-callback"></a>

### 2026-10-31 P3 注销callback候选修复离线交付

维护者已批准 [ADR-0024](../architecture/decisions/ADR-0024-logout-explicit-callback.md)。共享注销只新增固定callback请求参数，原严格JSON/JSONP、ecode/可选success、最终offline和幂等判据保持；不复制浏览器会话、不新增IP/ac_id、CAS退出或额外请求。网页材料与因果限制见[研究记录](../research/logout-browser-context-2026-10-31.md)：现有解析器已能接受网页封装，callback是待真实验证的候选修复，不能宣称唯一根因已经证明。

Windows聚焦SDK、auth-dev相关测试/vet、整链self-test及八项构建绑定审计通过；Linux相关测试/race/vet、三项有界fuzz和纯解析包race通过；六平台三个产品入口构建及诊断依赖隔离通过。汇总及失败记录保留在 `build/p3-logout-callback-20261031/`。最初两个聚焦失败来自新增夹具对空Body表示和既有用户名白名单的错误假设；已在测试层纠正，未更改产品边界。未重跑全仓安装门禁；Windows原生安装/junction缺口保持，macOS仅为交叉构建。

最终包 `build/dev-runs/windows/20261031T081905Z-d60ad4f5-auth-dev/ready/`，入口auth-dev-launch.exe；source `a396e11e95ec9705b03a2a19527de7070e9735bdca6151199a2cf611c6569669`，普通产品SHA-256 `3bb73e3da91e2d0bb4273b8939f0d6e0cf42d0f1bb98f3992b89efbd6e63f73d`，独立observed CLI SHA-256 `43f90dd275e22c4415d7776cf7f6998b98ce1aec968049e9cfda6329af4966d9`。211项源码快照与当前源码一致，五项哈希通过；相对前包只有auth.go及三份测试变化。外层schema3、子进程schema2不变，普通产品与观察EXE分别绑定。

本轮只执行合成/回环及零网络自检，没有读取真实profile/密码、操作Chrome或执行校园登录注销，也未运行真实拓扑preflight。下一步是另行确认的原生Windows私有终端password_core单次窗口；成功也只属于开发证据。正式password_core仍blocked / live_unverified，M3仍in_progress；旧包、Candidate和历史回执保持，未提交、推送或发布。

<a id="status-20261031-logout-envelope"></a>

### 2026-10-31 P3 新观察包真实回执：注销封装不匹配

维护者在私有终端确认单次窗口并执行 source `a4ef40dfb8dc66afa3db1b8f6832c7139008200d1ea74c62a7a9fc6a55a1cce2`。新 schema 3 回执严格解码通过，五项包哈希、210项原包源码快照和13项当前解码相关源码一致；仅保存闭集 JSON，不保存终端账号提示。证据在 `build/logout-observation-live-20261031/`，属于开发证据，不是 Candidate evidence。

登录、独立 online 和重复登录 AlreadyOnline 再次通过；首次注销 HTTP200 后返回 protocol_changed，SDK分类为 logout_response_parse。新增观察为 read=complete、envelope=other、structure=none、bom=absent：正文已完整读取，没有触发体积拒绝；未进入当前支持的 JSON 对象或简单 callback JSONP 对象入口，起始没有 UTF-8 BOM。不能将 other 写成 HTML、纯文本或具体成功标记，也不能把 structure=none 理解为对未知封装已验证有效。

字段 ecode/success 等及 parser 全为 unknown，表示观察器没有在支持的对象封装上做字段或业务投影，不表示字段缺失；SDK本身的业务解析拒绝由固定错误分类独立确认。当前不能依据这份回执新增响应成功白名单或放宽 PROTO-LOGOUT-001。

整套 password_core 仍 fail。cleanup_logout 仅有 status 请求、无第二次 logout 请求，最终 cleanup_status_offline 通过；正常第二次注销和正常最终状态步骤仍未完成。本窗口结束时观察到 offline，不代表后续持续状态。17条已记录 transport 调用不等于 OS DNS/TCP 流量；密码模式中的旧匿名计数字段为零不表示零请求。

此次单次真实授权已消耗。离线核对已确认当前与旧版的 HTTPS 端点、GET、action/username 键相同，均未请求 callback；旧版不解析业务正文，不能提供新成功契约。Referer、UA和前置步骤有差异，但没有证据证明其决定本轮响应格式。下一步仍需建立实际封装及其业务含义的依据，再形成最小修复或新的明确观察决定；不重复相同包、不保存原始响应、不因随后 offline 而绕过业务结果。正式 password_core blocked/live_unverified、M3 in_progress 和 Windows 安装/junction 缺口保持。

<a id="status-20261031-logout-observation"></a>

### 2026-10-31 P3 注销结构观察离线交付

维护者已确认 [ADR-0022](../architecture/decisions/ADR-0022-development-logout-structure.md)。独立 observed CLI 已实现仅注销响应的有界内存结构投影、固定 SDK 失败分类和严格新旧版本解码；普通产品注销业务成功与最终 offline 判据不变。旧回执不补造新字段，真实响应未采集。

最终开发包为 `build/dev-runs/windows/20261031T024632Z-b3a7065a-auth-dev/ready/auth-dev-launch.exe`，source `a4ef40dfb8dc66afa3db1b8f6832c7139008200d1ea74c62a7a9fc6a55a1cce2`；210项源码与五项包哈希一致。普通产品 SHA-256 为 `a7b52e8e9d7d28d2ca6fd0cb8fdd4a3a11111c3dd3b0e47628868a3bd1ebb946`，与前次开发登录成功包相同；观察 EXE 和启动器是新身份。整链零网络自检、合成 profile、缺失/额外/篡改文件拒绝和历史 v2 / 新 v3 回执解码通过。修复前中间包保留但不是本次交付入口。

Windows 专项、普通/诊断测试和 vet 通过；Linux 普通测试、全仓 race、vet、有界 fuzz 通过，六平台三入口产品构建及诊断依赖隔离通过。Windows 安装测试仍有既知 fresh install/upgrade 和 rollback junction 失败，未修改 ACL 或放宽期望；macOS 交叉构建不是原生认证验收。完整命令、结果、失败与跳过见 `build/logout-observation-p3-20261031/REPORT.md`。

下一步是另行授权后的私有终端 password_core，利用新包区分首次注销失败层。没有读取真实 profile/密码、运行校园请求、重试历史窗口、创建 Candidate、提交或发布。开发套件真实闭环仍未通过，正式 password_core 维持 blocked/live_unverified，M3 维持 in_progress。

<a id="status-20261031-password-progress"></a>

### 2026-10-31 P3 开发登录通过，注销响应处理仍失败

维护者确认 TUN 关闭、有线独占、初始 offline 后，在私有终端完成修复包 source `67985fe2d2a580608d81504565cae172e594eb60222aeebca8d11f1305aa43e0` 的一次 password_core。脱敏回执严格解码和原包哈希匹配，规范化转录见本地 `build/auth-key-relation-live-20261031/`。

初始 offline、首次登录、独立 online 状态和重复登录 AlreadyOnline 均通过。CAS POST 为302，activation为200，SDK与状态机共同确认开发登录成功。首次 logout 收到200但返回 protocol_changed，整套开发验证为fail；随后 cleanup_logout 只查状态即成功，最终cleanup_status_offline通过。正常第二次注销及最终状态步骤没有完成，清理不能替代套件验收。

当前失败定位在注销响应处理；现有回执不能区分正文超限与业务解析拒绝，也不能推断实际 ecode 表示。新增合成状态转换回归通过，不改变产品成功判据。下一步设计完整闭集注销结构观察，涉及新的正文投影先确认，不直接放宽parser。

正式 password_core 仍 blocked/live_unverified，M3 in_progress；P3已有开发登录成功证据，但整套闭环未通过。Windows安装/junction缺口继续保留。此次单次真实授权已消耗，后续没有新的网络操作，旧包及历史失败记录保持。

<a id="status-20261030-key-relation"></a>

### 2026-10-30 P3 登录使用关系修复离线交付

按 [ADR-0021](../architecture/decisions/ADR-0021-cas-key-use-relation.md) 实现共享动态选择器，完整检查允许来源，以有限局部使用关系与 RSA 身份唯一性选键。合成注释诱饵、冲突、来源失败与实际 POST 解密回归通过；这尚不证明真实 HTTP 500 已消除。产品不再使用旧正则，旧诊断包/回执不改写，新构建匿名真实入口停用。

Windows/Linux 专项普通、race（Linux）、vet、有界 fuzz、六平台产品依赖隔离/构建，以及独立 auth-dev 密码开发包整链自检和绑定校验已通过。204 项源码快照与当前文件一致，source `67985fe2d2a580608d81504565cae172e594eb60222aeebca8d11f1305aa43e0`。详细检查与限制见本地 `build/p3-key-relation-20261030/REPORT.md` 和 `final-checks.json`；旧失败记录保留。

Windows 普通与诊断全仓测试通过，独立安装器 fresh/rollback junction 执行失败仍未闭合，不能声明整仓全绿。没有六平台原生认证证据。维护者现已开启 TUN 并登录用于远程管理，本轮不运行现场 preflight 或真实认证，不自动恢复 offline。

下一步为准备测试现场后另行授权修复包的 password_core；密码只在维护者私有终端输入。正式 password_core 仍 blocked/live_unverified，M3 in_progress，P3 真实闭环未完成，P4/P5 未推进。

### 当前身份与证据等级

| 身份 | 值 | 本摘要的依据与限制 |
|---|---|---|
| 权威工作区 | BHKDesktop `D:\project\Go\IPGW-Meta` | Windows 原生开发/验证与同机 Ubuntu WSL 离线开发；不依赖笔记本传输。 |
| 开发分支 / HEAD | `codex/cas-submit-field-closure` / `b4eeb760945f0568ca163d957ddd55df2afa86c4` | 本次只读 Git 核验；存在后续未提交开发与文档改动，均保留。 |
| PR #28 开发 head | `a14cc291fed335a6b027f42f72a9775f603aa3cd` | 迁移交接支持的历史身份；本次未重新查询远端 PR。 |
| Candidate source | `dac245947b9bd11b5489e2fd87ca2823c2d2dac5` | 不等于开发 HEAD 或 PR head。 |
| 不可变 Candidate | `v1.0.0-dac245947b9b-33738413131.1` | 历史构建/安装/attestation 门禁通过由交接支持；本次未重建、复跑安装或在线复验 attestation。 |

不得把当前工作区开发字节标为既有 Candidate，也不能因构建门禁通过宣布真实登录成功。旧段落中的 `fbce605…`、“尚无 Candidate”及“下一步 RC-BUILD”已由后续交接覆盖，不是当前行动依据。

### 已完成与仍未闭合

| 事项 | 当前状态与实际含义 |
|---|---|
| 本机开发环境 | Windows 与 Ubuntu WSL 的固定 Go 1.25.0、独立缓存/产物和重复执行入口已建立。原生回执及后续开发报告支持 Windows 普通测试/vet/doccheck、Linux 普通测试/race/vet/doccheck/质量门禁与六平台开发构建通过；不是本次重跑结果，也不等于六平台本机认证验收。 |
| 历史 Candidate 与安装 | 历史 RC-BUILD、六平台安装及相关绑定检查已完成；原件和安装保留，正式登录门禁未随之通过。 |
| 正式 password_core | 仍为 `blocked`，password 能力保持 `synthetic_covered + live_unverified`；没有目标账号真实登录闭环通过证据。 |
| CAS-submit R3 | 交接记载六字段 POST 收到 HTTP 500，`login_succeeded=false`；诊断分类 `pass` 不表示登录成功。该结果来自插桩诊断，不是 Candidate live-gate 通过。 |
| key-provenance R1 | 交接记载零个通用 publicKey、两个不同 publicKeyStr；排除该窗口的“通用 publicKey 抢先匹配”假设，尚未证明实际用于加密的键，也未证明它导致 HTTP 500。 |
| 历史 key-provenance R2 | 原包已存在 runner 内置 R1 哈希与 R2 EXE 不配对的问题；具体生成环节未知，真实 R2 校园运行没有执行证据。保留原件与结果，不作为新开发阻塞，不再默认继续旧来源追查。 |
| D4–D15 | 诊断实现、合成验证与若干本地/匿名窗口已有记录；D15 仍 `collection=blocked`，references/resources 为空，未取得 CAS 页面与取键事实。该专项退出默认主线，不再作为修复产品认证的预设前置。 |
| 后续验收与发布 | Wi-Fi password、Terminal QR 正式闭环、完整发布矩阵及 promotion/release 没有完成证据；不能由开发测试或诊断观察代替。 |

上表历史认证详情引用迁移交接及其后追加记录，本次未读取旧聊天、私有 evidence bundle 或原始响应，因此不是原始证据复验。D15 的本地脱敏说明在 `build/desktop-dev-20261028/D15-LIVE-RESULT.md`；其 JSON 为 PTY 输出的规范化转录，明确不是字节精确 stdout。一次真实授权已经消耗。

D15 的 AuthRoot `any_match` 与 Disallowed `no_match` 仅为限定集合的成员观察，不能推出完整信任、最新性、吊销政策或某一指定末端证书的结论。受限诊断使用显式 ROOT 池，产品 Windows 默认系统验证路径与之不同；此静态差异既不证明产品 TLS 必败，也不证明必然可用。

### 里程碑与下一工作包

M1/M2 保留历史完成状态；若后续复现缺陷影响相应门禁，按证据重开对应工作包，不预先抹除历史验收。M3 继续 `in_progress`；M0 残余治理继续 `in_progress`，按 [ADR-0011](../architecture/decisions/ADR-0011-nonblocking-m0-governance.md) 不单独阻塞发布状态门禁。

P0 文档收敛已完成，回执位于 `build/roadmap-docs-20261029/`；其交付窗口没有启动 P1。维护者随后要求开始，P1 `WP-AUTH-BASELINE` 已完成本地只读研究，交付 [旧版/v1 行为矩阵与判别假设](../research/auth-baseline-2026-10-29.md)，回执位于 `build/auth-baseline-p1-20261029/`。此完成仅指源码研究退出条件满足，不表示真实根因或认证闭环已解决。

P1 确认默认系统 TLS/直接 IPv4、RSA 主封装与六字段集合存在共同点；差异集中在门户预访问/会话边界、UA、页面与脚本取键、form/action/service 和后续 ticket/activation。当前仍无法由这些差异断定 HTTP 500 根因。P1 交付后维护者要求实施 P2；下面记录其后续执行，不改变 P1 窗口范围。

### 2026-10-29 P2 开发路径交付

P2 `WP-AUTH-VALIDATION-PATH` 的实现与专项离线验收已完成，契约见[开发验证](../operations/development-validation.md)及 accepted [ADR-0017](../architecture/decisions/ADR-0017-development-auth-validation.md)。匿名模式直接复用 SDK 与同轮内存 CAS 会话，在 provider 边界停止；password 模式使用同快照开发产品、现有七步状态机和私有 prompt。产品认证实现、公共接口和正式四组门禁没有改变；新规则只适用于独立开发工具。

- 最终包：`build/dev-runs/windows/20261029T152626Z-bfab731e-auth-dev/ready/`；source digest `6a6f6312d7d930a711f4a19fcd6969d68b30392824ac1f5e8579d27284a086df`，163 文件源码快照。启动器自检、产品/观察程序/构建输入绑定和六种完整/篡改场景通过。产品标记 `dev-auth-local`，不是 Candidate。
- Windows P2 测试/livegate 回归/vet 与 Linux P2 测试/race/vet/有界 fuzz 通过；Linux 全仓 race、两端全仓 vet、六平台共 18 个产品构建及依赖隔离通过。完整命令、结果和范围见 `build/auth-validation-p2-20261029/REPORT.md`。
- **Windows 全仓基线未全绿**：当前源码与诊断组通过，安装器的 fresh/upgrade 与三个 rollback 场景经 junction 启动失败，实际目标路径执行成功。正式工具审批与 WSL 官方 Windows interop 中均可复现；本轮没有证明具体环境根因，没有放宽测试、修改安装器或权限。这是保留的安装回归/执行环境缺口，不用历史通过覆盖本轮失败。
- 修正开发测试入口的包枚举：根包、`cmd/`、`internal/` 是当前源码范围，`build/` 中无 go.mod 的历史扫描副本不参与当前测试。旧副本和首次失败日志保留，未删除测试求通过。
- 本轮只执行合成/loopback 测试、构建、文档与脱敏检查；未执行真实匿名/password 模式、本地真实 profile 或 CTL 读取，也未提交、推送、构建 Candidate 或发布。开发登录与正式 `password_core` 均无新增真实成功证据。

后续主线为 P3：先另行授权一次产品等价匿名观察，区分 SDK 默认系统 TLS、会话与动态取键/表单上下文，再决定证据支持的最小修复或私有开发 password 窗口。真实选键关系和 HTTP 500 根因仍未知，不能预选首键/第二键。Windows 安装回归缺口继续单独跟踪；P4 正式矩阵尚未迁移，M3 与 password `live_unverified` 不变。

<a id="status-20261030-anonymous"></a>

### 2026-10-30 产品等价匿名观察

维护者新的单次授权及现场条件确认后，Agent 在 BHKDesktop Windows PTY 执行已绑定 P2 包一次；source `6a6f6312d7d930a711f4a19fcd6969d68b30392824ac1f5e8579d27284a086df`，启动/结束绑定 verified，退出码 0。闭集回执及解释位于 `build/auth-anonymous-20261030/`，JSON 为 PTY 规范化转录，不冒称字节精确原始 stdout。

结果为 `observed / credential_boundary_reached`：6 次应用请求（status 1、discovery 2、CAS 3），本轮 CAS Cookie 已接收并回送，页面已观察，SDK 候选公钥来源为 script，动态 action 的 service 与入口相同，产品命令 0。产品默认系统 TLS 与前置会话在这个窗口能够走通；D 系显式信任池的历史失败不能据此继续作为产品无法访问 CAS 页面的前置阻塞。

provider 边界早于 RSA 解析/加密与 CAS POST。本次未观察键数量/值或实际浏览器选键关系，不能与历史两键做值级比对，也未解释历史 HTTP 500。service 相同不等于完整 action URL 相同；应用计数不等于系统 DNS/TCP 总量。未读取 profile/密码、未进入 POST/activation/logout，没有额外状态请求或自动重试。

下一步为另行批准的私有开发 password_core 验证；本次授权已消耗。正式 password_core 仍 blocked，password 仍 `synthetic_covered + live_unverified`，M3 不变。未创建 Candidate、提交、推送或发布。

### 2026-10-30 私有开发 password_core 首轮结果

维护者在私有终端执行同一绑定 P2 包，返回闭集回执：binding verified，开发结果 blocked / state_machine_failed；initial_status_offline pass（exit 0），login_logged_in blocked（exit 3 / network）。本地转录和分析见 `build/auth-password-20261030/`，不包含终端账号提示或凭据；这是维护者回传结果，不是 Agent 直接采集的原始 stdout。

product_commands 3 包括 profile show、初始 status 和 login。step_coverage complete 只表示状态机正常返回，不能表示七步通过；password 模式请求/Cookie/页面字段未采集，其零值不是零网络证据。未进入后续 online 检查，未取得 cleanup 所有权，没有注销或结束时 offline 的证据。network 无法区分传输失败、超时或 HTTP 非成功状态，不能认定本次仍为历史 HTTP 500。

下一最小工作是离线收敛产品等价的失败阶段观察契约与接入方式，再决定相应实现和新的私有窗口；不自动重试、不恢复 CTL 主线、不猜测选键。单次授权已消耗，正式 password_core / M3 状态不变。

### 2026-10-30 P3 失败分类开发工具离线交付

维护者已确认 [ADR-0018](../architecture/decisions/ADR-0018-development-auth-failure-observation.md)，完整被动观察方案已实现并完成专项离线验收。独立 observed CLI 复用原 CLI/app/SDK 与私有 prompt，普通产品不启用观察；请求角色、HTTP状态和固定错误类别单独封装，不读取正文/header/query，也不改变选键、请求、TLS或cleanup语义。

最终包位于 `build/dev-runs/windows/20261030T040449Z-27fa0356-auth-dev/ready/`，source `8fb7735263af6130c8106a15c02c15a794e0698d1c9fcbf508cac990fe74706c`，176文件快照，分别绑定普通产品与 observed CLI；新外层 schema 2 / auth-dev-observed 不重解释旧回执。Windows专项测试/vet/自检/绑定、Linux专项普通/race/vet/fuzz及六平台普通产品依赖隔离/构建通过，详见 `build/auth-observed-p3-20261030/REPORT.md`。

本次只是观测工具就绪，不是 P3 认证修复完成。尚无新包真实结果，前次 network 根因未知；正式 password_core、M3 和 Windows 安装/junction 缺口均不提升。下一步仅为另行单次授权的私有开发 password_core，不重复匿名或 CTL，不读取真实配置/凭据或自动重试。

### 2026-10-30 P3 私有观察窗口：CAS POST 500

维护者随后在私有终端确认并执行上述新包，回传 schema 2 闭集结果；source 与普通产品、observed CLI 身份对应已交付包。本地规范化转录、严格解码和文件哈希核验见 `build/auth-observed-live-20261030/`。这是维护者回传证据，不是 Agent 采集的原始 stdout；前段“尚无新包真实结果”保留为交付时状态。

initial_status_offline 通过，login 以 exit 3 / network 停止；新增 SDK 类别为 cas_http_status。login 轨迹是 status 200、两次 discovery 302、三次 cas_get 200、cas_post 500，均 response_received=true、transport=none。加上独立初始 status，共记录 8 次应用 RoundTrip，不能作为操作系统 DNS/TCP 总量。本轮明确失败在 CAS POST 响应，未记录 activation；此前 GET/TLS 路径可用，不再以 CTL 专项作为默认前置。

该轨迹不区分三个 cas_get 的页面/脚本身份，不观察正文、Cookie、选键及表单关系；HTTP 500 尚不能归因为密码、密钥、会话或服务端具体异常。旧内层 result 的请求/Cookie/页面零值仍是 password 模式未采集字段，不覆盖 observations。step_coverage complete 不表示七步通过；未取得后续目标身份 online 或最终 offline 证据，未取得 cleanup 所有权。

下一步收敛到 P1 矩阵中的提交上下文：先离线核对 form/action/service、动态状态与 CAS Jar 连续性、页面/脚本取键和加密输入契约；只对独立复现且有规范依据的缺陷修复。新增真实材料观察或产品语义变化须单独决策，不直接改选第二个键、不重复同包登录。此次授权已消耗；正式 password_core 仍 blocked / live_unverified，M3 仍 in_progress。

### 2026-10-30 P3 提交上下文合成检查

已完成 [提交上下文核对](../research/cas-submit-context-2026-10-30.md)。新增 CAS/SDK 合成测试覆盖 action 形状与重定向后解析、Cookie 轮换/删除/作用域、两轮失败后会话隔离、同轮动态字段、生成 RSA 的实际解密及 UTF-16 长度、提交前拒绝和 500 后不重试。Windows 聚焦普通测试及 Linux 聚焦普通/race/vet 全部通过，报告分别位于 `build/dev-runs/windows/20261030T105227Z-7260082f-focus/`、`build/auth-submit-context-20261030/`。

本轮未复现需要修改产品实现的缺陷，没有改变认证或选键行为，也没有新增真实请求。真实 POST 的表单/会话/键关联仍未知；现有来源分析器不实现 JS 使用关系，不能仅以“两键”解释 500。下一最小方案是 [待确认的匿名上下文观察](../operations/development-validation.md#p3-anonymous-context-proposal)，只分析 SDK 实际消费来源并区分当前 parser 命中与页面目标关系；未确认前不实施。正式状态与安装回归缺口保持。

### 2026-10-30 P3 匿名上下文离线交付

维护者随后确认完整范围，[ADR-0019](../architecture/decisions/ADR-0019-anonymous-submit-context.md) 已 accepted。独立 `auth-context-dev` 复用已有 SDK 匿名路径，只分析实际完整消费的成功响应；输出表单目标关系、来源类别和 parser_match，浏览器使用关系仍 unknown，不改变产品选键或 ADR-0018 的 password 观察范围。

Windows 专项普通/vet、启动器自检与九类包检查，Linux 普通/race/vet/有界 fuzz 均通过；六平台产品依赖隔离和构建通过，不等于原生认证验收。最终包为 `build/dev-runs/windows/20261030T112836Z-6bb2cf77-auth-context/ready/`，source `b0dc49ea88070b8837b561e76ee66648cc81bceeae144c489965d75ec79e17d9`，175 项快照与当前源码/包文件哈希匹配。Agent 使用正式工具审批账户执行本地 `--preflight`，ready、live_attempted=false、应用请求/产品命令均为零；报告见 `build/auth-context-p3-20261030/REPORT.md`。

本次只完成独立工具的离线交付，没有真实来源、真实选键或新的登录结论，也未修复或解释 HTTP 500。原 P2/P3、D 系包和中间开发包保留；没有 Candidate、提交、推送或发布。下一步仅为另行单次授权的匿名上下文窗口，不需要 profile/密码；正式 password_core 仍 blocked / live_unverified，M3 与安装回归缺口保持。

### 2026-10-30 P3 匿名上下文窗口结果

维护者授权并确认现场条件后，Agent 在 BHKDesktop 原生 Windows 运行上述最终包一次，退出码 0，binding verified。`build/auth-context-live-20261030/` 保存 PTY 规范化闭集转录及报告；原包严格解码、四项哈希和 175 项工作区/包内源码快照核验通过。没有保存真实页面、脚本、Cookie、URL、键值或凭据。

SDK 到达 credential_boundary_reached：6 次应用 GET（status 1、discovery 2、CAS 3），本轮 CAS Cookie 接收/回送为 true。表单 selection、lt、execution 均 unique，同源 HTTPS，action 与初始/最终页面完整目标相同，service 相同。仅对本次匿名窗口成立，不回溯证明前次 password POST 的动态上下文正确。

成功响应来源 3 个（页面及两份外部脚本）；分析拆分为 17 个来源单元，不代表 17 次请求。parser_match 映射至第二份已消费外部脚本 S17 的 occurrence 1，但分析为 partial / unsupported_syntax，该匹配的上下文、标识符及赋值分类均 unknown，键未检查，use_relation 仍 unknown。keys/relations 为空不表示零键，也没有建立与历史两个 publicKeyStr 的相等关系；“第二份脚本”不能解释成“第二把键”。

本次缩小了匿名表单目标关系的疑点，未解释 CAS POST 500、未证明实际用键或登录成功。单次授权已消耗；没有 profile/密码读取、POST、activation、logout、额外脚本或自动重试。下一步先离线审阅分析器的不支持语法降级边界并设计合成判别；现有材料不支持猜测真实语法，若扩大语法/输出或实际使用关系观察须先形成集中决策。正式 password_core 仍 blocked / live_unverified，M3 与 Windows 安装回归缺口保持。

### 2026-10-30 P3 语法覆盖边界离线核对

已完成[来源分析器的降级路径核对](../research/cas-submit-context-2026-10-30.md)。新增合成回归证明：词法遇阻后的代码、注释或字符串匹配可以投影为同一 unknown 回执；赋值语法失败与词法失败不同，早先来源的 partial 不会污染后一个外部脚本。当前真实回执不能辨识其具体触发原因，不能推断存在某种真实语法或直接改选键。

Windows 来源分析普通/vet/自检及 Linux 普通/race/vet/有界 fuzz/自检通过，具体回执见研究记录。没有修改产品或观察语义、重新打包或新增真实请求；unknown 符合现行保守契约，尚未发现规范支持的产品修复。后续需集中评估静态语法覆盖范围，避免逐项原因采集与反复校园复测；任何语义/输出扩展或真实窗口均不由本轮自动授权。正式 password_core、M3 与既有缺口不变。

<a id="status-20261030-static"></a>

### 2026-10-30 P3 完整静态语法离线交付

维护者确认后，按 [ADR-0020](../architecture/decisions/ADR-0020-bounded-static-cas-syntax.md) 完成成熟 parser 的固定版本评估、预算派生、独立观察和打包。新 schema 2 的 context 保留旧回执含义，syntax 分别报告词法上下文、直接字面量赋值、本轮相等标签与函数/控制等语法位置；不执行 JS、不改变产品选键，use_relation 始终 unknown。完整范围、验证和限制见[研究记录](../research/cas-submit-context-2026-10-30.md)及[开发契约](../operations/development-validation.md#p3-static-syntax-proposal)。

最终包 `build/dev-runs/windows/20261030T124136Z-538b1b67-auth-static-context/ready/`，source `1c4215a2c3fdd753aa0162072af06b8ffbd4e3b33d5fdb9f9a9acaf4401a7a36`。199 项快照和包哈希一致；Windows 专项普通/vet/自检/原生 Job 与硬终止/九类绑定检查、Linux 普通/race/vet/有界 fuzz，以及六平台三个普通产品构建和诊断依赖隔离通过。Agent 自动完成最终包零网络预检，binding verified / ready，未执行匿名入口；回执见 `build/auth-static-context-p3-20261030/REPORT.md`。

这关闭的是新增静态观察的离线交付工作包，不是 P3 认证修复。没有真实新材料，不能预言此前 S17 必可解析；旧采集守卫、不支持语法和资源边界仍可能停止。正则只识别词法外壳，不承诺完整浏览器语法合法性；源码/许可证与派生哈希可重验，维护该派生补丁仍是后续升级成本。

下一步可另行批准最终新包的一次产品等价匿名窗口，观察同轮 SDK 匹配的静态类别；不是自动重试密码或重新启动 CTL 路线。原真实授权已消耗，本次未新增校园请求、profile/凭据读取或认证成功证据。正式 password_core 仍 blocked / live_unverified，M3 in_progress，既有 Windows 安装/junction 缺口保留；未创建 Candidate、提交或发布。

### 2026-10-30 P3 静态语法单次窗口：产品文本命中注释

维护者新的单次授权和现场条件确认后，Agent 自动执行上述最终包一次，source `1c4215a2c3fdd753aa0162072af06b8ffbd4e3b33d5fdb9f9a9acaf4401a7a36`，binding verified，退出码 0。脱敏回执为 PTY 规范化转录，严格解码、四项包哈希与 199 项快照核验通过，见 `build/auth-static-context-live-20261030/REPORT.md`；不冒称字节精确 stdout，不保存真实页面或密钥。

本轮达到凭据边界：6 次应用 GET（status 1、discovery 2、CAS 3），本轮 Cookie 接收/回送，form/lt/execution 均 unique，HTTPS/初始/最终 action 及 service 关系均 same。三个成功消费来源分解为 17 个分析来源；这些 ID 是本轮局部索引，不是 17 次网络请求，也不能与历史来源做内容级比对。

旧 context 仍为 partial / unsupported_syntax；新 syntax 为 observed，parser_match 指向 S17（第二个已消费外部脚本）的 occurrence 1，**context=comment、assignment_class=none**。这直接支持“产品文本匹配会把脚本注释中的公钥文本作为候选”的具体问题，与现有生成密钥合成回归一致。新报告 keys 为空表示本观察未识别可确认的 RSA 字面量赋值，不表示真实页面不存在键、原匹配值不是 RSA，或后续未消费脚本也无键。

本轮没有密码、RSA 加密或 CAS POST，尚不能把 HTTP 500 唯一归因为该注释，也不能推断浏览器实际用键。下一步应离线形成针对注释/字符串/子串候选的产品取键契约与回归方案，再决定安全选键和脚本继续发现规则；不直接选择第二个键或把诊断 parser 自动接入产品。

本次授权已消耗，未重试、额外抓脚本、读取 profile/凭据、执行 JS、POST、activation 或 logout。正式 password_core、M3 和既有安装缺口保持；没有新增产品修复、Candidate、提交或发布。

## 当前实现快照

截至 2026-08-28，以下内容已进入重写后的 `codex/v1-freeze` 和已验证的 fresh clone，但“实现存在”不代表对应发布门禁已完成：

- 文档/agent 分层、稳定 ID 索引与 `doccheck` 已建立；工作树中的真实测试 fixture 已移除并由合成 fixture 替代，secret scan 配置及其合成 canary 已接入，自更新代码与可达入口已移除。
- 根 `ipgw` SDK façade、typed errors、接口枚举和 internal CAS/Srun/Dashboard 边界已实现；HTTPS-only、ticket HTTP 下一跳进程内截获、动态 CAS 表单/公钥、JSON/JSONP、最终身份核对、账号冲突、幂等 logout、挑战分类和脱敏 Observer 均已有合成或 `httptest` 覆盖。
- `ipgw`、`ipgw-meta`、`ipgw-legacy` 三入口、模式优先级、稳定 JSON envelope/退出码、named profiles、keyring/env/file/prompt provider、事务化双来源配置迁移及原子 bundle/安装脚本已实现。Config 读取、跨进程 mutation lock、pending-journal 写阻断、preview 目标代次核对和 keyring 提交失败清理均已加固并有确定性 race 覆盖。
- CI/release workflow、六目标交叉构建定义和 release gate 已写入仓库；实现与治理已通过 PR #1 以普通 merge 合入 `main` 的 `f927d7316885a26c8289ba77bc04ed27e379d3c8`（tree `4c5f253efebd5cea7bc85b0b5de5c2af84ed54f3`），但尚未形成 candidate-set、安装或真实校园网证据，因此不能外推为 M3 完成。

M1 的本地实现与自动化门禁已完成。M2 配置迁移、三入口、离线事务安装器与六平台原生 release-shaped asset 门禁均已完成；这些临时 PR artifact 不是 M3 candidate-set 或公开 release，不得外推为 M3 完成。

## 2026-08-28 冻结前只读核验

`WP-M0-PREFLIGHT` 已完成且没有创建提交、修改 refs 或接触远端写操作：

- GitHub CLI 认证和只读 API 请求成功；远端 `main` 与 `v0.1.0`–`v0.1.3` 均与本地记录一致。
- `main`/`v0.1.3` 为 `3721fd60dbdc8e6a35cab9569ec295b6fcc7bdea`；当前 HEAD tree 为 `a61feedcc1c5dd7f11b933e0d6235d5fe8565728`。
- 工作树共有 146 项变化（33 tracked、113 untracked），index 为空；存在两个指向 tree object 的 `refs/codex/**`，它们不得进入 mirror 推送范围。
- `core.autocrlf=true` 且尚无 `.gitattributes`，存在 LF/CRLF 机械污染风险；`WP-M0-FREEZE-AUDIT` 必须先建立明确 EOL 策略。
- 固定 `git-filter-repo 2.47.0` 脚本 SHA-256 为 `67447413E273FC76809289111748870B6F6072F08B17EFE94863A92D810B7D94`；gitleaks v8.30.1 可用且合成 canary 通过。
- 当前源码快照 secret scan 为 0；全历史有且仅有 4 条已知命中，分布在旧测试 fixture，尚未重写清除。
- `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...`、`go run ./cmd/doccheck --check` 均通过。核验机 Go 为 1.26.1；v1 构建基线仍以 `go.mod` 的 Go 1.25.0 和 workflow 的 `go-version-file` 为准。
- 维护者已于 2026-08-28 通过官方门户确认泄露会话失效；未保存截图或认证材料。

## 2026-08-28 历史重写与 rehome 核验

- 冻结提交 `ba7897d142a5a3d5a8a7c6219a6b735823d14bb8` 已映射为 `38fadd1bef3692f52a8c9a1b67db45819b57112c`，冻结 tree 保持 `9b3765dba53fe8e98d4c3f9b7c5d78c6edb26b91`。
- 隔离 mirror 的全历史扫描由 4 条已知命中降至 0；7 个 commit 的父图、元数据与消息保持一致，`fsck`、不可达对象、reflog、tag 和 tree 门禁通过。
- GitHub 上 `main`、`codex/v1-freeze` 与 `v0.1.0`–`v0.1.3` 已通过 6 条 per-ref lease 的单次 atomic push 更新并完成 refs/branch API 复核；无 PR 或 `refs/codex/**` 隐藏引用。
- `D:\project\Go\ipgw-meta-clean` 已从重写后远端 fresh clone；4 个重写前 commit 对象均不存在，全历史 secret scan、`fsck`、test/race/vet/doccheck 全部通过，成为后续唯一权威工作区。
- GitHub 对象 API 仍可直接访问 3 个已失去 ref 可达性的旧 commit，因此仍需维护者向 GitHub Support 请求清理缓存/悬空对象；该外部步骤完成前 `SEC-HISTORY-001` 和 M0 不得标记 complete。

## 2026-08-28 仓库小写改名、治理与 baseline 核验

- GitHub repository ID `1186323753`（node ID `R_kgDORrXdKQ`）已从 `UnbalancedCat/IPGW-Meta` 小写改名为 `UnbalancedCat/ipgw-meta`；default branch、两个 branch refs、`v0.1.0`–`v0.1.3` 四个 lightweight commit tags 及全部 SHA 均未漂移。当前权威工作区的 `origin` fetch/push URL 已更新为小写 canonical URL。
- 冻结 tree 原已统一使用 `github.com/UnbalancedCat/ipgw-meta`：Go module/import、源码 URL、文档、workflow、安装器和 release script 中没有旧 GitHub 仓库路径残留，因此本窗口没有对品牌名、旧工作区隔离路径、备份路径或 Windows 安装目录做机械大小写替换。
- 改名后已 fresh clone 到独立新目录 `D:\project\Go\ipgw-meta-govern-verify-20260828-r2`。该 clone 的 `main` 为 `5ad8c1fa05102fbcb249e7195f4801563b3d44a5`，验证分支为 `codex/v1-freeze` / `38fadd1bef3692f52a8c9a1b67db45819b57112c`，冻结 tree 为 `9b3765dba53fe8e98d4c3f9b7c5d78c6edb26b91`；refs、tag 类型、`fsck`、test/race/vet/doccheck、合成 gitleaks canary、全历史/reflog 与安全 tree scan 均通过，最终状态 clean。
- 合入后的精确 `main` `f927d7316885a26c8289ba77bc04ed27e379d3c8` / tree `4c5f253efebd5cea7bc85b0b5de5c2af84ed54f3` 已完成 `WP-BASELINE-VERIFY`：`go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...`、`go run ./cmd/doccheck --check` 全部通过；Windows/Linux/macOS 的 amd64/arm64 共 18 个三入口二进制交叉构建及 `internal/config` 六目标编译通过。
- 固定 gitleaks `8.30.1` 的上游 release assets 已同时按固定 SHA-256 与官方 checksums 核验；合成正负 canary、14 commits 的全历史/全部 refs/reflog 扫描及最终工作树扫描均为 0 命中。`git fsck --full` 通过，仅报告两个已知 dangling tree；全部本轮临时工具、cache 与构建目录已精确清理，工作树最终 clean。
- 首次 CI run `33166812911` 的 macOS 失败已按 docs-first 顺序修复：[`ADR-0010`](../architecture/decisions/ADR-0010-macos-trusted-system-path-alias.md) 由 signed commit `f1ca77c1096a84d5048af72395fd4449a34a9ffc` 建立；signed commit `907b3e754b67cf759421eb4326c44367b14fe78a` 使用逐组件 `openat(O_NOFOLLOW)`、只允许 Darwin 固定 `/var` → `/private/var` 受验证锚点，并增加原生 `/var/folders/...`、symlink 与 FIFO 回归测试。维护者登记 Signing key 后，post-clean signed commits `15fb31db059b660e03cf5460483cbf2f0aa0cbda`、`c562483cc21239246367d65a08687e20ea9c5356`、`5bee31b3ea07bb023c3b38b9465f7d12ed0caabb`、`f1ca77c1096a84d5048af72395fd4449a34a9ffc` 与 `907b3e754b67cf759421eb4326c44367b14fe78a` 当前均由 GitHub 验证为 `verified=true, reason=valid`。
- 最终 PR head `219da2f87112d54c8ad0f30ad13e45c39a0f8076` 的 CI run `33176975227` 六项精确 context 全部成功，且逐项复核均来自 GitHub Actions App `15368`；该提交由 GitHub 验证为 `verified=true, reason=valid`。PR #1 已以保留拓扑的普通 merge 合入，merge commit `f927d7316885a26c8289ba77bc04ed27e379d3c8` 的双亲精确为旧 `main` 与该 PR head，并由 GitHub 验证为 `valid`。
- active main ruleset `main-v1-protection`（ID `21733128`）无 bypass，现要求 PR、签名提交、严格六检查并禁止删除和 force-push；active tag ruleset `v-tag-protection`（ID `21733211`）无 bypass，禁止 `refs/tags/v*` 更新和删除。规则读回后 `main` 与四个既有 tag SHA 均未漂移。
- `WP-M0-RENAME-GOVERN` 与 `WP-BASELINE-VERIFY` 已完成。未对冻结提交或既有分支拓扑做追溯改写，也未使用 squash/rebase、force-push、release 或 tag 写操作。

## 2026-08-29 配置迁移事务收敛核验

- `WP-M2-CONFIG-CLOSE` 的 signed implementation commit `598850195a65167d121c2fc86477cf56676bb8df` 已补齐逐事务、无包级全局状态的 journal phase 失败注入，覆盖 `prepared`、`backups`、`keyring`、`config` 与 `marker_verified` 的自动回滚或已提交清理语义。
- 两类合成旧来源现同时验证 backup 固定写入 `BaseDir/migration-backups/`、名称精确为 `<source-kind>-<transaction-id>-<ordinal>.backup` 且完整保留原字节；Windows 原生 runner 验证受保护当前用户 DACL，Ubuntu 与 macOS 原生 runner 验证目录 `0700`、文件 `0600`。
- 合成 keyring backend 现覆盖“写入后返回错误”的补偿删除，以及补偿删除失败时 journal/backup 保留、错误脱敏和后续恢复重试；未调用真实系统 keyring 或接触真实凭据。
- PR #3 首轮 CI run `33224027909` 的六项 required checks 全部成功；本地 `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...`、`go run ./cmd/doccheck --check`、`internal/config` 六目标测试编译、固定 gitleaks `8.30.1` 合成 canary/15 commits 历史/精确 source tree 扫描及 `git fsck --full` 均通过，临时工具、cache、构建与扫描目录已清理。
- `WP-M2-CONFIG-CLOSE` 完成；M2 仍因安装器与原生安装矩阵未完成而保持 `in_progress`，后续工作包从 `WP-M2-INSTALL-UNIX` 继续。

## 2026-08-29 Unix 离线安装器收敛核验

- `WP-M2-INSTALL-UNIX` 已实现规范固定的 `--bundle`/`--bundle-sha256` 成对接口；离线来源必须为绝对、本地、非 symlink、`1..100 MiB` 且不可 group/world writable 的普通文件。安装器只从已打开的来源句柄复制到本轮私有 acquisition 目录，再对私有副本核对调用方 SHA-256；离线路径不初始化或调用下载器。
- 在线与离线 acquisition 后共用外层哈希、精确七成员名称/类型/大小、总量和压缩比限制、私有 bounded extraction、内部 `SHA256SUMS`、canonical manifest 与事务激活链。install root、bin dir 与 launcher dir 均执行绝对非根、不重叠和逐祖先无 symlink 校验；新建正式目录、version/binary/metadata 与私有 stage/backup/journal 使用规范模式。
- 激活事务现按已验证 version 发布、旧 active 分离、新 active 切换、三入口、launcher、PATH phase 和 commit 顺序写入受限 journal；失败时逆序恢复，无法安全恢复或命中 rollback failpoint 时保留 recovery materials 并 fail closed。测试控制仍只接受规范列出的四个变量，且仅在离线、当前用户私有测试根、精确 token 和全部输入/目标严格位于测试根时生效。
- Unix 测试覆盖 fresh install、upgrade、三入口 `--version`、launcher 默认/保持、公开与私有权限、离线无网络、输入/路径/归档拒绝、9 个固定前向 failpoint 及 3 个固定 rollback failpoint。WSL Ubuntu 本地完整矩阵通过；PR #4 implementation head `5dfe60ea152e4fd23677fe8cc18f4e2b59e151f5` 的 CI run `33227444605` 在 Ubuntu 与 macOS 原生执行上述测试，并与 Windows 全仓测试、race、vet/doccheck/secrets 和六目标 cross-build 一并六项成功。
- 本地 `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...`、`go run ./cmd/doccheck --check`、Git Bash 语法、Linux/macOS 两架构测试编译、固定 gitleaks `8.30.1` canary/全历史/reflog/工作树扫描及 `git fsck --full` 均通过；`fsck` 仅报告三个既有 dangling tree。
- `WP-M2-INSTALL-UNIX` 完成；合成 bundle 的原生 runner 测试不等于公开 release asset 的六平台 smoke，因此 M2 继续保持 `in_progress`，下一工作包为 `WP-M2-INSTALL-WINDOWS`，之后仍需 `WP-M2-INSTALL-NATIVE`。

## 2026-08-29 Windows 离线安装器收敛核验

- `WP-M2-INSTALL-WINDOWS` 已实现规范固定的 `-BundlePath`/`-BundleSha256` 成对接口；离线来源必须为固定本地磁盘上的绝对、非 reparse、`1..100 MiB` 普通文件，并拒绝 Users、Authenticated Users 或 Everyone 可写 ACL。安装器以不共享写入/删除的已打开句柄复制到本轮私有 acquisition 目录，再仅对私有副本核对调用方 SHA-256；离线路径不初始化下载 URL、TLS 或 HTTP client。
- Windows 在线与离线 acquisition 后共用外层哈希、精确七成员名称/类型/大小、总量和压缩比、私有 bounded extraction、内部 `SHA256SUMS`、canonical manifest 与事务激活链。install root、bin dir 与 config dir 均执行固定磁盘、绝对非根、不重叠、拒绝仓库/用户根和逐祖先无 junction/reparse 校验；install root 与 bin dir 还要求同卷，避免把跨卷复制伪装为原子入口发布。
- 当前用户安装布局使用受保护的当前用户、SYSTEM 与 Administrators ACL；事务依次发布已验证 version、分离/切换 `active` junction、原子发布三个 hard-link 入口、launcher、PATH phase 与 commit，并把受限状态持久化到原子替换的 journal。失败时逆序恢复；命中 rollback failpoint 或无法证明恢复安全时保留 transaction/backup/version recovery materials 并 fail closed。测试控制仍只接受规范列出的四个变量，且 PATH phase 在私有测试根内使用固定状态文件，不修改真实用户注册表 PATH。
- Windows 原生测试覆盖 fresh install、upgrade、三入口 `--version`、launcher 默认/保持、目标和文件 ACL、离线无网络、UNC/重叠/junction 祖先/宽松来源 ACL/异常归档拒绝、9 个固定前向 failpoint 与 3 个固定 rollback failpoint。首轮 PR #5 CI run `33237097662` 暴露从 pwsh 7 父进程继承的不兼容 `PSModulePath`；signed fix commit `b5a4d5c4c6e4f4e6fb48d3361fdb94a7b26905c0` 改为 .NET ACL API 并清理测试子进程继承路径后，run `33237394630` 的六项 required checks 全部成功。
- 本地 Windows PowerShell `5.1.26100.9168` 下完整安装器矩阵、`go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...`、`go run ./cmd/doccheck --check`、固定 gitleaks `8.30.1` 合成 canary/全历史/reflog/工作树扫描及 `git fsck --full` 均通过；`fsck` 仅报告三个既有 dangling tree。
- `WP-M2-INSTALL-WINDOWS` 完成；合成 bundle 的 Windows 原生测试不等于公开 release asset 的六平台 smoke，因此 M2 继续保持 `in_progress`，下一工作包为 `WP-M2-INSTALL-NATIVE`。

## 2026-08-29 六平台原生 release-shaped 安装矩阵核验

- `WP-M2-INSTALL-NATIVE` 经原生矩阵验收的 implementation head 为 signed commit `d54a30d085d9663c03970cd66b76f5df13216b0b`。package job 从精确 PR test-merge source commit `05035df77c4e75586e9cd5b03d569cc17a0a5e78`、tree `5eb94413504bda1d8231ec99a1081aaf7435666f` 生成版本 `native-05035df77c4e`；不可变 artifact ID 为 `9713909266`，digest 为 `sha256:51341d93b6eb09e758e2a62450312abca18400294d385d8a6e986a39dead9d5d`。
- 六个原生 runner 下载并核验同一精确 artifact，没有在 runner 上重建或重新打包。`linux-arm64`、`windows-arm64`、`darwin-amd64` 执行 release-shaped asset smoke；`linux-amd64`、`windows-amd64`、`darwin-arm64` 执行包含全部固定 failpoint、rollback failure、路径攻击和权限矩阵的 full 门禁。
- 首轮 Windows job 超时的根因是测试子进程环境 allowlist 遗漏 `PSModuleAnalysisCachePath`，导致 Windows PowerShell 每次启动重建 module analysis cache。signed fix `d54a30d085d9663c03970cd66b76f5df13216b0b` 仅恢复该缓存路径，未提高任何 timeout。
- implementation-head CI run `33249498526` 的六项 required checks 全部成功；native workflow run `33249498529` 的 package 与六个原生 runner 共七个 job 全部成功。
- 本地 `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...`、`go run ./cmd/doccheck --check` 全部通过；固定 gitleaks `8.30.1` 的合成 canary、implementation-head 全历史/全部 refs/reflog 与精确工作树扫描均为 0 命中。`git fsck --full` 通过，仅报告四个 dangling tree，且无损坏或 dangling commit。
- PR #6 已以普通 merge 合入 merge commit `989cfad32aaed7352c50fb9e80233ac137362616`；双亲依次为原 main `6d9cde533d4d3bc511cc8122a96bfcfa114988a8` 与 PR head `130a698399d83b2f7cda2ab7581ac6ef4c60b8a1`，tree 为 `8c280f9c92dc739e74425a4be52e6abdae364775`，GitHub 签名为 `verified=true, reason=valid`。merge 后 CI run `33254131004` 六项检查全部成功，native run `33254130929` 的 package 与六个平台共七个 jobs 全部成功。
- `WP-M2-INSTALL-NATIVE` 与 M2 已完成，下一工作包为 `WP-M3-LIVEGATE-SCHEMA`。本节 artifact 是仅用于本 PR 原生门禁的临时 release-shaped 验证产物，不是 M3 candidate-set、公开 release、tag 或发布资产。

## 2026-08-29 Live-gate schema 收敛核验

- `WP-M3-LIVEGATE-SCHEMA` 已按 docs-first 顺序固定 schema version 1：严格 18 个顶层字段与 5 个 step 字段、封闭枚举和四个精确环境 tuple、candidate/evidence ID 与小写 hash、canonical UTC time、capability transition、primary prefix/cleanup/result 以及产品/runner 退出码映射。
- `internal/livegate` 已实现严格 UTF-8/64 KiB/unknown/duplicate/trailing JSON guards、canonical 单 LF 编码、固定且不回显输入的验证错误，并以泄漏 canary 覆盖 JSON、枚举、ID、时间、状态机与 cleanup 拒绝路径；focused coverage 为 94.3%。
- 本地 `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...` 与 `go run ./cmd/doccheck --check` 全部通过。
- 固定 gitleaks `8.30.1` 的合成 canary、28 commits 全历史、全部 refs/reflog 与工作树扫描均为 0 命中；`git fsck --full` 通过，仅报告四个已知 dangling tree，且无损坏或 dangling commit。
- signed implementation commit `03ffa893dce68bf83886ca047b2c9c5760a351b9` 由 GitHub 验证为 `verified=true, reason=valid`。PR #8 已以普通 merge 合入 `01e4dc59bd7787cb382e9d2392f7e6c3052a569b`；双亲依次为 `0aaff5da9ac691bcb56538074a0b3c178b140808` 与 `03ffa893dce68bf83886ca047b2c9c5760a351b9`，tree 为 `6fa99e2f23efa0ef2ac6d1b269b52bcda0514148`，GitHub 签名为 `verified=true, reason=valid`。
- merge 后 CI run `33259519538` 六项 required checks 和 native run `33259519515` 的 package 与六个平台共七个 jobs 全部成功；独立 release push run `33259518906` 仍以零 job 失败，须在 `WP-M3-CANDIDATE` / promotion 前诊断。
- `WP-M3-LIVEGATE-SCHEMA` 完成；M3 保持 `in_progress`，下一工作包为 `WP-M3-LIVEGATE-RUNNER`。本工作包未实现 runner、candidate-set、attestation、release、tag 或 promotion，未启动校园网、QR、认证或其他网络会话。

## 2026-08-30 Live-gate runner 收敛核验

- `WP-M3-LIVEGATE-RUNNER` 已实现 maintainer-only runner：冻结 candidate/manifest/hash 绑定、固定 suite 状态机与清理权、私有三文件 evidence bundle、提交点与 durability 语义，以及不保留产品原始输出的结构化结果投影；公共 SDK、CLI、稳定 JSON、退出码、workflow、candidate/promotion 和安装器均未改变。
- Linux 证据目录发布以 dirfd、`openat(O_NOFOLLOW)` 和 no-replace rename 约束并发创建，既有非私有目录只验证且 fail closed；candidate identity 指纹纳入 ctime。Windows 新建 bundle 文件显式设置当前用户 owner 与受保护 DACL，外来或 defaulted owner 继续拒绝。非 Linux/Windows 的真实 evidence publication 继续精确 fail closed，未扩展平台能力。
- PR #10 首轮 CI run `33269797459` 暴露 Linux inode 复用/并发目录创建、macOS fail-closed 测试边界和 Windows 默认 owner 差异；signed fix commit `80462712f872519da73526927476fdad69edee32` 在不放宽验证器的前提下关闭这些问题。implementation commits `e5528f46792f7d9d3d087b2b59196106d6856976` 与 `80462712f872519da73526927476fdad69edee32` 均由 GitHub 验证为 `verified=true, reason=valid`。
- 修复后本地 `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...` 与 `go run ./cmd/doccheck --check` 全部通过；Linux/macOS 全仓交叉构建和 `internal/livegate` 测试编译通过。固定 gitleaks `8.30.1` 合成 canary、签名前 31 commits 的全历史/全部 refs/reflog 和精确 174 个工作树文件扫描均为 0 命中；`git fsck --full` 无损坏或 dangling commit，临时扫描与构建目标已清理。
- PR #10 最终 head 的 CI run `33271001499` 六项 required checks 全部成功，native run `33271001474` 的 package 与六个平台共七个 jobs 全部成功。PR 已以普通 merge 合入 `6196234374f72089affd5442d0b5c2c0193cf62d`；双亲依次为 `7585eb6c08d4e4471ae4783bd8c27fa10a9ebf23` 与 `80462712f872519da73526927476fdad69edee32`，tree 为 `b9383072cd8c32bf1f0aafd6596b4a6c0ae86077`，GitHub 签名为 `valid`。
- merge 后 CI run `33271253178` 六项和 native run `33271253166` 七项全部成功。同一 merge SHA 的 release push run `33271252579` 仍以精确零 job 失败，证明既有发布 workflow 缺口仍存在；它不是 main required check，进入 `WP-M3-CANDIDATE` 修改前必须先只读诊断，且不得据此创建 tag、release 或发布资产。
- `WP-M3-LIVEGATE-RUNNER` 完成；M3 继续保持 `in_progress`，下一工作包为 `WP-M3-CANDIDATE`。本工作包未创建 candidate-set、attestation、release、tag 或 promotion，也未启动校园网、QR、认证或其他真实网络会话；`REL-LIVE-MATRIX-001` 仍未执行。

## 2026-08-30 不可变 Candidate 流水线收敛核验

- 既有 `.github/workflows/release.yml` 的零-job failure 已完成只读根因复现：其 job 级 `env` 在 workflow 处理阶段引用不可用的 `runner` context；固定 actionlint `v1.7.12` 精确拒绝原第 74 行。该会在 tag 阶段重新构建并直接公开 release 的旧 workflow 已删除；CI 现固定安装 actionlint，并由 `internal/workflowguard` 同时锁定 workflow 触发、权限、action pin 与不可重建边界。
- `WP-M3-CANDIDATE` 已实现 maintainer-only `workflow_dispatch`：只接受精确 `v1.0.0` 与受保护 `main` 的完整 source SHA，验证 source 签名、tag 不存在及同 SHA 的六项 CI 和七项 Native checks；本轮按 [`ADR-0011`](../architecture/decisions/ADR-0011-nonblocking-m0-governance.md) 把状态 gate 收敛为 M1–M2，M0 不参与判定。单一 build job 按冻结 toolchain/build-input 只构建一次，生成 canonical full/release manifest、确定性六平台 bundle、两级 checksums、公开十资产与私有两 helper；六平台原生安装只下载同一不可覆盖 artifact，attestation job 重新下载并 full-verify 后仅为 candidate-set 与十个公开资产生成 11 个 subjects。
- 本地 focused candidate/CLI/installtest/workflowguard、actionlint、Linux 执行测试、Windows/Linux/macOS amd64/arm64 交叉构建、`go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...` 与 `go run ./cmd/doccheck --check` 均通过。固定 gitleaks `8.30.1` 合成 canary、全历史/全部 refs/reflog 与精确 205 个工作树文件扫描均为 0 命中；`git fsck --full` 无损坏或 dangling commit，扫描和交叉构建临时目标均已清理。
- PR #12 首轮 CI run `33279922238` 暴露 Unix rename ctime、测试暂存导致 executable mode 漂移以及 shellcheck 重定向/字面 Markdown 告警；signed fix `73f5aa30a29aebb60970319ea378bb20b62bdf06` 保留同一已打开对象的 identity/mode/size/mtime/hash 与目录精确成员验证，仅对 Unix 根目录 no-clobber rename 的 ctime 变化使用重命名后验证。implementation commit `74cb8f6e8dab02b2d1d935640acdb885ea19ff77` 与 fix 均由 GitHub 验证为 `verified=true, reason=valid`。
- PR #12 最终 head 的 CI run `33282461292` 六项 required checks 和 Native run `33282461260` 的 package 与六个平台共七个 jobs 全部成功。PR 已以普通 merge 合入 `382a8f994761a4a5d73f578d08f263718eac958c`；双亲依次为 `f907e270273bb376b37439e149567fa0398ed976` 与 `73f5aa30a29aebb60970319ea378bb20b62bdf06`，tree 为 `0206c0a06d3c8c46fc1225c5440d8a61b7ad2554`，GitHub 签名为 `valid`。merge 后 CI run `33282708893` 六项和 Native run `33282708924` 七项再次全部成功。
- `WP-M3-CANDIDATE` 完成，但 M0/`SEC-HISTORY-001` 仍为 `in_progress`，因此 `make candidate-gate` 按预期拒绝，Candidate workflow 未 dispatch，未创建正式 candidate-set、artifact、attestation、tag、release 或发布资产。M3 继续保持 `in_progress`，下一工作包为 `WP-M3-PROMOTION`；`REL-LIVE-MATRIX-001` 仍未执行，也未启动校园网、QR、认证或其他真实网络会话。

## 2026-08-30 Promotion 流水线收敛核验

- `WP-M3-PROMOTION` 已按 docs-first 顺序冻结 closed-world canonical `promotion-lock.json`、四份 public evidence summary、精确 release notes、source→promotion 白名单与 build-input 重算契约；Promotion workflow 只接受签名 annotated `v1.0.0` 和受保护 main 当前 tip，按精确 run/artifact API、原始 artifact digest、14 文件 Candidate、两级 manifest/checksum、11-subject attestation 与四份 evidence 绑定做封闭验证。
- Promotion 全程 no-build/no-repack：禁止 setup-go、build、package、strip、sign、clobber、单资产替换和自动 mutation 重试；只允许一次创建不可见 draft，逐项重新下载核对十个公开资产，再在同一步 mutation 前复核 gate/lock/tag/main/draft，最后一次公开并再次 fresh download 验证。stdlib verifier 及 7 组合成 happy-path/tamper 测试、workflowguard 与 CI actionlint 门禁均已落盘。
- 本地 `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...`、`go run ./cmd/doccheck --check`、actionlint 与六目标 cross-build 全部通过；固定 gitleaks v8.30.1 合成 canary、签名后 50 commits 的全历史、全部 refs/reflog 与精确 209 文件工作树扫描均为 0 finding。`git fsck --full` 无损坏，仅报告 1 个 dangling blob 与 7 个 dangling tree，全部临时扫描/cache/build 路径均已清理。
- signed implementation commit `466276fd28a855370a10ca8421117811b4e4ef13` 由 GitHub 两个 commit API 验证为 `verified=true, reason=valid`。PR #14 head CI run `33290209456` 六项 required checks 和 Native run `33290209465` 的 package 与六个平台共七个 jobs 全部成功；PR 已以普通 merge 合入 `e542f108a32e37f8313c9673cbd68254af25968c`，双亲依次为 `7660348949745ed4193a545c452097cbc91f9c92` 与 implementation head，tree 为 `0d2d0dcb9439bccc67945d1b94d367374a90f776`，GitHub 签名为 `valid`。
- merge 后 CI run `33290459305` 六项和 Native run `33290459311` 七项再次全部成功。`make promotion-gate` 因 M0 `in_progress` 按预期拒绝；Candidate/Promotion workflow 均未 dispatch，未创建或使用正式 candidate-set、artifact、attestation、promotion lock/evidence、tag、draft、release 或发布资产，也未启动校园网、认证、QR 或其他真实网络会话。
- `WP-M3-PROMOTION` implementation 完成；三个旧对象的 GitHub Support 外部待办完成前，`SEC-HISTORY-001` 与 M0 继续保持 `in_progress`，M0 仍禁止新 release。M3 也继续保持 `in_progress`；下一工作包为独立只读、无凭据且不创建 VM 的 `LAB-DISCOVER`。

## 2026-08-30 LAB-DISCOVER BHK 匿名路径核验

- ZOS 隔离能力尚未按 [`REL-LAB-002`](../runbooks/campus-lab.md#rel-lab-002发现与供应) 证明，未创建或启动 VM。BHK fallback 通过经人工核对 host key 的 WSL SSH 链路调用原生 Windows 网络栈；exact-main Windows/amd64 binary、受保护空配置、物理有线/Wi-Fi、TUN 与覆盖路由消失、非 fake-IP DNS、constrained-source 物理出口及 RealVNC 连通性均在请求前通过核对。
- 配置目录首次预检因 Windows 临时目录继承 DACL 而在创建 HTTP client 前 fail closed；两次命令均未发出请求。切换到仅当前用户、SYSTEM 与 Administrators 可访问的受保护空目录后，纯离线 `profile list` 成功且未创建配置文件。
- 维护者批准有线与 Wi-Fi 各一次替代 `status` 窗口；实际仅有线执行一次并返回有效 `protocol_changed` 契约，随后按停止规则终止，Wi-Fi 未执行，总真实请求数为 1。未使用凭据，未运行 `network scan`、login 或 logout，未采集、保存或输出原始响应；该结果按 [`REL-LIVEGATE-002`](../operations/live-validation.md#rel-livegate-002suite-状态机与清理权) 为 fail，不构成正式 live evidence，也不能外推为协议、认证或 M3 完成。
- 代码核对发现状态格式无法识别时 producer 使用未列入 CLI 脱敏 allowlist 的 `ProtocolPart="status"`，导致安全诊断字段被清空。docs-first 修复把该路径统一为规范允许的 `gateway_status`，未放宽解析器、改变 endpoint/transport 或增加响应暴露；合成 SDK/CLI 测试同时证明允许值保留、未知值清空且 response canary 不进入错误 JSON。
- signed implementation commit `8c1e24aa66824062580ec380d4bb2acce41ae991` 经 PR #16 的 CI run `33310700735` 六项 required checks 与 Native run `33310700716` 七项全部成功后，以普通 merge 合入 `ce57f4e2474055defba00195645db363729d3533`；merge 双亲、tree `cd9ee5cb7c41e0900917b37d2e2491eac5f92e8d` 与 GitHub 签名均已精确复验，merge 后 CI run `33310968357` 六项与 Native run `33310968406` 七项再次成功。
- implementation tree 的 focused test、`go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...`、`go run ./cmd/doccheck --check` 全部通过；固定 gitleaks `8.30.1` 的合成 canary、39 commits 全历史/refs/reflog 与 209 文件工作树扫描为 0 finding。`git fsck --full` 无损坏或 dangling commit，仅报告 1 个 dangling blob 与 14 个 dangling tree。M0、`SEC-HISTORY-001` 与 M3 均继续保持 `in_progress`。

- 维护者随后批准 exact-main 的单次有线、无凭据协议形状诊断。请求前再次核对远端 main、已知 rulesets 与 refs、source/tree/signature、helper hash、受保护空配置、system proxy、物理网卡、TUN/覆盖路由为零、非 fake-IP DNS、constrained-source 有线出口和 RealVNC；全部通过后仅执行 1 个新网关 round trip。BHK 匿名阶段累计真实请求数为 2，Wi-Fi 仍未执行。
- 封闭结果为 `http_class=http_2xx`、非重定向、`body_size_bucket=1_1024`、`leading_shape=other`、全部状态/身份/流量字段与 marker presence 为 false、`legacy_csv_shape=true`，产品分类为 `protocol_gateway_status` / `gateway_status`，包装器确认 `gateway_requests=1`。该结果只证明 HTTPS 返回了现有解析器未接受的单行多列 legacy CSV-like 形状，不证明字段含义、会话、认证或正式 evidence；未采集、保存或输出正文、响应头、URL、账号、IP 或原始 stdout/stderr。请求后 postflight 通过，本地 helper、WSL staging 与 Windows 受保护 helper 目录均已精确清理；旧 exact-main 产品 binary 与受保护空配置未改动。
- 本次 docs-only closeout tree 的 `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...` 与 `go run ./cmd/doccheck --check` 全部通过；固定 gitleaks `v8.30.1` 的合成 canary、40 commits 全历史/全部 refs/reflog 与工作树扫描均为 0 finding。`git fsck --full` 无损坏或 dangling commit，仅报告 1 个 dangling blob 与 14 个 dangling tree。

## 2026-08-31 legacy CSV 状态兼容核验

- docs-first 已固定 positional legacy CSV 的最小 online 身份不变量：去除外层空白后必须是单个无内嵌 CR/LF、至少 9 列的 CSV record，位置 0 为合法 username，位置 8 为全局单播 IPv4；其他列保持 opaque，不参与身份、online/offline 或冲突推断。
- 位置 6/7/11 只形成 all-or-nothing 的可选 summary 候选。流量与时长必须同时为 `int64` 范围内的非负十进制整数，非空余额必须可精确转换为 minor units；否则只保留已经成立的 online 身份并令 summary 整体为 `nil`，不返回部分摘要。该降级仅适用于 legacy CSV；JSON/JSONP 显式命名 summary 的无效、部分或 alias 冲突继续返回 `protocol_changed`。
- 实现只调整 `internal/srun` 的 legacy CSV 解析顺序并增加内部与公共 SDK 合成边界测试；公共 SDK/CLI/JSON/退出码、endpoint、transport、请求次数和脱敏边界均未改变。聚焦普通/race 测试与全量 `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...`、`go run ./cmd/doccheck --check` 全部通过。
- 固定 gitleaks `v8.30.1` 的合成 canary、41 commits 全历史/全部 refs/reflog 与工作树扫描均为 0 finding。`git fsck --full` 无损坏或 dangling commit，仅报告既有 1 个 dangling blob 与 14 个 dangling tree。
- signed implementation commit `dd7b563936a6c2472896fc4e2fd288e5d0cae536` 由 GitHub 验证为 `valid`；PR #19 head CI run `33323660497` 六项 required checks 与 Native run `33323660534` 七项全部成功，并以普通 merge 合入 main `2348f9a0bf208e56303222e03c0137618ff700fb`。merge 双亲依次为 `ea152221813020578c69bdfb604cc07ec00edcbe` 与 implementation commit，tree 为 `51abf026059e6a4e005952306ad046d785326764`，GitHub 签名为 `valid`；merge 后 CI run `33323891972` 六项与 Native run `33323892073` 七项再次全部成功。未执行新的校园网请求；任何真实有线 `status` 仍须取得新的单请求批准，M0、`SEC-HISTORY-001` 与 M3 继续保持 `in_progress`。

## 2026-08-31 exact-main 有线匿名复核窗口

- 维护者批准通过现有 `BHKDesktop-WSL` 使用最多 4 次 strict SSH/SCP 连接，重新取得并验证 exact-main Windows/amd64 artifact，并只在物理有线上以受保护空配置尝试一次无凭据 `status`；完整 preflight 不通过时必须保持网关请求为 0。请求前复核确认 main `147bfa9bcaf357b175bc183f6e67543223a62d8b` / tree `335677e39ca390f55eea576c72250aa6a3c8d372`、签名、完整 refs、rulesets 与 Native run `33324692028` 均无漂移；artifact `9735897428` 的外层 digest、manifest、release hashes、Windows archive、目标 SHA-256 与 build info 全部逐层匹配。该 artifact 不是 Candidate、tag 或 release。
- 第 1 次连接只创建固定 WSL 私有 staging，第 2 次只传输 exact binary 和封闭 runner；第 3 次通过远端成员/hash 门禁后，Windows topology preflight 以 `topology_preflight_internal` 在产品启动前 fail closed。封闭结果精确为 `request_attempted=false`、`status_invocations=0`、`gateway_request_upper_bound=0`，且 Windows 与 WSL 固定目标均确认清理。源码复核确认 `status` 只有一个 `client.Do`、无重试且拒绝 redirect；因此本窗口实际产品调用和网关请求均为 0。
- 剩余第 4 次连接仅重复只读 topology 诊断，不运行产品、不创建文件且网关请求为 0；其远端 stderr 非空，因此本地封闭输出契约拒绝结果，未读取、保存或回显错误正文，也未安全确定更细的失败步骤。4 次连接额度到此全部消耗，没有重试 `status`。该窗口不构成 live evidence，legacy CSV 兼容仍未获得 live 复核；继续远端诊断或新的有线 `status` 需要新的明确授权。M0、`SEC-HISTORY-001` 与 M3 继续保持 `in_progress`。
- 本轮远端 Windows/WSL 固定目标均已确认清理；本地已忽略的 artifact/runner 目录也在精确路径、零 reparse、零 tracked 成员门禁后删除并复核 absent。收尾 `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...` 与 doccheck 全部通过；固定 gitleaks `v8.30.1` 的合成 canary、63 commits/50 refs/178 reflog entries 全历史与 209 文件工作树扫描均为 0 finding。`git fsck --full` 无损坏或 dangling commit，仅报告 1 个 dangling blob 与 23 个 dangling tree；扫描临时目标残留为 0。

## 2026-08-31 BHKDesktop Win11 有线匿名兼容复核

- 维护者随后改用现有 Xftp/Xshell 会话在 BHKDesktop 的 WSL2 中人工执行自包含封闭 harness；Codex 未再直接 SSH/SCP。WSL2 只负责上传、staging、调用 Windows PowerShell、字节计数与清理，八阶段 topology 查询和 Windows/amd64 产品均使用 Windows 11 原生网络栈。zero-DNS 合成 preflight 与一次 real-DNS preflight 最终完整通过，且没有输出或保存 IP、接口、MAC、SSID、DNS 答案、URL、命令、路径或原始错误。
- status bundle 内嵌干净 clone 上以 Go `1.26.1` 构建的 exact-source Windows/amd64 binary：大小 `8242176`、SHA-256 `158357fde02a590ee0bee921627a21fded05e0c12ced49dc6c096d16c8070bd5`、`vcs.revision=147bfa9bcaf357b175bc183f6e67543223a62d8b`、`vcs.modified=false`。它不是 Native run `33324692028` 的 Go 1.25 artifact，也不是 Candidate、tag、release 或正式 evidence。
- 最终 bundle SHA-256 为 `e72659dfdbf1165142591eb6c2542b0aeed7572bd44ca345980ad50366d8ffb4`，远端 shell 自动校验返回 `upload_present=true`、`sha256_match=true`。封闭结果为 `result=pass`、`child_phase=complete`、`dns_invocations=1`、`status_invocations=1`、`gateway_request_upper_bound=1`、`status_outcome=success`、`network_state=reachable`、`session_state=online`、`cleanup_confirmed=true`；因此 legacy CSV 兼容实现已在该 exact-source build 的 BHKDesktop Win11 物理有线真实响应上通过匿名 parser compatibility 复核。
- 本次没有运行 login/logout、使用凭据、切换 Wi-Fi、执行 `network scan`、修改网络或读取原始产品输出/身份/网络标识。`session_state=online` 只表示匿名 status 观察到既有在线会话；按 [`REL-LIVEGATE-001`](../operations/live-validation.md#rel-livegate-001runner-接口与信任边界)，正式 suite 的 initial status 必须明确 offline，Agent 不得为测试自动 logout。本结果不提升认证能力、不进入 promotion evidence，也不改变 M0、`SEC-HISTORY-001` 或 M3 的 `in_progress` 状态。
- `cleanup_confirmed=true` 确认固定 WSL staging、Windows `%LOCALAPPDATA%` 目标和上传脚本均已清理。已忽略的本地 harness 通过 PowerShell 5.1 synthetic/AST/命令闭集、Bash mapper/E2E、确定性重建、gitleaks、doccheck、`git diff --check` 与 Git ignore/tracked/index 门禁，保持 0 个 tracked 路径且未暂存、未提交。

## 2026-09-01 LAB-DISCOVER ZOS 只读能力结论

- 维护者提供的非唯一性设备与版本信息确认测试对象为双口 Z4Pro，系统版本 `V1.0.0430216`、服务版本 `V1.0.0430398.2632`；只读 UI 显示当前为解绑 Bond，两个物理口各自拥有宿主 IP 并支持多网关。期间未记录账号、完整 IP/MAC 或其他网络标识。
- ZOS 官方 UI 同时提供“双网口加入同一网桥”和“双网口分别创建网桥（适用于软路由）”；后者可为两个物理口分别建立网桥并供 VM 指定网口。官方[网络设置相关注意事项](https://www.zspace.cn/help/?cid=0&articleId=100197)说明软路由网桥仍可由 ZOS 配置 IP，并警告使用主网口的 VM 会阻止该端口故障时的自动切换、可能导致 ZOS 管理面失联；官方[虚拟机功能说明](https://www.zspace.cn/help/?cid=1068&articleId=100167)也只证明桥接 vNIC 与物理口选择能力。
- 现有正式支持能力未证明 NAS 宿主可让测试物理口同时保持无 IP、DHCP 和默认路由，也未展示把该端口独占直通给 VM 的选项。依 [`REL-LAB-002`](../runbooks/campus-lab.md#rel-lab-002发现与供应) 的“能力不明即 blocked”规则，`LAB-DISCOVER` 只读窗口以 `blocked` 结果完成；没有切换网络模式、保存设置、创建网桥/VM、修改网络或启动认证。
- ZOS 后续仅用于离线 Linux 安装和无凭据合成测试，NAS 真实网络路径不进入 `LAB-PROVISION`。BHKDesktop Win11 物理有线匿名 parser compatibility 的既有 `pass` 结论保持独立，不构成 Candidate、正式 live evidence 或认证结果。

## 2026-09-01 M0 非阻塞发布门禁决策

- 维护者接受 GitHub 仍可按对象 ID 访问三个已失去 ref 可达性的旧 commit 这一残余治理风险，并决定 M0/`SEC-HISTORY-001` 继续保持 `in_progress`、继续跟踪 GitHub Support 与既有副本处置，但不再阻塞 Candidate、Promotion、release 或真实验收；不得把该决定改写成 M0 已完成。
- [`ADR-0011`](../architecture/decisions/ADR-0011-nonblocking-m0-governance.md)把状态判定固定为：Candidate 只要求 M1/M2 `complete`；Promotion 只要求 M1/M2 `complete` 且 M3 `in_progress`；release 只要求 M1/M2/M3 `complete`。M0 的缺失或任意状态均不影响这三个 gate 的结果。
- required CI、完整可达历史/refs/reflog 与工作树 secret scan、泄漏测试、source/main/tag 签名、ruleset、Candidate identity、build-input、artifact digest、attestation、真实网络 evidence、promotion lock 和逐资产复核均保持独立 hard gate；本决定不授权 Candidate dispatch、tag、draft、release、校园网或认证操作。
- 共用 `scripts/milestone-gate.sh` 已成为 Makefile 三个状态 gate 的唯一判定入口；53 个合成用例证明 M0 的 `not_started|in_progress|blocked|complete` 与缺失均不改变结果，同时覆盖 M1/M2/M3 缺失、重复和错误状态的封闭拒绝。当前正式 status 下 Candidate 与 Promotion gate 通过，release 因 M3 仍为 `in_progress` 按预期拒绝。
- 本地聚焦与全量普通/race 测试、vet、doccheck、actionlint、7 项 promotion verifier、六目标交叉构建、Bash AST、固定 gitleaks 合成 canary/43 commits 全历史/16 个改动文件扫描和 `git fsck --full` 均通过；fsck 无损坏或 dangling commit，仅报告既有 1 个 dangling blob 与 25 个 dangling tree。
- 本工作包只实现规范、状态 gate 与合成回归测试，并提交签名 PR 等待 required CI；不自动 merge。合入后下一工作包为 `RC-BUILD` 的独立授权窗口；ZOS `blocked`、BHK 既有会话 `online` 和正式 live matrix 的其他约束不因 M0 解耦而改变。
- PR #21 head 的 CI 六项与 Native 七项全部成功，并以普通 merge 合入 main `3786382bfe0b91c428d45166d4dbb046b57c720e`；merge 双亲、tree `f0a7cf1fff2a66d91e2939e3611502d1ae4923fe` 与 GitHub 签名均已精确复验，合入后 CI run `33463522316` 六项和 Native run `33463522437` 七项再次全部成功。维护者随后授权进入 `RC-BUILD`。
- 第一次 Candidate dispatch run `33464985630` 在任何 build/upload/attestation 前由 preflight 封闭失败：check-run hard gate 的 jq 使用了不可编译的 `all(.[] as $name; ...)` 形式；source identity、immutable inputs 与 checkout 已通过，后续所有 jobs 均 skipped，未创建 artifact、attestation、tag、draft 或 release。修复把 filter 拆为可直接执行合成测试的 `scripts/candidate-checks.jq`，保留精确 13 项 checks、最新同名 run、GitHub Actions app/source SHA 和 100 项分页上限约束；修复合入并通过新 main checks 前不得再次 dispatch。
- jq 修复的 signed commit `377ad0c587ae60555d190c7f08af3604d5303915` 经 PR #22 的 CI 六项与 Native 七项全部成功后，以普通 merge 合入 main `bd9df2d16edd1eaaf9c79d8fe2c7956606f321c5`；merge 双亲、tree `fa785418736d6b70b30f9446e667ae065facaa82` 与签名已复验，合入后 CI run `33466147317` 六项和 Native run `33466147353` 七项再次成功。
- 第二次 Candidate dispatch run `33466493360` 的 preflight 与唯一 build/upload job 成功，artifact `9785072853` 名称为 `candidate-set-v1.0.0-bd9df2d16edd-33466493360.1`、服务端 digest 为 `sha256:0c5bd5fd1ef68619f0c7444221ba03b347a563d6bc0630257ea367b05af496a5`。同一 artifact 在 Linux/macOS 四项原生验证通过，但 Windows amd64/arm64 均在安装前的 full verifier 阶段失败；根因是 workflow 传入 `D:\.../ipgw-candidate-set` 样式的绝对混合分隔符路径，而 CLI 未按 `ABS_PATH` 接口在调用核心 verifier 前执行 `filepath.Clean`。本地按精确 artifact 下载后以 clean absolute path 完整验证通过，Candidate-set SHA-256 为 `3fb183c76e9d00a43ac7b204d6b39df85a09f57cce7fe20680a31e3a54dd1763`；attestation job 按依赖 skipped，因此该 artifact 不是可接受 Candidate，不得进入 `LAB-TRANSFER`。
- CLI 路径适配的 signed commit `b2ef9bac5a6984e28ee2cdf4e5505a0c42db8d27` 经 PR #23 的 CI run `33467789575` 六项与 Native run `33467789580` 七项成功后，以普通 merge 合入 main `8f6831511aaaad04af8f3f5749310295427a2055`；merge 双亲、tree `25950fc821b9aebfcd89ba161074d6dc9d2d6b51` 与签名已复验，合入后 CI run `33468055045` 六项和 Native run `33468055065` 七项再次成功。
- 第三次 Candidate dispatch run `33468379539` 的 preflight、build/upload 和六平台 full verifier 均成功；artifact `9785669867` 名称为 `candidate-set-v1.0.0-8f6831511aaa-33468379539.1`、服务端 digest 为 `sha256:f42ccd0468a0e432f59df3b4ee3d0e871c90cd2d1f079fd9346644481f4ec8aa`。Linux/macOS 四项安装通过，但 Candidate-mode installtest 从同一未规范化 `IPGW_NATIVE_INSTALL_ARTIFACT_ROOT` 再次直接调用严格核心 verifier，导致 Windows amd64/arm64 在 installtest 阶段失败；attestation skipped。精确 artifact 的本地核心验证与修复后的 Windows/amd64 Candidate-mode 安装 smoke 均通过，Candidate-set SHA-256 为 `a3d91988d230716b25207bb91d9331a58cbedbc9a7911b59ba185ed3ce995ac5`。该 artifact 同样不是可接受 Candidate，不得进入 `LAB-TRANSFER`；修复只在 installtest 绝对路径环境适配层执行 `filepath.Clean`，核心 verifier 保持不变。
- installtest 路径适配的 signed commit `9844957d60a86c294351cb9d3b77e53010e892a5` 经 PR #24 的 CI run `33469870407` 六项与 Native run `33469870436` 七项成功后，以普通 merge 合入 main `d851cbf3b3e4500a3b9db57d37a8b3913e0e57c4`；merge 双亲、tree `663806ff014ecdf4ab1356c7f0af9a2a19e32c4f` 与签名已复验，合入后 CI run `33470174146` 六项和 Native run `33470174179` 七项再次成功。
- 第四次 Candidate dispatch run `33470568609` attempt 1 绑定受保护 main `d851cbf3b3e4500a3b9db57d37a8b3913e0e57c4` / tree `663806ff014ecdf4ab1356c7f0af9a2a19e32c4f`，preflight、唯一 build/upload、六平台 full/smoke 原生安装和 attestation job 全部成功。唯一 artifact `9786395563` 名称为 `candidate-set-v1.0.0-d851cbf3b3e4-33470568609.1`、大小 `55088183`、未过期，服务端 digest 与独立下载的 raw archive SHA-256 均为 `sha256:76fdb20f94053c02e807e54b15d1b105e44dd0a3ec75b981fc6f24d095652f13`。14 文件 closed tree、两级 checksums、full/release manifests、全部成员与 build-input 已由本地 full verifier 复验；candidate-manifest SHA-256 为 `d45772925464b80e1050ebb9cd81a38d7ed8646b437e07c1e6d1ac610c5801ea`，build-input SHA-256 为 `1ed7d8aef7778a12e5a699ea7de7fe23fe0afb218f35d0d6e3e3e9b829f97c9a`。artifact attestation `44319316` 与十公开资产 attestation `44319319` 共覆盖 11 个精确 subject，并已逐文件约束 repository、signer workflow、`refs/heads/main`、source SHA、run invocation 与非 self-hosted runner 独立验证；原始 ZIP 与 candidate-set 本地副本均已设为只读。该 Candidate 可进入 `LAB-TRANSFER`，但尚未传输、运行真实网络 suite、创建 tag、draft 或 release。

## 2026-09-01 LAB-TRANSFER BHK Windows

- 本地从只读 Candidate 精确提取 Windows/amd64 产品 binary、对应 maintainer helper、candidate manifest 和 checksums，生成 ignored 单文件 transfer 包；包大小为 `5286558`，SHA-256 为 `1a69d15c89776203b2433cd74dd00e2c0ec6eeb1e47fe15e145648102bc4bb4e`。归档成员闭集、payload/installer hash、PowerShell 5.1 AST/命令闭集、首次安装/精确复用/错误源/冲突目标、WSL 包装器 E2E、gitleaks、doccheck、ignore/tracked/index 与受保护文档前置 hash 门禁均通过；包不含源码、`.git`、token、凭据或原始日志。
- 维护者使用现有 Xftp/Xshell 会话把该单文件上传到 BHKDesktop-WSL2，shell 自动核验返回 `upload_present=true`、`sha256_match=true`、`stage_ready=true`。封闭 Windows 私有落盘结果为 `result=pass`、`candidate_id=v1.0.0-d851cbf3b3e4-33470568609.1`、`remote_reused=false`、`windows_payload_verified=true`、`product_invocations=0`、`network_requests=0`、`cleanup_confirmed=true`；WSL 上传包和 staging 已删除，Windows 当前用户 `%LOCALAPPDATA%` 下的固定私有 Candidate 保留。远端没有 build、package、strip、sign 或改写 Candidate。
- 该结果完成 `LAB-TRANSFER` 的可执行 BHK Windows 路径，只证明本地与 BHK 私有 Candidate 的一致 hash，不构成产品、校园网、认证或 evidence 运行。ZOS/NAS 因 [`REL-LAB-002`](../runbooks/campus-lab.md#rel-lab-002发现与供应) 隔离能力未获证明继续 fail closed；Linux payload 只在本地 ignored harness 中准备，未上传或执行。下一可执行工作包是 `LAB-PASSWORD-BHK`，但先前匿名状态为 online，真实 runner 所需 initial offline 尚未由维护者在私有界面建立并证明。

## 2026-09-01–02 BHKDesktop 管理链路与有线 password 诊断

- 维护者在 BHKDesktop 原生 Windows 私有 PowerShell 中完成 Candidate password profile 准备；修正 helper 的 profile identity 核对后返回 `profile_ready=true`、`profile_reused=true`、`product_invocations=2`、`network_requests=0`。机器上另有旧版 Meta 安装，因此人工观察不得用于证明 Candidate 行为；正式 runner 和后续诊断始终使用固定 Candidate 私有路径与 hash。
- 2026-09-01 首次 `password_core` 因 initial status 未明确 offline 而以 runner exit 11 / blocked 停止，没有进入 login。维护者随后通过 NAS 旧版 Meta 临时恢复同一上游路由器认证，暂停 LAB 并改造管理通道；该恢复只用于保持管理能力，不属于 Candidate、evidence 或测试成功。
- 管理工具最终把 Cloudflared 固定为原生 IPv6 路径，静态 audit 曾同时证明 TUN disabled、原生 IPv6、loopback 2233 代理、Cloudflared 与 RealVNC 进程以及所需配置形状。维护者在私有管理员 PowerShell 中分别应用 Cloudflared 和 RealVNC 配置并重启对应服务各一次，期间没有产品/Candidate 调用。账号再次 offline 后，Win-CF SSH 保持可重新连接而 RealVNC Viewer 一度不可用；维护者随后执行固定 `RollbackRealVNC`，NAS 恢复在线后 Viewer 重新连接。故后续 BHK 窗口以 Win-CF SSH 为主要管理面，RealVNC 只作次要通道，不视为 offline 可达硬证明。
- 在 Wi-Fi 关闭、TUN 关闭、单一物理有线、NAS 恢复路径 idle 且自动重登关闭、维护者从官方门户确认账号 offline 的条件下，2026-09-01 后续复测与 2026-09-02 重新精确绑定有线 profile 后的复测都通过 `initial_status_offline`，但 `login_logged_in` 以产品退出码 6 / `protocol_changed` 失败，runner 总体 exit 10 / fail。两次均未出现 password prompt；封闭 summary 与 NAS 路由器 offline 复核证明没有把其他接口在线状态误当作本次会话，也没有形成成功证据。
- 这两次失败发生在 credential provider 之前，没有读取或提交密码。未保存或回传 HTML、headers、URL、Cookie、ticket、LT、账号、IP、接口、SSID/MAC 或原始 stdout/stderr；M3 与 `AUTH-PASSWORD-001` 状态不提升。

## 2026-09-02 `ac_id` 与 CAS HTML 协议收敛

- 旧 Candidate 的零凭据 precredential 诊断先返回 `ac_id_not_discovered`。在同一精确有线/offline 环境中的封闭 shape 诊断证明：第一个匿名响应只给出同网关 host、无查询参数的重定向；有界跟随一次后，第二个同 host 重定向才在受支持 portal 路径的查询与页面 assignment 中提供一致的十进制 `ac_id`。诊断没有触达 credential provider，也没有保存原始响应。
- docs-first 修复只允许无凭据匿名发现跟随受限的同网关重定向，不放宽 CAS、ticket、Cookie 或凭据传输。signed commit `ab1cca6d506b9c569a6fe7fd50bc12e15ce1d234` 经 PR #25 的 CI run `33585698264` 与 Native run `33585698135` 全部成功后，以普通 merge 合入 main `ed367bb27ec073e18635fcfe727e1db735fb5f3d` / tree `a757f41f0552eecfafd4d3e45a2e294d1d89d184`；head/merge GitHub 签名均为 `valid`，merge 后 CI `33586054877` 与 Native `33586054884` 再次全部成功。
- Candidate run `33586390643` attempt 1 绑定该 merge source，preflight、唯一 build/upload、六平台 full/smoke 安装与 attestation 全部成功。Candidate ID 为 `v1.0.0-ed367bb27ec0-33586390643.1`，唯一 artifact `9830238126`、大小 `55128743`，服务端 digest 与独立 raw archive SHA-256 均为 `57dc4763a0bb82e3f2610b5179787b298e6782c7caeb510f53e26dc59f3ddf98`；candidate manifest SHA-256 为 `3358ba6b9dc6a09c82a6e87fe3fc4299c4242b0db02d2bac46fd4f308de3fe0e`，build-input SHA-256 为 `112762f92bd1b1dbc6617ad0201a088c41bdd53e55315b7152160ed4eddabd60`。本地 ignored 只读副本通过 full verifier 与 attestations 复核。
- 新 Candidate 的 Windows transfer 包大小 `5291655`、SHA-256 `3d51bd286d20ffa8860d1af02efa66ec1d548bb18b808b6cedfa7c8098cb7e97`。维护者人工上传后，固定 shell 自动验 hash、解包和 Windows 私有落盘返回 exact candidate ID、`remote_reused=false`、`windows_payload_verified=true`、`product_invocations=0`、`network_requests=0`、`cleanup_confirmed=true`；WSL 上传与 staging 已清理。
- 该 Candidate 的零凭据 precredential 不再卡在 `ac_id`，而是在四个匿名请求后以 `cas_page/page_parse_failed` 停止。进一步的封闭 CAS shape 只记录固定布尔/枚举：响应为完整 2xx HTML，可解析同源 login form/action，`lt`、`execution` 和 password input 存在，同源 login script candidate 存在，inline public key 不存在；challenge parse 失败。未输出或保存页面、脚本、URL、网络标识或认证材料。
- 根因是 CAS challenge 分类在检查 leading HTML 之前运行 JSONP 启发式，普通 HTML 的内联 JavaScript 函数括号与对象字面量使整页被错误送入严格 JSONP decoder。[`ADR-0013`](../architecture/decisions/ADR-0013-cas-html-before-jsonp.md) 与实现将可选 BOM/空白后的 leading `<` 固定优先分类为 HTML，同时保留严格 JSON/JSONP、active challenge 和认证失败的 fail-closed 语义。signed commit `96b5c810aea36d05427654f818e3592ea5a7f34a` 经 PR #26 的 CI run `33602941140` 全部成功；Native run `33602941165` 首次仅因 GitHub artifact 下载中间层 403 使 macOS amd64 job 失败，只重跑该 job 后七项全部成功。
- PR #26 已以普通 merge 合入受保护 main `fbce60521ba767b133e2dda70a2e2dc19aa06f3b`，双亲为 `ed367bb27ec073e18635fcfe727e1db735fb5f3d` 与 implementation commit，tree 为 `f455a4226b658eaf958f32cb19eda7b129305402`，GitHub 签名为 `verified=true, reason=valid`。合并后 CI run `33603787537` 六项与 Native run `33603787553` 七项全部成功。没有在该 source 上启动 Candidate、transfer 或任何真实请求；两个旧 Candidate 因不包含完整修复均不得继续形成正式 evidence。

## 2026-08-27 本地验收结果

- `go test -count=1 ./...`、`go test -race -count=1 ./...`、`go vet ./...` 与 `go run ./cmd/doccheck --check` 均通过。
- `ipgw`、`ipgw-meta`、`ipgw-legacy` 在 Windows/Linux/macOS 的 amd64/arm64 共 18 个二进制构建通过；`internal/config` 测试在同六目标编译通过，Windows/amd64 race 实际执行通过。
- `install.ps1` PowerShell AST、`install.sh`/`scripts/release.sh`/`scripts/test-gitleaks.sh` Git Bash 语法以及 `make -n ci` 均通过。它们不是六平台实际安装、升级、回滚证据。
- gitleaks 正/负合成 canary 与排除 `.git`、`build`、module cache 后的当前源码快照扫描通过。截至该次 2026-08-27 验收尚未执行历史重写；后续处理与复核结果见上节。

## 尚未完成与外部条件

- `SEC-HISTORY-001` 仍为 `in_progress`：会话失效确认、冻结提交、受限历史备份、全 refs 重写、tag 复核、全历史复扫、atomic per-ref lease 远端更新和本机 fresh clone 已完成；GitHub Support 缓存/悬空对象清理以及其他既有副本的重新克隆仍待完成。
- 仓库治理、只读 baseline、所有 M2 工作包、`WP-M3-LIVEGATE-SCHEMA`、`WP-M3-LIVEGATE-RUNNER`、`WP-M3-CANDIDATE`、`WP-M3-PROMOTION` 与只读 `LAB-DISCOVER` 均已完成。旧 source 上的 `RC-BUILD`/BHK `LAB-TRANSFER` 已分别形成并验证过两个不可变 Candidate，但后续 `ac_id` 与 CAS HTML 产品代码修复使它们不再有资格继续真实 evidence；必须从当前 main `fbce60521ba767b133e2dda70a2e2dc19aa06f3b` 重做 `RC-BUILD`。三个旧对象完成 GitHub Support 清理前，`SEC-HISTORY-001` 与 M0 继续保持 `in_progress`，但按 ADR-0011 不阻塞下一工作包。
- 六平台原生 release-shaped asset 安装矩阵已使用同一精确临时 PR artifact 完成；该 artifact 只构成 M2 原生门禁证据，不是 M3 candidate-set、公开 release、tag 或发布资产。
- 既有 `.github/workflows/release.yml` push runs `33259518906` 与 `33271252579` 的零-job 根因已复现并随 `WP-M3-CANDIDATE` 关闭；旧 workflow 已删除，candidate workflow 与 CI actionlint/workflowguard 门禁已合入。历史 Candidate `v1.0.0-d851cbf3b3e4-33470568609.1` 与 `v1.0.0-ed367bb27ec0-33586390643.1` 各自 artifact/attestation 仍保持不可变，但都早于当前 source，不得用于后续正式 suite 或 promotion。
- `WP-M3-CANDIDATE` 与 `WP-M3-PROMOTION` 已完成并冻结符合 [ADR-0007](../architecture/decisions/ADR-0007-immutable-candidate-promotion.md) 的一次构建、验证、attestation 与 no-build 原样晋升路径。当前 main 尚无可接受 Candidate，promotion lock/evidence、tag、draft 和 release 均未生成；M3 继续为 `in_progress`。下一阶段是在已批准的独立 `RC-BUILD` 窗口从 `fbce60521ba767b133e2dda70a2e2dc19aa06f3b` 生成新 Candidate；仍不创建 VM、不切线、不自动启动认证。
- `LAB-DISCOVER` 已完成：ZOS 路径因未能证明 [`REL-LAB-002`](../runbooks/campus-lab.md#rel-lab-002发现与供应) 要求的宿主测试口无 IP/DHCP/default route 而以 `blocked` 收敛，且没有通过改网或创建 VM 继续探测。BHK fallback 的 Win11 物理有线匿名 topology 与 exact-source parser compatibility 已按上节封闭通过，但 Wi-Fi、正式 Candidate/evidence 与认证矩阵均未执行；匿名结果观察到既有会话 online，不能作为 live-gate initial offline 或认证成功证据。M3 继续保持 `in_progress`。
- `REL-LIVE-MATRIX-001` 尚未完成：BHK campus wired 的旧 Candidate runner 已产生一个 initial-status blocked bundle 和两个在 credential provider 前以 `protocol_changed` 失败的 bundle，均不提升能力状态；campus Wi-Fi password 与 `network list/scan`、以及至少一种网络的 Terminal QR 闭环仍未执行。新 Candidate 必须先完成 transfer 与零凭据 precredential 门禁，才能继续正式 password suite。

## 当前已知能力限制

- v1 在具备可靠 network fingerprint 前禁用持久协议缓存 fallback；绑定 IP 不作为网络身份，每次操作执行动态发现。`WithProtocolStateStore` 仅为未来扩展缝，详见 [`SDK-CACHE-001`](../reference/go-sdk.md#sdk-cache-001协议状态存储)。
- `AUTH-SMS-001` 手机验证码仅为 `observed_anonymous + detected_only`，目前没有可触发的真实测试条件，不阻塞 v1。
- `AUTH-QR-001` Terminal QR 在真实校园网验证前保持 `live_unverified`。
- `AUTH-PASSWORD-001` 在当前候选完成有线、无线真实验收前保持 `live_unverified`。
- 当前可用性仅有 2026-08-27 的人工使用报告，不构成发布门禁证据。
- M0 残余治理继续 `in_progress`，但不再参与 Candidate、Promotion 或 release 状态 gate；其他独立 hard gate 仍必须全部通过。

## 更新规则

完成状态必须同时附带可复现的自动化结果或符合 `EVID-REDACT-001` 的真实网络证据。不得因“暂时没有复现”关闭安全或协议问题。
