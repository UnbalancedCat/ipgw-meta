---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# P5 候选源码准备记录（2026-11-01）

本页为执行记录；路线以 [roadmap](../upgrade/plan.md#roadmap-20261029) 为准，门禁以 [BHK 正式验收](../operations/bhk-acceptance.md) 为准。本阶段终点是同一新 Candidate 的验收及 promotion 材料准备，不创建最终 tag、不公开 release。准备期间未执行校园请求、真实 profile/凭据读取或认证。

## PR #29 与首轮 CI 修正

本节及之后的准备段落按执行时点保留；最新 Candidate 与本机验收结论见本文末节。

维护者批准的 344 文件清单已在新分支 `codex/bhk-v1-acceptance` 提交为 `2fc0d1644920a2aefb10a8859e4baebe84cf0501`，并推送、创建 [PR #29](https://github.com/UnbalancedCat/ipgw-meta/pull/29)。原分支及旧包保留；未合并、未启动 Candidate。下方“尚未提交”等段落是此次授权之前的准备快照。

首轮 CI（CI run `41266117777`、Native install run `41266117667`）不能判通过：Linux 普通测试/race、六平台交叉构建、资产打包、Linux 两架构及 macOS 两架构原生安装成功；文档、macOS 普通测试、Windows 普通测试及两架构原生安装失败。失败定位和修正均限于测试/文档：

- 文档检查：旧交接中的两个 ignored build 回执链接在干净检出不存在。保留路径和历史事实，改为本地产物代码文本，不把私有回执纳入仓库或放宽链接检查。
- macOS：新 livegate 测试引用只在 Linux/Windows 定义的私有文件系统夹具。把三个文件系统用例完整移入相同平台约束文件；纯 schema、进程投影和拓扑测试继续跨平台。v1 发布拒绝在文件系统操作之前验证。
- macOS：auxgate 接受路径夹具先解析临时目录的系统别名；生产路径检查不变。
- Windows：管理员进程新建目录可能默认由 Administrators 所有。仅在当前进程为管理员、初始所有者确为该组、且刚创建的测试目录为空时，将夹具所有者设为当前用户；其他情况拒绝。之后仍验证私有 ACL 和所有者，产品安装器不变。

修正后本地验证：

| 检查 | 结果与本地回执 |
|---|---|
| Windows P4 测试/vet | pass；`build/dev-runs/windows/20261101T025741Z-9c8cf4b5-p4/result.json` |
| Linux P4 测试/vet/race/有界 fuzz | pass；`build/dev-runs/linux/20261101T025800Z-ac3e3e03-p4/result.json` |
| Windows 普通进程安装回归 | pass；`build/dev-runs/windows/20261101T030037Z-21d3eec9-install/result.json`；未选择 Candidate 的 smoke 不计正式验收 |
| Windows 管理员安装回归 | 四组全部实际 pass、无 skip；`build/p5-preparation-20261101/a092bc24d9a64aba9c3b44fb2ba2132b.install.result.json`；新目录初始 owner 为 Administrators，测试程序 SHA-256 `12a83198e02ffa81136d76601ec10145fdc03395d205fbd31ec3a7ca1b91cdc1`，运行前后相同 |
| macOS 两架构受影响测试、Windows arm64 安装测试编译 | pass；`build/p5-preparation-20261101/ci-fixes-5621da3910f14ef6b407426ce4f66b45/result.json`；仅编译，不冒充原生执行 |
| 无 ignored build 的源码快照 doccheck | pass；同上 clean-source-doccheck 步骤 |

上述修正已获维护者授权，精确 8 文件提交为 `0d11f1ff4e04a98d1e47a38824ae7804d151c920` 并推送 PR #29。CI run `41270982316`、Native install run `41270982267` 共 13 项中 12 项成功，包含六平台原生安装；Windows 全仓测试仍失败。未合并或创建 Candidate，不能用本地通过覆盖远端失败。

第二轮 Windows 失败涉及诊断测试夹具：严格路径身份校验遇到 CI 临时路径的非规范表示；合成拓扑脚本运行达到现有 15 秒期限。追加修改只在测试中规范化临时路径，并为假 cmdlet 脚本禁用自动模块发现、显式加载所需的 inbox Utility 模块。后者限制合成测试的依赖，不声称证明真实拓扑查询在所有 CI 镜像可用；是否消除远端超时仍待新 CI。生产代码、真实来源、路径/重解析点/硬链接拒绝及超时保持。

追加验证：Windows 全仓普通/诊断/安装测试通过，回执 `build/dev-runs/windows/20261101T041344Z-00f51f7c-test/result.json`；vet 通过，回执 `build/dev-runs/windows/20261101T041715Z-d5006e6d-vet/result.json`。实际短路径 TEMP 的四项包验证独立通过、无 skip，回执 `build/p5-preparation-20261101/short-temp-929a5ced43e941978a48848d114e312c/result.json`。追加测试修改尚待具体清单提交/推送授权。远端结果快照在 `build/p5-preparation-20261101/ci-followup-checks.json`。

PR 的 `behind` 状态来自尚未纳入 main 的 PR #28 合并提交 `dac245947b9bd11b5489e2fd87ca2823c2d2dac5`；该提交与共同祖先 `a14cc291fed335a6b027f42f72a9775f603aa3cd` 的文件树相同，但祖先关系不同。strict checks 要求更新分支后重新验证，不能凭内容相同绕过；本轮未执行该合并。后续合并、Candidate workflow 和真实窗口仍需各自授权。

## P4 未覆盖项补齐

固定 Go1.25.0 离线编译 internal/config 测试程序，SHA-256 为 `f77b81cc5871a4e5f50b4197f1e39ef9e44eb871d4ada72c422f5ec3989a9bc4`。普通 Unbal 进程的五项测试仍 skip；随后通过 Windows UAC 在同一账户原生管理员进程运行相同字节，五项均实际 run/pass，退出码 0，无超时，程序前后哈希一致：

- TestReadRestrictedPasswordRejectsFinalReparsePoint
- TestValidateRestrictedWindowsParentsRejectsDirectorySymlink
- TestStoreLoadWindowsRejectsConfigReparsePoint
- TestStoreLoadWindowsRejectsDirectoryReparsePoint
- TestStoreSaveWindowsRejectsLockReparsePoint

只创建独立合成测试目录；没有启用 Developer Mode、修改永久账户权限、真实安装或仓库 ACL。回执为 `build/p5-preparation-20261101/00991b8b8061474091e20c427f4a0b22.result.json`。首次普通进程回执的 exit_code 为 null，是 PowerShell Start-Process 退出码获取问题；已改用显式 Process 对象，保留原回执，不据其零值推断通过。五项 skip 本身真实存在，管理员回执是独立重测。

发布手册 REL-LIVE-001 正文残留的 NAS 必需矩阵已改为三组 BHK 认证加两项辅助检查，历史日期记录保留。修正不改变已经批准的 P4 契约。

准备阶段质量检查通过：`build/dev-runs/linux/20261101T021602Z-c3083ec5-quality/result.json` 覆盖 promotion、milestone、Candidate、actionlint、gitleaks 规则及历史/工作树扫描；文档检查 `build/dev-runs/windows/20261101T021551Z-0ac8e63e-docs/result.json` 通过。没有修改产品代码，前次全仓/race/cross 报告继续作为对应源码的离线依据；后续新 PR 必须重新取得 CI 结果。

## 源码和远端基线

本地分支 codex/cas-submit-field-closure，HEAD b4eeb760945f0568ca163d957ddd55df2afa86c4；P3/P4 尚未提交。原 handoff 保持历史字节。拟提交包含共享取键/注销修复、对应回归、独立开发工具与历史诊断源码、P4 正式门禁及规范；不删除历史诊断，也不把它们装入产品。已有六平台依赖检查拒绝诊断和 JS runtime 进入产品闭包；静态解析器依赖保持锁定。

逐文件清单与 SHA-256 在 `build/p5-preparation-20261101/review-manifest.json`，可读路径在同目录 review-paths.txt。排除 build/release、私有材料、Python 字节码缓存；不执行 git add 全目录。清单仅用于审阅，不表示已暂存或已提交。

GitHub 只读核验：main 仍为 dac245947b9bd11b5489e2fd87ca2823c2d2dac5，签名 verified/valid；旧 PR #28 已合并，当前无开放 PR；最终 v1.0.0 ref 返回 404。main 的 22 项 checks 当时均成功，包含 Candidate 要求的 13 项，但仅属于旧 SHA。保护由 active main-v1-protection ruleset 实施，要求 PR、签名、strict checks，不能把旧 branch protection 字段解释为没有保护。新代码必须走新 PR；合并后对新 SHA 重新核验全部条件。

## 提交和后续检查点

准备一个新的 codex/bhk-v1-acceptance 集成分支和 PR，目标 main；不重开 PR #28、不强推。提交标题拟为 `fix: converge v1 authentication and BHK acceptance gates`。提交/推送/PR 的具体清单批准后才执行，合并及 Candidate workflow 另有检查点。

新 Candidate 到位后先完成原生安装辅助记录，再分别安排有线 password_core、terminal_qr。Wi-Fi 在切网前备齐包、依赖、命令和本地停止/恢复说明；维护者断开有线后独立运行辅助检查，通过后才运行已明确授权的一次 password_core。恢复有线只恢复管理连接，不证明账号已 offline；回执与身份逐项核验后才形成公开摘要。

正式认证仍 blocked/live_unverified，M3 in_progress。当前没有新 Candidate、五份正式证据或可用于发布的 promotion lock。任何旧开发包、旧 Candidate 和本页合成测试均不能替代它们。

## 正式 Candidate 与本机验收

2026-11-01 后续执行覆盖上述历史检查点：追加 15 文件提交为 `6ab3b7e248d172d90905c35ae8ba790c4a56a67d`，13 项 CI 全部通过。按明确授权纳入固定旧 main，分支 head 为 `c21094ab3e7b2a4064b66ae82cf2fbc2173f92cd`，tree 保持 `3eaa0bd52b82f27f2636d69f0ccec5d49d4a5424`；再次通过 13 项 CI 后经 GitHub squash 合并 PR #29。受保护 main 为 `1a32aabc2aa3125afe638b0dc3d3444bf7a95c7d`，签名 verified/valid、文件树一致，合并后的 CI `41274672274` 与原生安装 `41274672284` 合计 13 项成功。

维护者单独批准后，仅 dispatch 一次正式 Candidate workflow；[run 41275535736](https://github.com/UnbalancedCat/ipgw-meta/actions/runs/41275535736)、attempt 1 全部成功。Candidate `v1.0.0-1a32aabc2aa3-41275535736.1` 包含已合入的 P3/P4 修复。六平台原生安装按既定 full/smoke 矩阵通过：Linux amd64、Darwin arm64、Windows amd64 为 full，其余三组 smoke。非选中分支的条件 skip 不解释为已运行，也不影响所选门禁的通过记录。

| 身份 | 核验值 |
|---|---|
| artifact ID / 大小 | `13070928527` / 57845208 bytes |
| 原始 ZIP SHA-256 | `5444f60a0a076b2aba25bb50ce4ac3c2878c290b977f6e1ffe037f8fafca2663` |
| candidate manifest / set SHA-256 | `394a417bd3263d6ff4ed9573eb915a174292d379c214868bb94c205d8b96141b` |
| build-input SHA-256 | `480acf5a00beab8faaab8d31133f309d64b0d65942f6dfba85c1a7b628dec77a` |
| Windows helper SHA-256 | `c47bed1d8fb7a896e2e780e554eb17be84ce6bc69dce67e38dab6c847d43c0d8` |

本机核验原始 ZIP 与服务端摘要一致，安全解包后 14 文件、57843072 bytes 的闭集、manifest、校验清单及产物投影全部通过。使用校验官方发行摘要与 checksum 后的便携 gh 2.103.3，独立验证 ZIP 和十份公开资产的 attestation；固定仓库、workflow、main ref、source digest、GitHub hosted runner，以及仓库校验器要求的精确 run/attempt/subject 集合全部通过。只使用既有 Git 身份的内存凭据，没有记录 token 或修改系统 PATH。报告在 `build/p5-preparation-20261101/CANDIDATE-VERIFIED.md` 和同目录 `attestation-verification/`。

本地保留上述集成分支，未强制切到 squash main；fetch 后 origin/main 指向 Candidate source。14 文件原件保存在 `build/rc-build-v1.0.0-1a32aabc2aa3-41275535736.1`，另为原生验收建立同字节的独立私有副本，只收紧新副本 ACL。没有修改仓库/旧包 ACL、旧 Candidate 或真实安装。

### BHK 安装辅助执行记录

工具执行账户为 BHKDesktop 的 Unbal；记录真实 Windows 原生工具上下文，不冒充维护者私有终端或 CI。首次调用误用了不符合既有契约的输出目录，exit 2、没有 bundle；纠正为 `build/live-evidence` 后生成当日第 001 份辅助 bundle，结果 blocked。完整 ID 保存在私有回执，不从首次退出推断测试通过。

普通进程独立复现已编译测试程序的启动遭 Windows error 740，需要提升。经标准 UAC 运行同一 Candidate helper 后生成当日第 002 份辅助 bundle，仍 blocked。进一步在提升上下文复现 fresh 子测试：直接运行已安装目标版本成功，通过长测试路径的 active junction 启动失败。前两份辅助记录保留，不能用 CI 结果或目标 EXE 的成功替代 junction 验收；未改变安装安全规则或正式 helper 字节。

后续仅缩短独立测试 work-root，保持同一 Candidate、管理员上下文及完整验收；05:52:48–05:54:36 UTC 的第 003 份辅助记录 pass。十项检查全部 pass：Candidate、fresh install、三入口、active junction、默认 meta、升级保持、基本回滚、私有布局、真实用户 PATH 不变及 Candidate 前后不变。没有 skip，也没有用直接目标启动替代 junction 启动。此结果支持长测试路径是本机拒绝条件之一；不宣称所有 Windows 长路径安装均已支持。

已通过同一 helper 的 summarize 只读验证私有 bundle/checksums，生成 ignored 人工审阅副本 `build/p5-preparation-20261101/installation-summary-review.json`。bundle SHA-256 为 `19569e29a2d5e5ae38866151dd83ebfa4034b5e3be520ba94d4cbdd62c6c424a`；源码快照 SHA-256 为 `cfa42cde3d630f23c6a8666f05a140234ff17272016ec479f8a5f6081a50985b`，本地验证器 SHA-256 为 `dd773ec2c3d6647befa93df13a74f0ac8b521902a2ffc2d5c90b1bafd059e45b`，Go 精确 1.25.0 windows/amd64。验证器来自 Candidate source，属于本地辅助程序，不冒充 CI attested 资产。两份 blocked bundle 和首次无 bundle 的错误调用保留；公开摘要尚未经维护者审阅选入发布材料。

下一步为单独授权的同一 Candidate 有线 password_core，随后有线 terminal QR，再按预先准备的断网交接执行 Wi-Fi 辅助和 password_core。三份认证和 Wi-Fi 辅助证据尚未取得，本轮没有校园请求、profile/凭据读取、tag 或 release。正式认证与 M3 状态保持；本轮仅更新本地执行文档，没有追加提交或推送。

### 有线正式 password_core 回执

上述段落是构建与安装窗口的结束记录。后续维护者明确授权一次有线 password_core 并重新确认现场条件。首次原生终端入口因工具上下文 C 盘 LOCALAPPDATA 可见性差异，在产品哈希检查处停止，没有启动 runner 或认证；改用共享 D 盘新私有运行子集，复核相同产品、manifest 和 helper 字节后，维护者在自己的私有 PowerShell 终端完成本次正式窗口。未沿用开发 observed CLI，不保存终端账号提示或密码。

回执 `EVID-20261101-001`，schema v2，Candidate 为 `v1.0.0-1a32aabc2aa3-41275535736.1`；06:20:37–06:21:40 UTC（香港时间 14:20:37–14:21:40），result=pass、runner exit 0。平台/测试面为 windows-amd64 / bhk_windows / campus_wired，认证方式 password，套件 password_core。

七步均 pass、exit 0、error_code=null：初始 offline、login LoggedIn、独立 status online、重复 login AlreadyOnline、logout LoggedOut、独立最终 status offline、重复 logout AlreadyOffline。持续时间分别为 3/42/3/3/3/3/3 秒；没有 cleanup。此回执支持目标身份及该窗口的状态闭环，不表示所有未来校园请求或其他平台认证已通过。

Agent 使用同一冻结 helper 的 summarize 只读校验 bundle/schema/checksums，并逐项核验完整七步、零 cleanup、Candidate/source/manifest/helper 和实际普通产品哈希。bundle SHA-256 为 `1d64b8f07e8aeb717b94a357ca465f085fcfc5b529173281673211b4d0fc140c`；产品 SHA-256 为 `f1bab64a7074c993f75b447463ba22b9e1af806c90a9ed6a6a3b9a3fb071b85a`。其他 Candidate 身份与上表一致。脱敏审阅副本为 `build/p5-preparation-20261101/wired-password-summary-review.json`，逐项核验回执为同目录 `wired-password-verification.json`；未提交私有 bundle 或原始认证材料，公开摘要尚待维护者审阅选入发布集合。

该 tuple 的 capability_after 为 synthetic_covered + live_verified。有线 QR、Wi-Fi 辅助及 Wi-Fi password_core 尚未执行，Linux/macOS 无本 Candidate 校园证据；M3 保持 in_progress。单次授权已消耗，Agent 未重跑登录/注销或切换拓扑。本轮仅更新执行文档，未提交、推送、创建 tag/release。

### 有线正式 terminal_qr 失败回执

2026-11-01 后续解释修正（原回执及下文当时记录保留）：executeQRPair 会将 QR 成功后的 status 失败合并到 qr_login_logged_in，并累加耗时。所以下文“没有后续状态”不能理解为没有执行 status 请求，只能确认没有独立 status 步骤或通过证明。失败可能来自 QR 本身或后续状态检查；无 cleanup/清理权及最终状态未验证结论不变。具体依据及下一步见[QR 契约核对](qr-browser-contract-2026-11-01.md)。

维护者另行授权一次有线 QR 并重新确认现场条件，使用同一冻结 Candidate 的普通 CLI、配套 helper 和共享 D 盘私有终端入口。窗口为 07:14:41–07:17:30 UTC（香港时间 15:14:41–15:17:30）。回执 `EVID-20261101-002` 为 v2 / windows-amd64 / bhk_windows / campus_wired / terminal_qr，result=fail、runner exit 10。

仅有两个步骤：initial_status_offline pass / exit 0 / 3 秒；qr_login_logged_in fail / exit 6 / protocol_changed / 166 秒。没有后续状态、注销或 cleanup。runner 尚未通过 status_online 建立清理权，因此不能自行注销；既不能把用户扫码确认解释成网关目标身份已上线，也不能假定窗口后仍 offline。维护者只反馈“二维码出现，已扫码并确认”，未收集 QR payload、截图、链接或原始响应。

Agent 使用同一 helper 的 summarize 验证 bundle/schema/checksums，再核对两步前缀、失败码、能力不升级及运行文件哈希。bundle SHA-256 为 `36ab30c31a7edb6088e043bf7085b1e0eb848e4d7b870c5dd9500537ff545f04`，Candidate/source/manifest/普通产品/helper 身份与上一窗口一致。脱敏副本位于 `build/p5-preparation-20261101/wired-qr-summary-review.json` 与 `wired-qr-verification.json`；不作为必需 pass evidence，历史失败保留。

离线源码对照：SDK 在显示二维码后仍可能因轮询封装/状态不识别、确认跳转安全校验或页面解析、activation 响应等返回同一 protocol_changed；正式 human-mode runner 只记录错误分类，无法区分这些分支。166 秒不能单独证明轮询超时；正常二维码过期路径使用交互错误，不能据接近历史 180 秒期限改判为 expired。已提供的本地 login_neu.js 只展示调用 loginQRCode 并消费回调 redirect_url，不能证明实际轮询响应形式；本地未找到对应独立 QR 脚本，未主动获取站点资源。

下一工作应先确定失败阶段与网页 QR 的确认契约，采用脱敏静态材料或另行确定的开发观测范围；不重跑同包、不保存原始响应、不放宽成功/跳转规则。需要新的状态查询或真实诊断时单独明确范围；当前单次授权已经消耗。若后续产品修复改变 Candidate source，须走新 Candidate 和相应证据流程，不能把本轮 password/安装证据移作新包证据。此前有线密码及安装通过结论保留；Wi-Fi 两项仍未执行，正式发布集合及 M3 未完成。
