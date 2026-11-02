---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
status: approved
---

# 2026-11-02 QR 无 ticket 接续离线交付

本记录对应 [ADR-0028](../architecture/decisions/ADR-0028-qr-service-continuation.md) 与 [PROTO-QR-CONTINUATION-001](../architecture/protocol-correctness.md#proto-qr-continuation-001无-ticket-接续)。网页证据及限制继续保存在[QR 研究](qr-browser-contract-2026-11-01.md)。已完成共享修复、离线验收及开发 QR 五步真实闭环。下文保留首次窗口注销阶段门禁停止及离线交付历史；后续通过不解释或覆盖旧失败。正式 QR/M3 不提升，不将已修复的兼容差异写成旧正式失败的唯一根因。

## 后续窗口：QR 开发闭环通过

2026-11-02 维护者在新的单次授权并重新确认现场条件后，回传首因定位 v2 包的完整 pass 回执及终端退出码 0。脱敏回执与核验位于 `build/qr-live-pass-20261102/`；来源为维护者私有终端粘贴，不冒充 Agent 直接执行认证。包身份为下表当前交付的 source_digest `bb39bdf8ea6b2a68bdaf2ab9d3806a9cb6aedeeb5bd19848df44a090ea1fbb8c`。

| 命令 | 结果 | 退出码 | 耗时（秒） |
|---|---|---|---|
| initial_status | offline / pass | 0 | 3 |
| qr_login | logged_in / pass | 0 | 55 |
| status_online | 独立目标身份 online / pass | 0 | 3 |
| logout | logged_out / pass | 0 | 3 |
| final_status | offline / pass | 0 | 3 |

外层与内层 state 均为 pass，coverage=complete，gate_failure=null，无 cleanup。Agent 只读核验产品、worker、启动器、构建输入四项哈希及 225 项源码快照，并确认当前对应源码一致；现有严格 v2 解码器接受该回执。核验账户为 bhkdesktop\unbal；核验过程没有新增校园请求、profile 读取或认证操作。维护者报告终端总耗时约 88 秒，不能用五项命令耗时之和替代整体耗时。

此次已证明开发 QR 登录、目标身份核对、注销及窗口结束时 offline 完整闭合；不推定此后持续 offline。普通产品哈希与此前门禁 blocked 窗口相同，旧瞬时门禁失败未复现、原因仍未知，不能称为已定位或已修复。回执没有内部协议采集，不能仅凭通过断言实际采用了哪个接续分支。

本次单次授权已消耗，不自动重试。`release_evidence=false`，旧 Candidate 与正式失败回执保持；正式 QR 仍待新 Candidate 验收，M3 保持 in_progress。双 Wi-Fi 开发通过记录也保持各自身份。下一工作为统一 Wi-Fi/QR 改动的源码与测试缺口审阅，准备新的正式 Candidate 检查点；提交、推送、合并、正式构建及真实验收分别按后续授权推进。新 Candidate 必须重新取得四组认证和三份辅助证据，不能继承开发或旧 Candidate 的通过记录。

## 当前交付：门禁首因定位 v2

2026-11-02 维护者批准报告范围后，已完成实现、两端离线验收、独立新包及零网络预检。首因记录为命令、检查位置、包校验/查询不可用/快照差异分类；快照只输出六个固定变化槽，不输出值。正常命令完成的监测取消不登记为失败，后续检查不覆盖首因。普通产品及注销实现未改；不能由本次离线通过反推旧真实窗口的失败原因。

新入口为 `build/dev-runs/windows/20261102T100148Z-3e97df20-qr-dev/ready/qr-dev-launch.exe`，操作说明在同目录 `qr-window.md`。这是独立开发包；本节保留真实窗口开始前的离线交付事实，后续真实结果见上节。旧包和历史记录保留。

| 身份 | SHA-256 |
|---|---|
| source_digest | `bb39bdf8ea6b2a68bdaf2ab9d3806a9cb6aedeeb5bd19848df44a090ea1fbb8c` |
| 普通产品（与旧 QR 开发包相同） | `cd5e98420e94e75e61390281340ec415c0b52c25a9475c2311764c21d63ff8f2` |
| worker | `0fbcae313438b03fc31a9e165d9766c5bc2e2d3b5f38d74e7f9bbba33121a255` |
| launcher | `cc2600c3c4a038aade9c5f1bda78988d4b996b5d469b1f57592d23bdadb86338` |

225 项依赖源码快照、产品/worker/启动器/构建输入绑定和七种包审计通过。启动器外层 v2、内层 qr v1，旧外层 v1 显式只读解码不补造首因；版本/字段混配、缺失/重复/未知字段、矛盾状态及泄漏 canary 均有拒绝测试。详细契约见[开发验证](../operations/development-validation.md#qr-continuation-development)。

| 检查 | 结果 |
|---|---|
| Windows QR 聚焦/相关回归、vet、自检、打包及绑定审计 | pass，`build/dev-runs/windows/20261102T100148Z-3e97df20-qr-dev/` |
| Linux QR 相关回归、race、vet、有界协议 fuzz | pass，`build/dev-runs/linux/20261102T100151Z-b9324e06-qr-dev/` |
| 共享辅助/Wi-Fi/正式 runner 消费者，两端测试/vet、Linux race | pass，`build/qr-gate-observation-20261102/consumers/` |
| 新 v1/v2 回执解码器有界 fuzz | pass，Linux 5 秒、2 worker，最终记录 `build/qr-gate-observation-20261102/final-parser/` |
| 新首因工具配合真实本地门禁和假认证命令 | pass，`build/qr-gate-observation-20261102/guard-probe.json`；全部命令为假执行器，零网络/零 profile |
| 本机新包 preflight | ready / binding verified / exit 0；2026-11-02T10:08:53Z，工具审批账户 bhkdesktop\unbal，network_requests=0、profile_reads=0、live_attempted=false |

假命令复现使用首个 v2 构建，之后仅补严回执外层/内层一致性；门禁及停止规则未改，最终解码器已重跑两端回归与有界 fuzz，实际旧 v1 和最终新 v2 回执严格读取通过。假命令复现共有 16 次本地检查：15 次相同，1 次被正常完成取消；gate_failure=null，最终本地检查一致。该结果不代表真实 QR/注销闭环成功。首次 Windows 分类测试中，两个夹具的 Write-Output 触发受限环境下模块加载错误；改为纯 .NET 内存输出后分类矩阵通过，未改变 stderr 拒绝规则。首次失败日志保留于 `build/qr-gate-observation-20261102/windows-initial-failure.log`。

本轮未执行真实扫码、status/logout、网络配置调整、提交或发布。维护者上次报告仍在线，本轮没有重新查询；后续真实窗口需重新建立 offline 条件并另行单次授权。当前交付仅使下一次失败可定位，旧历史原因仍未知，正式 QR/M3 不提升。普通产品字节未变，不重复无关安装或六平台产品构建。

## 最新窗口：登录通过，注销阶段门禁停止

本节标题及链接保留；以下是首次窗口历史，后续 v2 包闭环通过见本文顶部，不覆盖此份 blocked 回执。

2026-11-02 维护者回传当前包单次私有终端结果，启动器 exit 11。source_digest、产品及 worker 与下述包一致；Agent 只读复核产品、worker、启动器哈希及 218 项源码快照通过。脱敏回执和核验记录位于 `build/qr-live-20261102/`，来源为维护者粘贴，不能冒充 Agent 直接执行记录。

| 命令 | 本次事实 |
|---|---|
| initial_status | pass，offline，exit 0 |
| qr_login | pass，exit 0，16 秒 |
| status_online | pass，独立目标身份 online 核对，exit 0 |
| logout | blocked，result=null；是否启动子进程或发出请求未知 |
| final_status | 无回执，最终会话状态未知 |

整体为 blocked / topology_changed，内部 coverage=partial、reason=topology_or_binding_changed；不是完整 QR 闭环通过，也不是 QR 命令协议失败。已证明开发 QR 登录及其后目标身份核对通过；没有内部协议观测，不能仅凭该结果断言实际走了哪条接续分支。

维护者说明运行期间没有手动切换网络、Clash/TUN，也未注意到重连。该说明不能排除系统后台变化或本地查询失败。静态核对 `internal/livegate/path_executor.go`、`internal/qrdev/run.go` 和启动器表明：命令前、期间或结束后的门禁失败均可产生 blocked/null；当前分类合并了快照不等及查询不可用，不能由标签证明实际路由改变。指纹包含全量路由、适配器状态和接口 metric；本次没有保存前后差异，不能回溯定位具体项。不得据此放宽门禁或断言注销未执行。

本次单次授权已消耗；最后已证明的账号状态是 online，当前状态未知。Agent 本轮仅保存脱敏回执、核验包及静态分析，未追加 status/logout、扫码或重试。下一离线工作是以合成场景检查门禁分类与快照稳定性，必要时准备只输出闭集差异类别的零网络排查；改变门禁语义或新增真实窗口另行决策。正式 QR、Candidate 验收及 M3 不提升，双 Wi-Fi 开发结果保持。

## 门禁排查：未复现，历史原因仍缺证据

维护者随后明确确认未注销且当前仍在线；这是维护者报告的会话状态，不是 Agent 新增校园查询。2026-11-02 本轮仅只读本地检查、假命令复现及合成回归，没有扫码、产品 status/logout 或网络配置变更。

- 原样提取当前 PowerShell 拓扑脚本，在原生工具审批账户下连续查询 8 次：均 ready、exit 0、无 stderr；6 个投影字段完全一致，每次约 2.06–2.22 秒。投影含 54 条路由、27 个适配器、1 项 DNS；地址、接口标识和原始投影仅在内存中比较，没有保存。
- 临时 Go 复现器复用当前 `qrdev.Run`、`GuardPath`、包校验和真实本地 `labpath.Capture`，将全部认证命令替换为假执行器。模拟 QR 等待 12 秒和其他命令各 1 秒；五步合成流程 pass，最终本地检查一致。共 21 次检查，其中 16 次一致，5 次因假命令正常结束而取消；这些取消未被误判为路径丢失。报告明确 commands_are_fake，不是第二次真实 QR 验证。
- 新增两项永久合成回归：正常完成取消正在执行的查询后仍可执行下一命令；命令前检查失败时不启动命令，恢复检查也不得绕过 sticky stop 执行 cleanup。连同现有主动路径丢失回归，Windows/Linux 各重复 20 次通过，Linux race 和两端 vet 通过。没有改变生产实现或放宽门禁。
- 报告位于 `build/qr-topology-investigation-20261102/`：`local-samples.json`、`guard-probe.json`、`windows-tests.json`、`linux-tests.json`。临时复现器不读取 profile、不启动产品进程，network_requests=0。本次工具执行上下文不能替代维护者历史窗口，也不能证明当时没有瞬时变化。

目前排除了本轮可复现的持续拓扑不稳定和普通命令收尾取消误报；不能排除历史瞬时查询失败、真实后台变化或其他未复现条件。门禁仅返回 bool 且不保留首个失败类别，原回执无法进一步还原。没有依据修改注销解析器、跳过门禁或重新解释旧 blocked 为 pass。

### 已批准：一次补齐门禁失败定位

建议只扩展独立 QR 开发报告及其内部门禁诊断，不增加认证内部阶段或 HTTP 观察。新增版本保留旧 v1 读取含义，正式 runner/evidence 和产品接口不变。

- 保留首个失败：固定命令枚举、检查位置（命令前/期间/后或最终检查）、失败类别（包校验失败/本地查询不可用/快照不一致）；结束时的检查不得覆盖首个原因。
- 查询不可用仅输出闭集类别，如启动失败、取消、期限、非零退出、stderr、输出超限、结构拒绝或本地规则拒绝；不输出错误文本。普通命令完成所引起的监测取消不登记为失败。
- 快照不一致仅输出所选地址/接口、DNS、路由、适配器等变化类别，不输出值、名称、索引、地址、GUID、MAC 或可关联指纹。不增加查询来源，不持久化原始快照。
- 保持原门禁接受/拒绝条件、15 秒查询期限、64 KiB 输出限制、取消及 cleanup 停止规则；不能通过忽略无关接口或放宽失败条件来求通过。
- 先以合成失败注入覆盖全部类别、首因保留及旧回执兼容，再执行真实本地门禁配合假认证命令，完成两端测试、绑定新包与零网络预检。本提案不授权真实扫码/status/logout；后续真实复核仍单次批准。

上述提案已获维护者确认；实施结果见下方当前交付，原排查过程和旧回执保持。

## 后续更新：残留默认路由修复及当前可测试包

2026-11-02 维护者授权一次私有终端 QR 窗口并确认现场后，旧包仍在零网络门禁停止。只读核验表明当前有线 Up/IPv4 Connected，Wi-Fi 适配器及其 IPv4 接口均为 Disconnected，但两者均残留默认路由。维护者要求继续；按 [ADR-0028 补充](../architecture/decisions/ADR-0028-qr-service-continuation.md#断开接口残留默认路由修正)修正默认路由计数，未改动系统路由、接口或代理。下文旧包/blocked 记录保留为修正前历史，不作为当前执行入口。

当前包为 `build/dev-runs/windows/20261102T073808Z-f6edc0c4-qr-dev/ready`：

| 身份 | SHA-256 |
|---|---|
| source_digest | `9e34a2476c9c75a7644fa7ca276d50d3713357b5c36a533c2041191090e69b13` |
| 普通产品（与上一包相同） | `cd5e98420e94e75e61390281340ec415c0b52c25a9475c2311764c21d63ff8f2` |
| worker | `a6814c4b0c676de3af1977f6f7d6293450e02a92b449e347dd8e1ffe2e8a50d8` |
| 启动器 | `b9b34cfe6282e37b69ff1c76f783399da77af242244a7b0e9388838ff7fb542f` |

Windows 原生测试以假 cmdlet 执行实际拓扑脚本，20 项矩阵覆盖断开残留、有线/Wi-Fi 交换、活动物理/虚拟竞争、状态矛盾/未知/缺失/重复、选定接口断开、无有效默认路由、半默认覆盖、读取失败和上限；另验证残留路由、接口 metric、重连变化进入指纹。所有查询均为测试假数据，没有校园请求。Windows QR 测试/vet/打包/self-test/绑定检查通过，报告位于新包父目录；Linux QR 测试/race/vet/fuzz 报告为 `build/dev-runs/linux/20261102T074150Z-7fa60fb7-qr-dev`。其他共享组件消费者补充测试/vet（Linux 另 race）分别见 `build/qr-implementation-20261102/windows-20261102T074256Z-825b5714` 及本轮 Linux 消费者报告。产品字节未变，不重复无关安装或六平台产品构建。

新包在 2026-11-02T07:43:58Z 经 `bhkdesktop\unbal` 工具审批上下文零网络预检：binding=verified、state=ready、exit 0，network_requests=0、profile_reads=0、live_attempted=false；回执为新包父目录 `local-preflight.json`。ready 仅证明当时本地门禁通过，不证明账号 offline，也不保证随后路径不变。

下一步由维护者在私有终端按该包 `qr-window.md` 或 `build/qr-implementation-20261102/QR-PRIVATE-WINDOW-r2.md` 执行尚未开始的一次窗口，在终端重新确认现场条件及单次 token。此次交付未执行 QR/校园状态查询；新包身份已明确，旧包保留，不混用工具。真实结果、正式 Candidate 及 M3 均未提升。

## 实现与拒绝边界

- 精确接受本轮 service 的唯一 ac_id、无 ticket 形式；空/重复 ticket、额外字段、fragment（含空 fragment）、路径变形、目标错配均不进入新增分支。
- 只查询一次既有 HTTPS 网关 API，不跟随其 HTTP 重定向；独立严格 JSON 解码器仅接受四个已确认字段。验证 HTTPS CAS 和无查询 service 基址后，用本地端点和本轮 ac_id 重建 CAS 地址。
- 延续本轮 Jar 的域隔离；复用 ticket 截获、HTTPS activation、业务结果和最终目标身份校验。新增分支 60 秒覆盖至最终 status，更短父期限优先，既有轮询 180 秒不变；无备用路线或循环。
- 新 qr-dev 启动器使用普通产品，复用 profile 保护、有线绑定/拓扑监测、私有 TTY 与 terminal_qr 状态机。命令轨迹在 QR/status 合并前记录；闭集回执校验轨迹和摘要一致性，拒绝截断、未知字段及矛盾结果。中断保留已校验命令事实，绑定失效不得发布可信结果。
- 公共 SDK/CLI/JSON、退出码和正式 runner/schema 未变。Windows 专用开发工具未进入产品依赖；不继续扩展 CTL 或浏览器观察。

## 离线验证记录

工具为本地 Go 1.25.0，GOTOOLCHAIN=local、GOPROXY=off；用户配置/凭据目录隔离，测试仅用假 transport、生成数据或既有 loopback 夹具。上述环境变量不是操作系统防火墙声明。执行账户为正式工具审批上下文，未修改 owner/ACL/safe.directory。

| 检查 | 结果与报告目录（仓库内 build/） |
|---|---|
| Windows 最终 QR 聚焦测试、vet、自检、构建及绑定审计 | pass；`dev-runs/windows/20261102T054339Z-b89bfd8f-qr-dev` |
| Linux 最终 QR 聚焦测试、vet、race、5 秒双 worker fuzz | pass；`dev-runs/linux/20261102T054356Z-b3838163-qr-dev` |
| Windows 全源码普通测试、vet | pass（另列 skip）；`qr-implementation-20261102/windows-20261102T052817Z-9187bfb7` |
| Linux 全源码普通测试、vet、race | pass（另列 skip）；`qr-implementation-20261102/linux-20261102T053538Z-2734b2ad` |
| 六平台三产品入口构建及依赖隔离 | pass；`dev-runs/linux/20261102T053201Z-3564395f-cross`；macOS 仅交叉构建 |
| 最终快照/产品/工具/启动器绑定 | pass；最终目录 `binding-audit.json`：有效包、非 TTY、工具/产品/输入篡改、缺失、额外成员七项 |
| 文档、链接与派生索引 | doccheck pass；`qr-implementation-20261102/final-document-check.json` 汇总最终复核 |
| 定向脱敏扫描 | pass；`qr-implementation-20261102/source-scan.json` 固定 90 项改动源码及哈希，排除 ignored 私有材料 |
| 本机零网络 preflight | blocked / topology_unavailable，binding=verified，exit 11；最终目录 `local-preflight.json` |

全源码验证明确排除未改动的 `internal/installtest`，不无条件重跑安装套件；所有其余源码包均纳入。Windows 五项 config symlink 测试因当前执行令牌没有创建符号链接权限而 skip；Linux `TestAssembleRejectsWrongPackagerHost` 因本机正是受支持构建主机而 skip。无测试文件包与真实 skip 在 `coverage-summary.json` 分列，不把它们计为测试通过；不重解释既有 P4 原生安装/管理员覆盖。

保留的首次失败：Linux 全源码夹具误放在 Windows 挂载盘，Unix 配置目录检查读到 777 权限并拒绝；报告为 `qr-implementation-20261102/linux-20261102T052835Z-49df51da`。修正检查脚本为 WSL 本地临时目录后通过，没有放宽产品权限规则。全源码检查后新增的精确 service 拒绝用例及提前非 TTY 拒绝，由上述最终两端 QR 聚焦回归覆盖；未对无关套件作重复运行。

协议合成覆盖同轮 Cookie 延续及域隔离、完整成功链、非法 service/Redirect/JSON、网关拒绝自动跳转、响应超限、各接续阶段取消、短父期限、activation/身份/status 失败及禁止后续请求。既有直接 ticket/CAS 确认、pending/过期/挑战回归保留。泄漏 canary 不进入事件，开发回执只含固定分类；未保存用户脚本、公钥、Cookie、ticket、UUID 或原始响应。

## 最终开发包身份

包目录：`build/dev-runs/windows/20261102T054339Z-b89bfd8f-qr-dev/ready`。

| 身份 | SHA-256 |
|---|---|
| source_digest | `acf9848927dc4b47cb3a1ef232ba1256fd37566b3832357fa1d7d528e4e112b7` |
| 普通产品 ipgw-meta.exe | `cd5e98420e94e75e61390281340ec415c0b52c25a9475c2311764c21d63ff8f2` |
| worker qr-dev.exe | `79b7c8b555b5476239b7f97026ba0df9bcf0e8b4958ee611fb860f176a7f6419` |
| 启动器 qr-dev-launch.exe | `be1355d98b19b79c88ca8fba7ab543419f4c98a9ed8f539a00a7edeb0c67e8d9` |

`source/` 为构建依赖源码快照，`payload/build-inputs.json` 与 `build-receipt.json` 固定来源和产物。普通产品版本为 dev-qr-local，release_evidence=false；背景 HEAD 为 `c21094ab3e7b2a4064b66ae82cf2fbc2173f92cd`，不能将 dirty 快照称为该提交或正式 Candidate。临时早期包保留，不作为最终测试入口。

本机 preflight 于 2026-11-02T05:53:35Z 在 `bhkdesktop\unbal` 的工具审批上下文执行；本地路径未达到当前门禁，不能仅凭该分类判定是 TUN、接口还是路由原因。没有自动修改配置或反复尝试，network_requests=0、profile_reads=0、live_attempted=false。此结果不能证明账号 offline，也不能代替后续维护者原生终端可见性检查。

## 下一次私有终端窗口

操作说明位于同一最终包的 `qr-window.md`，完整中文步骤另存 `build/qr-implementation-20261102/QR-PRIVATE-WINDOW.md`。当前只交付说明，尚未取得真实运行授权。

依赖顺序为：维护者重新建立并确认有线独占/TUN 关闭/目标 offline → 同包零网络 preflight ready → 新的单次 QR 开发窗口授权 → 私有终端确认和扫码一次 → 独立核对 QR、目标 status、logout、final status 回执。profile 仅真实模式读取，二维码仅终端呈现，不回传截图/原始输出。失败不重试，拓扑变化/取消后不继续请求；cleanup 仅在既有所有权和有效上下文下执行。

若仍失败，仅根据命令级记录集中说明缺口，不声称定位 QR 内部阶段，也不自动新增观察或复测。真实开发通过仍需后续统一新 Candidate 和全部必需正式证据。旧正式 Candidate、旧失败回执、双 Wi-Fi 开发结果和历史文档均保留；本轮无校园请求、提交、推送、Candidate 构建或发布。
