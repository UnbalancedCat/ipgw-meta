---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
adr: ADR-0017
status: accepted
---

# ADR-0017：产品等价开发验证与正式 Candidate 证据分离

2026-10-29 roadmap 与 P2 两项确认决定：直接复用 SDK 凭据前流程和同轮内存 CAS 会话；优先产品默认传输与系统 PKI，选择有线 IPv4 源地址绑定和本地拓扑检查，接受系统 DNS/PKI 的实际边界。D 系显式信任池与自定义网络路径不再作为产品等价性前提。

新增独立、源码/产物哈希绑定的开发工具，匿名观察到 provider 边界停止；开发 password 验证执行同快照产品并复用已有 password_core 状态机、私有 prompt 及精确身份判据。两者均需新的真实窗口授权，离线开发批准不授予真实运行。

开发回执只用于修复反馈，不是 Candidate evidence，不改变能力或正式四组门禁。ADR-0007 的不可变 Candidate/晋升约束仍完整适用于正式发布；原“真实认证只验证冻结候选”的执行范围增加本决策定义的独立开发例外。P4 尚未完成，不据此切换正式矩阵。

不新增产品 API/CLI/JSON，不修复未经证明的取键语义，不引入 CTL/微软签名者依赖或自建信任引擎。具体预算、停止条件与验收见 [产品等价开发认证验证](../../operations/development-validation.md)。
