---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
status: accepted
---

# ADR-0025：BHK 正式验收 v2 与辅助门禁

2026-10-31 维护者确认完整 P4。旧四组认证矩阵不再作为新版本发布政策；历史 v1 的格式、矩阵和解释保持。新政策须在规范、工具和离线验收全部闭合后切换，不把批准等同实现完成。

## 决策

- Evidence、公开摘要和 promotion lock 使用 v2；新生成和当前晋升只接受 v2，历史 v1 仅在独立只读入口核验。
- 必需认证为 BHK Windows amd64 有线 password、Wi-Fi password、有线 terminal QR。NAS 有线 password/QR 与 BHK Wi-Fi QR 可独立记录，不替代必需项。
- Evidence 新增 runner_sha256，执行前后绑定 Candidate manifest 中对应 helper。保留 Candidate/release manifest v1、六平台资产和两种 helper。
- BHK 正式 runner 显式绑定选定 IPv4，按有线/Wi-Fi 核验并监测拓扑；产品默认系统 DNS/PKI 保持。拓扑变化停止后续操作，包括 cleanup。
- 安装与 Wi-Fi 辅助验收独立记录；promotion 必须引用两项通过摘要。安装复用隔离测试，Wi-Fi scan 仅在枚举恰好一个选定地址时执行，不改公共 scan。
- P4 包含 Windows junction 既有缺口的离线定位与契约内修复；不得绕过 junction、放宽 ACL 或删除失败断言。

完整 wire、执行与验收契约见 [BHK 正式验收](../../operations/bhk-acceptance.md)。本决策不授权真实操作、正式 Candidate、提交或发布。产品 SDK/CLI/JSON/认证判据保持；Windows 证据不得提升 Linux/macOS 的真实能力状态。

## 后果

新 runner 不能搭配旧 Candidate。P3/P4 实现须进入新的受保护源码基线，P5 再构建并验收不可变 Candidate。自校验与本地回执是可审计的工具链绑定，不是硬件远程证明。
