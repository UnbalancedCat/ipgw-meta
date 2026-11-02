---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
status: approved
---

# Wi-Fi 与 QR 统一 Candidate 源码准备

本页记录 `WP-AUTH-CANDIDATE-ACCEPTANCE` 的源码审阅准备，不生成 Candidate 或正式证据。当前路线以[roadmap](../upgrade/plan.md#roadmap-20261029)为准；[双 Wi-Fi 开发结果](wifi-bound-progress-2026-11-01.md)和[QR 五步闭环](qr-continuation-progress-2026-11-02.md#后续窗口qr-开发闭环通过)已明确，单次授权均已消耗。

## 后续：签名替代合并及 main 身份测试失败

维护者授权的签名替代已完成：[PR #31](https://github.com/UnbalancedCat/ipgw-meta/pull/31) 从固定 main 创建 GitHub 签名提交 `9f695d46763f3fdece02e16892ad1386387c87d2`，13 项检查全部通过后标准 squash 合并。新 main 为 `f56c1ded009b36f3da8a1c16608f524e4a38f08b`，签名 valid、父节点为 `1a32aabc2aa3125afe638b0dc3d3444bf7a95c7d`，文件树精确等于已审阅的 `9d58bc00b5c09a9a36f6e8ca1697cbe19b68b81a`。PR #30 已关闭为被替代，原分支与提交保留；未强推、修改保护或安装签名密钥。此前本地提交未签名是准备检查遗漏，不能以 PR CI 通过替代签名检查。

合并后的 [main CI](https://github.com/UnbalancedCat/ipgw-meta/actions/runs/41452754606) 与[原生安装](https://github.com/UnbalancedCat/ipgw-meta/actions/runs/41452754779) 合计 12/13 通过；Windows 的 `TestPublishCandidateDirectoryRejectsReplacedStageIdentity` 失败。它不是此前 PowerShell 超时，未自动重跑；新 Candidate 保持阻塞，正式认证及 M3 不提升。

离线调查确认测试使用 `os.Lstat`，而生产 `Assemble` 从 `seal.rootDirectory.info` 取得已打开句柄的身份。固定本地 Go 1.25.0 的 Windows 实现会对部分路径 FileInfo 延迟加载文件 ID；独立合成复现中，旧 Lstat 对象在路径被替换后与替换对象 SameFile=true，句柄快照为 false，原目录仍存在于移走位置。这证明测试的身份采集不等价于生产路径，不是生产门禁需要放宽。只修改发布测试使用相同快照，并令替换目录的大小、模式和修改时间相同，验证身份本身必须拒绝；生产代码、门禁及协议不变。

精确合并回执、复现、测试报告与后续差异位于 `build/p5-unified-preparation-20261102/`。修正仍为本地未提交内容；新提交/PR 及远端检查须单独检查点，不能把本地回归写作 main CI 已修复。首次 Linux 检查误把临时目录置于 Windows 挂载盘，相关失败保留；后续使用既有 Linux 原生临时目录，未修改测试判据来兼容错误的执行位置。

## 后续：PR #30 与 Windows 合成夹具修正

原 99 文件清单经维护者授权后，已从固定 main 创建 `codex/bhk-wifi-qr-closure`，提交 `35dd57183d667d1766d0223772a598e0052dbd60` 并推送为 [PR #30](https://github.com/UnbalancedCat/ipgw-meta/pull/30)。该提交父节点为 `1a32aabc2aa3125afe638b0dc3d3444bf7a95c7d`，PR 精确包含一个提交、99 文件；四个未跟踪 Python 缓存保留且未提交。下文未提交描述保留为该授权之前的检查点。

首轮 [CI](https://github.com/UnbalancedCat/ipgw-meta/actions/runs/41440505354) 与 [原生安装](https://github.com/UnbalancedCat/ipgw-meta/actions/runs/41440505329) 合计 12/13 项通过：六平台原生安装、打包、Linux/macOS 测试、race、交叉构建及文档/vet/脱敏均通过；Windows 全仓测试失败于 `internal/labpath` 的合成 PowerShell 测试。两个固定分类及三个接受场景受到 15 秒查询期限影响，不能将超时当成对应规则的验证通过。

新 `syntheticTopology` 夹具遗漏了既有 `authdev`、`casacquire` 测试采用的模块隔离：关闭测试进程内自动模块发现，仅从固定 `$PSHOME` 路径导入系统 Utility 模块。所有网络 cmdlet 继续由合成函数覆盖。补充修改只影响测试；生产拓扑查询、15 秒期限、停止规则和产品字节均不变。同时将拒绝矩阵从“任意错误即通过”收紧为精确 `local_rule_rejected` 或 `local_query_failed`，避免超时掩盖未执行的规则。

CI 日志证明发生了超时，但没有直接测量镜像模块发现耗时；该因素与既有夹具处理的差异构成修复依据，远端是否消除失败仍须追加提交后的 CI 验证。不能将此次测试夹具问题解释为此前真实 QR 门禁瞬时失败的已证明原因。追加差异及验证材料保存在 `build/p5-unified-preparation-20261102/ci-followup/`；追加提交/推送需新的精确清单授权，合并、Candidate 和校园操作未获本次授权。

追加修正的本地验收：Windows `windows-20261102T121806Z/result.json` 中 labpath 全包合成测试、vet、doccheck 通过，43 个测试/子测试 pass 事件，无 skip；Linux `linux-20261102T121810Z/result.json` 中同包普通/race/vet/doccheck 通过，普通测试 13 个 pass 事件、无 skip。Linux 不执行 Windows 专属脚本测试。此次仅测试夹具及三份状态文档有差异，产品及生产工具源码未改，不重复已通过的原生安装或六平台构建。新的远端 Windows CI 仍未执行，不能以本地通过替代。

## 基线与拟提交范围

2026-11-02 本地仍为 `codex/bhk-v1-acceptance`，HEAD `c21094ab3e7b2a4064b66ae82cf2fbc2173f92cd`，暂存区为空。GitHub 只读查询 main 得到 `1a32aabc2aa3125afe638b0dc3d3444bf7a95c7d`，与本地 origin/main 一致；两者提交身份不同，但 HEAD 与该 main 的文件树完全相同。此次没有重新核验远端保护、required checks、tag 或 attestation；相关检查仍须在合并/正式构建检查点进行。

拟从上述 main 创建新分支 `codex/bhk-wifi-qr-closure`，保留当前所有工作区改动，仅按逐文件清单提交。文件树相同是无覆盖切换的前提，操作前必须再次确认；如发生漂移则停止，不 reset、强制 checkout 或强推。旧分支、Candidate、开发包和回执保持。

| 范围 | 实际变化与审阅重点 |
|---|---|
| 产品扫描 | [ADR-0026](../architecture/decisions/ADR-0026-bound-scan-dual-wifi.md)：显式 bound 扫描只选择唯一绑定项；默认 all、公共 SDK/JSON/退出码及认证判据保持。 |
| 正式验收 | 认证/摘要/promotion v3、辅助 v2，四组认证与三份辅助，SSID/接口绑定、旧版本显式只读核验；不补造历史字段。 |
| QR 接续 | [ADR-0028](../architecture/decisions/ADR-0028-qr-service-continuation.md)：精确无 ticket service 标识、一次 HTTPS 网关指示、同轮 CAS、ticket 截获与 HTTPS activation；严格拒绝/预算/最终身份回归。 |
| 开发工具与路径 | 独立 Wi-Fi/QR 开发包、包与源码绑定、私有 TTY、拓扑停止、QR 命令分别记录及首因定位；不以开发工具替代正式 helper。 |
| 文档与证据解释 | 当前入口及执行记录；保留旧失败、历史 Candidate、原始交接。未把用户脚本、账号、IP、二维码、Cookie、ticket 或原始响应放入提交。 |

审阅产物位于 `build/p5-unified-preparation-20261102/`：`review-manifest.json` 为精确路径/大小/SHA-256 清单，`review.diff` 包含修改与新增文件，`source-snapshot.json` 记录检查用源码集合，`PR-BODY.md` 为 PR 草稿，`COMMIT-CHECKPOINT.md` 为下一授权步骤。`source-review/` 是仅供检查的临时源码快照，不是第二份权威仓库。构建目录、缓存及私有 evidence 不纳入清单；未跟踪 Python 字节缓存保留并排除，不删除。

## 验证与覆盖解释

本次在只包含跟踪源码和拟新增文件的快照上执行相关跨组件检查，防止本地缓存或遗漏的未跟踪文件掩盖干净检出问题。固定 Go 1.25.0，GOTOOLCHAIN=local、GOPROXY=off，隔离测试配置与临时目录；测试采用合成请求/loopback，不把这些环境变量宣称为操作系统网络隔离。

| 检查 | 结果与本地报告 |
|---|---|
| Windows 干净快照集成测试、vet、doccheck | pass；`build/p5-unified-preparation-20261102/windows-20261102T110701Z/result.json` |
| Linux 同一快照测试、race、vet、doccheck | pass；`build/p5-unified-preparation-20261102/linux-20261102T110704Z/result.json` |
| promotion、milestone、Candidate check-run 合成；规则测试、actionlint、历史与工作区脱敏 | pass；`build/dev-runs/linux/20261102T111007Z-dcb41c1a-quality/result.json` |
| 既有完整源码普通/vet、Linux race | 保留 [QR 交付报告](qr-continuation-progress-2026-11-02.md)；不把当时 skip 计为 pass。本次补检覆盖后续共享路径与独立工具变化。 |
| 既有六平台三入口构建及产品依赖隔离 | 保留 `build/dev-runs/linux/20261102T053201Z-3564395f-cross/result.json`；此后产品源码/依赖没有变化，门禁与开发工具未进入产品依赖闭包。macOS 不是原生认证验收。 |
| Windows 五项 config symlink | 原生管理员五项实际 pass 回执为 `build/p5-preparation-20261101/00991b8b8061474091e20c427f4a0b22.result.json`；相关 config 及依赖源码未改。本轮未重跑或称其为本轮结果，普通令牌 skip 记录保留。 |
| 安装与正式 CI | 安装代码及套件未改，本轮不重复既有安装回归；新 PR required checks、新 Candidate 六平台原生安装及 BHK 安装辅助仍待执行。 |

本次集成检查不等于对全部历史诊断代码重新进行独立安全审计，也不能替代 PR CI。新临时快照与最终清单的源码一致性、文档检查及扫描汇总见同目录 `final-review.json`。

覆盖审计 `coverage-audit.json` 进一步确认：本轮 Windows 记录 1078 个测试/子测试 pass 事件、Linux 普通与 race 各 955 个，所选集成范围无测试级 skip；这些事件数不代表独立功能数。21 项 config 测试依赖源码与 main 基线一致，61 项产品源码与既有六平台构建覆盖一致；本轮没有重复管理员 symlink 或六平台产品构建。临时覆盖脚本首次因 Python Windows 默认 GBK 解码 Go JSON 失败，显式 UTF-8 后核验通过，未改产品或测试判据。

## 剩余检查点与停止条件

1. 维护者审阅精确清单、差异和 PR 草稿后，单独授权创建新分支、按清单提交、推送及创建 PR。提交前复核 main/HEAD/工作区/清单身份，禁止混入缓存或额外变动。
2. PR required checks 全部通过并完成审阅后，单独授权合并；再次核验保护、签名和合入文件树，不绕过保护。
3. 固定新的受保护 main source，单独授权一次正式 Candidate workflow；核验冻结产物、manifest/checksums、六平台原生安装、artifact digest 和 attestation。不修改旧 Candidate 或单独替换 helper。
4. 同一新 Candidate 重新取得安装辅助、有线 password、有线 QR、NEU 辅助/password、NEU-2.4G 辅助/password；每个真实窗口重新确认现场与授权，密码/二维码只在维护者私有终端处理。
5. 人工审阅公开摘要，准备 promotion v3 与精确 release-notes；最终签名 tag、晋升及公开 release 保留独立授权。

旧 QR 门禁瞬时失败原因仍未知，本轮 pass 不证明原因已消除。需要扩大认证/信任/扫描契约或出现来源漂移时停止相应动作并报告，不为得到 pass 放宽门禁。正式 QR 未取得新 Candidate 证据，M3 保持 in_progress；NEU-Mobile、自定义路由器 SSID 及跨平台真实认证均未被本轮验证。
