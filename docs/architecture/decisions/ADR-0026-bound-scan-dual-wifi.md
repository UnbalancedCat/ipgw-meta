---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
status: accepted
---

# ADR-0026：定向扫描与双 SSID 验收

2026-11-01 维护者批准本决策。规范批准不表示实现或真实验收已完成。

## 决策

- 增加显式 `network scan --scope all|bound`；默认 all 保持全接口行为。bound 必须显式指定 --bind-ip，本地枚举恰有一个匹配项后才对该项调用 Status；不从 profile 选择地址，不回退全扫。SDK、产品 JSON、退出码和认证判据不变。
- BHK 验收核对选定物理 Wi-Fi 地址唯一，允许无关虚拟地址；调用 bound 扫描，核对全部结果。network list 不过滤。
- 只在验收层通过 Windows Native Wi-Fi API 读取选定接口的当前 SSID。NEU、NEU-2.4G 分别验收；NEU-Mobile 已知但暂不支持，未知、断开或权限拒绝停止。SSID 变化纳入拓扑取消，不扫描热点、不连接、不修改权限。
- 认证 evidence/公开摘要/promotion lock v3，辅助记录/摘要 v2。wifi_ssid 为 neu、neu_24g、neu_mobile 或 null；有线/安装为 null。正式成功只允许前两个无线值。旧版本仅显式只读核验，不能补造 SSID 或混入当前发布。
- 必需集合为有线 password、有线 QR、NEU password、NEU-2.4G password，以及安装、NEU 网络辅助、NEU-2.4G 网络辅助。Candidate/release manifest v1 及来源绑定保持。
- 先交付开发包和另行授权的双 Wi-Fi 开发验证；QR 暂缓，后续修复收敛再统一生成新正式 Candidate。现有包与证据保留，不能跨 Candidate 复用。

学校公布三个名称及不同接入方式，见[官方指南](https://xwb.neu.edu.cn/2019/0909/c10010a262842/pagem.htm)。SSID 是场景身份，不是 TLS 信任或认证成功证明，不映射固定 ac_id。

## 验收

默认全扫兼容；bound 缺失/重复/无匹配及取消零请求；其他接口不被探测；SSID 读取失败/变化停止；证据缺项、重复、错 SSID、跨 Candidate 和历史混入拒绝。两端合成测试、race/vet、文档/脱敏及六平台构建通过后才交付开发包，真实窗口另行批准。
