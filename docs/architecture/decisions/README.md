---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# 架构决策记录

| ADR | 状态 | 决策 |
|---|---|---|
| [ADR-0001](ADR-0001-product-boundaries.md) | accepted | 单 Go SDK、三入口、NEU-first 内部边界 |
| [ADR-0002](ADR-0002-compatibility-and-cutover.md) | accepted | 工作流兼容与 1.x legacy 切换策略 |
| [ADR-0003](ADR-0003-secure-transport.md) | accepted | ticket 截获与 HTTPS-only |
| [ADR-0004](ADR-0004-challenge-model.md) | accepted | QR 与验证码统一人工挑战模型 |
| [ADR-0005](ADR-0005-docs-authority.md) | accepted | `docs/` 唯一规范源、`agent/` 派生索引 |
| [ADR-0006](ADR-0006-transactional-config-migration.md) | accepted | 旧凭据显式决策、无秘密 journal 与可恢复迁移事务 |
| [ADR-0007](ADR-0007-immutable-candidate-promotion.md) | accepted | 一次构建的不可变候选与原样晋升 |
| [ADR-0008](ADR-0008-offline-transactional-installer.md) | accepted | 离线 acquisition 与事务安装共用验证链 |
| [ADR-0009](ADR-0009-separated-live-test-plane.md) | accepted | 真实认证分离管理面、测试面与私有交互面 |
| [ADR-0010](ADR-0010-macos-trusted-system-path-alias.md) | accepted | macOS 固定 `/var` 系统别名作为受验证路径锚点 |
| [ADR-0011](ADR-0011-nonblocking-m0-governance.md) | accepted | M0 残余治理继续跟踪但不阻塞发布流水线 |
| [ADR-0012](ADR-0012-bounded-anonymous-acid-redirect.md) | accepted | `ac_id` 只在无 Cookie 的同网关匿名窗口内有界跟随一次 |
| [ADR-0013](ADR-0013-cas-html-before-jsonp.md) | accepted | CAS challenge 先识别完整 HTML，再接受严格的整响应 JSON/JSONP envelope |
| [ADR-0014](ADR-0014-cas-rsa-compatibility-envelope.md) | accepted | 仅在官方同源 HTTPS 内接受动态 CAS RSA-512 兼容封装 |
| [ADR-0015](ADR-0015-cas-submit-closed-fields.md) | accepted | CAS 密码 POST 使用六字段 closed-world allowlist |
| [ADR-0016](ADR-0016-cas-provenance-before-key-selection.md) | proposed | 先建立有界来源诊断与工具绑定，真实选键语义保持待决 |
| [ADR-0017](ADR-0017-development-auth-validation.md) | accepted | 产品等价开发验证与正式 Candidate 证据分离 |
| [ADR-0018](ADR-0018-development-auth-failure-observation.md) | accepted | 独立开发入口的被动认证失败分类与身份隔离 |
| [ADR-0019](ADR-0019-anonymous-submit-context.md) | accepted | SDK 实际消费集合的匿名上下文与 parser 匹配投影 |
| [ADR-0020](ADR-0020-bounded-static-cas-syntax.md) | accepted | 有界静态语法、字面量候选与运行时使用关系分离 |
| [ADR-0021](ADR-0021-cas-key-use-relation.md) | accepted | 完整来源与有限登录使用关系共同决定动态取键 |
| [ADR-0022](ADR-0022-development-logout-structure.md) | accepted | 独立开发入口的有界注销响应结构观察 |
| [ADR-0023](ADR-0023-browser-logout-comparison.md) | accepted | Chrome 注销对照、维护者先脱敏与独立状态确认 |
| [ADR-0024](ADR-0024-logout-explicit-callback.md) | accepted | 共享注销显式callback，保持既有业务与offline判据 |
| [ADR-0025](ADR-0025-bhk-formal-acceptance.md) | accepted | BHK v2 正式认证、执行者绑定与辅助发布门禁 |
| [ADR-0026](ADR-0026-bound-scan-dual-wifi.md) | accepted | 显式定向扫描、真实 SSID 核对及双 Wi-Fi 验收 |
| [ADR-0027](ADR-0027-browser-qr-contract.md) | accepted | 网页 QR 契约优先、无痕桌面对照与独立状态证明 |
| [ADR-0028](ADR-0028-qr-service-continuation.md) | accepted | QR 无 ticket service 的有界接续与独立开发验证 |

新增或改变公共行为时新建 ADR，不改写已接受 ADR 的历史结论；被替代的记录标为 `superseded` 并链接后继记录。
