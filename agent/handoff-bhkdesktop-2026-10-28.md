---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# 本轮接手记录（迁移上下文）

当前执行入口见 [2026-10-29 认证收敛 roadmap](../docs/upgrade/plan.md#roadmap-20261029)；下文迁移正文和逐次交接按发生时间保留，末尾记录本次文档工作包。历史“下一步”与授权不自动成为当前任务。

## 2026-10-31 当前执行补记：P4 离线交付

2026-11-01 第二轮 CI 交接：已按授权提交并推送精确 8 文件修正，HEAD `0d11f1ff4e04a98d1e47a38824ae7804d151c920`；远端 12/13 成功，全部原生安装通过，Windows 全仓夹具修正仍在本地。相关回归、差异和剩余分支祖先检查点见 [P5 执行记录](../docs/research/p5-readiness-2026-11-01.md#pr-29-与首轮-ci-修正)。下一步是追加清单的提交/推送及新 CI，不自动合并或构建 Candidate；未执行校园请求。以下交接保留发生时状态。

2026-11-01 后续交接：维护者已授权并完成精确 344 文件提交、推送及 PR #29 创建，当前分支 `codex/bhk-v1-acceptance`、HEAD `2fc0d1644920a2aefb10a8859e4baebe84cf0501`。首轮 CI 失败定位、修正和两端回归见 [P5 执行记录](../docs/research/p5-readiness-2026-11-01.md#pr-29-与首轮-ci-修正)。下一步为具体补充清单的提交/推送检查点，随后复核新 CI；未授权合并或 Candidate，未执行校园请求。旧分支、旧包及以下历史快照保留。

2026-11-01 开始 `WP-AUTH-CANDIDATE-ACCEPTANCE` 准备：维护者批准先补齐五项 Windows symlink 覆盖、审阅源码和修正文档残留，再准备具体提交/PR/新 Candidate。各 Git mutation、正式 CI 与真实窗口保留明确检查点；终点为验收和 promotion 材料，不签最终 tag、不公开 release。Wi-Fi 使用切网前备齐本地材料的断网交接。当前网络条件不视为未来持续授权；以下 P4 历史交付保留。

同日准备结果：五项 symlink 原生固定测试均 pass，当前发布矩阵已收敛；远端 main 仍是旧 Candidate source、PR #28 已合并。拟提交清单和新 PR 检查点见 [P5 准备记录](../docs/research/p5-readiness-2026-11-01.md)。未提交/推送/合并，未构建 Candidate，未运行校园请求；下一步先审阅具体文件清单。

`WP-AUTH-BHK-GATES` 的 P4-A–P4-F 已完成；规范以 [BHK 正式验收](../docs/operations/bhk-acceptance.md) 为准，验收报告及未覆盖项见 [P4 记录](../docs/research/p4-bhk-gates-2026-10-31.md)，正式状态见 [status](../docs/upgrade/status.md#status-20261031-p4)。Windows/Linux 回归与平台隔离通过；原生安装/junction 缺口已在测试准备和私有证据目录权限处理范围修复，五项 symlink 特权用例仍因环境 skip。旧包、Candidate、历史 handoff 和本地既有改动保留，未提交或推送。

下一工作包 `WP-AUTH-CANDIDATE-ACCEPTANCE` 依赖 P3、P4 及现行发布 hard gate；新 Candidate、正式安装/Wi-Fi 辅助采集和三组认证仍未执行。没有读取真实 profile/凭据或运行校园请求。复核入口为 `python scripts/dev.py p4 --context tool`、各平台 test/vet（Linux race）、quality/cross/docs；详细报告路径与 P5 授权边界统一引用上述交付记录，不在 agent 中复制规范。

- 本文保存父任务提供的完整交接正文；下方历史授权不是当前授权，历史验证不是本次新测试证据。
- 开始：2026-10-28，BHKDesktop，`D:\project\Go\IPGW-Meta`。本轮仅接手、保存交接、Git 安全对齐和只读基线验证，不开发功能、不运行产品/校园网诊断/构建测试、不安装、不读凭据、不修改网络或安全权限、不提交或推送。
- 按本轮明确要求保留 `agent/handoff.md` 与 `docs/upgrade/status.md` 的历史快照原字节；本文件承载此次迁移接手的开始与结束记录。
- 本轮结果另附在正文末尾；以下为收到的交接正文。

---

# IPGW-Meta：BHKDesktop 新接续任务交接

整理日期：2026-10-28。下文历史诊断日期按原任务时间戳的 UTC 日期记载。

这是新任务的工作交接，不是完整复制旧聊天。原任务「继续 IPGW-Meta RC-BUILD」ID：`01a0611f-3fec-73c3-ac1f-3c9ffea543c8`，保留在 BHKLAPTOP。交接已通过受支持的 `read_thread` 接口分页读取 67 个相关 turn，覆盖 2026-09-03 当前协议修复/候选构建至 2026-10-13 最新诊断，以及 2026-10-28 文档提交。没有读取应用内部数据库或 sessions。较早历史用仓库规范和提交链交叉核对。

## 1. 当前任务目标与授权边界

用户已经批准采用“交接说明 + BHKDesktop 新接续任务”，不再依赖原任务跨主机 Hand off。后续主要开发和校园网验证都落在 BHKDesktop；其位于所需校园网环境。不得继续沿用“BHKLAPTOP 改代码，再 SSH 传到台式机测试”的旧默认流程。BHKLAPTOP 保留原件和历史参考，BHKZ4Pro NAS 是维护者自己的恢复/状态复核面，不是默认测试或控制对象。

新任务第一阶段应核对项目、最新提交、现有文件和工具环境，安全对齐 Git 状态，然后报告下一步计划。不得因读取旧聊天中的授权或助手建议，就自动启动旧诊断、登录、注销、校园网请求、安装、网络调整、发布或工作流 dispatch。

历史一次性测试授权已经对应到历史窗口，不能继承为当前授权。当前的迁移与接续授权不等于新的真实网络测试授权。尤其没有证据表明 key-provenance R2 的真实校园网诊断已获授权或已执行。

人工操作时使用完整主机名 BHKLAPTOP、BHKDesktop、BHKZ4Pro，避免含糊的“BHK”。维护者在 2026-10-13 明确要求：每轮只给一个完整、自包含的操作/命令，等返回输出后再继续；不要一次堆叠多个需人工执行的步骤。

## 2. 项目、目录和 Git 基线

台式机应用已识别的项目：`IPGW-Meta`，project ID `df9d39a0-f8ff-4162-b72e-70565c26e535`。

主要工作目录：`D:\project\Go\IPGW-Meta`（BHKDesktop）。这是迁移后的规范源码副本，不能因旧聊天曾把同名路径称为“旧仓库”而误判当前目录；原台式机旧内容已经另行归档。

目录保留情况：

- BHKDesktop 新项目：`D:\project\Go\IPGW-Meta`
- BHKDesktop 原项目完整归档：`D:\project\Go\IPGW-Meta.archive-20261028-165538`
- BHKDesktop 迁移暂存与回执：`D:\project\Go\IPGW-Meta.migration-20261028T083505Z`
- BHKLAPTOP 原项目：`D:\project\Go\ipgw-meta-clean`，未删除。

最近核实的 Git 状态：

| 字段 | BHKLAPTOP 源端 | BHKDesktop canonical checkout |
|---|---|---|
| 分支 | `codex/cas-submit-field-closure` | 相同 |
| HEAD | `b4eeb760945f0568ca163d957ddd55df2afa86c4` | `a14cc291fed335a6b027f42f72a9775f603aa3cd` |
| 工作区 | 干净 | 仅以下两份文档未暂存 |
| origin fetch/push | `https://github.com/UnbalancedCat/ipgw-meta.git` | 完全相同 |
| upstream | `origin/codex/cas-submit-field-closure` | 相同 |
| fetch refspec | `+refs/heads/*:refs/remotes/origin/*` | 相同 |
| 类型与子目录 | 普通 checkout，项目在仓库根 | 相同 |

2026-10-28 用户在原任务明确要求提交并推送未提交内容；原任务完成签名提交 `b4eeb760...`，记录远端 SHA 核对一致、文档检查和秘密扫描通过。它仅原样提交两份文档，没有更新其历史内容。迁移执行者之后独立比较了全部 219 个跟踪文件：与迁移前清单没有任何字节差异。

两份文件的当前字节基线：

```text
agent/handoff.md
SHA256 F89D3DC64965B2D9B7014893A263DC2085FE1123CB89DC6ADF766DFB1CCF9E2D

docs/upgrade/status.md
SHA256 D9D0EF4A44A7C3216B4D5A0FE10759A952FC13CC5EFCB26B23001674789BF0C1
```

新任务需要安全对齐台式机提交状态到 `b4eeb760...`：先重验所有文件、两份文档哈希、工作区/index、分支祖先关系和目标提交；保留可恢复的当前文档/diff，再选择不丢工作区内容的对齐方法。不能直接 `reset --hard`、强制 checkout、清理文件，不能为了变干净而丢掉两份文档，也无需重复提交相同内容。若发现新增修改或哈希漂移，先报告而不是覆盖。旧 `main` 的本地缓存不能替代当前工作分支。

注意区分三种身份：开发分支最新 HEAD 是 `b4eeb760...`；PR #28 的开发 head 为 `a14cc291...`；不可变 Candidate 的 source 是合并提交 `dac245947...`。不能混为同一 SHA，也不能把当前工作区重新构建的程序冒充已有 Candidate。

## 3. 规范来源和历史快照问题

项目为东北大学校园网的 Go SDK 与命令行客户端，正执行 `IPGW-META-V1`，revision `2026-08-28-r2`。三个入口是 `ipgw`、`ipgw-meta`、`ipgw-legacy`。

以 `docs/` 为产品规范唯一来源；`agent/` 是执行索引与交接，不得改写产品规范。开始前完整阅读：

- `AGENTS.md`
- `docs/upgrade/plan.md`
- `docs/upgrade/status.md`
- `docs/upgrade/migration-matrix.md`
- `agent/plans/stabilization-v1.md`
- `agent/handoff.md`
- `docs/architecture/protocol-correctness.md`、`security.md`
- `docs/operations/release.md`、`live-validation.md`、`offline-install.md`
- `docs/evidence/README.md`、`docs/runbooks/campus-lab.md`
- [ADR-0007](../docs/architecture/decisions/ADR-0007-immutable-candidate-promotion.md) 不可变 Candidate 晋升、[ADR-0009](../docs/architecture/decisions/ADR-0009-separated-live-test-plane.md) 测试平面分离、[ADR-0011](../docs/architecture/decisions/ADR-0011-nonblocking-m0-governance.md) M0 治理，以及 [ADR-0012](../docs/architecture/decisions/ADR-0012-bounded-anonymous-acid-redirect.md)/0013/0014/0015 协议修复。

`agent/handoff.md` 和 `docs/upgrade/status.md` 仍包含 9 月 2 日执行快照；`b4eeb760...` 的提交说明也是保留该快照。里面的当前分支、远端 main、下一工作包等历史描述不能直接当成今天状态。不要为了“继续”重做已完成的 RC-BUILD 或旧 LAB-TRANSFER。后续若要更新正式进度，应以核实证据更新，而不是把本交接的历史摘要当成新测试结果。

已知规范边界：HTTPS-only，不禁用 TLS 验证、不加入 HTTP 凭据降级；不能用公网可访问替代目标账号登录成功；不能恢复固定公钥或宽松认证回退；修改 SDK/CLI/JSON/配置/认证能力须 docs-first；不得保存或输出密码、验证码、Cookie、ticket、LT、execution、原始认证响应或带查询的认证 URL。

## 4. 最新开发与诊断的实际进度

### 4.1 协议修复与当前 Candidate

提交链已读证实：

- 2026-09-02：PR #25 有界 `ac_id` 发现重定向；PR #26 CAS HTML 优先于 JSONP 分类。
- 2026-09-03：PR #27 动态 CAS RSA-512 envelope 兼容，合并提交 `1142977d0f7e4e9fc73076d8a024fcef6b93a259`；不是恢复历史固定公钥。
- 2026-09-03：匿名 form-contract 诊断发现额外具名控件；PR #28 将 CAS POST 固定为精确六字段闭包，不再透传其余五个控件。开发 head `a14cc291...`；普通 merge 后 source `dac245947b9bd11b5489e2fd87ca2823c2d2dac5`，tree `9c8332ef4f2be8233f52fbd5d3f090e40bb0de74`。原任务核验签名 valid，合并后 CI `33736442732.1` 六项、Native `33736442713.1` 七项成功。
- 2026-09-03：新 RC-BUILD 成功，Candidate `v1.0.0-dac245947b9b-33738413131.1`，run `33738413131` attempt 1。源门禁、唯一构建上传、六平台原生安装、attestation 共九个 job 全部成功；11/11 attestation subjects 独立验证通过。
- 2026-09-04：此 Candidate 的 Windows LAB-TRANSFER 与 password_core 工具安装完成，有对应用户返回的 pass/cleanup 记录。

Candidate 身份（来自当时独立验证记录，非本次重新查询 GitHub）：

```text
artifact ID: 9886736664
raw archive size: 55164995
raw archive SHA256:
def3f2d7d1d8cde20db1912680da912f58e5452bddc1762470d0722b906284c6
candidate-manifest SHA256:
63ad82767c3406cfa683faef5c0d1aa553a6de9e5a04cf6e9f7bbb6c34c5edcb
release-manifest SHA256:
6fcdc2dfdc847e4da8ea63c968e042e57f4d7b7c320186d78eaa946d2f645b5a
build-input SHA256:
a7819ea4e0e61df89d74fc1ab8a8f8be143199c6c1072a164b9279d5748d8905
attestation IDs: 44947814 / 44947820
```

这证明构建与原生安装门禁通过，不证明真实校园网登录成功。旧 `d851...`、`ed367...`、`fbce...`、`1142977...` Candidate 不得冒充当前正式测试对象。

### 4.2 2026-09-04 正式 password_core 仍 blocked

证据 ID `EVID-20260904-001` 绑定上述 dac245 Candidate。原任务独立验证了私有三文件闭集、canonical JSON、摘要模板、checksum 和 Candidate 绑定。

- `initial_status_offline=pass`
- `login_logged_in=blocked`，产品退出码 `3/network`
- runner exit 11；没有证明 online，也没有取得成功会话清理权。
- 未执行 cleanup/logout，维护者另行返回 `portal_offline_confirmed=true`。
- 能力仍是 `synthetic_covered + live_unverified`；blocked evidence 不能生成公开成功 summary。

该次正式 evidence 的 `network` 结果本身没有区分 HTTP 500 与传输错误，不能倒推为当时已证明 500。私有 evidence bundle SHA256 为 `8ff3f336c621f7c0b2dde706f47272974c07518f820bf9942f55ab620d0bd4cf`。本次源码迁移没有搬运该私有 evidence。

### 4.3 2026-10-13 CAS-submit 诊断 R1/R2/R3

这些版本号属于 CAS-submit 诊断工具，不是后文 key-provenance 的 R1/R2。

- CAS-submit R1：安装/离线 self-test 通过；正式 runner 因合法省略的 `credential.reference` 被 StrictMode 读取而在 preflight 停止。该失败没有凭据提示、网络请求或 CAS submit。
- CAS-submit R2：修正 profile parser，安装/self-test 通过；真实窗口发出 5 个匿名请求后，诊断 allowlist 误拦同源 CAS Cookie，返回 `request_target_rejected/request_outside_allowlist`。没有读密码，也没有发 CAS POST。它仍已消耗一次诊断运行，不是零副作用安装。
- CAS-submit R3：修正诊断门禁，安装/self-test 通过；维护者另行授权一次正式窗口后执行。返回精确六字段 CAS submit 已发送，收到明确 HTTP 500，`transport_failure_phase=none`，无 redirect/ticket/activation/logout，`login_succeeded=false`。共 8 个网络请求、3 次 DNS、1 次私有密码提示。原任务随后收到维护者 `portal_offline_confirmed=true`，该窗口闭合，不应自动重跑。

R3 中 `result=pass` 表示诊断分类成功，不是登录成功。凭据操作执行者是 `instrumented_sdk_source_tree`；Candidate binary 只被校验并用于一次本地 profile show，不能把 instrumented diagnostic 当作正式 Candidate live-gate 通过。

### 4.4 2026-10-13 key-provenance R1 的结论

CAS-submit 六字段闭包仍得到 HTTP 500后，进入零凭据公钥来源诊断。更早“通用 publicKey 抢在 publicKeyStr 前被选中”的假设已被 R1 排除，不应继续当成根因。

R1 用户返回的实际结果：

- `publicKey` 赋值数为 0；有两个不同且有效的 `publicKeyStr`。
- 当前产品选择第一个匹配值；PKIX RSA、512–1023 bit 桶、指数 65537。
- 3 个脚本候选、抓取 3 个；总计 7 个匿名 GET、3 次 DNS。
- 0 凭据提示/读取、0 CAS POST、0 activation、0 logout；未输出敏感值或原始响应。
- R1 没有给出第二个赋值的位数桶/脚本序号，因此尚不能判断哪个才是密码提交应使用的密钥。
- 未使用历史固定密钥或固定密钥 fallback。

不能凭两个键的存在就改选第二个键，也不能硬编码旧公钥。需要补充可审计的公开脚本语义和安全来源信息。

### 4.5 key-provenance R2：安装已完成，真实诊断未执行

R2 新增安全赋值序号、脚本序号、位数桶、等价分组和是否当前选中项；不输出密钥值、密钥 hash、脚本路径或正文。

原任务记录已完成本地 Go test/race/vet、PowerShell 5.1/7 的 21 cases / 144 assertions、确定性封包和 gitleaks；这是历史验证，本次交接没有重跑。

R2 产物身份：

```text
诊断包：ipgw-meta-lab-keyprovenance-tools-dac245947b9b-33738413131-r2.tar.gz
size: 2954173
SHA256: 59a04bf45dd5d84db25319cf93696d2d3896216b2a19cd9d469cb0133a61bdc2

引导脚本：install-keyprovenance-direct.ps1
size: 30888
SHA256: aeecf6f269166112a82be8a2e56a3b165465d5e5ff558db99b697121cb56ff8f
```

它们当时位于 BHKLAPTOP `build/lab-keyprovenance-dac245947b9b-33738413131-r2/` 和 `build/lab-keyprovenance-direct-dac245947b9b-33738413131-r2/`；已在旧任务中上传到 BHKDesktop，远端闭集/hash/ACL/只读核验通过。

2026-10-13 09:48:57 UTC，维护者返回 R2 安装完成 JSON：`result=pass`、`stage=complete`、`tools_installed=true`、`candidate_verified=true`、`self_test_verified=true`、`product_invocations=0`、`candidate_invocations=0`、`diagnostic_invocations=1`、`credential_reads=0`、`network_requests=0`、`cleanup_confirmed=true`。这里的一次 diagnostic invocation 是离线 self-test，不是校园网 R2 运行。

随后助手仅请求“授权一次 BHKDesktop Windows / campus wired / keyprovenance R2 零凭据诊断；portal_offline_confirmed=true”。在此之后至最新提交，没有找到该授权或真实诊断结果。不能执行旧消息中的建议，也不能宣称 R2 已测过校园网。

历史安装位于该用户 LOCALAPPDATA 的私有工具目录；当前存在性、完整性和安装 manifest 应在 BHKDesktop 只读核实。不要仅凭历史 pass 就执行，也不要因新代码目录没有工具包就重复安装。此次迁移没有搬运这些 ignored 诊断源码/包；它们在笔记本原件仍保留。若后续确需补充材料，应明确选择文件、哈希与敏感边界，不能复制整个 build/。

## 5. 真实测试尚未完成的事项

截至已读取证据：当前 Candidate 未通过真实 password_core 闭环；key-provenance R2 真实诊断尚未执行；未证明正确密码密钥选择；没有因这些诊断而完成新的协议修复或新 Candidate；Wi-Fi/QR 等真实矩阵不能标为完成；没有完成 promotion/release。

原任务在 Candidate 构建时记录未创建 v1.0.0 tag/draft/release/promotion；这是当时记录，发布前仍须重新查远端，不能把历史不存在当成永久事实。

下一项可能的真实工作是一次新的、有独立授权的 key-provenance R2 零凭据诊断，而不是重跑 password_core 或 CAS-submit R3。计划应先核对 R2 工具实现及当时边界：匿名 GET 最多 12、DNS 最多 12、Candidate 仅一次本地 profile show；密码提示/读取、CAS submit、activation、logout 均为 0。旧预算仅供制定新计划，不自动授权执行。

任何新的真实窗口前重新确认 BHKDesktop 原生 Windows、单一校园有线 UP、Wi-Fi UI OFF且无 UP 无线接口、TUN OFF、无 IPv4 /1 覆盖路由、默认路由仅有线、精确绑定 profile、无其他自动登录客户端；维护者私下确认账号 offline，管理/恢复路径可用。2026-10-13 的 offline 和拓扑结果已经过期，不能沿用。不要让用户把账号、密码、原始响应或网络身份材料贴到聊天中。NAS 的复核/恢复由维护者自行控制，不自动连接它。

## 6. BHKDesktop 工具环境

`go.mod` 基线是 Go 1.25.0。普通测试/构建说明包含 `go test ./...`、`go vet ./...`、`go test -race ./...`、`go run ./cmd/doccheck --check`；这里仅记录，不是本轮执行指令。

正式 `candidate-build` 在 Makefile 中要求本地精确 `go1.25.0`、Linux/amd64，使用 `GOTOOLCHAIN=local`；完整 CI 涉及 Bash/sh、make、Python 3、race 工具链及固定 gitleaks 8.30.1。Windows 原生开发不自动满足正式 Candidate 构建条件；不要把工作区 build 代替不可变 CI Candidate。

迁移期间台式机检查曾观察 Windows Go 1.23.4（由父任务目标机检查提供），不足以直接认定满足当前模块要求。新任务应只读确认实际 PATH/版本和现有 Linux/WSL 环境，不自动安装或升级。2026-10-13 的 Windows SSH 用户上下文曾返回空的 WSL 列表，故不能假定存在可用 WSL 开发环境；当前 WSL 构建能力仍未验证。旧 Windows-only 诊断安装明确不依赖 WSL；WSL 也不是校园网产品测试平面。

## 7. 2026-10-28 迁移证据和保留范围

迁移从规范 source 工作区构建新副本：219 个跟踪文件、两份工作区文档、63 个 Git refs；Git bundle 仅包含这些 refs 的可达历史，不复制源 hooks/config/reflog/不可达对象。origin/upstream 已在目标正确重建，未指向临时 bundle。选定 Candidate 共 15 个文件（原始 ZIP + 14 文件 candidate-set）迁入：

`D:\project\Go\IPGW-Meta\build\rc-build-v1.0.0-dac245947b9b-33738413131.1`

这 15 个文件总计 110327854 字节，21 条 SHA256SUMS 引用复核通过；源快照、可达 Git 历史和选定 Candidate 的 gitleaks 扫描退出码为 0、命中 0。扫描零命中不是无秘密的绝对证明。私有 live-evidence、其他诊断版本、管理链路材料、凭据、OS keyring、未知备份和缓存未被无差别迁移；原件未删除。

迁移包为 113694720 字节，共 240 个传输文件，远端逐项大小/hash、219 源码文件、15 Candidate 文件、63 refs、HEAD/branch/status 校验及 `git fsck` 通过。旧目录无可见命令行/程序路径引用后，使用拒绝覆盖的目录改名完成归档和接替；没有终止进程。完整文件占用检查不在该检查能力范围内，但改名均成功。

```text
迁移包 SHA256:
2f337b9bc456856ed84edc3fa65a2d924aa2a88b01d2bbd5c0812b8315443199
源码 manifest SHA256:
fb4c60e1c02ec85fed8df43f61b735cca7f1b5e36dc4992714a86da2b570dceb
payload-manifest SHA256:
fb9c09e7b7ba74ff7eaa9836981f15632ae4c14f29c61c8ccdce63e8b524895f
Git bundle SHA256:
717a5f64bb4a6eac7272797a390dcd5afea571243e6056a654a87a1fee2981ba
```

台式机暂存目录内可读证据：`validation.json`、`cutover-result.json`，以及 `payload/manifest.json`、`payload/candidate-manifest.json`、`payload/payload-manifest.json`、`payload/scan-summary.json`、`payload/git-state.json`。这些迁移清单刻画 a14cc291 + 两份 dirty 文档时点；后来的 b4eeb760 已提交相同字节，清单不是新的 Git HEAD 声明。

源侧回执副本和本交接文件位于 BHKLAPTOP 任务目录 `C:\Users\Unbal\Documents\Codex\2026-10-28\task`；它们是否已被父任务传入目标，应以新任务实际附件/文件为准。本交接整理本身没有再次传输源码或创建新任务。

迁移时成功管理入口是用户确认的 NAS 端口转发 `neu6.unbalancedcat.cn:10001`，SSH 登录名 `bhkdesktop\unbal`，严格已有 host-key 校验和 BatchMode 下认证成功。这里仅记历史管理路径，不作为后续开发默认；曾经使用的 `192.168.3.10:22` 在此次笔记本环境与本机地址重合，不能盲目复用。不得接受未知/变更密钥、关闭校验或读取私钥内容。

## 8. 项目权限修复与 Hand off 的最终结论

父任务在 BHKDesktop 已处理迁移后两个目录 owner 问题：经用户明确批准并完成 UAC，仅 `D:\project\Go\IPGW-Meta` 和其 `.git` 目录 owner 改为 Unbal，非递归，DACL/group/inheritance 未变。普通 Unbal Git 可用；目标应用 `list_projects` 正式识别此路径、project ID 和 `isGitRepository=true`。不要重复修改 ACL/owner，也不要全局放宽 Git safe.directory。

Hand off 曾先被 pending composer state 阻止，后因 `No matching saved project was found on BHKDesktop.` 多次在前置检查停止；每次均无 operation ID，且只在用户提供实质新证据后重试。最终即使目标本地项目已识别，笔记本跨主机匹配仍失败。没有创建移交 worktree，原任务仍在笔记本。这不表示源码或 Git 仓库损坏。

用户已批准放弃此路线，改为目标项目新任务。不要继续重复 Hand off、复制内部会话存储、擅自更改项目记录或要求用户反复切设备。新任务不是原聊天的完整副本，本交接与保留的原任务是上下文来源。

## 9. 新任务推荐的第一轮输出

1. 确认运行主机确为 BHKDesktop、项目/cwd 为目标 canonical 路径；只读列出 Git 状态和工具环境，不启动产品。
2. 验证最新提交 b4eeb760、目标 a14cc291 + 两份文档的相同字节，按第 2 节安全对齐提交；如有额外改动先保全并报告。不要凭陈旧 handoff 文档切回旧分支。
3. 核对选定 Candidate 的位置/hash/只读属性；只读确认历史 key-provenance R2 安装是否仍存在、对应 manifest 和工具 hash 是否匹配。不要自动重装、运行 self-test 或访问凭据。
4. 汇报“已完成的历史诊断、未执行的 R2 真实诊断、现有环境缺口”。在没有新授权时停在这里，不运行校园网诊断。
5. 若用户要继续真实排障，再提出单次 R2 零凭据窗口的具体拓扑、请求预算、输出闭集和停止条件。R2 结果未出来前，不猜根因、不切换密钥、不硬编码、不重跑 CAS-submit/password_core、不晋升 Candidate 或发布。

关键原任务证据 turn：PR #28 merge `01a0667d-e56f-7df1-b908-e1f96e6d6c5e`；Candidate RC-BUILD `01a0668e-63ae-7852-a9a4-350af9b5dd06`；正式 evidence 复核 `01a06a77-ae04-70c3-84c2-cdf21a6e793`；CAS-submit R3 HTTP 500 `01a13427-f268-7f30-8609-cf0bb8b77efc`；key-provenance R1 `01a13464-a1e4-7631-8be8-06c1cabccf69`；key-provenance R2 安装/self-test `01a134aa-4992-7c12-856d-236eb7697190`；最新文档提交 `01a181f2-230e-7841-8661-38e9a9637c52`。需要复查时用受支持任务读取接口，不搜索内部会话文件。

---

## 本轮结束：2026-10-28 接手完成

以下仅为本轮迁移与本地完整性核验，不是校园网、登录、安装、自测或发布证据。未来继续维护以本任务和 BHKDesktop 当前目录为入口；不重试旧 Hand off，不默认从笔记本传输开发结果。

- 主机实测 `BHKDesktop`；cwd/Git 根为 `D:\project\Go\IPGW-Meta`；应用项目 `df9d39a0-f8ff-4162-b72e-70565c26e535` 仍识别为本地 Git 仓库。
- 已阅读本目录 AGENTS、计划/状态/矩阵/工作包索引、历史 handoff、协议/安全/发布/真实验收/离线安装/evidence/实验室规范与 [ADR-0007](../docs/architecture/decisions/ADR-0007-immutable-candidate-promotion.md)/0009/0011–0015。计划为 `IPGW-META-V1 / 2026-08-28-r2`；不存在本地 `.agents`，故没有可读的 `.agents/skills`。本轮是明确授权的接手迁移，不重开已完成的 RC-BUILD 或 LAB-TRANSFER，也不宣称执行完整 `WP-BASELINE-VERIFY` 测试门禁。
- 初始受限账户 `CodexSandboxOffline` 被 Git ownership 检查阻止；经正规工具审批在普通用户上下文完成 Git 核验、fetch 与对齐。未修改 owner、ACL 或 safe.directory。build 回执目录的受限写入也经正规审批解决；没有审批拒绝或未决 mutation。
- origin fetch/push 均为 `https://github.com/UnbalancedCat/ipgw-meta.git`。仅从既有 origin 获取当前工作分支；确认远端为 `b4eeb760945f0568ca163d957ddd55df2afa86c4`，旧 HEAD `a14cc291fed335a6b027f42f72a9775f603aa3cd` 是其祖先。未查询当前 main、发布状态或 GitHub attestation 有效性。
- 219 个原有跟踪文件逐项核对 Git 原始 blob 和迁移源清单的 size/SHA256/mode 全部一致；两份 dirty 文档精确匹配上文给定 SHA256，index 无暂存变化。对齐前保存 diff、完整 index、两份原字节副本、refs/status 和全部跟踪文件哈希；临执行再次核对无漂移后使用 `git reset --mixed` 仅对齐引用/index。对齐后 HEAD/upstream 均为 `b4eeb760...`，219 文件字节不变，tracked clean，`git fsck --full` 退出 0。未创建提交、未推送。
- 独立恢复回执位于 `build/takeover-bhkdesktop-20261028/`，处于既有 build ignore 范围。新交接是唯一非忽略的未跟踪文件；旧 `agent/handoff.md` 与正式 `docs/upgrade/status.md` 均保留原字节。
- 选定 Candidate 的原始 ZIP 与 14 文件集合共 15 个文件、110327854 bytes，全部只读，raw archive/candidate-manifest/release-manifest SHA256 与交接一致；manifest 12 项大小/hash 和两级 21 条 checksum 引用通过。build-input 字段匹配历史值，但没有重新计算 source build-input，也未重新验证线上 artifact 生命周期/attestation。
- LOCALAPPDATA 中 dac245 Candidate 四文件闭集与仓库 Candidate manifest、目标产品 hash、helper hash 和三条 transfer checksums 一致，全部只读。未执行任何产品命令。
- key-provenance R2 目录仍存在，精确三文件为 `ipgw-meta-keyprovenance-windows-amd64.exe`、`run-keyprovenance.ps1`、`TOOL-SHA256SUMS`，全部只读；两个工具的大小分别为 7150080/22374 bytes，SHA256 分别为 `02f3aff22894dc7ca4a32bb24c9af66fa5859dd88ebd04f566641342a7823971` / `c0aeeb6c7f5f7e15055e001434b675d48ad805a7e81aa9ff15e6d0bbb8311859`，与既有 TOOL-SHA256SUMS 一致。安装目录未见独立 JSON manifest；此次只验证现有校验清单的自洽性，没有原始 R2 tar.gz/引导脚本可用于重新建立独立包绑定。未重装、未读脚本正文、未运行 self-test 或真实诊断。
- 工具实测：Git `2.45.2.windows.1`；Go `1.23.4 windows/amd64`（`C:\Program Files\Go\bin\go.exe`，仅在进程内设置 `GOTOOLCHAIN=local` 读取版本）；PowerShell `7.6.5` / Windows PowerShell `5.1.26100.8655`；Python `3.12.4`；Git Bash `5.2.26`；GCC 为 Strawberry 的 `i686-posix-sjlj 4.9.2`。PATH 未找到 make、sh、gitleaks、rg、gh。Git Bash 存在但未作为构建环境验证；未证明该 GCC 可满足 race 工具链要求。
- WSL 列表实测 Ubuntu-22.04、docker-desktop、docker-desktop-data 均为 WSL2/Stopped，running 列表为空。未启动发行版或检查其内部工具链，因此 Linux/WSL 构建能力仍未核实。Go 1.23.4 不满足模块 Go 1.25.0；正式 Candidate 构建还要求精确 Go 1.25.0/Linux amd64。本轮未安装/升级任何依赖。
- 未运行 Go test/race/vet/doccheck、全仓秘密扫描、产品、self-test、校园网诊断、登录/注销或任何真实网络测试；未读取凭据/profile/私有原始 evidence。未改网络/安全权限，未创建发布或 workflow dispatch。
- 当前仍无成功登录闭环证据；key-provenance R2 的真实校园网运行仍未授权、未执行。后续若继续真实排障，先另行制定并取得单次 R2 零凭据窗口授权，同时重新确认拓扑/账号 offline/恢复路径和工具可信来源。开发测试及工具升级也不属于本轮授权。

接手状态：完成并停止；剩余为环境能力缺口与未来独立授权事项，无需创建其他接续任务。

---

## 2026-10-28 R2 来源审计窗口

- 执行者：Codex；范围为 `LAB-PASSWORD-BHK` 阻塞项的前置来源审计，不执行正式 suite，不重开 `RC-BUILD` 或 `LAB-TRANSFER`。开始时分支/HEAD 与接手基线一致，仅本交接未跟踪；旧 `agent/handoff.md` 和 `docs/upgrade/status.md` 原字节保持不变。
- 授权边界：只读检查既有工具、明确原件和构建记录，并保存本次审计回执；不修复、安装或运行诊断/self-test/Go 测试，不读 profile/凭据，不启动 WSL，不请求校园网，不提交或推送。规范引用 [REL-LIVEGATE-001](../docs/operations/live-validation.md)、[REL-APPROVAL-001](../docs/operations/release.md)、[EVID-REDACT-001](../docs/evidence/README.md)。
- 本地核验：R1/R2 三文件及本地 checksums 均一致；R2 runner 内置诊断哈希精确匹配 R1 EXE，不匹配 R2 EXE。静态控制流证明，若更早的前置检查通过，它将在拓扑查询、profile show 和诊断启动之前返回 `diagnostic_invalid`。未执行 runner。
- 来源链状态：`blocked`。当前安装集合已核验；原始 R2 包、引导脚本、源码/模板、构建与安装自测覆盖记录尚未独立核实，不能确定引入环节。指定本地 build/迁移 payload 文件名检索未找到原件；受支持原任务读取接口返回主机不可用。此状态不改变正式里程碑或认证能力。
- 结束回执：`build/r2-source-audit-20261028/audit.json` 保存来源核验表、完整工具哈希、缺口和待审修复范围；这是 ignored 本地审计记录，不是正式 live evidence 或成功摘要。当前安装文件、Candidate 和两份历史快照未修改。
- 下一步仅由维护者在 BHKLAPTOP 只读核对交接列明的两份原件大小/hash，返回封闭结果后再决定选择性补齐。每次只提供一个完整、自包含步骤；原件不可取得或校验失败继续 `blocked`，不得直接改已安装 runner 哈希或重建冒充原件。
- 后续材料核验：维护者返回 BHKLAPTOP 两份原件的 metadata_match，大小/hash 精确匹配交接；回执保存为 build/r2-source-audit-20261028/laptop-originals-metadata.json，明确标注为维护者返回，尚未在 BHKDesktop 独立读取原始字节。来源链继续 blocked，待原包成员、bootstrap 和源码/构建记录核验；下一步只选择性接收该两文件到同目录 originals/，不执行。

- 2026-10-28 original-package follow-up: both received originals independently match the handoff size/SHA256. All three packaged payload files are byte-identical to installed R2. The original packaged runner already embeds the R1 EXE hash (line 14), so the mismatch existed before installation; the precise source/template/build step remains unproven. Static bootstrap review confirms direct EXE --self-test only (lines 384-427, 681-688), with no runner binding check. Receipt: `build/r2-source-audit-20261028/original-package-static-audit.json`. Source chain remains blocked pending source/template/build records. No extraction to disk, diagnostic/test execution, installation or network requests.

## 2026-10-28 BHKDesktop 独立开发测试启动

- 维护者已确认并要求实施独立开发测试计划；唯一源码目录为 BHKDesktop `D:\project\Go\IPGW-Meta`。不再要求 BHKLAPTOP 材料；R2 原包错误已证明，历史生成环节未知仅作为历史缺口，不阻塞新开发。
- 开始基线：分支 `codex/cas-submit-field-closure`，HEAD `b4eeb760945f0568ca163d957ddd55df2afa86c4`，跟踪文件无变化，仅本交接未跟踪；旧 handoff/status 原字节保留。
- 当前执行工作包 A：先验证非 TUN 管理连接。脱敏被动基线及阶段状态位于 `build/desktop-dev-20261028/management-before.json`、`state.json`。等待维护者关闭 TUN 后返回当前聊天；先保留 Clash 核心、系统代理与当前节点。失联时由维护者重新开启 TUN 恢复。
- 工作包 A 未通过前不开始环境下载/安装或测试。未来开发测试按已确认计划分离管理连接、依赖准备、合成测试与真实校园网验证；本阶段不创建 Candidate、不提交/推送/发布、不执行校园网诊断。此记录不改变正式里程碑和 password_core blocked 状态。

- 工作包 A：维护者关闭 TUN 后，新聊天和本地工具调用正常；Mihomo 网卡/路由已退出，显式回环代理的 OpenAI 文档、Go 模块代理、Go 版本元数据 GET 和归档 HEAD 探测通过 TLS 校验。Go 目录 HEAD 405 已明确为方法不支持，未作为连通失败；功能验收通过。未独立证明所有流量的 IPv6 上游或计费路径，未访问校园网接口。开始工作包 B 环境盘点。

### 本独立开发窗口结束

- 管理连接：非 TUN 功能验证通过；全部流量 IPv6/计费未独立证明。保持维护者当前非 TUN 设置。Ubuntu 已按授权启动，未启动 Docker、未改 WSL NAT/代理/防火墙。
- 工具与入口：两端独立 Go 1.25.0、Linux 固定检查工具及离线模块缓存已就绪。开发操作见 [本机开发说明](../docs/operations/development.md)，逐次结果和完整失败/skip 清单位于 `build/desktop-dev-20261028/summary.json`，可读报告为 `REPORT.md`。
- 已验证：Windows CAS/SDK 聚焦测试、vet/doccheck/开发构建；Linux 普通测试、race、vet/doccheck、质量门禁和六平台交叉构建通过。合法日期型 evidence 实例的 doccheck 误判已按现有规范修复并增加回归。
- 当前阻塞：Windows 四个安装器叶用例仍在 active junction 执行合成 EXE 时失败；短路径没有解决。独立 junction 探针不足以解释完整失败，安装器试改已撤回，产品认证行为不变。Windows 符号链接权限与原生资产条件 skip 未当作通过。
- CAS：8 个纯合成场景记录顺序依赖、页面/脚本优先和标识符子串匹配；未改变选键规则，未产生真实网络或登录证据。
- 下一工作包：保留失败 synthetic fixture，定位版本目录、junction 解析与执行环境，再针对性修复和回归；不依赖笔记本，不放宽安全约束。旧 handoff/status 与既有 Candidate/R2 哈希复核不变，未提交/推送/发布。

### Windows 安装测试续接窗口启动

- 维护者要求继续；复核分支、HEAD 和工作区变化与上一窗口结束一致。沿用独立开发授权，仅检查合成安装 fixture、修复有证据支持的缺陷并回归；不执行真实校园网操作。
- 首先补充失败时的 junction/目标文件诊断，在自动清理前记录现场；旧 handoff/status 保持原字节。

### Windows 安装测试续接检查点

- 聚焦失败已复现并在清理前保存现场元数据：目标文件存在且直接执行成功，经 active 执行失败。独立最小探针表明 PowerShell 5/7 与 cmd 创建的新 junction 同样失败，填充 Print Name 或使用扩展路径不能解决；未证明是安装器缺陷，未修改 install.ps1。
- 新增开发入口 install 阶段和可选测试名过滤，仅用于聚焦回归，不能替代全仓基线。文档检查与 diff whitespace 检查通过。
- 等待维护者在桌面原生 PowerShell 执行 `build/desktop-dev-20261028/native-junction-check.ps1` 返回封闭 JSON，以区别工具执行环境和本机原生行为。脚本已在工具内验证，仅运行固定哈希的合成打印程序；不修改权限、网络或已有安装。结果更新见同目录 REPORT.md 续接节，上一窗口 summary.json 保留。

- 原生终端首次返回：LOCALAPPDATA 下原探针 EXE 不存在，脚本在读取哈希时停止，没有执行 junction 测试。新增阻塞为跨执行上下文的 C 盘文件可见性差异；此前工具链就绪与测试通过仅证明工具执行上下文，不能宣称桌面原生开发环境已就绪。探针改为读取同机 D 盘脚本旁经固定哈希校验的合成 EXE，下一次同时返回旧探针和 Go 路径是否可见；不复制工具链或改权限来猜测修复。

- 原生终端二次返回已核对共享 D 盘回执：direct 与 junction 均成功，旧探针和 Go 均不可见。此前四个失败仍是工具上下文中的失败；未确立安装器缺陷，install.ps1 保持原样。下一步是维护者在原生终端执行已准备的 `build/desktop-dev-20261028/native-windows-bootstrap.py --execute`：校验本地官方归档与固定缓存清单、安装隔离 Go、离线准备依赖、运行 Windows 原生合成基线。253 个缓存输入及全新空缓存的离线准备已在工具上下文预检通过；原生环境就绪仍待回执，不改系统 PATH 或沙盒/ACL。

- Windows 原生准备及 focus/test/vet/docs/build/cas-observe 回执已逐项读取核对，全部通过。原生 test 回执 `build/dev-runs/windows/20261028T141653Z-e358438b-test/result.json` 中此前四个安装失败用例均明确 pass，无 fail；6 个条件 skip 保留。Windows 原生环境现已就绪，安装器未改，模块文件未变。下一步仅从桌面启动已有 Ubuntu-22.04 执行 `build/desktop-dev-20261028/native-linux-check.py`，验证已有 Linux 离线检查在该上下文的可用性，不安装或下载。

### 独立开发与测试计划最终交接

- 维护者从桌面启动 Ubuntu-22.04 的七阶段回执已逐项核对：普通测试、race、vet/doccheck、质量门禁、六平台三个入口共 18 个交叉构建、CAS 观察全部通过。两端原生工具与入口均已实际验证，完整索引为 `build/desktop-dev-20261028/final-summary.json`。
- Windows 六项条件 skip、Linux 普通/race 各两项条件 skip 在最终索引中单列；未提升权限或选择 Candidate 资产来消除 skip，不外推为完整发布原生矩阵通过。四个历史安装失败在 Windows 原生基线明确通过，install.ps1 未改；工具执行上下文与原生 C 盘可见性差异仍作为环境注意事项保留。
- 当前分支/HEAD 未变，所有开发改动未提交/推送。旧 handoff/status 原字节核对通过；没有重建 Candidate、修改历史 R2、真实校园网诊断或发布。正式里程碑和 password_core blocked/live_unverified 不变。
- CAS 八场景两端逐字段相同；分析与下一工作包位于 `build/desktop-dev-20261028/CAS-GAPS.md`。下一阶段先做新开发诊断的来源绑定/离线验证及必要规范决策；真实网络操作另行取得单次授权。
- 本独立开发窗口验收完成并交接，当前无需维护者操作。当前开发说明见 docs/operations/development.md，最终结果见本地 REPORT.md；各早期阶段记录保留为历史，不按其旧“下一步”重复执行。

### CAS 来源诊断离线设计窗口启动

- 维护者要求开始下一阶段；主机/分支/HEAD 与上一窗口相同，现有未提交改动全部保留。范围为新诊断的离线设计、ADR 草案和合成行为观察，不接入真实网络，不读取 profile/凭据，不修改公共认证行为。
- 重点为来源语法/使用关系的证据等级、脱敏闭集输出、独立工具身份、runner/EXE 来源绑定以及整条离线自测；缺少真实依据的选键语义保持待决。

### CAS 来源诊断离线设计窗口结束

- 已形成 [ADR-0016](../docs/architecture/decisions/ADR-0016-cas-provenance-before-key-selection.md) 与 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 草案，均为 proposed。新诊断工具/runner 尚未实现；产品选择首键/次键/多键拒绝等规则未决定，现有认证实现未修改。
- 现有内部 cas-observe 扩展为 19 个内建合成场景。Windows/Linux 工具上下文输出逐字段一致，真实网络请求与真实凭据读取为零；后缀与转义标识符场景未取到键，注释/字符串/模板/未调用函数等场景被当前文本正则匹配。结果只证明合成行为，不重新打开已被历史 R1 排除的真实假设，也不是生产协议正确性证据。
- 回执：`build/desktop-dev-20261028/cas-design-summary.json`；详细运行分别为 `build/dev-runs/windows/20261028T144006Z-0bb4ec02-cas-observe/` 与 `build/dev-runs/linux/20261028T144008Z-c0c08a6b-cas-observe/`。生成稳定 ID 索引后 doccheck 通过，明确变更文件集合的脱敏 gitleaks 扫描通过。未重复完整基线，未创建新 Candidate、提交/推送或执行真实诊断。
- 下一依赖已满足的工作包为 WP-CAS-PROVENANCE-OFFLINE：先收敛草案输出 enum/分析覆盖契约，实现仅内建合成输入的分析器与报告，随后另阶段实现独立入口校验器/runner 绑定及整链自测。实现范围引用草案，不把来源观测变为产品选键决策。旧 handoff/status 继续保留原字节。

### CAS 来源分析器第一离线实现窗口启动

- 维护者要求自主继续；本窗口在既有环境执行必要合成验证，不需要其重复操作。当前分支/HEAD 与已知基线一致，所有现有改动保留。
- 先在草案收敛有限词法/直线赋值子集、未知与拒绝状态及闭集报告；未打包原型明确 not_packaged，不伪造来源 verified。实现仅限内建合成入口及分析/报告测试，不接入真实网络，不改变产品取键。

### CAS 来源分析器第一离线实现窗口结束

- WP-CAS-PROVENANCE-OFFLINE 的分析器/闭集报告子阶段已实现，入口为 `internal/cmd/cas-provenance`，重复验收为 `python scripts/dev.py cas-provenance --context tool`（Linux 使用 python3）。边界引用 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)，[ADR-0016](../docs/architecture/decisions/ADR-0016-cas-provenance-before-key-selection.md) 仍为 proposed；包绑定/可信启动器/runner 整链自测未实现，整个工作包尚未完成。
- Windows/Linux 工具上下文最终聚焦运行均通过：12 个顶层 Test、2 个 fuzz seed 目标，各 65 个 pass 事件（含子用例），零 fail/skip；Linux 另通过 race 和两个有界 fuzz，最终运行分别 25120 / 93027 次执行。两端 vet、11 场景内建 self-test 通过，脱敏示例 JSON 逐字段相同。此次没有重复原生安装基线或要求维护者操作。
- 最终回执分别为 `build/dev-runs/windows/20261028T151613Z-d57aa30e-cas-provenance/` 和 `build/dev-runs/linux/20261028T151624Z-0b48b8e1-cas-provenance/`；汇总和源文件摘要位于 `build/desktop-dev-20261028/cas-provenance-summary.json`、`CAS-PROVENANCE-REPORT.md`。这些是开发验证记录，不是来源已验证的安装包或正式 live evidence。
- 本窗口未修改公共 SDK/CLI/认证实现、安装器、模块文件或旧 handoff/status；既有未提交改动保留，无提交/推送/发布/Candidate 构建。Linux 一次启动报告 systemd 用户会话警告，但目标命令成功；最终运行未复现，未更改服务配置。正式 password_core 仍 blocked/live_unverified。
- 下一窗口为独立开发包生成、外部信任锚点、启动校验器/runner 与整链负例；开始前继续收敛同一草案的打包字段和检查/使用窗口，不碰历史 R2。真实采集及产品选键决策保持独立待审。

### CAS 来源绑定与整链离线验证启动

- 维护者明确要求持续推进，不再因 30 分钟左右强制停止；当前分支/HEAD 不变，已有未提交文件保留。当前执行约定已补入 [REL-WINDOW-001](../docs/operations/release.md)。
- 工作范围引用 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)：先收敛声明式 runner、四文件开发包及包外信任锚点，再实现可信启动器和实际文件锁/整链负例；不安装、不修改历史 R2/Candidate、不联网、不改变产品选键。

### 2026-10-29 CAS 开发绑定链完成记录

- 本次连续完成开发包生成、独立可信启动器、声明式 runner、严格元数据/外层回执、Windows 祖先目录与文件锁，以及实际分析器整链自测。工作包依赖和边界仍引用 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)；没有按时长中断。安装/转移、静态使用关系和真实采集未实现或执行，不能把开发包回执当作正式 evidence。
- Windows 最终回执为 `build/dev-runs/windows/20261028T160206Z-93559525-cas-binding/result.json`：92 个 pass 事件（包含子用例）、零 fail/skip；实际新包 runner=passed、EXE 尝试/启动均为 1、11 场景分析器自测通过。拒绝路径用独立合成执行标记和零尝试/启动计数验证；启动前和子进程存活期间写入/重命名均被拒绝，取消后进程及锁均回收。
- Linux 最终回执为 `build/dev-runs/linux/20261028T155822Z-88c49404-cas-binding/result.json`：75 个 pass 事件、零 fail/skip，元数据/严格报告 race、vet 和有界 fuzz 通过；Linux 不提供保护执行能力，Windows 专属测试未编入 Linux，不列作 Linux 已通过项。
- 四成员 ZIP 与目录逐项字节一致，16 项源码输入与构建快照一致，构建器拒绝重复输出并保留原包。首次 C 盘工具夹具因最终句柄路径不一致被拒绝，改为共享 D 盘独立开发夹具后通过；未修改路径约束、ACL、sandbox 或 safe.directory。中间 vet 的跨包非命名字段初始化已修复；最终回执为修复后结果。
- 汇总及详细记录为 `build/desktop-dev-20261028/cas-binding-summary.json` 和 `CAS-BINDING-REPORT.md`。旧 handoff/status、产品认证、安装器、模块和历史 R2/Candidate 保持原样，所有源码改动未提交/推送。正式 password_core 仍 blocked/live_unverified。
- 当前授权的开发绑定与整链验证范围已完成，无需维护者操作。后续安装或真实采集应另拟明确工作包；产品选键与静态使用关系不得根据本次绑定成功直接改变。

### 2026-10-29 开发安装合成验收启动

- 维护者要求继续；当前分支/HEAD 与前次一致，现有改动保留。推进 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 的安装/自测剩余项：实现严格归档验证、新目录独占写入、整链复核与封闭回执。
- 实际执行仅限本次 D 盘开发测试目录，不安装到既有用户工具目录、不触碰历史 R2/Candidate、不访问校园网、不改变产品选键；已更新对应规范的第三阶段边界。

### 2026-10-29 开发安装合成验收完成

- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 完成严格 ZIP 验证、新目录独占安装、整链自测和安装回执；可重复入口为 `python scripts/dev.py cas-install --context tool`。Windows 实际安装仅发生在本轮 D 盘合成目录，用户工具目录未安装或升级。
- Windows 最终回执：`build/dev-runs/windows/20261028T162621Z-4e616e52-cas-install/result.json`，37 个顶层 Test/fuzz seed 目标、134 个 pass 事件（含子用例），零 fail/skip；完整 ZIP→新目录→runner→分析器自测通过。错误锚点不创建目录，已有目录不覆盖，失败残留保留；写入后篡改在 EXE 尝试前拒绝。
- Linux 最终回执：`build/dev-runs/linux/20261028T162709Z-e8954671-cas-install/result.json`，28 个顶层目标、101 个 pass 事件，零 fail/skip；race、vet、metadata/ZIP 有界 fuzz 通过。Windows 专属测试未编入 Linux，不计作 Linux 安装通过。ZIP fuzz 本次仅 5 次执行，不声称穷尽覆盖。
- 本轮开发包从 20 项明确输入构建，身份 `dev-e86205a98469a6222008864197f305d94452b9c3f7b4d303b7b08971a2487e36`。四成员安装结果与构建目录逐项一致，原始 ZIP 不变；包外摘要只支持本地审阅来源，不是远端 attestation。
- 完整索引及说明为 `build/desktop-dev-20261028/cas-install-summary.json` 与 `CAS-INSTALL-REPORT.md`。当前阶段只完成离线开发安装验收；静态使用关系仍 unknown，真实采集器未实现，[ADR-0016](../docs/architecture/decisions/ADR-0016-cas-provenance-before-key-selection.md) 仍 proposed，password_core 仍 blocked/live_unverified。
- 分支/HEAD 未变，已有工作树改动保留且未提交/推送。旧 handoff/status 原字节、模块、产品认证与产品安装器均未改；不创建 Candidate、不发布、不操作历史 R2。后续真实采集须先收敛适配器边界和单次审批方案；当前无需维护者运行命令。

### 2026-10-29 采集适配器边界复核启动

- 维护者要求继续，沿用 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 的独立工具和离线边界。分支/HEAD 未变，已有工作树变化保留；只读核对产品入口、匿名发现、Cookie、脚本和分析器依赖。
- 按已启用的 grilling 技能准备具体决策草案，先分清产品请求上下文等价与匿名会话限制。未实现或运行真实采集，未修改旧快照、公共协议或产品选键。

### 2026-10-29 采集适配器决策准备完成

- 只读事实复核与 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 待决草案已完成，记录 D1 入口上下文、D2 无会话/URL 子集两项选择及离线验收矩阵。它们尚未被当作接受；未增加网络能力或修改已验证工具行为。
- 已纠正诊断文档首段的旧“安装未实现”说明，以及现行计划 3.2 节残余的 35 分钟强制停点；保留旧 handoff/status 原字节。当前等待维护者回答设计选择，非重复申请已授权离线测试，也不是新的校园网运行许可。
- 此阶段只做文档与源码审计，未重新运行 Go 测试或构建；文档校验通过不代表采集器实现/验收通过。正式 password_core 和 [ADR-0016](../docs/architecture/decisions/ADR-0016-cas-provenance-before-key-selection.md) 状态保持不变。

### 2026-10-29 采集回放实现启动

- 维护者明确接受 D1 保留产品 service/ac_id 上下文、D2 无会话且脚本采用严格 URL 子集。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 第四阶段实现纯内存回放及合成验收，不加入真实传输或扩展已验证安装包。

### 2026-10-29 纯内存采集回放完成

- 已实现 `internal/cascollect` 和仅内建 self-test 的 `internal/cmd/cas-collect`，重复入口 `python scripts/dev.py cas-collect --context tool`；边界引用 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)。无真实 transport、resolver、socket、URL/file 输入，模拟请求计数与实际全零 operations 分开。
- Windows 最终回执 `build/dev-runs/windows/20261028T170524Z-68ae4d1a-cas-collect/result.json`，Linux 最终回执 `build/dev-runs/linux/20261028T170606Z-2e0ce9d3-cas-collect/result.json`。两端各 29 个顶层 Test/fuzz seed 目标、122 个 pass 事件（含子用例），零 fail/skip；vet/self-test 通过，Linux race 与两个有界 fuzz 通过（19240 / 77062 次执行）。这些是聚焦回归，不替代完整发布门禁。
- 合成验证覆盖动态 ac_id/service、一次受限匿名跳转、无 Cookie 回传、脚本 URL/MIME/大小边界、全部源预检、去重及局部来源关联、计数/取消/闭集报告。两把新生成 RSA 的页面/外部来源及顺序互换证明有效页面键不短路采集、同轮等价关系正确且无密钥原文输出；没有决定产品该选哪个键。
- 完整索引与说明为 `build/desktop-dev-20261028/cas-collection-summary.json`、`CAS-COLLECTION-REPORT.md`。本次代码不在此前四文件开发包中，不重新生成或安装该包；旧工具/候选保持原样。真实传输、TLS、DNS 报文/物理连接计数和真实采集尚未实现或验证。
- 分支/HEAD 和历史 handoff/status 保持不变，无提交/推送/发布。下一依赖为真实传输适配器的独立设计和离线隔离验收，执行校园网请求仍需单次批准；[ADR-0016](../docs/architecture/decisions/ADR-0016-cas-provenance-before-key-selection.md) 仍 proposed，password_core 仍 blocked/live_unverified。

### 2026-10-29 连续推进至原生可测试包

- 维护者要求一次推进到可亲自测试阶段；已确认 D1/D2，不再按离线子阶段停止。HEAD/分支未变，已有变更保留。范围为 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 的独立传输实现、合成验收、来源绑定和原生交互入口；Agent 不执行真实校园网请求。

### 2026-10-29 原生匿名诊断开发包就绪

- 已连续完成 WP-CAS-ACQUIRE-DEV，验收边界引用 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 第五阶段；重复入口为 `python scripts/dev.py cas-acquire --context tool`。真实传输仅在 loopback UDP/HTTP/TLS 夹具中验证，真实校园网诊断尚未执行；本条不更新正式里程碑或 password_core 状态。
- Windows 最终回执 `build/dev-runs/windows/20261028T181220Z-4c60d487-cas-acquire/result.json`：70 个顶层 Test/fuzz seed 目标、245 个 pass 事件，零 fail/skip。包含本机只读公共证书加载、假拓扑、绑定负例/文件锁、假子进程确认门禁、vet、新快照构建/整链 self-test 与构建审计。
- Linux 最终回执 `build/dev-runs/linux/20261028T181239Z-2578efc6-cas-acquire/result.json`：54 个顶层目标、188 个 pass 事件，零 fail/skip；race、vet、self-test 通过，两个有界 fuzz 分别执行 114048 / 116650 次。Windows 专属门禁未编入 Linux，不算 Linux 已验证原生门禁；未重复全仓/发布矩阵。
- 最终包为上述 Windows 目录下 `ready/`，34 项精确输入形成 source digest `a68cb1b374ac2639fb1baf5659cd11ff32ed7e5ee0e6d333be3482a733fa4db6`；可信启动器 `cas-acquire-launch.exe` SHA-256 为 `1dbe8279595a6fdda23f3acc0f238011a5d4e32b8c8ef1bfa55fcf695a87d862`。payload 摘要与独立启动器内置绑定、源码快照及当前输入均一致；五项实际构建审计通过且原 payload 未变。本地来源验证不声称线上 attestation。
- 维护者可在原生 PowerShell 执行该启动器 `--run-once`，它先验证/锁定包并零网络自测，再显示单次窗口范围与确认词；无确认或非交互输入不启动采集。只接受已确认的 D1/D2；实际拓扑/链完整性/页面适配仍可能 blocked，不自动改网卡、Cookie 或采集子集。首次真实输出须返回脱敏 JSON 后再判断，不自动重试。
- 产品认证、模块、历史 R2/Candidate 以及旧 handoff/status 保持原样；当前分支/HEAD 未变，所有开发变化未提交/推送/发布。使用关系仍 unknown，[ADR-0016](../docs/architecture/decisions/ADR-0016-cas-provenance-before-key-selection.md) 仍 proposed，正式 password_core 仍 blocked/live_unverified。汇总及原生操作步骤见 `build/desktop-dev-20261028/CAS-ACQUIRE-REPORT.md`。

### 2026-10-29 首次原生执行与路径门禁修订启动

- 维护者返回已绑定 source a68cb1b… 的首次原生单次执行回执：completed、collection=blocked、reason=topology_rejected，GET/DNS/connect 尝试全部为零；这不是校园网登录或来源采集成功。原包及其回执保留。
- 本轮工具上下文只读观察：1 个活动有线硬件接口、0 个活动无线硬件接口、9 个活动虚拟接口，唯一 IPv4 默认路由在有线接口，半默认路由为 0；WSL/Docker 发行版均已停止。该观察不冒称与维护者执行瞬间完全相同。
- 独立复现发现最小环境与继承环境的 EncodedCommand 都产生 PowerShell Progress CLIXML stderr；旧代码会将该输出当成失败。门禁本身也会因活动虚拟接口拒绝；不能只修 stderr 后声称拓扑已通过。
- 维护者明确批准改为指定有线接口、socket 接口绑定与逐目标路由核验，保留虚拟接口。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 更新诊断契约、合成验收和独立新包；不自动重试校园网请求，不改路由/网卡/服务。

### 2026-10-29 指定有线路径门禁修订完成

- 已按维护者选择修订 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)：保留不影响目标的虚拟接口/路由；固定有线 index/LUID 与源地址，DNS/TCP socket 设置并读回接口绑定，逐目标拒绝其他接口同等或更具体的匹配路由。接口状态/地址归属变化即停；没有修改任何系统网卡或路由。
- 同时修复 PowerShell Progress stderr 误判。回归通过相同最小环境与 EncodedCommand 协议重现旧失败，验证抑制进度后通过、真正 stderr 仍拒绝；不以忽略 stderr 的方式绕过。
- 新增只读 `--preflight`，不触发真实运行确认或域名解析/连接/发送。工具上下文通过原生 PowerShell 5.1 执行最终包预检，binding=verified、preflight_completed、collection=ready，live_attempted/live_started=false，所有操作计数零。该结果不代替维护者原生上下文核验或真实 DNS/CAS 可达性。
- Windows 最终回执 `build/dev-runs/windows/20261028T184734Z-88491630-cas-acquire/result.json`：75 个顶层目标，261 个 pass 事件，0 fail/skip；包含真实 Windows loopback UDP/TCP 接口绑定、失效接口发送前拒绝、整链自测与构建负例。Linux 回执 `build/dev-runs/linux/20261028T184752Z-87db0426-cas-acquire/result.json`：57 个顶层目标，198 个 pass 事件，0 fail/skip；race/vet 通过，两个有界 fuzz 分别执行 100922 / 8264 次。
- 最终独立包为该 Windows 目录的 `ready/`，37 项源码输入，source digest `14c88f294720880e0296e5336ebc3231787ecf18b3cbaab13ed2a11074b7abed`；可信启动器 SHA-256 `1561da842915087d8b0ffcb20bd85b0255a3ffe01726a6bdb2bc679817149ca8`。旧包及维护者首次零网络失败回执保留，不覆盖或冒称已修复旧包。
- 下一步仅请维护者运行新包 `--preflight` 并返回脱敏 JSON；真实采集重试尚未批准或执行。Linux 回归使用既有 Ubuntu WSL，没有启动 Docker、安装工具、修改产品认证/公共契约或创建 Candidate；分支/HEAD 未变，未提交/推送。正式 password_core 仍 blocked/live_unverified。详情见 `build/desktop-dev-20261028/CAS-ACQUIRE-PATH-REPORT.md`。

### 2026-10-29 维护者原生预检回执核验

- 本窗口接收维护者原生 PowerShell 的 `--preflight` 输出并仅核验/记录回执；沿用 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 第五阶段，不自动运行真实采集。
- 返回的 source digest 与指定修订包一致，binding=verified、state=preflight_completed、collection=ready、reason=none，live_started/live_attempted=false，attempts/operations 全零。记录为 `build/desktop-dev-20261028/cas-acquire-native-preflight-result.json`，来源明确为维护者返回的 JSON，不冒称 Agent 独立执行。
- 启动器当前 SHA-256 与既有构建记录一致。沙盒账户 CodexSandboxOffline 的只读 Git 检查触发 dubious ownership 后，经正式工具审批在 Unbal 上下文完成核验：分支 codex/cas-submit-field-closure、HEAD b4eeb760 未变，已有工作区改动保留；没有修改 owner/ACL/safe.directory。
- 原生本地前置条件现有通过回执；真实 DNS、CAS 目标路径及来源采集仍未由本次预检验证。下一步可交付同一包的单次交互运行命令，维护者须按规范当次确认窗口并输入确认词；本回执不是重试授权，不自动重试。只追加执行记录，未改变诊断代码、旧快照、Candidate 或正式 password_core 状态。

### 2026-10-29 匿名发现失败回执与离线分类修订启动

- 维护者返回 source 14c88f29… 的单次匿名执行：binding=verified、state=completed、collection=blocked、reason=discovery_failed、context_constructed=false，GET/DNS/connect 各 2；profile/credentials/CAS POST/activation/logout 全零。这是维护者返回的脱敏 JSON，不是正式登录 evidence。
- 静态核对该精确构建的规划路径：允许一次同主机无 query 跳转后，在第二份发现响应处理中停止；未请求 CAS 页面。旧原因合并多条分支，无法确定实际 ac_id 缺失、非法/冲突或跳转上限中的哪一项，不能补造响应内容或选择硬编码 ac_id。
- 分支 codex/cas-submit-field-closure、HEAD b4eeb760 不变，已有改动全部保留。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 先细分封闭失败类别、再做离线回归与独立新包；不扩大采集范围、不自动重试校园网、不改产品认证或历史包。

### 2026-10-29 匿名发现分类修订完成

- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 细分六个固定原因，仅改变诊断失败分类，不改变发现接受边界、请求预算、两步上限、产品选键或认证行为。旧 discovery_failed 回执保留原义；本次真实失败的具体子原因仍未知。
- Windows 回执 `build/dev-runs/windows/20261028T190548Z-cdb4415d-cas-acquire/result.json`：80 个顶层目标、294 个通过事件；Linux 回执 `build/dev-runs/linux/20261028T190615Z-019b3b2f-cas-acquire/result.json`：62 个顶层目标、231 个通过事件。测试用例无失败/跳过；两端 payload main 包无独立测试文件，其包级 skip 不算用例通过，入口由 self-test/构建整链覆盖。Linux race/vet 与有界 fuzz（44901/37673 次）通过。
- 新回归覆盖六类失败的一/两步内存分类及第二步 loopback 真实传输停止；证明没有第三次请求、Cookie 回传或响应内容输出。原成功/拒绝矩阵通过。文档校验与构建五项正负审计通过；没有重跑全仓发布矩阵。
- 新包位于上述 Windows 目录 `ready/`，source digest `843d532cc10a531425d8281c1b4c65e8ec3874786534e2722e9bf3a8d6698835`，启动器 SHA-256 `a18ca362e47645e4d333d0f3e9a28de3fc9120bcecd27eef721e9e65730b6256`。工具上下文 PowerShell 5.1 本地 preflight=ready、binding=verified、live 标志 false、计数全零；没有执行新包真实采集。
- 维护者上次网络回执记录为 `build/desktop-dev-20261028/cas-acquire-native-discovery-result.json`。下一步由维护者在同一约束的独占窗口，使用新包 --run-once 并当次确认；没有自动重试许可。汇总见 `build/desktop-dev-20261028/CAS-ACQUIRE-DISCOVERY-REPORT.md`。旧包、Candidate、历史 handoff/status、产品与模块保持原样，未提交/推送，password_core 仍 blocked/live_unverified。
- 脱敏扫描首次命中新测试内自造的合成 ticket 值；改用仓库既有 synthetic-canary 后，仅测试夹具变化，产品/包输入未变。补跑 `build/dev-runs/windows/20261028T191108Z-464d318a-cas-collect/result.json` 的集合测试/vet/self-test 通过；保留首轮扫描日志，不改扫描规则。一次 WSL 调用出现 systemd 用户会话警告，gitleaks 已实际执行，未修改服务配置。

### 2026-10-29 query 失败回执与分类补充启动

- 维护者返回 source 843d532c… 的已确认单次执行：binding=verified、completed、blocked/discovery_query_rejected、context_constructed=false，requests/dns/connects 各 2，其余操作零。第二份发现响应的受支持同主机 Location 包含 ac_id，但重复/非法值/额外字段三者尚不能区分；不能将本回执重标为额外字段导致。
- 只读对照发现 SDK 与诊断的 query 键集合要求不同。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 补充仅分类的三个互斥停止原因，先完成离线验证及独立新包，不修改接受语义，不重试网络。分支/HEAD 保持既有基线，所有工作区改动保留。

### 2026-10-29 query 分类补充完成

- 固定按 duplicate → invalid_value → extra_fields 分类；只有最后一类证明 ac_id 唯一且合法。对应细节和未来接受语义决策边界引用 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)。grilling 只读复核亦确认产品/诊断差异，不将 D1 自动解释为删除诊断拒绝规则的授权。
- Windows `build/dev-runs/windows/20261028T191755Z-be7e920f-cas-acquire/result.json`：80 个顶层目标、312 个 pass 事件；Linux `build/dev-runs/linux/20261028T191840Z-f012b193-cas-acquire/result.json`：62 个顶层目标、249 个 pass 事件，测试用例零 fail/skip；payload main 无独立测试文件的包级 skip 单列。两端 vet/self-test、Linux race 和有界 fuzz（63837/48220 次）通过。新用例覆盖编码重复键、复合错误优先级、空键及非 ASCII 数字，loopback 仍在第二次响应后停止。
- 新包为上述 Windows 目录 `ready/`，source digest `6dfefd43147ec5061816499bd2121e34d594138c1b5f3a507ef98407b563e166`，启动器 SHA-256 `f9adac79c73ccaa9ff7cbf8e0d53307bf7f8101d98cb3a9f3183c90165fdcbfd`，37 项构建输入，整链 self-test/五项绑定审计通过。未运行新包校园网入口。
- 维护者回执原样字段记录于 `build/desktop-dev-20261028/cas-acquire-native-query-result.json`，具体 query 子原因仍未知。后续仅交付新包单次交互命令，真实执行须维护者当次确认；旧包和历史回执均保留。最终来源/扫描及说明见 `build/desktop-dev-20261028/cas-acquisition-query-summary.json` 与 `CAS-ACQUIRE-QUERY-REPORT.md`。本次未修改产品认证或正式里程碑，未提交/推送，password_core 仍 blocked/live_unverified。

### 2026-10-29 额外 query 已确认：D3 决策准备

- 维护者返回 source 6dfefd43… 的单次执行回执：verified / completed / blocked / discovery_query_extra_fields，context_constructed=false，GET/DNS/connect 各 2，其余操作零。按对应构建的检查顺序，这次已证明 ac_id 唯一且格式合法，另外存在 query 键；未知其名称、内容和语义。此前回执保留原分类，不追溯改写。
- 回执保存为 `build/desktop-dev-20261028/cas-acquire-native-extra-fields-result.json`，明确标记维护者回传。尚未进入 CAS 或完成登录；不重复请求同一诊断包。
- 本窗口只准备 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 的 D3 待确认提案：从受限 Location 只取合法 ac_id，其他参数丢弃但明确认证/会话标记仍停止。该提案改变诊断接受规则，须 grilling 决策确认，未实施或授权真实重试；产品行为及所有现有产物保持原样。
- 准备完成：D3 已明确键名语法/保留名碰撞、固定认证标记、未知值仅丢弃的限制及合成验收矩阵。只读复核与 doccheck 通过（`build/dev-runs/windows/20261028T193859Z-9b19f1b8-docs/result.json`），分支/HEAD 和已执行启动器 hash 与交付一致。当前等待维护者选择；未修改诊断 Go 代码、未运行 Go 测试或生成新包，未再请求校园网。额外字段具体语义仍未知，password_core 仍 blocked/live_unverified。

### 2026-10-29 D3 离线实施启动

- 维护者明确同意开始；按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 已确认 D3 实现仅消费 ac_id、普通额外字段丢弃、异常键及认证标记停止。依次完成合成/loopback、两端回归和独立包；不自动执行真实重试，不修改公共产品认证或旧包。

### 2026-10-29 D3 实施与 DNS 加固完成

- D3 已实现：独立诊断仅提取唯一合法 ac_id，未知普通 query 不透传/不输出，固定认证标记及异常键名拒绝；内置 self-test 也覆盖正反路径。边界及确定性失败顺序引用 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)，产品 SDK/CLI/认证实现未改。
- 首轮 Linux `build/dev-runs/linux/20261028T194443Z-22d65247-cas-acquire` 的 FuzzDNSPacket 在 16291 次执行后发现当前 dnsmessage SVCB 越界 panic；该轮 failed，后续步骤未执行，不能计作通过。生成样本 `internal/casacquire/testdata/fuzz/FuzzDNSPacket/a92e9bef76c5ada7`、失败日志和首轮 Windows 中间包均保留，不将中间包交付为最终版本。
- 独立诊断 DNS 改为先核验 header/question/长度，再仅解释 A/CNAME，其他资源体只核验 framing 后跳过，避免调用无关类型解析器；不是捕获所有 panic 后伪装通过，不更新依赖、不改产品 DNS。增加该样本及各 section 的 SVCB/HTTPS opaque framing 回归，畸形长度仍拒绝。
- 修复后 Windows 最终回执 `build/dev-runs/windows/20261028T194900Z-e8fb5c9a-cas-acquire/result.json`：86 个顶层目标、558 个 pass 事件；Linux `build/dev-runs/linux/20261028T194928Z-e5b0f49e-cas-acquire/result.json`：68 个顶层目标、495 个 pass 事件。测试用例零 fail/skip；无独立测试文件的 payload main 包级 skip 单列。两端 vet/self-test、Linux race 通过；最终 DNS/report 有界 fuzz 为 73460/7524 次，不声称穷尽。
- 最终包为上述 Windows 目录 `ready/`，38 项构建输入，source digest `ae1f513d06653ba8d9fa2d8a8b733574d22bf29cbb2e655890d3db15a4b7c20e`，启动器 SHA-256 `ac22cd95a45291030268f16625a6bc645f4de5f68866354bfa51d62c0eba45d1`。整链自测和五项绑定审计通过；工具上下文 PowerShell 5.1 preflight=ready / verified，live 标志 false，操作计数零。
- 下一步仅维护者按原生有线独占窗口单次确认执行最终包，Agent 未再请求校园网。具体额外字段未知，仍可能触发 D3 阻断或后续协议边界；password_core 保持 blocked/live_unverified。最终汇总及旧包/来源/定向扫描见 `build/desktop-dev-20261028/cas-acquisition-d3-summary.json`、`CAS-ACQUIRE-D3-REPORT.md`。未提交/推送/发布或创建 Candidate；旧 handoff/status 原字节保留。

### 2026-10-29 CAS 上下文已构造：传输分类启动

- 维护者返回 source ae1f513d… 单次回执：verified / completed / blocked / request_failed，context_constructed=true，GET/DNS/connect 尝试各 3，其余操作零。首次证明本轮 D3 已通过匿名发现并构造 CAS 上下文；不证明 CAS GET 已写入或服务器已返回响应，不是登录成功。
- 当前 request_failed 合并 deadline 设置、TLS 握手/证书校验、HTTP 写入或 body 读取等路径，不能由耗时推断根因。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 增加封闭的阶段/typed-error 分类及离线验收；不自动重试、不改变 TLS/证书来源/预算、不升级依赖或改产品认证。

### 2026-10-29 传输分类修订完成

- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 增加 22 个固定原因及严格回执验证，保留旧 request_failed 解码。合成证书证明拒绝发生在 GET 前；阶段故障和报告闭集回归通过。没有放宽 TLS、改证书来源或调整请求预算。
- Windows `build/dev-runs/windows/20261028T200136Z-a38d5c7f-cas-acquire/result.json`：91 个顶层目标、592 个 pass 事件；Linux `build/dev-runs/linux/20261028T200140Z-ae89d588-cas-acquire/result.json`：73 个顶层目标、529 个 pass 事件。用例零 fail/skip，无测试文件的 payload main 包级 skip 单列；vet/self-test、Linux race、DNS/report 有界 fuzz（70158/21679）及五项绑定审计通过。
- 新包为上述 Windows 目录 `ready/`，39 个构建输入，source digest bfeca3ece21aa4bb39e2c1a0aa90761db9140cd6da5ef52b8cce798c7d5a7e94，启动器 SHA-256 215294a5c348303eaca6109919c94f5cdd59a99551d7c946d9cd4f4bce2de693。工具上下文 PowerShell 5.1 local preflight 为 verified/ready、网络/操作计数零；未执行新包真实采集。
- 维护者回执及解释见 `build/desktop-dev-20261028/cas-acquire-native-request-result.json`；修订报告 `CAS-ACQUIRE-TRANSPORT-REPORT.md`，摘要/定向脱敏扫描 `cas-acquisition-transport-summary.json`。下一步仅维护者在原生独占窗口当次确认单次运行并回传 JSON；旧失败具体阶段仍未知，password_core 仍 blocked/live_unverified。
- 分支/HEAD 不变；产品 auth/client/internal/cas/internal/transport、go.mod/go.sum、旧 handoff/status 未改，旧包保留。无提交/推送/发布/Candidate，也未向 BHKLAPTOP 取件。

### 2026-10-29 TLS unknown authority 回执核验启动

- 维护者返回 source bfeca3ec… 的单次回执：verified/completed，context_constructed=true，blocked/tls_unknown_authority，requests/dns/connects 各 3，其余操作零。包摘要与交付一致；按固定 Go 1.25.0 的 TLS 调用顺序，第三次尝试在证书链验证阶段、CAS GET 写入前停止，不能认定缺根、缺中间证书或服务器配置错误。
- 原回执另存 `build/desktop-dev-20261028/cas-acquire-native-tls-result.json`，明确 maintainer_returned_json。只读核对 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 与实现：现有 ROOT/Disallowed 离线来源不含 CA 中间证书缓存，系统自动补链未开启。
- 继续准备失败后的本地链对照方案；改变证书读取来源需先明确决策，当前未实施、未运行新网络/安装/证书操作，分支/HEAD 和历史包保留。

### 2026-10-29 TLS 回执复核完成、D4 待决策

- 已核对固定 Go 1.25.0 源码：TLS 自动验证只接收 peer intermediates，正常验证失败时两个验证回调均不会运行。新增 CA 缓存不能通过普通回调修复自动验证；也不能把 CA 加进 RootCAs 来绕过失败。
- D4 具体待确认方案及离线验收已写入 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)：原失败后关闭连接，使用同一根/拒绝快照作只读本地中间证书对照，任何结果都保持 blocked。已通过 grilling 提交这一读取范围决策；尚未收到确认，未实施 Go 代码或构建新包。
- 本轮复核报告 `build/desktop-dev-20261028/CAS-ACQUIRE-TLS-RECEIPT-REVIEW.md`；原 TLS 失败仍未证明缺根/缺中间证书或服务器配置错误。未重试真实网络、读取凭据、修改信任库或放宽校验，password_core 保持 blocked/live_unverified。

### 2026-10-29 D4 离线实施启动

- 维护者明确选择“同意推荐：只做失败后的离线对照”。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 实现只读 CA 中间证书对照、保持原 TLS 失败与 blocked，完成合成/loopback、两端回归及独立新包。真实复测仍由维护者当次确认，Agent 不执行。
- 开始时分支 codex/cas-submit-field-closure、HEAD b4eeb760 未变，已有改动及旧包全部保留，不修改产品认证、Candidate、证书存储、Git refs 或权限。

### 2026-10-29 D4 离线实施完成

- 已按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 完成 D4：原 TLS unknown-authority 后先关闭 socket，再只读 CA 作本地中间链对照；结果始终 blocked，根池不变，不继续 GET。原自动 TLS 校验及证书信任来源未放宽。
- 最终 Windows 回执 `build/dev-runs/windows/20261028T203812Z-e8a1cb60-cas-acquire/result.json`：98 个顶层目标、631 个 pass 事件；Linux `build/dev-runs/linux/20261028T203816Z-75238641-cas-acquire/result.json`：78 个顶层目标、561 个 pass 事件。用例零 fail/skip，payload main 无测试文件的包级 skip 单列。vet/self-test、Linux race、DNS/report 有界 fuzz（109588/5866）、五项绑定审计通过；新增 Windows 内存证书库测试通过，没有写入系统证书库。
- 最终包在上述 Windows 目录 `ready/`，42 项源码输入，source digest 8c9c60b9804497ee4a8dc96dc7a6065140c2fcaa5fe0df0e67c1d5669f24d0ab，启动器 SHA-256 e2300889aab8e86f1fbb1af0b4fc040511627120c4713d7df3ad1b20b5d09104。工具上下文 PowerShell5.1 local preflight 为 verified/ready，全零操作；没有执行真实 D4 对照，也不证明本地缓存能解决真实 CAS 链。
- 汇总见 `build/desktop-dev-20261028/CAS-ACQUIRE-D4-REPORT.md` 与 `cas-acquisition-d4-summary.json`；首轮开发产物保留，交付上述最终包。下一步只由维护者原生终端当次确认单次复测；仍不得将可补链分类当作 TLS/登录成功。
- 产品 auth/client/internal/cas/internal/transport、go.mod/go.sum、旧 handoff/status 未改。没有真实请求、凭据读取、证书安装、提交/推送/发布/Candidate；password_core 仍 blocked/live_unverified。

### 2026-10-29 D4 unavailable 回执与阶段分类启动

- 维护者返回 source 8c9c60b9… 的已确认单次回执：verified/completed、context_constructed=true、blocked/tls_chain_probe_unavailable，requests/dns/connects 各 3，其余操作零。仍在 CAS GET 前停止；不足以判断缓存可补链或缺根，不能由约 9 秒耗时认定期限原因。
- 当前分支/HEAD 与包摘要匹配交付，已有改动保留。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 仅补齐 D4 内部失败阶段分类、合成回归与新包；不扩大来源/上限、不改 TLS 或产品行为、不自动重试。

### 2026-10-29 D4 不可用阶段分类完成

- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 增加13个固定原因，原 unavailable 保留历史解码；输入、基线复验、库打开/枚举、缓存解析/限额、同名候选及取消/期限可分别停止。未放宽任何对照接受规则或资源上限。
- Windows `build/dev-runs/windows/20261028T204606Z-744f52d9-cas-acquire/result.json`：100个顶层目标、648个pass事件；Linux `build/dev-runs/linux/20261028T204610Z-c9af7e0a-cas-acquire/result.json`：80个顶层目标、578个pass事件。用例零fail/skip，payload main无测试文件的包级skip单列；vet/self-test、Linux race、DNS/report有限fuzz（114881/9163）及五项绑定审计通过。
- 最终包为上述Windows目录的ready，source digest 5c2de5a2818f08abaa3cbade50a05c743424414fcf5997be4b2e71b1673e9ac4，launcher SHA-256 24fe38f17e3c76a07feeff41e93197bad938dfea98da895a5a58fa9bb5611aac，42项构建输入。工具上下文PowerShell5.1 preflight为verified/ready、全零操作，未执行真实重试。
- 原回执另存 `build/desktop-dev-20261028/cas-acquire-native-d4-unavailable-result.json`；汇总为 `CAS-ACQUIRE-D4-STAGES-REPORT.md`、`cas-acquisition-d4-stages-summary.json`。旧不可用具体原因仍未知，下一步由维护者当次确认单次复测，不自动连续运行。
- 产品认证/依赖、旧handoff/status、旧包和Candidate未改，无证书安装、提交/推送/发布。password_core保持blocked/live_unverified。

### 2026-10-29 本地 CA cache_invalid 回执核验启动

- 维护者返回 source 5c2de5a2… 的 verified/completed、context_constructed=true、blocked/tls_chain_probe_cache_invalid，requests/dns/connects各3，其他操作零。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)，本地 CA 条目结构/DER 解析失败导致对照提前停止；不代表 CAS 证书无效或证明某个解析原因。
- 原回执另存 `build/desktop-dev-20261028/cas-acquire-native-cache-invalid-result.json`。分支/HEAD和已执行包摘要一致，保留原产物；只做源码核对和合成复现，准备改变缓存失败策略前的具体决策，不直接跳过无效条目、不删除或安装证书、不重试校园网。

### 2026-10-29 缓存解析合成复现完成、D5 待确认

- 新增 Windows 内存证书库单测试，证明 Windows 可枚举 Go 默认解析器拒绝的生成 DER；固定 Go1.25、GODEBUG 为空，回执 `build/desktop-dev-20261028/cache-parser-repro-20261028T205304Z-71524bac/result.json` exit0。只证明机制，不判定现场负序列号或实际证书损坏；无真实证书/网络读取。
- [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 已准备 D5：仅有界 DER 解析失败可被排除并标部分缓存，同时新增零网络缓存检查入口。结构错误/库错误/上限/取消仍停止，ROOT 不变、永远 blocked。两项均改变现行约定，待维护者明确决定后再实施。
- 报告 `build/desktop-dev-20261028/CAS-ACQUIRE-CACHE-REVIEW.md`；本轮未改生产 Go 代码或重建包，不要求继续运行校园命令，无证书修改/发布/提交。password_core 保持 blocked/live_unverified。

### 2026-10-29 D5 离线实施启动

- 维护者明确选择“同意推荐：部分缓存＋零网络检查”。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 实施可解析缓存子集、绑定启动器本地检查、合成回归和独立开发包；先检查本地缓存，再决定下一步，不自动执行校园网。
- 分支 codex/cas-submit-field-closure、HEAD b4eeb760 及既有改动保留。范围仅诊断工具、其测试/规范及当前交接，不修改产品认证、旧历史快照、证书库、Candidate 或旧包。

### 2026-10-29 D5 实施完成、本地检查返回部分缓存

- 已按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 实现 D5。Windows 最终回执 `build/dev-runs/windows/20261028T210921Z-66a738b9-cas-acquire/result.json`：104 个顶层测试、680 个 pass 事件；Linux `build/dev-runs/linux/20261028T210923Z-425b6242-cas-acquire/result.json`：83 个顶层测试、609 个 pass 事件。用例零 fail/skip，payload main 包级 skip 单列。两端 vet/self-test、Linux race、有界 DNS/report fuzz（59397/853）和九项绑定审计通过。
- 最终开发包在上述 Windows 目录 ready，source digest f516fdd233da212bc1591eb25bfb29aa43146c553214352b687d824c406cd9d9，launcher SHA-256 04e3403cd3cb89858714d8dbc87c2d12ebd4eeb5f014ae7ee000cab54a3dbc4c，43 项构建输入。旧 204606 包保留，身份不变；产品认证、依赖、旧 handoff/status 未改。
- 工具正式审批账户 bhkdesktop\unbal 运行本地 --cache-check，回执 `ready/tool-local-cache-check.json` 为 verified/cache_check_completed、blocked/cache_partial，live 标志 false，所有尝试/操作计数零。只证明该上下文缓存部分可解析，不识别具体条目，也不证明真实 CAS 可补链。
- 报告 `build/desktop-dev-20261028/CAS-ACQUIRE-D5-REPORT.md`，绑定与定向脱敏扫描摘要 `cas-acquisition-d5-summary.json`。下一步只让维护者原生终端执行该包 --cache-check 并返回 JSON，再判断是否需要新的校园窗口；不自动重试。未读凭据、修改证书、安装、提交/推送/发布或创建 Candidate，password_core 仍 blocked/live_unverified。

### 2026-10-29 D5 原生本地检查回执确认

- 维护者返回同一 source f516fdd2… 的原生 --cache-check 回执：verified/cache_check_completed、blocked/cache_partial，live 标志 false，全部尝试/操作计数零。原件按 maintainer_returned_json 保存于 `build/desktop-dev-20261028/cas-acquire-native-d5-cache-check.json`；分类与工具上下文一致，不证明两次缓存内容相同或真实 CAS 可补链。
- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)，本地前置检查已完成，无需改代码或重新构建。下一步可交付同一包的单次匿名对照，由维护者重新确认窗口并在终端授权；只判断可解析子集是否改变离线补链分类，对照仍 blocked，不恢复失败的 TLS 或继续该 GET。Agent 不执行真实请求，不沿用旧窗口授权，password_core 保持 blocked/live_unverified。

### 2026-10-29 D5 部分缓存对照未解决：回执核验与分类启动

- 维护者返回 source f516fdd2… 的 verified/completed、context_constructed=true、blocked/tls_chain_partial_cache_unresolved，requests/dns/connects 各3，其余操作零。原回执另存 `build/desktop-dev-20261028/cas-acquire-native-d5-unresolved-result.json`。已进入第二次 Verify；原 TLS 已失败且 socket 已关闭，CAS GET 未发送。不能由该合并类别判定缺根、具体缓存条目或服务器证书错误。
- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 仅细分第二次 Verify 的公开错误类型，补离线回归并交付独立新包；不改变证书来源/读取时机/信任规则或网络预算，不重试真实请求，不重标旧回执。

### 2026-10-29 D5 失败类型分类完成

- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 完成第二次 Verify 的14类完整/部分缓存固定失败分类；旧 unresolved 保留解码。公开类型/枚举之外不推测原因，不输出证书或错误原文，不改变任何连接接受规则或预算。
- Windows `build/dev-runs/windows/20261028T212040Z-f7c5e531-cas-acquire/result.json`：107个顶层测试、707个pass事件；Linux `build/dev-runs/linux/20261028T212041Z-578e2a53-cas-acquire/result.json`：86个顶层测试、636个pass事件。用例零fail/skip，payload main包级skip单列；两端vet/self-test、Linux race、有限fuzz（60724/108253）及九项绑定审计通过。
- 最终包为上述 Windows 目录 ready，source digest 444b33da4ddf12841ad9f04df2abb0cca3ee9f5b11f8a026fccb0bed36f470bb，launcher SHA-256 a9af475fbad9a49f1522c5913418c1dd9a6c63951c00658703547a6adbb816e1，44项构建输入。新包工具本地 --cache-check 为 verified/cache_partial，全零操作；未执行新包真实请求。原 f516fdd2… 包保留。
- 报告 `build/desktop-dev-20261028/CAS-ACQUIRE-D5-TYPES-REPORT.md`，审计汇总 `cas-acquisition-d5-types-summary.json`。下一步只由维护者重新确认窗口、单次运行新包并回传固定 JSON；Agent 不自动重试。产品认证、依赖、历史快照未改，无证书写入、提交/推送/发布/Candidate；password_core仍blocked/live_unverified。

### 2026-10-29 部分缓存仍为 UnknownAuthority：核验启动

- 维护者返回 source 444b33da… 的 verified/completed、context_constructed=true、blocked/tls_chain_partial_cache_unknown_authority，请求/DNS/连接尝试各3，其余操作零。原件另存 `build/desktop-dev-20261028/cas-acquire-native-d5-authority-result.json`，标记 maintainer_returned_json。来源及新旧包哈希复核一致，分支/HEAD 不变。
- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)，这明确第二次 Verify 仍为公开 UnknownAuthority，但不能据此判断缺根、缺中间证书或被隐藏的候选失败；CAS GET 未发送。现有分类已到信息边界，不要求重复执行原包；先只读核对下一阶段方案，需要改变诊断约定时再由维护者决定，不自动联网或扩大信任来源。

### 2026-10-29 UnknownAuthority 核验完成、D6 待决定

- 固定Go源码核对支持同源关系观察有信息增益，但CheckSignatureFrom不是完整链验证；Windows逻辑库/系统链引擎会引入其他来源与行为，当前不切换。事实复核见 `build/desktop-dev-20261028/CAS-ACQUIRE-AUTHORITY-REVIEW.md`。
- [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 已写入D6待确认提案：复用原证书集合、额外有界根快照与签名关系观察，保持blocked、无新网络/信任源。新增计算与输出需维护者决定；本轮未实现Go代码、运行真实诊断或构建新包，不要求重复旧包。旧包/历史回执保持，password_core仍blocked/live_unverified。

### 2026-10-29 D6 离线实施启动

- 维护者明确选择“同意推荐：同源、有界关系观察”。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 完成原根快照、关系观察、闭集报告、合成验收及独立新包；真实运行仍由维护者当次确认，Agent不执行。
- 实施前分支codex/cas-submit-field-closure、HEAD b4eeb760未变；保留既有改动和所有旧包，不改产品认证、原信任来源或Git refs。

### 2026-10-29 D6 离线实施完成、待维护者单次运行

- 已按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 完成 D6、闭集兼容及有界合成验收。Windows `build/dev-runs/windows/20261028T215647Z-663075fc-cas-acquire/result.json`：116个顶层测试、750个pass事件；Linux `build/dev-runs/linux/20261028T215053Z-41adbbb9-cas-acquire/result.json`：95个顶层测试、679个pass事件。用例零fail/skip，payload main包级skip单列；两端vet/self-test、Linux race、有界fuzz（71053/184）、九项绑定审计和doccheck通过。Windows最终回归还覆盖启动器传递纯合成D6回执，无真实请求。
- 新包位于上述Windows目录ready，46项构建输入；source digest f08f01d11ce5410752bf0898853db6b08312de62874c618acd375eb146541aa4，launcher SHA-256 343cd059098b3e78bdb86c94087157ce110ebbbc5c4e0942574690266fd3a937。当前源码与快照逐项一致，原444b33da…包保留不变。
- 工具正式审批账户bhkdesktop\unbal执行新包本地 --cache-check：verified/cache_check_completed、blocked/cache_partial，live标志false，全部尝试/操作零。无新包校园执行证据；本地检查不能推断peer关系。
- 报告 `build/desktop-dev-20261028/CAS-ACQUIRE-D6-REPORT.md`，审计及定向脱敏扫描汇总 `cas-acquisition-d6-summary.json`。下一步仅交付该包，由维护者重新确认原生有线独占、TUN关闭、账号离线窗口并在终端单次授权。Agent不自动运行或重试。产品认证、依赖、历史快照未改，无证书写入、提交/推送/发布或Candidate；password_core仍blocked/live_unverified。

### 2026-10-29 D6 循环观察回执核验启动

- 维护者返回 source f08f01d1… 的 verified/completed、context_constructed=true、blocked/tls_chain_relation_observed；关系complete/partial，仅cycle_observed=true，其余七项false。请求/DNS/TCP尝试各3，其余操作零。原件标记maintainer_returned_json另存 `build/desktop-dev-20261028/cas-acquire-native-d6-cycle-result.json`。
- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 核对当前分类解释与合成反例；不把循环等同服务器配置错误、不安装根、不重复运行校园命令。分支/HEAD不变，保留旧包和既有改动。

### 2026-10-29 D6 循环解释核验完成、D7待决定

- 原包三项身份哈希与46项源码/快照一致。纯生成证书反例证明：不在根池中的自签名链尾与多证书循环都可产生本次八项事实。聚焦11子测试通过，回执 `build/desktop-dev-20261028/d6-cycle-review-20261028T221018Z-09f5c5f4/result.json`；无真实证书读取/网络。仅增加测试，不改生产Go代码或构建新包。
- 报告 `build/desktop-dev-20261028/CAS-ACQUIRE-D6-CYCLE-REVIEW.md`；[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 已加入D6解释边界及D7待确认提案。下一决定为一次性收敛末端类别、成员来源与原根表示关系；不得仅由循环输出决定安装根或放宽TLS，不要求重复原包。实施及真实执行授权分别处理，password_core仍blocked/live_unverified。

### 2026-10-29 D7 离线实施启动

- 维护者明确选择“同意推荐：一次补齐同源关系分类”。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 实施、合成验证并交付独立开发包；不新增证书来源/验签/Verify/网络或改变连接接受规则。
- 分支codex/cas-submit-field-closure、HEAD b4eeb760未变，保留已有修改及全部旧包；真实运行仍由维护者当次确认，Agent不执行。

### 2026-10-29 D7 离线完成、独立新包待原生运行

- 已按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 完成D7及旧报告兼容。Windows `build/dev-runs/windows/20261028T222217Z-3cb04d2d-cas-acquire/result.json`：120个顶层测试、779个pass事件；Linux `build/dev-runs/linux/20261028T222217Z-fdc7b567-cas-acquire/result.json`：99个顶层测试、708个pass事件。内部诊断套件用例零fail/skip，payload main包级skip单列；vet/self-test、Linux race、有界fuzz（106472/15442）、九项绑定审计和doccheck通过。
- 新包为上述Windows目录ready，47项构建输入；source digest fb0c3e658cc3e90d2b1e73f8c2df4123f2a097845a7939e4cb8b330df57b0fbb，launcher SHA-256 d39cf1bd45524ae31b991f58996e1183c533ee72975cfa093342df1690795106。旧D6包保留不变。工具审批账户bhkdesktop\unbal执行新包 --cache-check 返回verified/cache_partial，live标志false，全部尝试/操作零；无D7原生校园运行证据。
- 报告 `build/desktop-dev-20261028/CAS-ACQUIRE-D7-REPORT.md`，绑定及定向脱敏扫描汇总 `cas-acquisition-d7-summary.json`。下一步只交付该包的完整单次运行步骤，由维护者原生终端重新确认窗口并授权；Agent不自动请求或重试。产品认证、依赖、旧handoff/status未改，无证书写入、提交/推送/发布或Candidate；password_core仍blocked/live_unverified。

### 2026-10-29 D7 原生末端观察回执核验启动

- 维护者返回source fb0c3e65…，verified/completed、blocked/tls_chain_topology_observed。D6 complete/partial仅cycle=true；D7 complete仅末端self_loop present、peer_only_member=true，根表示匹配及其他类别均false。尝试3/3/3，其余操作零。原件以maintainer_returned_json保存为 `build/desktop-dev-20261028/cas-acquire-native-d7-terminal-result.json`。
- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 复核解释与信任来源审计方向；不由未验证peer中的自签名对象建立信任，不重复校园诊断或改动证书库。HEAD及既有工作区变更保留。

### 2026-10-29 D7 回执核验完成、转零网络审计提案

- 包哈希与47项输入一致，核验摘要 `build/desktop-dev-20261028/cas-acquire-native-d7-check.json`。报告 `CAS-ACQUIRE-D7-TERMINAL-REVIEW.md`记录可证明的peer-only末端自环及不能外推的Windows/产品信任结论；原peer未保存，无法在后续本地审计中回溯匹配。
- [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 新增D8待确认：仅同一固定注册表ROOT/Disallowed/CA的零网络读取/解析/拒绝覆盖审计，不增加信任来源。下一步等待维护者决定，不重复D7网络窗口。未改Go或新建诊断包，无证书读取/写入、凭据/校园请求或Git refs操作，password_core仍blocked/live_unverified。

### 2026-10-29 D8 离线实施启动

- 维护者明确选择“同意推荐：先做零网络覆盖审计”。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 实施独立本地工具、合成测试、绑定及开发包；不新增信任来源、不改变旧采集命令白名单或产品TLS。
- HEAD b4eeb760及已有修改保留，旧D7与Candidate原件保留；本阶段无校园请求、安装/写证书、凭据、发布或Git refs操作。

### 2026-10-29 D8 离线实施完成、待原生本地审计

- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 完成独立工具、闭集覆盖报告、绑定和开发执行入口。Windows `build/dev-runs/windows/20261028T230122Z-6c26c462-trust-audit/result.json`：50个顶层测试、178个pass事件；Linux `build/dev-runs/linux/20261028T230127Z-1a4e4610-trust-audit/result.json`：31个顶层测试、120个pass事件。零fail/skip、无包级skip；两端vet/self-test、Linux race、5秒有限fuzz（3次执行）和14项绑定检查通过。最初测试语法失败记录保留，修正后重跑通过。
- 新包为上述Windows目录 `ready/trust-audit-launch.exe`，29项构建输入；source digest 9e44de46c7bacd6c22d2111d08b1b485d97d410f29ec3b5b42362a7218cc141d，launcher SHA-256 b0cf4efcfb38324d2a0836e95a44c03ac9426b12131045ab279a030cb010aa0c。当前源码与快照一致，旧D7三项身份摘要不变。执行入口和用途见 [开发说明](../docs/operations/development.md)。
- 报告 `build/desktop-dev-20261028/TRUST-AUDIT-D8-REPORT.md`，汇总 `trust-audit-d8-summary.json`，最终文档/定向扫描回执 `trust-audit-d8-final-checks.json`。所有证书测试使用生成数据或内存库，未读取实际本机证书库；下一步由维护者原生执行一次 `--trust-audit` 并回传JSON，无需改变TUN、账号或网络。此审计不能回溯匹配已释放peer，也不代表TLS恢复。
- 产品认证、模块文件及旧handoff/status未改；无校园请求、实际证书写入、安装、提交/推送/发布或新Candidate。password_core仍blocked/live_unverified。

### 2026-10-29 D8 原生覆盖回执核验启动

- 维护者返回source 9e44de46…，verified/audit_completed、result complete；ROOT完整读取且全部可解析，Disallowed完整读取且空，CA完整读取但解析partial；拒绝重叠none、剩余解析根present，全部操作零。原件标记maintainer_returned_json保存为 `build/desktop-dev-20261028/trust-audit-native-d8-result.json`。
- 维护者说明TUN一直关闭，作为用户声明记录，不宣称Agent重新核验拓扑。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 本地审计不要求切换TUN；本轮不重复审计或校园请求，先核验绑定和解释边界。

### 2026-10-29 D8 原生覆盖核验完成

- 三项包摘要、29项源码/快照及回执精确形状均匹配，摘要 `build/desktop-dev-20261028/trust-audit-native-d8-check.json`；事实、解释边界及下一方向见 `TRUST-AUDIT-D8-NATIVE-REVIEW.md`。当前固定ROOT无读取/解析缺口、固定Disallowed空，CA解析partial；不能关联已释放D7 peer，亦不证明完整Windows信任或产品TLS成功。
- 完成诊断与默认产品TLS的源码差异表及官方API文档审查。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 停止同源重复复测，下一工作包为明确来源及副作用边界的受控对照设计；系统链引擎/额外来源尚未授权，普通SystemCertPool.Verify不能作为零网络检查。未形成可安全执行的新真实窗口，不交付重复校园命令。
- 本轮未改Go代码、运行诊断/Go测试、读取实际证书库、获取peer或创建新包。TUN保持维护者当前状态，无需为D8切换；无凭据、实际证书写入、Git refs或发布操作，password_core仍blocked/live_unverified。

### 2026-10-29 TLS 受控对照设计启动

- 维护者要求继续。分支codex/cas-submit-field-closure、HEAD b4eeb760保持，既有改动保留。本窗口按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 只做源码/官方API契约核对及可审阅设计；D8不包含Windows链引擎、额外来源或新peer授权。
- 依grilling并行核对自定义链引擎的来源/副作用边界与避开链引擎的本地固定来源备选。设计目标是收敛一次有信息增益的后续对照，不重复同源审计或先试运行真实API。

### 2026-10-29 D9 设计完成、待整体范围决定

- 源码/官方契约审查见 `build/desktop-dev-20261028/TLS-D9-DESIGN-REVIEW.md`。仅靠内存库、exclusive根和cache-only参数，无法证明排除默认系统库读取/全部副作用；不在本机试运行Windows链引擎。
- [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 新增D9待确认提案：七个固定额外本地来源、网络前材料门禁、同轮peer的有界内存关系/拒绝观察、离线验收及独立开发包。范围/预算/输出限制集中写在规范中；不复制为agent产品规则。
- 下一步由维护者决定该整体范围，再实施；真实运行仍须新包当次确认。当前没有新运行命令，无Go实现/测试、新来源读取、证书API执行、校园请求或新包。旧产物和password_core blocked/live_unverified保持，TUN不需切换。

### 2026-10-29 D9 离线实施启动

- 维护者明确同意实施D9并完成离线验收、打包。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 锁定闭集输出、七来源门禁和同轮有限关系分析；真实运行授权仍独立。
- 保留既有修改及所有旧包；本轮不读取新增实际来源、不发校园请求、不修改产品认证或Git refs。

### 2026-10-29 D9 离线验收与独立包完成

- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 完成七来源材料门禁、同轮有限比较、严格报告、内建合成自测和 launcher 传递。工作范围及执行入口见 [开发说明](../docs/operations/development.md)，没有更改产品认证契约。
- 最终 Windows `build/dev-runs/windows/20261029T073249Z-6081b6be-cas-acquire/result.json`：136个顶层测试、833个pass事件；Linux `build/dev-runs/linux/20261029T073254Z-cc9c162e-cas-acquire/result.json`：112个顶层测试、754个pass事件。测试fail/skip为零；两端命令包cas-acquire无独立测试文件，包级skip单独记录，并由self-test/构建/launcher链覆盖。两端vet、Linux race、三项有限fuzz和9项绑定检查通过；早期夹具失败日志保留。
- 新包位于上述 Windows 目录 `ready/cas-acquire-launch.exe`，53项快照输入；source digest `79b28b1e0c767c14622eb2055fc46b67e75a272fc61a6686b7aaa28733c7a031`，launcher SHA-256 `901bf8634633bd2b0e2a13c0912213dd8dffb4db5be383c1eb9376f57b60edf6`。旧D7/D8三项身份摘要均保持，产品/模块/历史handoff/status无差异。
- 交付报告 `build/desktop-dev-20261028/CAS-ACQUIRE-D9-REPORT.md`、汇总 `cas-acquisition-d9-summary.json`、文档/定向扫描回执 `cas-acquisition-d9-final-checks.json`。Windows既有根池测试只读原ROOT/Disallowed；D9新增实际来源未读取，测试使用假provider、生成证书和内存库/loopback。
- 下一步等待维护者在原生终端用新包当次确认；不自动执行校园窗口或替用户确认。无校园请求、凭据、实际证书写入、安装、提交/推送/发布或Candidate变化；password_core保持blocked/live_unverified。

### 2026-10-29 D9 原生回执核验启动

- 维护者返回 source 79b28b1e… 的单次运行 JSON；按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 只核对包绑定、来源差异及解释边界，不沿用本次一次性授权重试。回执显示 AuthRoot 有同轮匹配，其他新增策略来源空/缺失；仍 blocked。

### 2026-10-29 D9 原生回执核验完成

- 三项包摘要、53项源码/快照及精确回执字段匹配；保存 `build/desktop-dev-20261028/cas-acquire-native-d9-result.json`，标记maintainer_returned_json；检查回执 `cas-acquire-native-d9-check.json`。维护者报告网络尝试各3、其余禁止操作零；Agent没有重跑。
- AuthRoot可解析子集存在本轮可达DER重叠及末端同名同SPKI对应，两项为独立存在性，不推导同一证书或可信根。其余固定策略来源空/缺失；原ROOT仍无关系路径，CA/AuthRoot解析partial，结果blocked、analysis=null。
- 解释及后续研究缺口见 `CAS-ACQUIRE-D9-NATIVE-REVIEW.md`。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 不重复同类校园窗口；已只读核对官方CTL来源与验签API的信任前提。下一方案必须先固定独立可信签名者依据，不能以缓存存在或API的trusted参数自证信任；新增CTL读取/验证/peer关联尚未授权。
- 本轮只更新交接与脱敏核验材料，无Go实现/测试、新包、实际证书读取/写入、校园请求、凭据、Git refs或发布变化。历史handoff/status保持；password_core仍blocked/live_unverified。

### 2026-10-29 CTL 信任依据离线设计启动

- 维护者要求继续。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 从D9正匹配推进可信依据设计；grilling事实调查分别核对独立签名者锚点、CTL签名/成员/用途/拒绝语义。本窗口只读源码和官方资料，不读取新增本机来源、不下载证书材料、不运行CryptoAPI或校园窗口。

### 2026-10-29 D10 离线设计完成、待整体决定

- 事实审查见 `build/desktop-dev-20261028/TLS-D10-DESIGN-REVIEW.md`；[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 已新增D10待确认提案，先做独立零网络CTL材料及固定签名层级检查，不安排新的校园握手。公开固定身份与数学签名关系不等于完整用途、吊销、最新列表或TLS信任。
- 下一步由维护者整体决定新增来源、计算及独立包范围，再进行离线实现/合成验收/打包；实际本地读取仍交付原生单步入口。不自动执行新来源读取、更新信任或校园窗口。
- 本窗口仅有提案/审查/当前交接变更；无Go代码或测试、新包、实际证书读取/写入、凭据、Git refs及发布变化。旧包和历史handoff/status保留，password_core仍blocked/live_unverified。设计检查回执为 `build/desktop-dev-20261028/tls-d10-design-checks.json`。

### 2026-10-29 D10 离线实施启动

- 维护者明确选择“同意推荐：实施D10、完成离线验收与打包”。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 实施独立本地工具、固定来源和有限签名观察，先锁定输出后实现；不自动读取实际新增来源。
- 开始核验分支codex/cas-submit-field-closure、HEAD b4eeb760保持，既有改动保留。工作包沿WP-CAS-ACQUIRE-DEV内部诊断边界推进；无产品认证/历史快照/发布状态修改。

### 2026-10-29 D10 离线验收与独立包完成

- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 完成独立CTL材料工具、固定签名观察、闭集输出、先保护后读取及绑定启动器；开发入口见 [开发说明](../docs/operations/development.md)。没有产品认证契约变化。
- 最终Windows报告 `build/dev-runs/windows/20261029T091316Z-c7b66083-ctl-check/result.json`：60个顶层测试、203个pass事件；Linux `build/dev-runs/linux/20261029T091319Z-9b8e1c53-ctl-check/result.json`：40个顶层测试、149个pass事件。两端fail/skip零，无包级skip；两端vet、Linux race与两个有界fuzz、Windows原生内存签名/Job及15秒强制终止、14项包绑定检查通过。
- 独立包为上述Windows目录 `ready/ctl-check-launch.exe`，32个源码输入；source digest `058c63f5576ce901f2878b368fc0b0af96f3a2f10d3e41b2a80d2ed717251477`，launcher SHA-256 `40ec94e9dadcb78e3aa7aab3119ce42ea299d6b1285575c9da6bc5469f6b1a33`。交付报告 `build/desktop-dev-20261028/CTL-CHECK-D10-REPORT.md`；文档、定向扫描及旧包摘要检查记录在 `ctl-check-d10-final-checks.json`。
- 下一步仅维护者原生执行该包的 `--ctl-check` 并返回脱敏JSON，无需切换TUN或调整账号。实际CTL及证书来源本轮未读取；没有校园请求、真实证书写入、安装、Git refs、发布或Candidate变化。password_core仍blocked/live_unverified，旧包与历史handoff/status保留。

### 2026-10-29 D10 原生回执核验启动

- 维护者返回source 058c63f5… 的本地检查回执：绑定verified，两个来源read=complete但parse_failed，尚无结构/层级/签名结论，operations全零。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 只核验回执/包身份并离线检查解析分类，不重读真实来源或沿用校园授权。
- 分支codex/cas-submit-field-closure及HEAD b4eeb760保持，现有修改保留。当前事实不能解释为坏证书、缺根或签名失败；停止重复同包运行。

- 随后维护者明确要求 Agent 自动运行当前零网络本地检查。按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 更新执行安排：完成修订离线验收/绑定后可由正式工具运行 `--ctl-check`，记录账户与工具上下文；无需浏览器刷新或账号状态前提。校园请求及扩源仍未授权。

### 2026-10-29 D10 解析分类与自动本地检查完成

- 按 [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md) 细分闭集解析原因，接受集合和来源范围保持；自动执行账户 `bhkdesktop\unbal`，上下文 `codex-supported-tool`，不假定原生桌面等价。
- 两次工具执行回执保存在 `build/desktop-dev-20261028/ctl-check-tool-d10-stage-result.json` 与 `ctl-check-tool-d10-unsupported-result.json`。最终 AuthRoot 停在不支持的尾部扩展形状，Disallowed 停在不支持的成员算法；具体解释见 `CTL-CHECK-D10-PARSE-REVIEW.md`，不能解读为损坏或签名失败。
- 最终 Windows `build/dev-runs/windows/20261029T093158Z-57b37a63-ctl-check/result.json` 与 Linux `build/dev-runs/linux/20261029T093330Z-5410d090-ctl-check/result.json` 的测试、vet/self-test及平台对应 race/fuzz/构建/绑定检查通过；汇总与定向文档/扫描验收入口为 `build/desktop-dev-20261028/finalize-d10-parse.py`。
- 本轮实际读取已批准本地来源；校园请求、凭据读取、证书/注册表写入均零。当前本地执行不再等待维护者复制命令。扩展接受语义仍需具体规范决定；历史快照、旧包、Candidate与password_core blocked/live_unverified保持。

### 2026-10-29 D11 实施启动

- 维护者明确要求实施完整D11工作包，按[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)完成结构兼容、v1/v2报告、平台隔离、离线验收和自动零网络本机复核。
- 启动核验分支codex/cas-submit-field-closure、HEAD b4eeb760保持；既有修改及旧包保留。没有校园请求、凭据、信任修改或Git refs授权。

### 2026-10-29 D11 实现与验收完成，本机签名观察仍blocked

- 按[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)完成四成员算法、有界扩展、原生合成证实的属性省略、v1/v2闭集兼容及CMS拒绝阶段分类。产品未依赖诊断包；Windows本地读取与纯Go测试分层。
- 最终Windows `build/dev-runs/windows/20261029T101221Z-74ad4fa3-ctl-check/result.json`：65个顶层测试、287个pass事件；Linux `build/dev-runs/linux/20261029T101226Z-affcad6e-ctl-check/result.json`：45个顶层测试、233个pass事件。fail/skip零；vet/self-test、Linux race/两项fuzz、原生内存验签/Job/硬期限及14项绑定检查通过。六平台18依赖检查与18开发构建通过，macOS仅构建覆盖。
- 新包位于上述Windows目录ready，source `b2f173c3a780b04407fc28a19dfccd79ebdf2a7930af7807d59156eaf6e6bb93`，launcher `a43df0bfe125c74690b7c0aa0836fcc56454cd6663e8752a6ddede8897a4f88d`。33源码输入匹配；旧D9/D10/首轮D11包及历史快照保留。
- 已自动执行两个不同源码身份的零网络本机检查，账户bhkdesktop\unbal，工具上下文单独记录。最终AuthRoot为unsupported_signed_attribute_unknown，Disallowed为parse_failed_ctl_algorithm_unsupported；触及D11停止条件，不继续扩大接受范围。本地签名观察保持未完成，不推断具体OID/算法或可信性。
- 交付 `build/desktop-dev-20261028/CTL-CHECK-D11-REPORT.md`、`ctl-check-tool-d11-classified-result.json` 与 `ctl-check-d11-final-checks.json`。实际读取固定本地来源；网络、凭据、证书/注册表写入均零，没有Candidate、提交/推送或发布；password_core仍blocked/live_unverified。

### 2026-10-29 D12 后续事实核对与待确认提案

- 维护者要求继续；只读复核分支/HEAD及D11回执，既有改动保留。官方资料表明CTL成员标识不止四种SHA算法，CMS签名属性也可能超出现有白名单；仍不能由此猜测本机OID。
- 按grilling将[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)补充D12待确认提案：仅同两固定值的有界数字OID辨识，输出未经认证标识、不读取ROOT或验签、不改变接受范围。这个输出例外需维护者决定后才实施/运行。
- 本窗口未改Go、未创建新包或重读实际CTL、未发校园请求；旧结果及password_core blocked/live_unverified保持。调查记录见 `build/desktop-dev-20261028/CTL-D12-DESIGN-REVIEW.md`。

### 2026-10-29 D12 实施启动

- 维护者明确同意有界OID辨识和自动本机复核，按[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)实施独立入口、合成验收、绑定包和零网络执行；不增加验签或白名单接受。
- 分支codex/cas-submit-field-closure、HEAD b4eeb760保持；既有改动与旧包保留。固定两值之外不读取证书库，不发校园请求，不修改信任或Git refs。

### 2026-10-29 D12 实现、离线验收与自动本机辨识完成

- 按[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)完成独立ctl-identifiers入口、闭集报告、有界OID定位、保护执行及快照绑定。D11既有分析/命令与其构建快照一致，不改变其接受集合或信任。
- Windows报告 `build/dev-runs/windows/20261029T112342Z-4cdc6a2c-ctl-identifiers/result.json`：77个顶层测试、318个pass事件；Linux报告 `build/dev-runs/linux/20261029T112410Z-0ee1f4ea-ctl-identifiers/result.json`：55个顶层测试、262个pass事件。fail/skip零；vet/self-test、Linux race/两项有界fuzz、15项绑定检查通过。六平台18依赖隔离检查和18开发构建通过，macOS仅构建覆盖。
- 新包位于上述Windows目录ready；source `52fd756b6d43777a09a6c2373807c26da9d453928b72c91b7ac7f04ac56f26ab`，launcher `f0802e2aacefa1f3bb638d660fc130753b0e843ad5de33dec224bce152a5a578`；35项快照输入匹配。
- 已自动执行一次新包的零网络本机入口，账户bhkdesktop\unbal，上下文codex-supported-tool，不假定桌面会话等价。两个来源的限定标识提取complete，但signature仍not_checked、input_authenticity仍unverified、authentication仍blocked。七项operations全零。
- 官方名称核对已完成，实际数字OID与剩余兼容范围集中保存在 `build/desktop-dev-20261028/CTL-IDENTIFIERS-D12-REPORT.md`；执行回执 `ctl-identifiers-tool-d12-result.json`，核验 `ctl-identifiers-d12-final-checks.json`。不在交接复制产品规范或属性值。
- 下一工作包应先明确已识别SPC属性与Disallowed属性索引的结构兼容范围，合成验证后才继续原签名观察；本轮不擅自扩大算法/属性白名单。D9/D10/D11旧包和历史快照保留，Candidate及password_core blocked/live_unverified保持；没有校园请求、提交、推送或发布。

### 2026-10-29 D13 事实核对与待确认方案

- 维护者要求继续，复核分支/HEAD保持，保留现有变动。按grilling核对现有解析/验签/报告和官方定义，未重新读取实际CTL或执行诊断。
- [REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)新增D13待确认整体方案：固定两种表示仅作有界签名观察，明确opaque覆盖限制和v3结构含义，完成合成交叉验证后自动零网络复核。不在本轮先行扩大接受集合。
- 当前等待的是诊断范围与报告语义决定，不是再次请求PowerShell执行权限。产品、Candidate、历史文档和D12包保持；password_core仍blocked/live_unverified。

### 2026-10-29 D13 实施启动

- 维护者已明确批准D13整体范围，按[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)连续实现、合成验证、独立打包和自动本机零网络复核；不再等待执行权限确认。
- 保留既有修改、历史文档及旧包；无产品认证、校园网络、信任或Git refs变更。

### 2026-10-29 D13 实现、验收与自动本机有限签名观察完成

- 按[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)完成固定两表示的有界兼容、报告v3及严格v1/v2历史兼容；未实现成员匹配、内部策略或信任决策。
- Windows `build/dev-runs/windows/20261029T115548Z-f33dc450-ctl-check/result.json`：80顶层测试、343个pass事件；Linux `build/dev-runs/linux/20261029T115610Z-b0faa980-ctl-check/result.json`：58顶层测试、287个pass事件；fail/skip零。Windows原生与Go合成签名对照、vet/self-test、Linux race/两项fuzz、14项绑定检查通过。六平台18依赖隔离与18开发构建通过，macOS仅构建覆盖。
- 新包位于上述Windows目录ready，source `c404ba679290296cd08585ac44046989527088716fb18257a256285d9a656c01`，launcher `a60d60c9d23fa5ac946074779c5aed6d3ff2b393b946f37c239965dc812a70bb`；36项快照匹配。历史D9–D12包保持，旧handoff/status及Candidate未改。
- Agent已在bhkdesktop\unbal工具上下文自动执行一次新包零网络检查。两来源有限观察complete，固定层级observed、signature verified；两份window均unbounded，未解释覆盖显式保留。authentication仍blocked，purpose_policy/latest未建立，revocation未检查，七项operations均零。
- 交付 `build/desktop-dev-20261028/CTL-CHECK-D13-REPORT.md`、`ctl-check-tool-d13-result.json`及`ctl-check-d13-final-checks.json`。不将签名结果当Windows完整信任或校园登录证据；后续成员/策略/同轮peer关联需独立范围决定。password_core保持blocked/live_unverified，没有校园请求、提交、推送或发布。

### 2026-10-29 自动连续推进D14离线工作包

- 维护者再次明确要求自动继续。按[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)推进纯合成成员/策略分析，常规开发与测试不重复提问；事实查证和实施由Agent连续完成。
- 只用新生成证书/CTL和内存API；不读取实际新材料或重跑D13，不接入校园请求、产品认证或信任池。历史包、Candidate和工作区既有修改保留。

### 2026-10-29 D14纯合成实现与两端验收完成

- 按[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)完成有限成员关系原型、Windows生成材料内存对照、合成命令、快照构建和自动包审计；复跑入口为`python scripts/dev.py ctl-member --context tool`。没有实际CTL读取或校园执行。
- Windows `build/dev-runs/windows/20261029T123953Z-7b989332-ctl-member/result.json`、Linux `build/dev-runs/linux/20261029T123958Z-ebd305f0-ctl-member/result.json`分别为40/35顶层测试、198/193 pass事件，fail/skip零；vet、自测、Linux race/fuzz通过。新增整表上限回归另有Windows普通/Linux race通过日志。
- 独立合成包为上述Windows目录ready，21项快照，source `429f8bfb69bba03b1cbfc88737c7256b8d2912968c4a944b77caaf3dacdea0c7`；包自测与6种不支持模式拒绝通过。六平台18依赖隔离与18构建通过；macOS仅构建覆盖。
- 完整记录为`build/desktop-dev-20261028/CTL-MEMBER-D14-REPORT.md`及`ctl-member-d14-final-checks.json`。本次成员关系仅生成材料，不提升真实认证/信任状态；password_core仍blocked/live_unverified。旧包、历史handoff/status、Candidate与已有工作区修改保留，无提交或发布。

### 2026-10-29 D15自动实施启动

- 按[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)自动准备同轮peer成员观察、合成验收与新包；实际执行仍留在单次确认入口后。
- 只实施并测试代码，不读取实际CTL或产生校园请求；保留全部既有包及历史记录。

### 2026-10-29 D15离线验收与自动本地预检完成

- 按[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)完成固定材料准备、同轮peer成员观察、严格回执扩展、取消与硬期限保护；产品认证不变。
- Windows `build/dev-runs/windows/20261029T130551Z-b73caec4-cas-acquire/result.json`：187顶层/1045 pass；Linux `build/dev-runs/linux/20261029T130553Z-a10dd219-cas-acquire/result.json`：153顶层/956 pass；均零失败，包级skip仅无测试文件的cmd入口。vet、自测、Linux race/四项fuzz、9项绑定检查通过；六平台18依赖隔离与18开发构建通过。
- 新包在上述Windows目录ready，source `35d0a0ec861a20ff05d7afd7fae30e7c12a0146e808f7f02fae6ff9bdf618459`，launcher `bc8bbba6889cfb6d5c43ccdb57493006ad30b672a9ee9ad262f04c8c00192ede`。68项快照匹配。自动一次`--preflight`返回ready；本地ROOT/Disallowed读取允许，实际CTL读取及校园请求零，不代表账号offline。
- 交付 `build/desktop-dev-20261028/CAS-ACQUIRE-D15-REPORT.md`、`cas-acquire-d15-preflight.json`和`cas-acquire-d15-final-checks.json`。已连续做到可审阅单次真实窗口的阶段；真实执行仍待新授权，不自动输入历史确认。旧包/历史记录/Candidate保留，无提交、推送或发布，password_core仍blocked/live_unverified。

### 2026-10-29 D15本轮单次授权已执行并消耗

- 维护者本轮明确授权一次并确认当前条件。Agent经正式工具审批，在bhkdesktop\unbal原生Windows PTY启动精确D15包并输入一次确认；exit0，未重试，未改TUN/拓扑。
- 依[REL-CAS-DIAG-001](../docs/operations/cas-provenance-diagnostic.md)保存闭集结果：3次请求尝试/DNS/连接；AuthRoot成员any_match、Disallowed no_match；两窗口unbounded，策略覆盖未解释，collection/authentication仍blocked。无profile/credential读取、CAS POST、activation或logout。
- 记录在`build/desktop-dev-20261028/D15-LIVE-RESULT.md`和`cas-acquire-d15-live-once.json`；后者明确为PTY规范化转录，不冒称字节精确stdout，另作严格解码核验。工具上下文不等同于桌面会话的全面验收。
- 静态核对确认诊断显式ROOT池与产品Windows默认系统验证路径不同；不把诊断结果升级为产品TLS缺陷或成功证据。当前peer已释放，不保存/重构材料；系统链引擎、信任变更及后续真实窗口未授权。password_core保持blocked/live_unverified。

### 2026-10-29 认证收敛 roadmap 文档更新启动

- 当前工作包：`WP-AUTH-ROADMAP-DOCS`，范围及退出条件以 [P0](../docs/upgrade/plan.md#roadmap-20261029) 为准；本次只更新文档，不启动后续产品工作包。
- 只读核对分支 `codex/cas-submit-field-closure`、HEAD `b4eeb760945f0568ca163d957ddd55df2afa86c4` 与 origin 一致。沙盒账户 `BHKDesktop\CodexSandboxOffline` 触发 dubious ownership 后，经正式工具审批在 `bhkdesktop\unbal` 上下文核验；未修改 owner、ACL 或 safe.directory。
- 保存文档基线至 `build/roadmap-docs-20261029/before.json`：除本文件已增加启动记录和导航外，其余列入文件均为本次修改前字节；具体时点见该目录 `baseline-note.txt`。沙盒创建 build 目录被系统拒绝后，经正式审批完成保存；未改变目录权限。旧 `agent/handoff.md` 保持原样；维护者本次明确允许在 status 增加带日期的当前摘要，保留其历史正文。

### 2026-10-29 认证收敛 roadmap 文档交付

- `WP-AUTH-ROADMAP-DOCS` 已交付，权威入口为 [当前 roadmap](../docs/upgrade/plan.md#roadmap-20261029)，事实摘要见 [status](../docs/upgrade/status.md#status-20261029)，后续依赖见 [工作包索引](plans/stabilization-v1.md#auth-roadmap-work-packages)。此处只记录执行结果，不复制目标契约。
- 修改范围为八份文档：plan/status、工作包索引、本交接、文档中心、开发说明、诊断说明和 [ADR-0016](../docs/architecture/decisions/ADR-0016-cas-provenance-before-key-selection.md) 导航补充；ADR 仍为 proposed，旧诊断正文与包保留。原 plan/status 历史字节为纯新增方式保留；旧 handoff、live-validation、release、evidence、migration-matrix 和生成索引与本次修改前哈希一致。
- 使用既有 Go 1.25.0 离线 docs 入口执行 doccheck；首次发现研究规范链接指向错误，修正后通过。定向 gitleaks 扫描八份文档，零命中。详细最终命令、退出码、摘要和文件哈希保存在 `build/roadmap-docs-20261029/`。
- 下一开发工作包为 `WP-AUTH-BASELINE`（P1），本次未启动。没有产品修改或产品测试、真实 CTL/证书读取、校园诊断、凭据操作、Candidate 构建、Git 提交/推送或发布；保留此前全部未提交开发改动。

### 2026-10-29 P1 旧版与 v1 行为对照启动

- 维护者要求开始，当前执行 `WP-AUTH-BASELINE`，范围与退出条件见 [roadmap P1](../docs/upgrade/plan.md#roadmap-20261029)；只读固定源码并形成研究矩阵，不运行旧版认证或读取凭据。
- 本地 Git 只读核验：当前分支/HEAD 未变，所有既有改动保留；v0.1.3 解析为 `5ad8c1fa05102fbcb249e7195f4801563b3d44a5`，指定归档 HEAD 为 `819203b3d81d5960cde7e2777fa92991658ed04a`，归档工作区无变化。历史 `3721fd6…` 对象在两处本地仓库均不可读取，不以 tag 名称推断 tree 等价。

### 2026-10-29 P1 研究交付

- `WP-AUTH-BASELINE` 交付 [固定来源、行为矩阵及 P2 判别输入](../docs/research/auth-baseline-2026-10-29.md)，按 [PROTO-RESEARCH-001](../docs/research/active-implementations.md) 保留旧行为线索，不恢复不安全路径、不把源码差异当成真实根因。
- 更新研究导航和 status 当前摘要；源码身份/限定 blob 比较、328 份代码与配置文件的前后哈希基线、文档检查和定向脱敏回执保存于 `build/auth-baseline-p1-20261029/`。未复制旧源码正文或真实材料。
- 下一工作包为 `WP-AUTH-VALIDATION-PATH`（P2），本次未实施。未运行旧程序、产品测试、证书/CTL 审计或校园请求，未读取凭据/私有 evidence，未提交、推送或创建 Candidate；已有开发改动和旧包保留。正式里程碑以 [status](../docs/upgrade/status.md#status-20261029) 为准。

### 2026-10-29 P2 契约收敛启动

- 维护者要求开始 `WP-AUTH-VALIDATION-PATH`。按 [roadmap P2](../docs/upgrade/plan.md#roadmap-20261029) 先核对复用边界、形成规范/ADR，再实现离线验收；没有新增真实操作授权。
- 分支/HEAD 与 P1 一致，保留全部既有改动。正在核对现有 SDK 注入入口、正式 runner 的状态机/执行器与开发包绑定，避免复制认证流程或伪造 Candidate 身份。
- 常规开发权限不重复询问；仅尚未确定的实际会话/网络边界按已启用 grilling 形成集中决策。当前没有运行产品、校园请求、真实证书/CTL 读取或凭据操作。

### 2026-10-29 P2 实现与专项离线验收交付

- `WP-AUTH-VALIDATION-PATH` 按维护者本轮 Q1/Q2 确认落实 [ADR-0017](../docs/architecture/decisions/ADR-0017-development-auth-validation.md) 与 [REL-DEV-AUTH-001](../docs/operations/development-validation.md)。这里仅记录执行结果；窗口、传输、凭据和回执规范以 docs 为准。
- 源码位于 `internal/authdev/`、两个 `internal/cmd/auth-dev*` 入口；共用 `internal/livegate` 只增加可选执行上下文及开发步骤校验，正式默认行为回归通过。绑定复用 Windows 只读目录/文件句柄，历史 D 工具身份与模式不扩展。
- 最终独立包为 `build/dev-runs/windows/20261029T152626Z-bfab731e-auth-dev/ready/`，source `6a6f6312d7d930a711f4a19fcd6969d68b30392824ac1f5e8579d27284a086df`。通过 `python scripts/dev.py auth-dev --context tool` 完成两端专项验收、启动器合成自检与篡改拒绝；未运行真实入口。
- 汇总回执在 `build/auth-validation-p2-20261029/REPORT.md`。Linux 全仓 race、两端 vet、六平台产品交叉构建通过；Windows 全仓安装/junction 回归仍失败，直接目标执行成功，未证明根因。失败日志保留，不改权限、不以历史通过覆盖本轮失败。
- 发现 `go test ./...` 会收集 build 下历史扫描副本，开发入口改为当前根包/cmd/internal 全量枚举；所有当前包保留。Windows 束文件测试与 ACL/安装测试使用各自测试目录，没有修改生产 ACL 或所有者。
- 分支/HEAD 未变，保留全部既有改动；没有提交、推送或 Candidate。下一主线为 P3，真实窗口须新的单次授权。里程碑和缺口见 [status](../docs/upgrade/status.md#status-20261029)，本交接不提升认证能力。

### 2026-10-30 P2 产品等价匿名单次窗口启动

- 维护者本轮明确授权一次，并确认当前 BHKDesktop 原生 Windows、TUN 关闭、有线独占、账号 offline、运行期间拓扑不变；范围以 [REL-DEV-ANON-001](../docs/operations/development-validation.md) 为准，不延伸到 password 窗口。
- 正式工具审批账户为 `bhkdesktop\unbal`。主机/工作目录/分支/HEAD 与 P2 交付一致；启动器、观察程序、产品与构建输入四项哈希均匹配已交付身份 source `6a6f6312d7d930a711f4a19fcd6969d68b30392824ac1f5e8579d27284a086df`。
- 包内自检与 `--preflight` 返回 binding verified、ready，live_attempted false，应用请求和产品命令均为零。随后只执行该包一次匿名入口，不自动重试或执行后续认证。

### 2026-10-30 P2 匿名窗口完成，单次授权已消耗

- 精确包在 Windows PTY 输入一次确认后正常退出 0，binding verified，结果为 observed / credential_boundary_reached。应用请求 6（status 1、discovery 2、CAS 3）；CAS Cookie 接收/回送为 true，page_observed true，key_source script，action_service same，product_commands 0。
- 按 [REL-DEV-ANON-001](../docs/operations/development-validation.md) 保存闭集回执至 `build/auth-anonymous-20261030/`；result.json 明确为 PTY 规范化转录。未保存真实认证材料，没有额外 status、自动重试、凭据读取或后续认证。
- 该窗口已证明产品等价前置流程到达 provider，未证明 RSA/POST/登录成功；下一步与正式状态见 [status](../docs/upgrade/status.md#status-20261030-anonymous)。本轮授权已用完，不延伸到开发 password_core。

### 2026-10-30 私有开发 password_core 授权与就绪

- 维护者在匿名结果交付后明确授权下一次私有开发 password_core；按 [REL-DEV-PASSWORD-001](../docs/operations/development-validation.md) 使用既有 P2 包，凭据仍由维护者在私有终端输入。
- Agent 以正式审批账户 `bhkdesktop\unbal` 复核 BHKDesktop、既有分支/HEAD 和四项包哈希均未变；零网络 `--preflight` 返回 binding verified、ready，所有请求/产品命令为零。
- 已交付选择现有 prompt profile 的私有终端入口；现场条件由启动器当次确认。尚未执行 password 模式，未读取 profile/凭据，也没有新的登录或注销结果；等待维护者返回闭集 JSON，不自动重试。正式状态不变。

### 2026-10-30 私有 prompt 配置准备

- 维护者要求 Agent 创建配置后由其填写。已核验既有开发产品哈希，通过其本地 `profile add` 在 Windows 默认用户配置位置新增 `bhk-dev-prompt`，provider 为 prompt、switch 为 refuse，username 仅为待替换占位符；退出码 0。
- 使用产品配置事务保留既有条目，未导出原配置内容，未读取密码或调用认证/网络入口。真实账号由维护者在仓库外的私有配置中填写，密码不写入文件；随后仍使用既有单次私有窗口授权，不视为已经执行。

### 2026-10-30 私有开发 password_core 首轮完成，授权已消耗

- 维护者回传同一 P2 包的闭集 JSON，binding verified；initial_status_offline pass，login_logged_in blocked / exit 3 / network，product_commands 3。只保存脱敏 JSON 和分析至 `build/auth-password-20261030/`，不复制终端账号提示。
- 按 [REL-DEV-PASSWORD-001](../docs/operations/development-validation.md) 与现有状态机解释：没有后续 online 检查或 cleanup 所有权，没有最终 offline 证据；step_coverage complete 不是七步全部通过。没有自动重试或新增真实操作。
- 现有分类不足以定位网络错误阶段，后续先离线设计产品等价的闭集阶段观察，不预判 HTTP 500 或选键根因；正式进度见 [status](../docs/upgrade/status.md)。

### 2026-10-30 P3 失败分类离线实现启动

- 维护者已确认 [ADR-0018](../docs/architecture/decisions/ADR-0018-development-auth-failure-observation.md) 完整方案；本轮范围为离线实现、两端合成验收、独立打包和自检，不执行真实认证或读取用户配置。
- 现有分支/HEAD 不变，保留所有既有改动；普通产品与 observed CLI 分别绑定，旧包保持。预算与隐私边界以 [开发验证约定](../docs/operations/development-validation.md#p3-failure-observation-proposal) 为准。

### 2026-10-30 P3 失败分类离线验收与打包完成

- 按 [ADR-0018](../docs/architecture/decisions/ADR-0018-development-auth-failure-observation.md) 完成独立 observed CLI、被动观察器、严格子进程/schema2封装、四文件绑定和仅开发使用的 ProcessAdapter；普通产品及正式 runner 默认路径保持。
- Windows专项测试/vet/自检/整链与Linux专项普通/race/vet/fuzz通过，六平台普通产品构建和诊断依赖隔离通过。最终入口 `build/dev-runs/windows/20261030T040449Z-27fa0356-auth-dev/ready/auth-dev-launch.exe`，source `8fb7735263af6130c8106a15c02c15a794e0698d1c9fcbf508cac990fe74706c`；报告见 `build/auth-observed-p3-20261030/`。
- 只执行假响应、合成配置、loopback及零网络自检/预检；新包真实运行未授权。不得沿用上轮已消耗的password窗口授权；下一步与未闭合状态见 [status](../docs/upgrade/status.md)。

### 2026-10-30 P3 维护者回传观察窗口

- 维护者在私有终端完成新包当次确认后回传闭集 JSON；按 [ADR-0018](../docs/architecture/decisions/ADR-0018-development-auth-failure-observation.md) 保存规范化转录至 `build/auth-observed-live-20261030/`，不复制账号提示。后续对脱敏回执和原包做离线严格解码/哈希核验，没有执行产品或新的网络操作。
- 回执将 login 的 network 定位为 CAS POST HTTP 500，之前三次 CAS GET 均 200；本轮没有 activation 或最终 offline 证据。具体事实、限制和下一工作包见 [status](../docs/upgrade/status.md) 最新 P3 结果，不把历史两键或 CTL 观察当成本轮根因。
- 当前 HEAD 不变，既有改动及旧包保留。真实授权已消耗；下一步仅离线收敛提交上下文，正式里程碑不提升，不自动重试或注销。

### 2026-10-30 P3 提交上下文离线检查启动

- 维护者要求继续；在 `WP-AUTH-MINIMAL-FIX` 范围内审阅 [PROTO-DISCOVERY-001](../docs/architecture/protocol-correctness.md) 和 [ADR-0015](../docs/architecture/decisions/ADR-0015-cas-submit-closed-fields.md)，补齐有信息增益的合成会话/提交检查，不改变选键规则或被动观察边界。
- HEAD 与分支未变，保留所有既有改动。只使用新生成的合成密钥、凭据及 fake transport；真实授权不续用。新观察方案先核对已有分析能力再集中决策。

### 2026-10-30 P3 提交上下文离线检查交付

- 新增根 SDK `cas_submit_context_test.go`；Windows 聚焦普通与 Linux 聚焦普通/race/vet 通过。范围、结果与剩余证据缺口见 [研究记录](../docs/research/cas-submit-context-2026-10-30.md)，命令回执位于 `build/auth-submit-context-20261030/`。
- 没有修改产品实现、公共契约或现有包，也没有运行真实认证。现有 CAS Jar/字段流程在所测场景下成立；不据此断定真实上下文正确。
- 已准备 [下一匿名上下文提案](../docs/operations/development-validation.md#p3-anonymous-context-proposal)，并就新增投影范围发出集中 grilling 决策。该提案尚未生效，不沿用历史真实授权；正式状态以 [status](../docs/upgrade/status.md) 为准。

### 2026-10-30 P3 匿名上下文实施启动

- 维护者已确认有界匿名上下文观察和完整离线交付，先落实 [ADR-0019](../docs/architecture/decisions/ADR-0019-anonymous-submit-context.md)，然后实现独立入口、合成回归与绑定包。旧观察和旧包不改写。
- 本轮不运行真实匿名或密码模式；只使用 fake/loopback 和生成合成材料。产品选键与共享认证语义不变，使用关系保持 unknown。

### 2026-10-30 P3 匿名上下文离线交付完成

- 按 [ADR-0019](../docs/architecture/decisions/ADR-0019-anonymous-submit-context.md) 实现独立 `internal/authcontext`、对应两个命令、独立绑定与构建/验收入口；沿用原匿名守卫，不改 SDK 认证流程。来源分析仅增加分离的 parser 映射 API，旧分析回执含义保持。
- 最终包 `build/dev-runs/windows/20261030T112836Z-6bb2cf77-auth-context/ready/`，source `b0dc49ea88070b8837b561e76ee66648cc81bceeae144c489965d75ec79e17d9`；Windows 与 Linux 专项、六平台产品依赖隔离/构建通过，175 项快照及文件哈希一致，零网络预检 ready。完整范围和回执见 `build/auth-context-p3-20261030/REPORT.md`。
- 审阅修正了观察器的主机名大小写及动态字段原值分类，两端回归和最终新包重新验收；前一包和预检记录保留。没有真实匿名/密码运行、profile/凭据读取或新的真实授权。
- 下一步与未闭合状态以 [status](../docs/upgrade/status.md) 最新摘要为准。没有修改产品选择、正式门禁、Candidate 或 Git 提交状态。

### 2026-10-30 P3 匿名上下文单次窗口启动

- 维护者明确授权最终包一次匿名运行，并确认 BHKDesktop 原生 Windows、TUN 关闭、有线独占、账号 offline、窗口期间拓扑不变。范围依 [ADR-0019](../docs/architecture/decisions/ADR-0019-anonymous-submit-context.md)，不延伸到登录或后续重试。
- 正式工具审批账户 bhkdesktop\unbal；主机、分支、HEAD 与交付一致。最终 source `b0dc49ea88070b8837b561e76ee66648cc81bceeae144c489965d75ec79e17d9`，启动器/观察程序/普通产品/build-inputs 四项哈希只读复核一致。即将只运行该包一次 --anonymous-once。

### 2026-10-30 P3 匿名上下文单次窗口完成

- 最终包执行一次后退出 0，binding verified，SDK 到达凭据边界；闭集结果与解释见 `build/auth-context-live-20261030/REPORT.md`。回执是 Agent PTY 规范化转录，严格解码、四项包哈希与 175 项源码快照通过；不保存真实材料。
- 依据 [ADR-0019](../docs/architecture/decisions/ADR-0019-anonymous-submit-context.md) 解释结果：本轮表单目标关系一致；parser 匹配可定位但语法分类部分覆盖，使用关系仍 unknown。具体事实与下一步以 [status](../docs/upgrade/status.md) 最新摘要为准，不据此修改产品选键或提升认证状态。
- 此次真实授权已消耗，没有重试、后续认证或额外 status；旧包与现有改动保留。接下来仅离线审阅分类边界，扩大语义/材料或新的真实操作需独立决定。

### 2026-10-30 P3 语法覆盖边界离线核对启动

- 维护者要求继续；沿 [ADR-0019](../docs/architecture/decisions/ADR-0019-anonymous-submit-context.md) 和现有来源分析契约，只审阅降级边界并新增有判别力的合成回归，不改变产品或分析语义。
- 现有改动与旧包保留；本轮不读取真实材料、不重新运行匿名窗口，不将历史授权续期。完成后记录可证明路径与仍需决策的缺口。

### 2026-10-30 P3 语法覆盖边界离线核对完成

- 在现有 [ADR-0019](../docs/architecture/decisions/ADR-0019-anonymous-submit-context.md) 边界内新增 `parser_boundary_test.go`，核对匹配映射、未知后缀、来源隔离与闭集报告不可区分性；实现和输出契约未变。
- Windows 普通/vet/自检与 Linux 普通/race/vet/有界 fuzz/自检通过；结果和下一决策依据见 [研究记录](../docs/research/cas-submit-context-2026-10-30.md) 与 [status](../docs/upgrade/status.md)。没有新包、校园请求或认证结论。

### 2026-10-30 P3 完整静态语法工作包启动

- 维护者已确认完整离线工作包，按 [ADR-0020](../docs/architecture/decisions/ADR-0020-bounded-static-cas-syntax.md) 连续进行可行性、规范、实现与验收，不授权真实窗口。
- 在独立 `build/static-syntax-feasibility-20261030/` 评估固定 parser；显式代理仅获取公开开发依赖并通过模块校验。Windows 初步合成 AST 差分、取消、预算和 source map 禁用检查通过，尚不代表独立工具可交付。

### 2026-10-30 P3 完整静态语法离线交付

- 按 [ADR-0020](../docs/architecture/decisions/ADR-0020-bounded-static-cas-syntax.md) 完成纯静态派生 parser、严格独立报告、同一 SDK 采集路径接入、预算/取消/原生资源保护和独立绑定；产品取键不变，旧回执和包保留。
- 最终入口 `build/dev-runs/windows/20261030T124136Z-538b1b67-auth-static-context/ready/auth-static-context-launch.exe`，source `1c4215a2c3fdd753aa0162072af06b8ffbd4e3b33d5fdb9f9a9acaf4401a7a36`；199 项源码快照与包哈希一致。两端专项验收、Linux race/有界 fuzz、六平台普通产品构建及诊断依赖隔离通过；详细回执见 `build/auth-static-context-p3-20261030/REPORT.md`。
- 正式工具审批账户 bhkdesktop\unbal 自动执行最终包零网络预检，binding verified / ready；未执行匿名或密码入口、读取真实材料或 profile。上轮单次授权已经消耗，新包真实窗口尚未授权。
- 剩余边界与下一步以 [status](../docs/upgrade/status.md#status-20261030-static) 为准；不提升正式认证、M3 或安装回归状态，没有 Candidate、提交或推送。

### 2026-10-30 P3 静态语法单次匿名窗口启动

- 维护者已授权最终新包一次匿名观察，并确认当前 BHKDesktop 原生 Windows、TUN 关闭、有线独占、账号 offline、拓扑保持。执行范围依 [ADR-0020](../docs/architecture/decisions/ADR-0020-bounded-static-cas-syntax.md)，不延伸至重试或密码验证。
- 正式工具审批账户 bhkdesktop\unbal；主机、分支、HEAD 不变，四项固定包哈希一致，本地零网络预检 ready。执行 source `1c4215a2c3fdd753aa0162072af06b8ffbd4e3b33d5fdb9f9a9acaf4401a7a36` 的最终启动器一次；后续只处理闭集回执，不保存真实材料。

### 2026-10-30 P3 静态语法单次匿名窗口完成

- 最终包一次运行退出 0，binding verified；规范化闭集回执、严格解码、四项哈希和 199 项快照核验见 `build/auth-static-context-live-20261030/REPORT.md`。真实材料未落盘，未新增请求或重试。
- 本轮 parser 匹配经新静态分类落在注释中；事实边界、认证含义与下一离线工作见 [当前状态](../docs/upgrade/status.md#status-20261030-static)。遵守 [ADR-0020](../docs/architecture/decisions/ADR-0020-bounded-static-cas-syntax.md)，不据此自动更改产品取键或扩大来源。
- 此次真实授权已消耗。后续只整理结果和文档检查，没有 profile/凭据读取、POST、activation、logout 或产品修复；正式认证与里程碑不提升。

### 2026-10-30 P3 浏览器对照离线可行性启动

- 维护者明确确认采用独立 Playwright/CDP，在本地合成页面验证隔离、请求阻断和调用链观察；不授权校园导航，不复用个人 Chrome 会话，不修改产品选键。范围和判别目标见 [研究记录](../docs/research/cas-submit-context-2026-10-30.md)。
- 只读盘点已发现 bundled Playwright 及既有 Chrome/Edge；优先复用现有工具，不安装或修改系统/浏览器设置。离线实验输出独立放在 build 中，不能当作真实流程或认证验收。

### 2026-10-30 P3 登录使用关系修复启动

- 维护者已批准实施 [ADR-0021](../docs/architecture/decisions/ADR-0021-cas-key-use-relation.md)；先更新规范，再修改共享选择器和开发包。旧历史 handoff 不改写。
- 正式工具审批账户 bhkdesktop\unbal 核验分支及 HEAD 未变；保留全部已有修改和未跟踪文件。没有更改权限或 safe.directory。
- 本轮连续完成离线回归与打包；不读取私有 profile/凭据，不运行校园网络，不升级正式里程碑。旧匿名入口在新构建中冻结，旧包不变。


### 2026-10-30 P3 登录使用关系离线交付

- [ADR-0021](../docs/architecture/decisions/ADR-0021-cas-key-use-relation.md) 已实现；执行范围、检查与缺口见 `build/p3-key-relation-20261030/REPORT.md` / `final-checks.json`，正式状态见 [当前摘要](../docs/upgrade/status.md#status-20261030-key-relation)。旧 handoff/Candidate/包不改写，未提交或推送。
- 新开发入口 `build/dev-runs/windows/20261030T143328Z-f77e3182-auth-dev/ready/auth-dev-launch.exe`，source `67985fe2d2a580608d81504565cae172e594eb60222aeebca8d11f1305aa43e0`；204项绑定与整链合成自检通过。没有运行真实模式，匿名入口在新构建中拒绝。
- 维护者已报告当前 TUN 开启、账号在线用于远程管理；本轮继续离线工作，不使用或清理其会话。后续真实操作先单次授权及现场条件确认，Windows junction 安装缺口继续单列。


### 2026-10-31 P3 修复包首次 password_core 回执

- 维护者私有终端执行单次授权窗口；JSON规范化转录、严格解码与原包绑定校验见 `build/auth-key-relation-live-20261031/REPORT.md`。不保存账号提示或认证材料。
- 当前进度和下一工作包以 [status](../docs/upgrade/status.md#status-20261031-password-progress) 为准：开发登录成功，套件在注销响应处理失败，清理最终offline。本次授权已经消耗，Agent没有重跑或新网络操作。
- 新增合成注销状态转换测试，未修改产品实现、旧包或成功判据；后续先收敛新的结构观察边界，不根据未知正文直接改parser。

### 2026-10-31 P3 注销观察提案整理

- 正式工具审批账户 bhkdesktop\unbal 核对分支和 HEAD 未变，保留全部已有修改。依据当前失败轨迹，只读核对响应上限、业务 parser 和观察器；未重跑真实窗口。
- 新范围见 [ADR-0022](../docs/architecture/decisions/ADR-0022-development-logout-structure.md) 和[开发验证提案](../docs/operations/development-validation.md#p3-logout-structure-proposal)，当前状态 proposed。确认前不实现正文观察，不改变产品成功判据。
- 本轮只准备文档与对应检查；旧包、正式状态和 Windows 安装缺口保持。下一步由维护者决定新材料投影范围，而非重复申请 PowerShell 或既有离线验证权限。

### 2026-10-31 P3 注销结构观察实施启动

- 维护者已确认 [ADR-0022](../docs/architecture/decisions/ADR-0022-development-logout-structure.md) 完整离线范围；连续实施、回归、独立打包与零网络自检，不扩大产品契约。
- 原真实授权已消耗，本轮没有真实 profile/密码读取或校园请求授权。旧包与历史回执保持。

### 2026-10-31 P3 注销结构观察离线交付

- 按 [ADR-0022](../docs/architecture/decisions/ADR-0022-development-logout-structure.md) 完成实现、合成等价/预算/并发与严格版本回归、独立打包及零网络自检。执行账户 bhkdesktop\unbal；没有改变普通产品判据或真实操作权限。
- 最终入口与身份以 [status](../docs/upgrade/status.md#status-20261031-logout-observation) 为准；汇总回执在 `build/logout-observation-p3-20261031/REPORT.md`。中间包保留，下一窗口只使用最终包。
- 下一步是维护者另行单次授权的私有终端 password_core；本轮没有运行该入口或读取真实 profile。Windows 安装/junction 缺口与正式里程碑保持，不以观察器交付替代认证闭环。

### 2026-10-31 P3 注销观察新包单次回执

- 维护者私有终端执行新包一次；严格解码、原包绑定和解码源码核验通过，闭集结果见 `build/logout-observation-live-20261031/`。未保存终端账号提示或响应材料。
- 当前事实与下一工作以 [status](../docs/upgrade/status.md#status-20261031-logout-envelope) 为准：开发登录再次成功，首次注销封装未被现有对象入口识别，套件仍失败；清理仅查询即已 offline。
- 单次授权已消耗；Agent只整理证据和只读核对，没有重跑、产品修改或新正文观察。后续契约变化遵循规范和实质性决策确认，不自动扩大 [ADR-0022](../docs/architecture/decisions/ADR-0022-development-logout-structure.md)。

### 2026-10-31 P3 Chrome 注销对照启动

- 维护者已批准 [ADR-0023](../docs/architecture/decisions/ADR-0023-browser-logout-comparison.md) 的准备与静态范围；真实窗口另行授权。使用当前Chrome分工、维护者先脱敏、旧版仅源码对照，不执行旧debug。
- 正式工具审批账户 bhkdesktop\unbal 复核分支/HEAD不变，保留全部已有修改。已只读盘点Chrome/DevTools窗口，未截图或读取页面；维护者随后确认Sources已就绪。
- 研究矩阵见[注销对照记录](../docs/research/logout-browser-context-2026-10-31.md)；下一步只核对可安全查看的已加载源码。旧包/历史回执/产品判据及正式里程碑保持。

### 2026-10-31 P3 Sources 截图核对与短信入口线索

- Computer Use因无法可靠识别当前浏览器URL而中止，未成功读取页面，未绕过限制。维护者提供局部Sources截图，事实与缺口已记入[研究记录](../docs/research/logout-browser-context-2026-10-31.md)。当前匹配属于CAS登录form action，尚未取得网页注销处理代码。
- 维护者要求记录的短信输入/按钮结构仅作为静态前端线索；既有能力状态不变，不触发短信发送或提交，不改变注销主线。
- 下一步等待已有网关Sources的安全材料；没有新的真实窗口授权、产品实现改动或认证操作。历史交接、包与正式里程碑保持。

### 2026-10-31 P3 本地网页脚本静态对照

- 已只读分析维护者提供的三份桌面脚本；文件身份、请求/业务判断差异及静态证据限制见[研究记录](../docs/research/logout-browser-context-2026-10-31.md)。未复制原件、执行脚本或请求资源。
- 网页分支配置仍缺；下一步仅核对已加载源码中的三个开关，不收集完整CONFIG。真实窗口申请前需明确双栈和CAS会话退出范围，见[浏览器约定](../docs/operations/browser-comparison.md#browser-logout-comparison)。
- 未修改产品解析器或认证判据，未运行真实操作；历史包、正式里程碑及安装缺口保持。

### 2026-10-31 P3 网页配置缺口闭合

- 维护者提供的配置确认normal分支及双栈关闭，事实和证据限制见[研究记录](../docs/research/logout-browser-context-2026-10-31.md)。只保存必要布尔值，不保存实际地址、完整CONFIG或截图。
- 下一步是已有Network材料的脱敏对照；缺少记录时另行申请明确含CAS会话退出影响的真实窗口，不再重复要求配置查找。本轮未进行浏览器操作或网络认证，产品和正式状态保持。

### 2026-10-31 P3 网页注销请求及预览核对

- 维护者提供自行操作后的Network材料；必要脱敏投影和证据边界见[研究记录](../docs/research/logout-browser-context-2026-10-31.md)。未转存请求头、原始正文、动态值或截图；不追认为完整三次status验收或持续授权。
- 只读核对表明当前SDK可接受符合其严格封装的相应业务字段，不能直接归因于业务字段名不同。下一步仅确认现有Response完整封装，再形成最小方案；不要求重跑登录/注销，不修改产品或正式状态。

### 2026-10-31 P3 网页封装对照离线结论

- Response封装已由维护者提供；只保留脱敏结构，用独立合成数据验证现有解析器，Windows两个纯解析包测试通过。报告及证据限制见[研究记录](../docs/research/logout-browser-context-2026-10-31.md)。未修改产品实现或执行真实请求。
- 当前材料收集结束，不再重复要求网页登录或源码查找。下一实质性决策是研究记录中的callback单变量候选修复，尚待确认；新真实验证仍另行授权，正式状态不提升。

### 2026-10-31 P3 callback候选修复启动

- 维护者已批准[ADR-0024](../docs/architecture/decisions/ADR-0024-logout-explicit-callback.md)对应完整离线工作包；在自己的SDK/CLI和绑定开发包中实施，不操作Chrome，不沿用真实运行授权。
- 正式工具审批账户bhkdesktop\unbal只读复核分支和HEAD不变，保留全部既有改动；沙盒账户所有者隔离未通过时没有修改safe.directory或权限。旧handoff保留历史用途。
- 先规范再实现，连续完成相关Windows/Linux回归、独立包与自检；正式里程碑和安装缺口不提升。

### 2026-10-31 P3 callback候选修复离线交付

- [ADR-0024](../docs/architecture/decisions/ADR-0024-logout-explicit-callback.md)范围已完成；两端相关测试、Linux race/fuzz、六平台构建与依赖隔离、独立绑定包/自检通过。最终身份与限制见[当前状态](../docs/upgrade/status.md#status-20261031-logout-callback)，汇总见 `build/p3-logout-callback-20261031/REPORT.md`。
- 正式工具审批账户bhkdesktop\unbal执行；保留初次夹具失败记录及后续通过结果，没有修改既有空正文/用户名/成功判据以绕过失败。旧handoff、包及Candidate保持。
- 下一步是另行单次确认的私有终端password_core；本轮没有读取真实profile、拓扑preflight或执行校园请求。正式状态与原生安装缺口保持，未提交或推送。

### 2026-10-31 P3 新包单次开发闭环通过

后续执行入口：2026-10-31 维护者批准完整 P4，按 [ADR-0025](../docs/architecture/decisions/ADR-0025-bhk-formal-acceptance.md) 连续实现和离线验收。当前进度见 [status](../docs/upgrade/status.md#status-20261031-p4)。正式审批账户 bhkdesktop\unbal 已复核分支/HEAD 不变，已有修改保留；旧 handoff、Candidate 和真实授权边界保持。

- 维护者私有终端回传新包七步全pass；严格解码、五项包哈希和211项源码快照核验通过。脱敏回执及身份检查见 `build/logout-callback-live-20261031/`，当前结论见[开发闭环状态](../docs/upgrade/status.md#status-20261031-password-dev-pass)。
- 首次注销的JSONP业务解析与offline复核均通过，最后幂等注销不发送额外logout；没有cleanup。只记录窗口末尾offline，不能推断当前持续状态。原始终端账号提示未保存。
- 单次授权已消耗；Agent只做离线证据复核和文档更新，没有重跑或新增产品实现。下一工作包依[roadmap](../docs/upgrade/plan.md#roadmap-20261029)进入P4规则收敛，正式Candidate、里程碑和既有平台/安装缺口保持。

### 2026-11-01 P5 正式 Candidate 构建与离线验收

- P4 补充覆盖、PR #29 合并及一次正式 Candidate workflow 已按各自具体授权执行；新 Candidate 来源、六平台安装和 11 项独立 attestation 复核通过。完整身份及本机安装执行结果见 [P5 记录](../docs/research/p5-readiness-2026-11-01.md#正式-candidate-与本机验收)，正式状态见 [status](../docs/upgrade/status.md#status-20261031-p4)。
- 本地仍在 codex/bhk-v1-acceptance；Candidate 使用已合入的 squash main，不混用本地分支提交身份。旧分支、旧包、首次失败回执和未跟踪缓存保留。新执行文档变更尚未提交；本轮不自动推送。
- 后续依赖为同一 Candidate 的 BHK 安装辅助验收闭合，再进入另行授权的有线 password/QR 与预先准备的 Wi-Fi 窗口。未读取真实 profile/凭据、未执行校园请求，未创建最终 tag/release；原 handoff 的历史快照不重写。
- 收尾：安装辅助第 003 份回执及其只读 bundle 核验通过；前两份 blocked 保留。执行环境要求标准提升上下文及短的独立测试 work-root，来源和结论边界见上述 P5 记录。下一待授权项为有线正式 password_core，密码仍只由维护者在私有终端输入；当前认证状态不从早先现场条件推断。
- 后续单次窗口：维护者已授权该 Candidate 的一次有线正式 password_core，并重新确认现场条件。普通产品从已验证 Windows 归档选择性提取至新私有目录，产品/manifest/helper 哈希核验通过；私有终端入口为 `build/p5-preparation-20261101/run-wired-password-once.ps1`。Agent 仅执行 CheckOnly，没有读取 profile、查询校园状态或启动认证；等待维护者在私有终端执行并回传末尾脱敏 JSON。入口一次性标记阻止盲目重跑，原始认证输出不落盘；本条不表示正式验收已通过，也不授权 QR/Wi-Fi 窗口。
- 原生终端反馈上述 LOCALAPPDATA 产品不存在；失败发生在首个哈希检查，尚未启动 runner、读取 profile 或消耗真实认证运行。工具上下文的 C 盘可见性检查不能代替维护者终端可见性。已从 D 盘已验证原件准备新的共享 D 盘私有运行子集（普通产品、manifest、对应 helper），原件及旧入口保留；当前入口改为 `build/p5-preparation-20261101/run-wired-password-shared-once.ps1`。CheckOnly 三项固定哈希通过，仅收紧新目录/文件权限，未变更产品或 helper 字节；等待维护者执行同一次尚未开始的授权窗口。
- 窗口收尾：维护者已完成有线正式 password_core，七步全部通过、无 cleanup；Agent 对新 v2 bundle、步骤及 Candidate/产品/helper 绑定的只读复核通过，见 [正式回执](../docs/research/p5-readiness-2026-11-01.md#有线正式-password_core-回执) 和 [状态](../docs/upgrade/status.md#status-20261031-p4)。本次授权已消耗，不重复运行；下一待授权项为有线 terminal QR，密码与二维码仍只在维护者私有终端处理。尚未完成 Wi-Fi 两项、摘要人工审阅或 promotion，M3 不提前完成。
- 有线 QR 准备：维护者已单独授权同一 Candidate 的一次 terminal_qr，并重新确认当前现场条件。新入口 `build/p5-preparation-20261101/run-wired-qr-shared-once.ps1` 使用已验证的共享 D 盘私有运行子集；语法、固定 suite、三文件哈希和独立一次性标记检查通过。Agent 仅执行 CheckOnly，未读取 profile、启动网络或查看二维码；等待维护者在私有终端扫码及回传末尾结果行。该授权不覆盖 Wi-Fi，不重复有线密码窗口。
- QR 窗口收尾：回执 fail，初始 offline 通过，QR 登录返回 protocol_changed；维护者确认已展示、扫码并确认。Agent 完成只读 bundle/步骤/文件绑定复核，无 cleanup、无最终状态证据，详见 [QR 记录](../docs/research/p5-readiness-2026-11-01.md#有线正式-terminal_qr-失败回执)。单次授权已消耗，不重跑、不自动注销或进入 Wi-Fi；下一依赖是失败阶段与网页确认契约的证据，正式 Candidate 保持冻结，既有通过回执不改写。
- 执行顺序调整：维护者要求暂缓 QR 并明确批准一次 Wi-Fi 辅助、仅完整通过后一次密码套件。已在共享 D 盘现有私有运行根补齐同字节 14 文件 Candidate，不覆盖旧文件或 evidence；入口 `build/p5-preparation-20261101/run-wifi-window-once.ps1` 和本地 `WIFI-OFFLINE-HANDOFF.md` 已准备，语法及 CheckOnly 文件校验通过。Agent 未查询校园状态、切网或开始测试；维护者按说明断开有线并在私有终端独立执行，恢复管理连接后回传脱敏结果。QR 失败及发布必需性保持，当前 Wi-Fi 尚未验证。
# 2026-11-01 双 SSID 工作包开始

维护者已批准 [ADR-0026](../docs/architecture/decisions/ADR-0026-bound-scan-dual-wifi.md) 对应工作：先规范，再定向扫描、SSID 核对、证据迁移及离线开发包。QR 暂缓；不执行校园请求、不改网络配置、不提交或生成 Candidate。现有四份文档修改与缓存保留。Wi-Fi 辅助首次运行在 selected_address_unique blocked，password 未启动；旧包和回执保留。当前只进入离线实施，不声明完成。

组件交接：定向扫描、Native Wi-Fi 读取及 Snapshot SSID 比较已实现；组件回归、Linux race/vet/fuzz 和文档检查通过，Windows 安装长路径失败及短路径复核单列。证据版本迁移、promotion、预检及开发包仍未完成，下一步继续这些依赖，不请求真实复测。具体结果及剩余项见[实施记录](../docs/research/wifi-bound-progress-2026-11-01.md)。

后续离线交付：上述剩余链路现已完成，最终两端验收、skip/环境限制、开发包身份及本地零网络 blocked 回执均见[实施记录](../docs/research/wifi-bound-progress-2026-11-01.md)。下一依赖是另行授权 NEU、NEU-2.4G 的开发窗口；QR 仍暂缓，后续统一新 Candidate，不跨包继承证据。分支/HEAD 未变化，未暂存/提交/推送或执行校园请求，未改网络/Clash。旧 handoff、四份先存文档修改、缓存及旧包保留。

NEU 首次维护者执行在零网络预检 blocked；维护者说明 TUN 未关，并明确要求同包再试一次。Agent 仅记录回执和核对启动器哈希，下一步仍是维护者私有终端窗口；范围及停止条件见[实施记录](../docs/research/wifi-bound-progress-2026-11-01.md)。自有路由器/自定义 SSID 需求只记录于 roadmap，留后续设计，不在本轮实施。

NEU 后续维护者回传已完成：最终同包开发窗口辅助与密码七步全部 pass，窗口结束时 offline，非正式 evidence。本次关闭 Clash 的环境说明及退出码未回传等证据边界已记录于[实施记录](../docs/research/wifi-bound-progress-2026-11-01.md)。NEU 单次窗口已结束，不重复运行；下一依赖为另行授权 NEU-2.4G 窗口。QR、正式新 Candidate 与 M3 状态不提升。本次仅更新执行记录与当前入口，不修改产品或网络配置。

NEU-2.4G 后续已获单次授权并由维护者完成：同包八项辅助与密码七步全部 pass，启动器 exit 0，窗口结束时 offline；两种 SSID 的开发验证结果现已明确。环境及浏览器门户跳转仅作脱敏观察，见[实施记录](../docs/research/wifi-bound-progress-2026-11-01.md)。两个窗口均已结束，不重试或扩大授权；下一发布依赖为 QR 收敛，再统一新 Candidate 及正式验收。本次仅更新文档，未操作校园网络、提交或发布。

2026-11-01 后续文档工作：维护者要求记录三网络手动试用及 ac_id 观察，并在开发更新时参考其他实现。已只读核对三个公开仓库并固定版本，结论及证据边界见[研究记录](../docs/research/ac-id-comparison-2026-11-01.md)，后续执行入口见[开发说明](../docs/operations/development.md#外部实现对照)。不改产品行为、不沿用真实授权；现有未提交修改与旧包保留。

2026-11-01 QR 工作开始及离线交付：维护者批准网页契约优先的计划。已新增 [ADR-0027](../docs/architecture/decisions/ADR-0027-browser-qr-contract.md)、研究矩阵与浏览器窗口约定，给旧回执追加合成步骤解释修正，并增加精确两路径同摘要的假执行器回归。没有修改产品、正式 schema 或旧回执；缺失官方 QR 脚本/真实状态关系，暂不生成修复包。后续入口与检查报告见[QR 记录](../docs/research/qr-browser-contract-2026-11-01.md)。未校园导航、扫码、status/logout、提交或发布；旧包和现有未提交改动全部保留。

2026-11-02 QR 后续：维护者在单次授权窗口完成无痕网页扫码、目标 online 核对、普通产品注销及最终 offline。已取得 login-qrcode.js 静态材料，发现无 ticket service 接续与当前 SDK ticket 要求的兼容差异；证据等级、历史失败不可唯一回溯及剩余接续材料见[QR 记录](../docs/research/qr-browser-contract-2026-11-01.md)。真实主动动作已结束，仅分析已有材料；不追加测试、不修改产品，正式 QR/M3 不提升。

2026-11-02 后续离线执行：维护者批准 [ADR-0028](../docs/architecture/decisions/ADR-0028-qr-service-continuation.md) 完整工作包。已完成共享接续修复、独立普通产品 QR 开发入口、两端合成验收、快照/绑定包与零网络预检；结果、skip 和环境失败复核见[实施记录](../docs/research/qr-continuation-progress-2026-11-02.md)。当前本地 preflight blocked，尚无新包真实 QR 结果。下一依赖为维护者测试路径重新就绪及新单次授权，不沿用浏览器地址/代理路径，不重复旧包测试。既有未提交 Wi-Fi 改动、旧 Candidate/回执和缓存均保留；本轮未提交或发布。

2026-11-02 窗口前门禁修正：维护者已授权并确认现场，但旧包零网络预检仍停止；只读定位为断开接口残留默认路由。按随后“请继续”完成 [ADR-0028 补充](../docs/architecture/decisions/ADR-0028-qr-service-continuation.md#断开接口残留默认路由修正)对应离线修复、两端回归、新包与本地 ready 预检。当前身份和私有终端步骤见[实施记录](../docs/research/qr-continuation-progress-2026-11-02.md#后续更新残留默认路由修复及当前可测试包)。没有改系统路由/代理或执行校园请求；下一步为维护者执行尚未开始的一次窗口，不沿用旧包启动器。旧 Candidate、历史回执和未提交改动保持。

2026-11-02 当前 QR 窗口收尾：维护者回传 QR 命令及独立目标身份 online 核对 pass，注销阶段门禁 blocked，最终会话未知；哈希和 218 项快照核验通过，详细事实及门禁可观测性限制见[窗口结果](../docs/research/qr-continuation-progress-2026-11-02.md#最新窗口登录通过注销阶段门禁停止)。单次授权已消耗，下一步仅离线核对门禁原因，不自动 status/logout 或重试；不提升正式 QR/M3。旧包、回执及已有改动保留，本轮未修改产品、提交或发布。

2026-11-02 门禁排查启动及收尾：按维护者“开始排查”仅进行本地只读快照比较、假命令复现及永久合成回归。原生 8 次查询稳定、假认证五步通过，两端重复回归/vet 和 Linux race 通过；未定位历史失败的具体项，未修改生产逻辑或真实会话。报告和待确认的首因诊断方案见[排查记录](../docs/research/qr-continuation-progress-2026-11-02.md#门禁排查未复现历史原因仍缺证据)。维护者报告账号仍在线；不复用已消耗的单次授权，不提交或发布。

2026-11-02 首因定位工作启动：维护者已批准独立 QR 开发报告扩展；按 [ADR-0028](../docs/architecture/decisions/ADR-0028-qr-service-continuation.md) 补充连续实施、回归、打包及零网络检查。当前账号在线为维护者报告，不执行真实状态查询、扫码或注销；旧包及现有改动保留。

2026-11-02 首因定位离线交付：按已批准 [ADR-0028](../docs/architecture/decisions/ADR-0028-qr-service-continuation.md) 补充完成实现、两端 QR/共享消费者回归、Linux race/fuzz、独立包绑定及零网络 ready 预检。当前入口、身份和历史限制见[交付](../docs/research/qr-continuation-progress-2026-11-02.md#当前交付门禁首因定位-v2)。普通产品字节未变，不修改旧包/回执；下一步仅为另行授权的真实私有窗口，需重新建立 offline。本轮未读取 profile、扫码、status/logout、提交或发布。

2026-11-02 QR 开发窗口收尾：维护者回传首因定位 v2 包五步完整 pass、退出码 0，注销及最终 offline 均通过，无 cleanup 或门禁失败。Agent 仅保存脱敏回执、核验四项包哈希/225 项快照并用现有解码器严格读取；详见[闭环记录](../docs/research/qr-continuation-progress-2026-11-02.md#后续窗口qr-开发闭环通过)。单次授权已消耗；旧门禁失败原因仍未知，历史包/回执及所有未提交改动保留。下一步依赖为统一源码审阅及新 Candidate 准备，正式里程碑保持，未追加校园操作、提交或发布。

2026-11-02 统一 Candidate 准备启动：维护者要求开始下一步；仅核验 Wi-Fi/QR 未提交源码、现有覆盖和发布依赖，准备审阅清单及提交检查点。沿用当前 roadmap，不修改历史交接，不提交/推送/合并，不构建正式 Candidate 或执行校园请求。

2026-11-02 统一源码审阅收尾：两端干净快照集成/vet/doccheck、Linux race、发布工具合成门禁和脱敏检查已通过；既有产品跨平台与 config 管理员覆盖的源码一致性已核对，不重复无关安装。当前精确清单、差异、PR 草稿及下一授权边界见[准备记录](../docs/research/unified-candidate-preparation-2026-11-02.md)。拟从与当前 HEAD 同树的固定 main 创建新分支，保留全部未提交改动及排除缓存；尚未创建分支、暂存、提交、推送或构建正式 Candidate，没有校园请求。

2026-11-02 已授权 Git 检查点执行：原 99 文件清单已创建新分支、提交和推送，PR #30 首轮 CI 为 12/13，通过及 Windows 合成夹具失败详情见[准备记录](../docs/research/unified-candidate-preparation-2026-11-02.md#后续pr-30-与-windows-合成夹具修正)。当前仅补齐离线测试隔离与拒绝分类，生产路径和预算保持；追加提交需新清单授权，不合并、不创建 Candidate 或执行校园操作。四个缓存文件与旧包/回执保留。

2026-11-02 CI 追加修正离线收尾：Windows labpath 合成全包/vet/doccheck、Linux 同包/race/vet/doccheck 已通过，无测试级 skip；结果与剩余远端验证边界见[准备记录](../docs/research/unified-candidate-preparation-2026-11-02.md#后续pr-30-与-windows-合成夹具修正)。当前四文件补充清单包含一份 Windows 测试和三份执行文档，尚未提交/推送；待精确清单授权后更新 PR #30 并复核 CI，不重跑校园窗口。

2026-11-02 签名替代执行及 main 失败调查：已授权 PR #31 签名替代、CI、标准合并及关闭旧 PR #30 完成，分支和历史保留。合并后新 Windows 测试失败及本地修正依据见[准备记录](../docs/research/unified-candidate-preparation-2026-11-02.md#后续签名替代合并及-main-身份测试失败)。当前执行身份仍为原本地分支，未强制切换；仅补齐测试回归和新的审阅清单，不自动追加提交、Candidate 或校园操作。正式状态以 status 为准。
