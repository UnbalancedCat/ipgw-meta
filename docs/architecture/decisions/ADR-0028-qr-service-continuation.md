---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
adr: ADR-0028
status: accepted
---

# ADR-0028：QR 无 ticket service 的有界接续

2026-11-02 维护者批准按已观察网页链路修复，选择保留 HTTPS 网关接续 API、独立窄范围 QR 开发启动器、严格失败边界及明确时间预算。证据见 [QR 研究](../../research/qr-browser-contract-2026-11-01.md)，不将此兼容差异写成旧正式失败的唯一根因。

产品规则见 [PROTO-QR-CONTINUATION-001](../protocol-correctness.md#proto-qr-continuation-001无-ticket-接续)。不执行 JS，不发送 HTTP service，不读取浏览器 Cookie，不自动 fallback；保持公共 API/CLI/JSON、退出码、正式证据格式及目标身份成功判据。

开发入口使用普通产品和既有 QR 状态机，合并步骤前独立记录命令结果；不扩展 password observed CLI，不记录 HTTP 正文或二维码。开发报告为独立 schema v1、release_evidence=false；既有历史报告不重解释。边界和验收见[开发验证](../../operations/development-validation.md#qr-continuation-development)。

本批准仅涵盖实现、合成验证及开发包交付；真实扫码、提交、Candidate 与发布另行授权。

## 断开接口残留默认路由修正

2026-11-02 维护者确认测试条件后，旧开发包预检仍拒绝，未发起校园请求。只读核验发现两条默认路由：有线适配器 Up/IPv4 Connected；另一条对应断开的 Wi-Fi，适配器与 IPv4 接口均为 Disconnected。维护者随后要求继续修正门禁；不通过删除路由、忽略活动竞争接口或绕过预检解决。

新默认路由排除规则见 [REL-BHK-PATH-001](../../operations/bhk-acceptance.md#rel-bhk-path-001原生测试路径)。仅两项唯一匹配且同时明确断开的残留默认项可排除，未知/缺失/重复/矛盾状态仍拒绝；全量路由及连接状态纳入变化检查。Windows 原生合成测试执行实际 PowerShell 脚本，用假 cmdlet 返回数据，覆盖有线/Wi-Fi 双向切换、活动竞争路由、状态冲突及重连指纹；不访问真实校园或改变配置。

接口状态字段依据微软 [Get-NetIPInterface](https://learn.microsoft.com/en-us/powershell/module/nettcpip/get-netipinterface?view=windowsserver2025-ps) 的 Connected/Disconnected 定义；单次快照不等于路由锁或全部流量隔离。共享验收组件修复只进入新源码及新开发包，旧 Candidate/helper 不替换。新包身份须在后续真实窗口前明确，真实执行仍由维护者私有终端完成。


## 开发门禁首因诊断补充

2026-11-02 维护者批准补齐首个门禁失败的命令、检查位置和闭集原因，按[开发报告约定](../../operations/development-validation.md#qr-continuation-development)升级独立启动器外层 v2，内层命令报告与正式回执保持。只复用现有本地来源，原门禁及取消/清理权规则不变；不增加 HTTP 或 QR 内部观察。先完成离线失败注入、两端验收、独立新包和零网络检查，不授权新的真实登录/注销。
