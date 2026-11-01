---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# BHK 正式验收 v2

## REL-BHK-GATES-001：政策与切换

[ADR-0025](../architecture/decisions/ADR-0025-bhk-formal-acceptance.md) 已批准；实施状态见 [status](../upgrade/status.md)。P4 全部验收前不得宣称工具已完成切换。日期明确的 v1 回执及旧包不重写。以下是 v2 的权威增量契约；未变的状态机、持久化、安全和 Candidate provenance 契约继续适用。

新发布要求三项 windows-amd64/bhk_windows：campus_wired/password/password_core、campus_wifi/password/password_core、campus_wired/terminal_qr/terminal_qr。可采集集合另外包含 Windows Wi-Fi QR 以及 linux-amd64/nas_vm/campus_wired 的 password、QR。补充项不进入三项必需 ID 集合。

## EVID-BHK-002：版本与绑定

私有 evidence 与公开摘要 schema_version=2，在原字段之外必需 runner_sha256（64 位小写十六进制）。步骤、能力前后数组、时间及 ID 规则不变。字段集合按版本严格校验；v1 不接受新字段或 Windows QR，v2 缺 runner hash 拒绝。历史读取与当前生成/晋升验证分离，新 publisher 不生成 v1。

runner manifest projection 在旧必需字段外读取 test_tools，精确两项 Linux/Windows helper（固定名称、平台、正整数大小、SHA-256），仍拒绝各层重复字段和非法边界。实际运行文件必须匹配对应 helper；开始前和发布前复查文件身份、私有元数据、大小和哈希。初始失败 exit12 无 bundle；结束漂移 exit13 无有效 bundle。不读取 profile 以建立尚未成立的身份。

历史 bundle 审计只读三文件、v1 schema/摘要/checksum；历史 lock 审计只读 v1 旧四组和旧 release-notes。仅输出固定结果与检查范围，不联网、不改写、不产生 Actions 发布输出；缺少本地 provenance 材料时不得声称完整来源复验。

## REL-BHK-PATH-001：原生测试路径

BHK `ipgw-live-gate run` 必需 --bind-ip IPv4，沿用既有其余参数。所选 IPv4 必须唯一对应声明网络类型的活动硬件接口；唯一 IPv4 默认路由属于该接口，拒绝两条半默认覆盖路由。允许不影响测试路径的虚拟接口，另一类物理测试接口不得同时活动。保留有界接口/路由/DNS 快照；每命令前及运行期间复核，变化 sticky 停止，包括 cleanup。不修改接口、路由、DNS、代理或防火墙。

全部产品命令（含私有 TTY QR）透传同一绑定。网络/拓扑信息仅内存处理，不写回 evidence；系统 DNS/PKI 仍可能联网和更新缓存，不承诺逐目标路由锁或 OS 总流量计数。password 七步、QR 五步和清理所有权保持。NAS 的既有隔离实验室规则不变，不套用 Windows 探针。

## EVID-BHK-AUX-001：辅助验收

辅助记录独立 schema_version=1，kind 仅 bhk_windows_native_install 或 bhk_wifi_network；绑定 plan/revision、Candidate ID、manifest hash、source commit、build-input、验证器身份、原生执行上下文、时间、固定步骤和 pass/fail/blocked。缺失、重复、skip、未知步骤或截断不为 pass。私有 bundle 采用 evidence.json、summary.md、SHA256SUMS 的原子发布；公开 canonical 摘要仅保留闭集结果与必要哈希，不包含账号/IP/接口名/index/路径/原始错误，不使用真实地址哈希替代脱敏。

辅助 JSON 为紧凑 UTF-8、字段按以下顺序、末尾一个 LF：`schema_version, plan_id, revision, summary_id, kind, candidate_id, candidate_set_sha256, source_commit, build_input_sha256, runner_sha256, execution_context, result, started_at, finished_at, checks, verifier`。`summary_id` 为 `AUX-YYYYMMDD-NNN`，日期等于 started_at 的 UTC 日期，NNN 为 001–999。execution_context 仅 windows_native。公开摘要在末尾增加 bundle_sha256，沿用三文件按 ASCII 名升序的 `<name>NUL<decimal-size>NUL<sha256>LF` 摘要算法；没有 steps 正文或原始输出。checks 每项仅 name/result，失败记录只能是完整通过前缀加一个 fail/blocked 末项，pass 必须包含全部固定步骤。

安装步骤顺序为 candidate_verified、fresh_install、three_entry_versions、active_junction_execution、fresh_default_meta、upgrade_preserves_launcher、basic_rollback、private_layout、real_user_path_unchanged、candidate_unchanged。verifier 为 source_tree、snapshot_sha256、go_version、test_binary_sha256；源码 tree、快照及测试程序的三个哈希身份来自本地验证，go_version 精确 `go1.25.0 windows/amd64`。pass 必须具备 verifier；准备失败允许 null，不假造缺失身份。

Wi-Fi 步骤顺序为 candidate_verified、selected_wifi_topology、list_complete、selected_address_unique、bound_status_valid、scan_scope_valid、scan_complete、selected_result_valid、topology_unchanged、candidate_unchanged；verifier 必须 null，执行者绑定由 runner_sha256 提供。selected_wifi_topology 同时要求私有 profile 的目标身份可读且初始 status 为 offline；只读命令不调用密码提供器。

安装验证器从 Candidate source 的 Git blobs 生成隔离临时快照（非权威仓库），核验 source/tree/build-input。固定本地 Go1.25.0，GOTOOLCHAIN=local、GOPROXY=off、GOWORK=off、GOENV=off，拒绝外部 GOFLAGS/overlay 注入。编译并绑定测试程序哈希，记录快照与工具链；本地验证器不冒充 Candidate attested 资产。强制 Candidate 模式和 required，使用 Candidate installer/产品；固定安装测试及子测试必须完整执行且通过，不能仅凭 exit0。隔离 HOME/配置/临时目录，使用既有测试 PATH 替身，验证真实用户 PATH 未变；覆盖 fresh、三入口、active junction、默认 meta、upgrade 保持、基本 rollback 和私有布局。不能绕过 junction 执行。

安装总期限 15 分钟，测试程序自身 10 分钟；测试 JSON 输出最多 8 MiB，单事件最多 1 MiB，stderr 最多 64 KiB，仅内存处理。Windows helper 使用退出时关闭的 Job，退出后终止残留子进程；保护建立失败停止。Git tree 最多 10,000 文件、每 blob 16 MiB、累计 256 MiB；拒绝链接、路径越界和大小写冲突。固定测试为 TestNativeReleaseAssetSmoke 及 fresh_install_and_launcher_default、upgrade_preserves_launcher、basic_rollback_restores_baseline 三个子测试；每项 run/pass 和包级 pass 均必须出现，skip/fail/未知项/重复结束拒绝。退出码 0 不替代这些要求。

Wi-Fi 辅助检查使用普通 Candidate、固定所选地址和 offline 前提：network list 必须只含该一个地址；绑定 status 必须是有效 offline；再次核对枚举后才能 scan，结果必须完整、同一地址且无逐项错误。多地址时零 scan 停止。整个子进程有超时/输出上限及拓扑监测；不登录/注销或调用密码 provider。scan 的 --bind-ip 不作为过滤能力。此辅助检查不证明同窗口认证成功，真实执行在 P5 独立授权。

Wi-Fi 总期限 120 秒、每检查最多 30 秒，list/scan 输出最多 1 MiB，最多 256 个枚举项；扫描前只允许一项。拓扑监测每秒发起一次有界本地快照，快照读取自身最多 15 秒；这不是即时内核锁。每个产品命令开始前和结束后检查，检测到变化即取消并阻止后续命令。时间、输出或取消边界触发时不发布 pass。

## 维护者入口与人工审阅

仅使用同一 Candidate test_tools 内的 helper；以下是参数模板，P4 不执行任何真实采集。路径参数必须为绝对路径，不从环境猜测 Candidate。

```text
ipgw-live-gate check-install --candidate-root <目录> --out <私有输出目录> --repository <权威仓库> --go <Go1.25.0.exe> --module-cache <已准备缓存> --work-root <仓库之外的测试专用目录>
ipgw-live-gate check-network --candidate-root <目录> --out <私有输出目录> --candidate <普通产品EXE> --profile <私有prompt名称> --bind-ip <选定Wi-Fi IPv4>
ipgw-live-gate summarize --bundle <私有bundle目录>
ipgw-live-gate audit-v1 --bundle <历史v1私有bundle目录>
python scripts/promotion.py history --lock <历史v1锁> --notes <历史release-notes>
```

check-install/check-network 返回 0 表示完整 pass；1 表示已发布非 pass 记录；2 表示安全/环境/取消/持久化拒绝，无可用通过记录。summarize 只读并输出当前 canonical 摘要，不自动写入仓库；维护者仍须审阅 bundle 与摘要再选择闭集 ID。audit-v1/history 只读本地结构和哈希，不查询线上 artifact/attestation，不生成当前发布输出。真实 Wi-Fi 与三组认证必须在 P5 单独批准；安装辅助记录也须绑定 P5 新 Candidate，不能使用本轮合成安装测试代替。

## REL-BHK-PROMOTION-002：晋升

lock schema_version=2；evidence_ids 精确三项认证 ID，另必需 installation_summary 与 wifi_summary，各为 {summary_id,sha256}，ID 导出固定文件名，禁止任意路径。两摘要分别为安装/Wi-Fi kind、pass、同 Candidate/source/build-input；认证摘要 runner_sha256 必须等于该 Candidate helper。历史与补充摘要不自动参与晋升。Candidate/release manifests、资产集合、attestation、祖先及 promotion 白名单、无重建晋升保持。

精确 release-notes 改为 Windows 两种网络 password 和有线 QR；Linux/macOS 未有本 Candidate 校园证据时保持 live_unverified。辅助结果不改变 capability。M3 只有 P5 正式验收和发布条件全部闭合后才能完成。
