---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# 双 SSID 离线交付与开发验证结果

2026-11-02 后续：QR 开发闭环也已通过，当前转入[统一 Candidate 源码准备](unified-candidate-preparation-2026-11-02.md)；下文 11 月 1 日“QR 尚未闭合”等表述保留为当时状态。两份 Wi-Fi 开发回执身份不变，不移作新 Candidate 的正式证据。

依据 [ADR-0026](../architecture/decisions/ADR-0026-bound-scan-dual-wifi.md)。本页记录实现和验证，不替代规范。离线实现与开发包已交付；NEU、NEU-2.4G 开发窗口均已回传通过结果，本工作包的双 Wi-Fi 开发验证已闭合。下述离线检查的 skip/环境限制继续保留，正式验收未完成。QR 尚未闭合，M3 保持 in_progress。

## 当前交付（2026-11-01）

- network scan --scope all|bound 已实现，默认 all 保持原行为；bound 显式绑定、唯一匹配，只有选中地址可进入 Status。公共 network list 不过滤。
- 正式认证/摘要 v3、辅助 v2、promotion v3 已贯通。NEU/NEU-2.4G 独立身份与四组认证/三份辅助门禁有正负合成覆盖；辅助绑定选定接口 index，只调用 bound 扫描。预检与辅助复用同一完整列表/选定项校验，防止预检忽略随后辅助必拒绝的本地材料。
- 历史认证 v1/v2、辅助 v1 及 promotion v1/v2 有显式只读审计，不补造 SSID，不写当前摘要；旧矩阵与 release-notes 语义保留。Candidate/release manifest 仍为 v1。
- 独立 wifi-dev-launch 使用普通开发产品、正式辅助逻辑和原 password_core 状态机，不使用 observed CLI；仅辅助全 pass 才进入密码套件。包前后绑定、选定 SSID、命令期间监测、窗口末尾复核、取消/硬期限及私有密码规则已接线。
- 零网络预检具备独立进程硬期限和闭集输出。不会读取认证 profile、调用 status、连接 Wi-Fi 或读取无线密码。预检不是账号 offline 证明。

## 最终开发包

目录：`build/dev-runs/windows/20261101T090253Z-1af89a9a-wifi-dev/ready/`。

| 身份 | SHA-256 |
|---|---|
| source_digest | 718583a7b888cbf49a2acc5717baaa551907daf53815c032b7ba44571964d6ce |
| 普通产品 | 700c25ac6718c9b4668f709a455c82e0b3a21c7d87a325b1bd42bff51a02b202 |
| 启动器 | bd62526da18dd36a359f4bcf806ce6e90ebc702e56158223660b6697a165a82e |

快照、worker、产品、启动器和 build-inputs 绑定、自检以及缺失/额外/篡改文件拒绝通过；详见同次 result.json、binding-audit.json、ready/build-receipt.json。普通产品版本 dev-wifi-local；release_evidence=false，不冒充 Candidate。两份独立说明为 ready/neu-window.md 与 ready/neu_24g-window.md，包含固定包命令、前提、停止及恢复步骤。旧包和本轮中间包保留。

正式审批工具账户 bhkdesktop/unbal 在 BHKDesktop 执行本包一次 `--preflight neu`：binding=verified，blocked/topology_unavailable，exit 11；network_requests=0、profile_reads=0、live_attempted=false，辅助/密码均未启动。回执为同目录 local-preflight.json。只能证明当前工具上下文未取得符合规则的 Wi-Fi 路径，不据此断言 Clash 是根因、账号状态或实际 SSID 权限已验证；不重复同包同条件检查。

## 集成验收

| 检查 | 结果与报告（build/dev-runs 下） |
|---|---|
| Windows 全仓普通、诊断、安装回归 | pass：windows/20261101T085809Z-e327c8d0-test；使用既有短测试根，未改真实安装/PATH/ACL |
| Windows 全仓 vet | pass：windows/20261101T090033Z-ba8128cf-vet |
| 最终 Windows schema/历史/辅助/路径 tests+vet | pass：windows/20261101T090237Z-f76e741d-p4 |
| 最终 Windows Wi-Fi tests/vet/self-test/build/binding | pass：windows/20261101T090253Z-1af89a9a-wifi-dev |
| Linux 全仓 test/race/vet | pass：linux/20261101T085902Z-11a8e701-test、20261101T085915Z-6a9a02c2-race、20261101T085932Z-142bf367-vet |
| 最终 Linux Wi-Fi 与 P4 tests/race/vet/fuzz | pass：linux/20261101T090252Z-96b39c64-wifi-dev、20261101T090310Z-f181061e-p4 |
| 六平台三入口构建及产品依赖隔离 | pass：linux/20261101T085938Z-c7ddd49c-cross；没有 Windows 诊断、SSID 接口或 JS 运行时进入产品闭包 |
| promotion/milestone/Candidate 合成、actionlint、gitleaks | pass：linux/20261101T090611Z-2606f86f-quality；最终源码和交付文档已扫描 |
| doccheck / 差异 / 快照一致性 | pass：windows/20261101T090615Z-46e5e6e1-docs；git diff --check 通过，257 项包源码与当前文件逐项一致 |

上述全仓入口的“pass”指命令成功，不隐去跳过：本轮 Windows 五项需 symlink 特权的安全测试 skip，TestNativeReleaseAssetSmoke 因没有传入正式 Candidate 资产而 skip；其既有 P5 原生通过记录仍只绑定旧源码/旧 Candidate，本轮没有重跑或冒用。无测试文件的包级 skip 另计，不是失败用例被删除。安装长路径失败及后续短路径复核见下方历史记录。macOS 这里只是交叉构建，不是原生认证验收。最后改动集中在内部预检/历史审计，其后已补相应两端回归，无需无条件重复不受影响的产品构建。

## 较早组件检查记录（保留）

- 公共 CLI 增加显式 network scan --scope all|bound；默认全扫保持。bound 本地唯一匹配、显式 IPv4、取消和错误分类回归覆盖；只向匹配项调用原 Status。
- labpath 增加 Native Wi-Fi 当前连接读取：通过接口 index → LUID → GUID 关联，读取真实 SSID；只输出校园闭集标签。NEU-Mobile 拒绝进入验收；未知、权限拒绝、断开、畸形长度均拒绝。取消在调用前后检查，不遗弃后台调用；完整预检进程硬期限仍待接线。
- Wi-Fi Snapshot 包含选定 index 和 SSID，原有周期/命令边界快照比较可检测 SSID 变化。没有主动调用真实 Wi-Fi 读取或发出校园请求。

## 已执行验证

- Windows 普通测试和独立诊断测试通过；本轮初次新测试类型错误及帮助契约断言已修复，失败日志保留。
- Windows p4-tests/p4-vet 通过：build/dev-runs/windows/20261101T083805Z-c2ec7c8c-p4/result.json。
- Linux internal/cli、internal/labpath 的普通/race 与 vet 通过，使用固定本地 Go 1.25.0、GOTOOLCHAIN=local、GOPROXY=off。
- 新 ADR 使用 doccheck 既有生成入口更新 stable-ids；历史 handoff 原件未修改。
- Windows 完整 test 入口的普通/诊断分组通过，安装分组在默认长临时路径的 active junction 执行失败，不能称全仓入口全绿。使用既有短测试根仅复测 fresh/rollback 失败项通过：build/dev-runs/windows/20261101T084135Z-99e44115-install/result.json；未改安装代码、ACL 或断言。原失败报告保留于 build/dev-runs/windows/20261101T083839Z-a410db78-test/result.json。
- Linux 纯内存 SSID 解析 10 秒有界 fuzz 通过；文档检查 build/dev-runs/windows/20261101T084109Z-1392b655-docs/result.json 通过。仍需最终集成后的完整验收，不能以组件覆盖宣称计划完成。

## 剩余事项与下一窗口

维护者原生终端第一次执行最终包 --run-once neu 后回传：binding=verified，blocked/topology_unavailable，exit 11；live_attempted=false，network_requests=0、profile_reads=0、auxiliary_checks 为空、password=null。维护者随后说明忘记关闭 TUN，并要求同包再试一次。仅据该报告不能确定 TUN 是唯一根因；没有开始校园请求或认证。新一次范围仍仅 NEU，关闭 TUN、断开有线、目标账号 offline/独占后由维护者私有终端执行；预检或辅助失败即停止，不自动重试。

维护者提出的自有路由器/自定义 SSID 支持已纳入 [roadmap](../upgrade/plan.md) 后续需求，本轮只记录，不改白名单、拓扑规则或包身份。SSID 限制属于验收层，不代表公共产品只支持这两个网络名称。

### NEU 开发窗口回执（2026-11-01）

维护者在原生私有终端重新执行同一最终包并回传 JSON：binding=verified、state=pass、reason=none、wifi_ssid=neu、live_attempted=true、release_evidence=false。源码摘要为 `718583a7b888cbf49a2acc5717baaa551907daf53815c032b7ba44571964d6ce`，普通产品 SHA-256 为 `700c25ac6718c9b4668f709a455c82e0b3a21c7d87a325b1bd42bff51a02b202`。此记录来源是维护者回传，不是 Agent 独立执行或重新认证。

- 预检 ready：local_items=5、selected_items=1；本次场景覆盖了存在其他本地地址时唯一选择 Wi-Fi 的路径。
- 八项辅助检查全部 pass：selected_wifi_topology、list_complete、selected_address_unique、bound_status_valid、scan_scope_valid、scan_complete、selected_result_valid、topology_unchanged。
- password_core 七步全部 pass，均 exit 0：initial_status_offline、login_logged_in、status_online、second_login_already_online、logout_logged_out、final_status_offline、second_logout_already_offline。product_commands=7、step_coverage=complete，无额外 cleanup。
- 最后两个状态步骤证明该窗口结束时 offline，不替代未来窗口开始时的状态检查。
- 回传未包含 `$LASTEXITCODE` 的数字；仅记录 JSON pass 和各产品步骤 exit 0，不补造启动器退出码，也不因此要求重跑。
- 预检的零请求计数仅描述预检。普通产品密码套件中的默认观察计数和标记不代表整场零请求，不能用于否定实际登录、注销或推算系统 DNS/PKI 流量。

维护者说明本次关闭了 Clash，原因是此前浏览器也无法进入门户。将其记录为本次环境事实；尚未隔离 TUN、系统代理、DNS 或其他因素，不宣称关闭 Clash 是所有环境的必要条件，也不将上一回执的 topology_unavailable 唯一归因于 TUN。本轮未修改代理或网络配置。

### NEU-2.4G 开发窗口回执（2026-11-01）

维护者获得独立单次授权后，在私有终端执行同包 `--run-once neu_24g`，回传 binding=verified、state=pass、reason=none、wifi_ssid=neu_24g、live_attempted=true、release_evidence=false，并明确回传启动器退出码 0。源码摘要和普通产品哈希与上述 NEU 包完全相同。

- 预检 ready，local_items=5、selected_items=1；上述八项辅助检查全部 pass。
- password_core 七步全部 pass，各步骤 exit 0/error_code=null；依序用时为 1、11、1、1、1、1、1 秒。product_commands=7、step_coverage=complete，无额外 cleanup。
- 最后两步确认窗口结束时 offline。维护者说明本次仍关闭 Clash；不据此推导其他代理配置下必定失败或要求所有使用者关闭 Clash。
- 证据级别为维护者回传开发回执；预检零请求与普通产品未启用观察的计数含义同上。未生成正式 evidence，未重跑认证。

维护者另报告，连接开放的两个校园 SSID 时，浏览器会自动跳转到 HTTP 门户 `/index_15.html`，链接包含客户端/接入设备信息及连通性检测重定向上下文。本记录只保留路径与行为分类，不保存完整 URL、查询值或设备标识。该观察不证明 SDK 使用了此门户、不证明浏览器与 SDK 同路径，也不证明该 HTTP 页面承载了凭据提交；Agent 未访问或重放链接，不据此增加 HTTP 认证或改变 HTTPS 安全约束。

### 下一依赖与未验证项

后续维护者使用普通开发产品的 `ipgw-dev` 临时命令，反馈有线、NEU、NEU-2.4G 手动登录/注销均无问题；浏览器分别观察到 ac_id 为 1、15、15。该补充没有逐命令回执，不冒充受控七步或正式证据；完整证据边界及外部源码对照见[三网络手动记录](ac-id-comparison-2026-11-01.md)。

两个单次开发窗口均已结束，不自动重试。NEU 与 NEU-2.4G 场景已实际通过 Native Wi-Fi 当前连接读取、选定项扫描及完整密码闭环；不泛化为所有账户/权限环境均可用。自定义 SSID、NEU-Mobile 仍未实现或验收，当前没有新正式 Wi-Fi pass。下一发布依赖为 QR 问题收敛；不把本次 Wi-Fi 授权沿用到 QR。

QR 未闭合前不生成新正式 Candidate。所有旧 Candidate/回执/历史交接及原有未提交修改保留；后续新 Candidate 必须重新取得完整正式集合，不能继承本开发包或旧 Candidate 的通过记录。
