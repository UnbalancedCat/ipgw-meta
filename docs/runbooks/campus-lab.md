---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# 校园网实验室运行手册

2026-10-31 P4：当前 evidence/promotion v2、三组 BHK 必需认证及辅助门禁见 [BHK 正式验收 v2](../operations/bhk-acceptance.md)。当前生成与晋升使用 v2；旧四组和旧格式仅在显式历史审计中生效，历史回执不重写。Candidate/release manifest 仍为 v1；格式迁移不提升真实认证能力。

本手册约束双网口 NAS 和 BHK Windows 的真实网络测试面。任何无法证明的拓扑状态都按不安全处理。

2026-10-29 P2 的独立 BHKDesktop 开发窗口以[开发验证契约](../operations/development-validation.md)及 [ADR-0017](../architecture/decisions/ADR-0017-development-auth-validation.md)为准：同机管理/开发、有线 IPv4 产品传输、系统 DNS/PKI 和私有凭据交互，使用开发包身份。Candidate 冻结、私有交互及安全规则继续生效；BHK 正式路径采用 [REL-BHK-PATH-001](../operations/bhk-acceptance.md#rel-bhk-path-001原生测试路径)。下述双 VM 要求仅适用于可选 NAS 采集，不要求 BHK 停止不影响测试路径的虚拟接口。运行前授权须包括系统 PKI 可能的材料获取/缓存更新，恢复路径为重新开启原 Clash TUN，保留核心、代理与节点。

## REL-LAB-001：分离拓扑

- 本地 Windows 负责 GitHub artifact 下载、attestation/hash 验证、协调与 SCP。
- NAS 管理 VM 建议 Ubuntu 24.04、2 vCPU、4 GiB、20 GiB，只连接管理网并可安装 Codex。
- 一次性测试 VM 建议 Ubuntu 24.04、2 vCPU、2–4 GiB、20 GiB，不安装 Codex；管理 vNIC 只有私网静态路由，无默认路由和 DNS。
- 测试 vNIC 使用固定 MAC 并连接独立物理口，是测试 VM 唯一默认路由和 DNS。NAS 宿主不得在该物理口配置 IP、DHCP 或默认路由。
- 只使用 ZOS 官方桥接或直通功能，不 root hack，不使用 Docker host network。Docker 仅用于无凭据合成测试。

## REL-LAB-002：发现与供应

`LAB-DISCOVER` 是独立只读窗口，只确认 ZOS 是否能把管理网络和测试物理口分别交给 VM；不创建 VM、不切线、不启动认证。能力不明或需要非官方修改时记录 blocked。

`LAB-PROVISION` 在单独批准后创建管理 VM、测试 VM 和干净快照，只执行匿名 topology/status 预检：核对接口、路由、DNS、宿主测试口无地址、管理通道在测试默认路由变更后仍可达。匿名预检不得携带账号、Cookie、ticket 或凭据，不得把公网可达解释为校园网登录状态。

ZOS 无法实现隔离时 fail closed：NAS 只执行离线 Linux 安装与无凭据测试；真实认证全部转移到 BHK Windows。不得用容器 host network、宿主策略路由或临时 root hack 绕过门禁。

## REL-LAB-003：单次实验窗口

1. 开始前确认 candidate ID/hash、固定 suite、testbed/network、维护者私有 TTY 和官方门户应急入口。
2. BHK 检查选定有线/Wi-Fi IPv4、唯一默认路由和拓扑监测，TUN 关闭、账号初始 offline、测试窗口独占；NAS 可选采集时，测试 VM 从干净快照启动；检查管理接口无默认路由/DNS、测试接口是唯一默认路由/DNS，NAS 宿主测试口仍无 IP/DHCP/default route。
3. 按 [`REL-TRANSFER-001`](../operations/live-validation.md#rel-transfer-001候选下载与远端传输) 接收并复核候选，不在远端构建。
4. 维护者在私有 TTY 输入密码或扫描二维码；Codex、SSH 转录、shell history 和 evidence 不得接触这些材料。
5. 只运行一个 suite。结束前确认 runner 已停止、没有 QR 轮询、最终 status offline；若清理失败，立即停止并由维护者在官方门户处理。
6. NAS 可选采集时关闭测试 VM 后才能更换物理线缆或上游网络。每轮完成后恢复干净快照。
7. 真实校园网窗口中，BHK WSL 只做 candidate 准备与 hash；Windows 原生 suite 由维护者在私有 PowerShell 中执行。另行批准的本机开发窗口可使用同一台式机 WSL 执行 Linux 合成测试和开发构建，见[本机开发说明](../operations/development.md)，不得将其产物或结果当作冻结 Candidate 的真实验收。

不得在真实测试面保存 GitHub token、秘密历史、敏感备份、pcap、认证截图或原始认证日志；一次性 NAS 测试 VM 不保存源码或 `.git`。2026-10-28 起 BHKDesktop 同时承担开发主机职责，可在权威开发目录保存源码和 `.git`，但必须与冻结 Candidate、真实测试配置和私有凭据分开管理。
