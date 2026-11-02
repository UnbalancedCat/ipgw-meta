---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# BHKDesktop 本机开发与合成测试

2026-11-01 Wi-Fi 增量入口：`python scripts/dev.py wifi-dev --context tool` 执行合成测试、vet、自检；Windows 构建独立 wifi-dev-launch 包并审计源码/文件绑定与篡改拒绝，Linux 加 race 和有界 SSID fuzz。`p4` 覆盖 schema/辅助/路径，`quality` 覆盖 promotion 及其他质量门禁，`cross` 检查六平台产品依赖隔离。依据 [ADR-0026](../architecture/decisions/ADR-0026-bound-scan-dual-wifi.md)；包内分别提供 NEU/NEU-2.4G 操作说明。`--preflight` 是零网络、本地只读；`--run-once` 必须另行单次授权和私有终端确认，合成入口不执行它。普通产品开发身份 dev-wifi-local，release_evidence=false，不使用 observed CLI 或冻结 Candidate 字节。

## 当前优先级（2026-10-29）

2026-11-01 当前 QR 工作：按 [ADR-0027](../architecture/decisions/ADR-0027-browser-qr-contract.md) 先完成网页契约对照，不直接复用 password observed CLI 或增加 QR 诊断体系。旧 runner 合成步骤的解释与针对性合成回归见[研究记录](../research/qr-browser-contract-2026-11-01.md)；真实浏览器和后续修复包验证分别授权。无确定产品缺陷时不生成新包或无条件重跑全仓构建。

唯一后续路线见[认证收敛 roadmap](../upgrade/plan.md#roadmap-20261029)，当前事实见[状态摘要](../upgrade/status.md#status-20261029)。Windows/WSL 开发基线及 P1 行为对照已建立；P2 入口与边界见[产品等价开发验证](development-validation.md)，不重复安装或环境初始化。

D4–D15 证书/CTL 专项退出默认开发主线，以下工具入口、命令和历史执行约定保留供追溯，不构成继续扩展或自动重跑的任务。只有直接阻塞产品认证的新证据才触发重新安排；真实单次授权不继承。

开发认证验证与正式 Candidate 验收分离、产品默认系统 PKI 观测按 [ADR-0017](../architecture/decisions/ADR-0017-development-auth-validation.md) 和 P2 验收记录推进；P4 的 v2 格式、helper 绑定、安装/Wi-Fi 辅助检查见 [BHK 正式验收](bhk-acceptance.md)，当前验收状态见 [P4](../upgrade/status.md#status-20261031-p4)。当前没有因此获得新的开发登录或网络许可，产品与当前正式门禁不变。

P2 合成入口为 `python scripts/dev.py auth-dev --context tool`（Windows 或同机 WSL）；Windows 阶段额外创建独立 `ready/` 包并运行启动器 `--self-test` 与篡改拒绝检查。`scripts/build-auth-dev.py --output ABS_NEW_BUILD_DIR` 必须在该入口提供的固定 Go/离线环境中使用；不覆盖旧目录。payload 不是人工执行入口，真实模式只从已核验的 `auth-dev-launch.exe` 发起。

`--preflight` 只读本地拓扑；`--anonymous-once` 和 `--password-core PROFILE` 是未来单次授权入口，均不在 `auth-dev` 合成阶段执行。password 模式只在维护者私有终端输入凭据；不要把该模式输出/交互转录当作无人值守测试。

## 外部实现对照

2026-11-01 维护者要求：后续开发更新将外部实现对照作为常规检查项。优先参考 neucn/ipgw、Neboer/ipgw-py-manager、7aaauTools/neu-login，必要时结合本地 legacy。首次 ac_id 对照及固定提交见[研究记录](../research/ac-id-comparison-2026-11-01.md)。

- 每次开发更新先判断相关性；涉及门户发现、CAS、注销、QR、网络选择或兼容行为时，核对相关仓库当前提交及对应源码/调用链。纯文档措辞或无关实现调整可记录“不涉及协议对照”，不机械重复抓取全部仓库。
- 记录查询日期、完整提交、文件/函数、与上次的相关差异，以及采纳/不采纳/未知的依据。不可访问时明确缺口或沿用的历史版本，不宣称完成最新复核。
- 外部代码、README 和 issue 是研究线索，不能替代 docs 契约、本地合成回归或单独授权的真实证据。借鉴行为前先检查安全边界与适用范围；复制代码前另查许可证及来源要求。
- 默认只读源码，不运行外部登录器，不复制凭据、公钥快照或原始认证材料，不自动下载/更新依赖，也不建立定时任务。本流程不增加真实校园请求授权。

## P4 离线开发入口（2026-10-31）

`python scripts/dev.py p4 --context tool` 验证 livegate/candidate/auxgate/labpath/helper；Linux 同入口增加 race 和有界 fuzz。它不调用 check-network、check-install 或真实拓扑。
Windows 安装回归可使用 `python scripts/dev.py install --context tool --test-temp-root ABS_PRIVATE_TEST_DIR`；目录须预先指定为仓库外测试专用路径，不改系统 TEMP。测试仅对新 fixture 收紧 DACL，不修改仓库 ACL、owner 或实际安装。正式辅助检查使用 Candidate 自带 helper，参数见 [维护者入口](bhk-acceptance.md#维护者入口与人工审阅)。

本轮 Windows 安装缺口定位为工具执行环境与测试准备问题：D 盘私有测试目录中 direct/junction 均可执行；重复写已经正确的 owner 会额外要求 WRITE_OWNER。测试先核对 owner 后只写 DACL，PowerShell 错误必须终止；原安装器检查和 junction 断言全部保留。完整合成安装回归通过不等于新 Candidate 原生安装验收通过。已有权限不足导致的符号链接测试 skip 须单独列出。

## 当前开发入口（2026-10-28）

2026-10-30 新增独立匿名上下文入口：`python scripts/dev.py auth-context --context tool`。该阶段连续执行专项合成测试/vet、Linux race/fuzz、Windows 快照构建/启动器自检/篡改拒绝，不执行真实匿名模式。协议及有限 parser_match 的含义见 [ADR-0019](../architecture/decisions/ADR-0019-anonymous-submit-context.md) 和 [开发验证](development-validation.md#p3-anonymous-context-proposal)。新包的 `auth-context-launch.exe` 与旧 `auth-dev-launch.exe` 独立，不能混包或沿用旧真实授权。

权威源码目录为 BHKDesktop `D:\project\Go\IPGW-Meta`；Ubuntu-22.04 WSL 使用同一目录 `/mnt/d/project/Go/IPGW-Meta`。不复制第二份权威源码，不依赖笔记本传输开发结果。旧 `agent/handoff.md`、`docs/upgrade/status.md` 的日期明确的历史快照保持原样，当前执行记录位于 `agent/handoff-bhkdesktop-2026-10-28.md`。

历史 R2 包已证明内置错误诊断哈希；生成环节仍未知。此缺口不阻塞新的本机开发，但不能据此宣称重现了历史 R2 构建。保留原包、已安装工具和不可变 Candidate。

## 工具链与依赖准备

两端使用官方归档、官方 SHA-256 校验后的 Go 1.25.0，并保留原有系统 Go/PATH。开发入口要求：

- Windows：`%LOCALAPPDATA%\ipgw-meta-dev\toolchains\go1.25.0\go`。
- Linux：`~/.local/share/ipgw-meta-dev/toolchains/go1.25.0/go`；另需 GCC、make、Bash、Python 3、jq。
- Linux 检查工具位于开发目录的 `bin/`：actionlint 1.7.12、gitleaks 8.30.1，与 CI 一致。
- 两端分别使用开发目录下的 `cache/mod`、`cache/build`、`cache/gopath`。依赖准备是独立联网步骤；使用官方 Go module proxy/sumdb 校验，不更新 `go.mod` 或 `go.sum`。
- Windows 可以下载并校验 Linux 归档、模块缓存和工具，再由同机 WSL 使用。不得为了依赖下载擅自扩大代理监听、防火墙权限或切换 WSL 网络模式。

开发入口不下载安装工具，也不自动修复缺失缓存；缺项时明确失败。不要通过更换模块版本或放宽测试期望解决环境缺口。

同一个主机名、账户名和路径字符串不证明不同执行上下文看到同一文件。工具终端内安装成功后，须在桌面原生终端确认工具可见并实际运行开发入口，才能认定 Windows 原生环境已就绪。若工具上下文的 junction 执行失败而原生最小探针成功，保留两组结果并继续原生回归，不能直接据此修改安装器、ACL 或沙盒策略。`scripts/dev.py` 的 `--context desktop-native` / `--context tool` 仅记录调用者声明，不能充当隔离或身份保证。

## 管理连接与测试边界

Codex、官方依赖下载可使用 Clash 的显式本机代理及维护者已有的 IPv6 上游。关闭 TUN 后分别验证新聊天、工具连接和实际下载能力；HEAD 成功不代表完整 GET 成功。失败时先恢复已知连接，再处理实际失败通道。不能由连接可达推导所有流量使用 IPv6、计费状态或校园账号登录成功。

WSL NAT 且 `autoProxy=false` 时，Linux 的回环地址不能直接当作 Windows 的回环代理地址。开发过程不修改这一设置。

合成测试允许假 RoundTripper、loopback TLS 服务、临时 Git 仓库，以及测试专属目录内的安装器、ACL、junction 和子进程操作。入口隔离 HOME、APPDATA、LOCALAPPDATA、XDG 和临时目录，清除产品与候选安装相关环境，禁用模块自动下载，并为通常使用环境代理的客户端设置不可用代理、仅放行 loopback。**这些进程设置不是系统防火墙**：执行前仍须检查测试没有外部网络/真实凭据入口；清空代理环境也不会消除 TUN 路由。

真实认证遵守[真实验收规范](live-validation.md)和[校园实验室规范](../runbooks/campus-lab.md)，需另行授权。产品默认传输使用直接 IPv4，管理代理连接不能作为产品连接或登录证据。

## 重复执行

以下为可选开发阶段，不是每次必须全量运行的队列。仅执行当前工作包所需检查；P0 文档交付只运行 docs、链接及定向脱敏检查，不构建诊断包或运行产品测试。

在仓库根目录执行；Windows 使用 PowerShell，Linux 使用 Ubuntu WSL：

```powershell
pwsh -NoProfile -File scripts/dev.ps1 -Stage focus
pwsh -NoProfile -File scripts/dev.ps1 -Stage install
pwsh -NoProfile -File scripts/dev.ps1 -Stage test
pwsh -NoProfile -File scripts/dev.ps1 -Stage vet
pwsh -NoProfile -File scripts/dev.ps1 -Stage docs
pwsh -NoProfile -File scripts/dev.ps1 -Stage build
pwsh -NoProfile -File scripts/dev.ps1 -Stage cas-observe
pwsh -NoProfile -File scripts/dev.ps1 -Stage cas-provenance
pwsh -NoProfile -File scripts/dev.ps1 -Stage cas-binding
pwsh -NoProfile -File scripts/dev.ps1 -Stage cas-install
pwsh -NoProfile -File scripts/dev.ps1 -Stage cas-collect
pwsh -NoProfile -File scripts/dev.ps1 -Stage cas-acquire
pwsh -NoProfile -File scripts/dev.ps1 -Stage trust-audit
pwsh -NoProfile -File scripts/dev.ps1 -Stage ctl-check
```

```bash
bash scripts/dev.sh focus
bash scripts/dev.sh install
bash scripts/dev.sh test
bash scripts/dev.sh race
bash scripts/dev.sh vet
bash scripts/dev.sh docs
bash scripts/dev.sh quality
bash scripts/dev.sh cross
bash scripts/dev.sh cas-observe
bash scripts/dev.sh cas-provenance
bash scripts/dev.sh cas-binding
bash scripts/dev.sh cas-install
bash scripts/dev.sh cas-collect
bash scripts/dev.sh cas-acquire
bash scripts/dev.sh trust-audit
bash scripts/dev.sh ctl-check
```

各步骤写入唯一 `build/dev-runs/<platform>/<run-id>/`，记录命令、工具版本、退出码和日志。`test`/`race` 保存 Go JSON 事件，须分别检查失败和 skip；退出 0 不表示未配置的原生 Candidate 资产测试也执行过。检查失败立即停止该入口，不把跳过算通过。目录下的开发构建统一标记 `dev-local`，不能冒充 Candidate。

Windows 执行原生普通测试；Linux 承担 race、shell/Python gates、actionlint、合成泄漏规则、脱敏 Git 历史/非忽略工作树扫描和六平台交叉构建。工作树扫描不读取 ignored 私有 evidence/build 内容。隔离用户目录位于开发目录的 `runs/`，Windows 测试临时目录通常使用同一开发目录下较短的 `t/<run-id>`；产品安装器测试目录不得位于仓库内。新开发包的绑定/安装合成夹具使用下述独立 D 盘报告目录，不属于产品安装位置。后续清理由维护者检查后单独处理。

安装器聚焦排障可执行 `python scripts/dev.py install --run TestWindowsOfflineFreshInstallAndUpgrade`（Linux 使用 `python3` 和对应测试名）。只有 `install` 阶段接受测试名过滤；完整基线仍执行 `test`，不得用过滤运行代替全仓验收。

## CAS 合成观察的含义

2026-11-02 QR 无 ticket 接续独立入口为 `python scripts/dev.py qr-dev --context tool`（WSL 使用 python3）。它执行共享协议/状态机/开发回执聚焦回归和 vet，Linux 增加 race 与解析器有界 fuzz；Windows 构建快照、普通产品与 qr-dev-launch，执行 self-test 和七类绑定/TTY 检查，不执行真实模式。`--preflight` 是独立零网络本地检查，不读取 profile。开发包、全源码检查的安装排除/skip 以及后续单次私有终端说明见[实施记录](../research/qr-continuation-progress-2026-11-02.md)，不复用 password observed CLI 或旧匿名诊断。

当前 P3 静态语法工作包使用 `python scripts/dev.py auth-static-context --context tool`（Linux 使用 python3）。它先运行纯静态解析与原 SDK 假响应回归、上游 AST 差分、资源/脱敏/schema 检查；Linux 增加 race 和有界 fuzz，Windows 增加原生 Job/硬终止验证及独立快照构建、三成员绑定与九类拒绝矩阵。该开发命令不运行真实匿名入口。

新身份 `auth-static-context-dev` / schema 2 保留旧 context 的原义，另输出 syntax。语法赋值、局部相等标签和函数/控制位置不是实际用键；默认产品没有解析器依赖。派生来源与许可证见 `internal/jsparser/upstream.json` / `LICENSE`，用 `scripts/derive-jsparser.py --upstream <固定版本模块目录>` 只读复核可重复补丁。来源/预算及停止条件见 [ADR-0020](../architecture/decisions/ADR-0020-bounded-static-cas-syntax.md) 和[开发验证](development-validation.md#p3-static-syntax-proposal)。旧包保持，真实运行另行单次批准。

`cas-observe` 是内部维护工具，无真实网络模式、文件输入或历史公钥。每次生成两把合成 RSA 密钥，向 SDK 注入完全封闭的假 transport；输出仅为 A/B 标签、计数和稳定错误码。合成 CAS POST 故意返回 500，以阻止进入 ticket/activation/成功路径。

场景包括单键、同键重复、不同键顺序互换、标识符子串、页面与脚本差异、两个脚本的顺序差异。输出记录当前实现的行为，不把选择首键、次键或歧义拒绝确立为规范，也不作为真实登录或 R2 运行证据。实际正确选键仍需独立证据；涉及认证语义变化时先更新规范与 ADR。

来源设计窗口另补充标识符后缀/大小写、注释、字符串、模板、未调用函数、点属性、转义标识符和通用/特定键名顺序场景。该比较器仍使用 SDK 与完全封闭的假 transport，不能复用为真实匿名采集器；独立工具提案见 [CAS 来源诊断设计草案](cas-provenance-diagnostic.md)。

`cas-provenance` 阶段运行独立分析器与命令入口的聚焦测试、vet、内建 self-test；Linux 另执行 race 和两项各五秒、两个 worker 的有界 fuzz。它只分析内建合成来源，不调用 SDK、不执行 JavaScript，也不接收文件、URL、标准输入或认证参数。原型仅支持 `--self-test`，输出 `synthetic-prototype / not_packaged`，使用关系始终 unknown；不确定语法保留 partial，输入拒绝清空内容数组。此阶段自身不验证包绑定，另用 `cas-binding` 验证完整调用链，不能把任一入口的成功当作真实登录证据。

`cas-binding` 阶段验证新四文件开发包。Windows 运行绑定负例、句柄锁/子进程取消测试、vet；随后从明确源码快照离线构建分析器和包外启动器，核对四成员 ZIP/源文件摘要、拒绝复用已有输出目录，再通过声明式 runner 运行实际分析器。manifest 期望值来自本次已审阅本地构建流程；这不授权相信任意下载包旁边的 sidecar。Linux 只运行元数据/输出合成测试、race、vet 和有界 fuzz，不宣称具备 Windows 保护执行能力。

绑定及新开发包安装测试使用 `build/dev-runs/windows/<run-id>/t` 的独立 D 盘夹具，因为工具执行上下文可能重定向 C 盘 LOCALAPPDATA，严格句柄最终路径比较必须拒绝这种差异；不因此放宽路径、reparse 或共享锁检查。这些夹具不改既有安装位置。每次开发包保存 `package/`、`reviewed-launcher/`、`source-snapshot/`、ZIP 和构建回执，全部属于开发产物；`cas-binding` 不安装。包外启动器必须单独审阅和固定摘要，不能从未知包启动一个“自证可信”的程序。

`cas-install` 在绑定检查基础上增加严格 ZIP 验证、新目录安装和失败保留测试。Windows 以本轮构建记录的显式归档/manifest 摘要调用包外启动器，将新包写入本轮报告的 `binding-artifacts/synthetic-install/` 并运行整链自测；同时检查错误锚点不创建目录、重复安装不改字节、原始 ZIP 不变。它不安装到用户工具目录，不升级历史 R2 或产品，不读 profile。Linux 只执行共享解析/回执测试、race、vet 和 metadata/ZIP fuzz，不提供 Windows 安装能力。开发安装使用固定参数序列 `--install-archive ZIP --archive-sha256 HASH --manifest-sha256 HASH --destination NEW_DIRECTORY`；摘要须由包外审阅获得。失败残留目录保留且拒绝复用，不自动删除或修复。

`cas-collect` 运行独立纯内存采集回放：聚焦测试、vet、内建 `--self-test`；Linux 另跑 race 和两个有界 fuzz。D1/D2 已确认：保持动态 service/ac_id 构造，不创建 Cookie Jar，脚本严格限制同源 HTTPS 无查询参数；只消费内建 `.invalid` 目标和假响应，没有 HTTP/DNS、真实地址输入或 transport 默认实现。输出的 simulated_attempts 是模型计数，operations 仍全零。显式静态集合 complete 不表示分析完整、Cookie 无关、页面与 SDK 相同、正确选键或真实登录；现有分析器使用关系仍 unknown。此入口尚未纳入任何四文件包或安装目录，既有包不得冒称包含本次回放代码。

`cas-acquire` 执行新匿名采集工具的 loopback UDP DNS/HTTP/TLS 测试、预算/取消/输出检查、vet 与零网络内建自测；Linux 另执行 race，Windows 另验证拓扑假数据、文件锁、显式确认控制和新快照构建/整链 self-test。该开发阶段从不执行真实采集。Windows 输出 `ready/cas-acquire-launch.exe`、精确源码快照、双文件 payload 和构建摘要；必须使用本轮审阅固定的包外启动器，不能混用旧包。

D9 在同一开发入口增加七个固定来源的假 provider/生成证书/内存库回归、同轮关系比较及闭集报告测试，Linux 增加 D9 报告 fuzz（与原 DNS、报告 fuzz 各五秒）。Windows 的既有根池检查会只读原 ROOT/Disallowed 来验证显式 Go 池不启用系统链引擎；D9 新增来源只在合成测试中替代，不在开发阶段读取实际策略库或 AuthRoot。新包 `--self-test` 只用生成数据，不读取证书库或开 socket；`--preflight` 仍仅验证原前置条件，不代表 D9 材料门禁通过。

维护者确认 `--run-once` 后，新包才读取 D9 额外来源；无可比较材料或读取失败则零网络停止。只有原失败连接关闭、D6/D7完整后才作来源差异比较；新增材料不进入根池、不继续 GET，结果始终 blocked。若原 TLS 已通过则在 HTTP 写入前停止。来源覆盖、比较失败和未执行含义见 [D9](cas-provenance-diagnostic.md#d9-已确认同轮-peer-与固定额外本地来源的差异观察)。

`trust-audit` 是 D8 独立零网络覆盖审计的开发验收入口：两端执行合成测试、vet、自测，Linux 另执行 race 和有界报告 fuzz；Windows 使用生成的内存证书库验证解析与原始拒绝记录、执行绑定负例，再输出 `ready/trust-audit-launch.exe`。此开发阶段不读取实际证书库；合成子进程回归只返回生成的数据。维护者另用已审阅包的 `--trust-audit`，只读固定 CurrentUser/LocalMachine 注册表 ROOT、Disallowed、CA，无需改变 TUN、网络或账号状态；包没有真实采集入口。`complete` 仅表示有界枚举完成，仍须查看解析覆盖和拒绝对照，不表示 TLS 可用；详细限制见 [D8](cas-provenance-diagnostic.md#d8-已确认现有注册表来源的零网络覆盖审计)。

`ctl-check` 是 D10 独立零网络 CTL 材料检查的开发入口：两端执行生成证书/CTL/CMS 的聚焦测试与 vet，Linux 另跑 race 和报告/DER 两项有界 fuzz；Windows 在专用受限子进程中验证原生内存签名 API、子进程禁止和绑定/硬期限，再生成 `ready/ctl-check-launch.exe`。该阶段只执行 `--self-test`，实际 CTL/ROOT/Disallowed 均不读取。维护者另行原生运行新包 `--ctl-check` 才读取固定材料；不需要切换 TUN 或账号状态，没有校园入口。签名层级、时间窗口、用途覆盖与完整 Windows 信任不同，结果保持 authentication=blocked；见 [D10](cas-provenance-diagnostic.md#d10-已确认独立零网络-ctl-材料与固定签名层级检查)。

2026-10-29 后续执行约定已取代上述手工交接：Agent 可在合成验收及绑定完成后自动执行固定来源的零网络 `--ctl-check`，记录账户/工具上下文。D11 在同一开发入口加入四种成员算法、条目属性省略和CTL扩展的合成矩阵；报告v2将成员算法、条目属性、扩展覆盖分开，保留v1历史解码，不代表完整信任。Windows原生内存API作独立对照，Linux运行纯Go测试；本地读取不支持Linux/macOS，不向产品引入微软CTL依赖。`cross`阶段对六平台三产品入口检查诊断依赖隔离并交叉构建；macOS结果仅为构建覆盖。范围与停止条件见[D11](cas-provenance-diagnostic.md#d11-已确认本地-ctl-结构兼容与签名观察)。

`ctl-identifiers` 是 D12 独立零网络标识辨识的开发入口：两端执行合成回归、vet和self-test，Linux另跑race与报告/定位器各五秒的有界fuzz；Windows额外验证绑定、硬期限及禁止跨工具模式，输出 `ready/ctl-identifiers-launch.exe`。开发阶段不读取真实材料。Agent在验收及快照绑定完成后可自动执行新包 `--ctl-identifiers`，只读原两个固定CTL值，只输出有界成员算法与签名属性数字OID。结果始终 `input_authenticity=unverified / signature=not_checked / authentication=blocked`；不继承ctl-check的验签含义、不自动扩展其白名单。无需切换TUN或刷新门户；本地来源读取只支持Windows，纯Go辨识层支持跨平台合成测试。详见[D12](cas-provenance-diagnostic.md#d12-已确认独立零网络-ctl-标识辨识)。

D13在`ctl-check`阶段增加固定SPC属性外层与Disallowed属性索引的有界签名观察；生成的报告v3明确`structure=bounded_checked`及两项覆盖字段，保留v1/v2历史回执形状。SPC内部和属性索引含义未解释，不做成员匹配、信任判断或链接访问。Windows原生内存验签与Go合成回归均须通过，失败不能互相覆盖。新包完成快照/绑定后由Agent自动执行一次零网络`--ctl-check`；旧包保留，产品依赖不变。具体范围见[D13](cas-provenance-diagnostic.md#d13-已确认固定两种表示的有界签名观察兼容)。

`ctl-member` 是D14纯合成成员关系实验入口：`python scripts/dev.py ctl-member --context tool`（WSL使用python3）。两端执行生成材料的签名/成员/策略覆盖回归、vet与self-test；Linux另跑race和五秒fuzz，Windows以受限内存API独立核对RSA签名属性散列，再生成`ready/ctl-member.exe`并自动执行一次`--self-test`及模式拒绝审计。该单文件开发包仅接受`--self-test`，没有真实读取、文件输入、校园请求或信任接口；源码快照与摘要用于可复核性，不冒称真实诊断启动器。命中、未命中与未知策略均不产生可信结论。它不重跑D13真实来源，也不改变产品行为，范围见[D14](cas-provenance-diagnostic.md#d14自动推进纯合成成员与策略分析)。

D15沿用`cas-acquire`开发入口，加入固定签名CTL准备和同轮CAS失败peer成员观察的合成回归、严格回执扩展、原生内存对照、准备阶段硬停止及loopback先关闭连接检查。该入口仍不调用真实材料准备器或`RunOnce`；完整验收后输出独立新采集包，历史包保持原样。其`--self-test`只生成材料，`--preflight`仅执行原本地拓扑/ROOT/Disallowed检查，不读取CTL或验证账号offline。只有单次真实确认后才准备CTL并尝试网络；准备失败零网络停止，比较结果始终blocked、不加入信任。window分类来自本轮准备时刻，不冒称比较时重新验证或缓存最新性。详见[D15](cas-provenance-diagnostic.md#d15同轮peer成员观察的离线实现与待执行包)。

维护者先在 BHKDesktop 原生 PowerShell 用修订版采集启动器的 `--preflight` 做仅本地核验：包绑定与自测、有线接口选择、DNS 目标的本地路由及公共证书库；不解析域名、不建连接、不发送请求。ready 不能替代真实 DNS/CAS 验证。2026-10-29 已批准保留虚拟接口，改为明确有线源地址与接口绑定、逐目标路由拒绝歧义；不再因无关虚拟网段存在就拒绝。旧包继续保留其旧门禁，不原地修改。

真实窗口仍使用 `--run-once`，在包校验与自测完成后显示明确范围和确认词。仅在 TUN 关闭、有线独占、账号已 offline 且恢复路径明确时输入确认，工具不会自动停服务或改路由。恢复方式是手动重新打开原 Clash 的 TUN 开关，保留核心、系统代理和节点。失败后先返回脱敏 JSON，不自动重复执行；state=completed 只说明工具返回有效报告，result.collection 仍可能 blocked，任何结果都不等于真实登录成功。详细范围见 [REL-CAS-DIAG-001](cas-provenance-diagnostic.md)。


## 2026-10-30 P3 取键修复入口

当前离线专项为 `scripts/dev.py key-relation`，实现契约见 [ADR-0021](../architecture/decisions/ADR-0021-cas-key-use-relation.md)，下一真实窗口边界见 [开发验证](development-validation.md#p3-key-use-relation)。静态 parser 可进入跨平台产品依赖，运行时与诊断依赖仍禁止；六平台 build/cross 入口执行精确依赖检查。旧匿名包只保留历史用途，不再作为默认调试路线。


### P3 注销观察离线入口（2026-10-31）

2026-10-31 后续主线已进入 [ADR-0024](../architecture/decisions/ADR-0024-logout-explicit-callback.md) 的callback候选修复，使用同一auth-dev入口和既有观察投影；最终开发包与验证范围见[最新交付](../upgrade/status.md#status-20261031-logout-callback)。不回到Chrome取证或CTL扩展；真实password_core仍单次确认并私有输入密码。

`logout-structure` 阶段集中验证独立观察器的有界结构分析、Read/Close 等价、SDK/cleanup 对照及严格版本；Linux 另含 race 和有界 fuzz。`auth-dev` 阶段生成并验证完整开发包。材料、版本、预算和真实窗口边界见[注销观察开发验证约定](development-validation.md#p3-logout-structure-proposal)，当前交付见[状态摘要](../upgrade/status.md#status-20261031-logout-observation)。这些入口均不会自动执行真实认证。
