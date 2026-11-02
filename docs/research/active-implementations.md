---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
snapshot_date: 2026-08-27
---

# 活跃实现与官方页面研究快照

本页只记录可复核的协议线索与风险，不把第三方实现当作规范。运行时不依赖这些仓库；更新结论时必须记录新的固定提交和日期。

2026-11-01 QR 扩展：已定位 neucn/neust、CreamPig233/neu_wisedu2wakeup、RekaYOO/NEU-JWXT-Toolkit 的实际微信 QR 代码。固定提交、目标服务差异及不能采纳的宽松判据见[QR 核对](qr-browser-contract-2026-11-01.md)；不再把最初三个密码项目的检索结论泛化为“其他项目都没有 QR”。

2026-11-01 补充：[三网络手动测试与 ac_id 对照](ac-id-comparison-2026-11-01.md)记录有线/NEU/NEU-2.4G 的维护者反馈、浏览器 1/15/15 观察，以及 neucn/ipgw、Neboer/ipgw-py-manager、7aaauTools/neu-login 的完整固定提交和源码差异。下方历史表格与日期保持；后续按[开发说明](../operations/development.md#外部实现对照)进行相关更新对照。

2026-10-29 补充：已完成 [P1 本地旧版/v1 行为对照](auth-baseline-2026-10-29.md)。当前本地 v0.1.3 解析为 `5ad8c1fa05102fbcb249e7195f4801563b3d44a5`；下表历史 `3721fd6` 引用保留，但两处本地仓库均无法读取该旧对象，不能据 tag 名称宣称 tree 映射已验证。研究优先级见[当前 roadmap](../upgrade/plan.md#roadmap-20261029)，本页旧快照不作为当前实际页面或认证成功证据。

2026-10-30 补充：[P3 提交上下文离线核对](cas-submit-context-2026-10-30.md)记录真实 POST 500 后的源码检查、合成会话/表单覆盖与仍缺少的关系证据；未改变产品选键或取得真实成功。

| 来源 | 固定版本 | 可参考事实 | 不采纳的做法 |
|---|---|---|---|
| 本项目 legacy v0.1.3 | `3721fd6`（2026-03-24） | 已实际使用过的动态 `ac_id` redirect 截获路径；复杂协议问题久未收敛时优先作为历史行为线索 | 按网卡猜测 `ac_id`、无界自动 redirect、认证 Cookie 复用、敏感日志和弱成功语义 |
| [neucn/ipgw](https://github.com/neucn/ipgw) | `8a52d79` | 旧命令、配置与 Dashboard 工作流兼容基线 | 旧成功语义和不安全副作用 |
| [Neboer/ipgw-py-manager](https://github.com/Neboer/ipgw-py-manager/commit/3583017) | `3583017`（2026-05-14） | 动态 form action、`lt`、`execution`、公开 JS 公钥、动态 `ac_id`、JSONP、logout 参数 | HTTP gateway、明文密码、字符串成功判断、详细敏感日志 |
| [DoraTiger/NEU_IPGW](https://github.com/DoraTiger/NEU_IPGW/commit/f0b608b) | `f0b608b`（2026-06-23） | HTTPS gateway 的近期可用性线索 | 硬编码公钥、跳过 TLS 验证、弱业务成功判断 |
| chiyuki0325/NEU-IPGW | `70b3f03`（2026-05-10） | 移动 SSO / TGT / SMS 端点只作为研究线索 | 保存或打印 TGT；未验证的移动协议不得进入 v1 |
| clodite 浏览器实现 | `f6d6552`（2026-03-18） | 浏览器路径可作为人工 canary | 不作为 SDK 正确性或 headless 支持证据 |

## 官方匿名观察

- [统一身份认证页面](https://pass.neu.edu.cn/tpass/login)在快照日可匿名观察到动态表单字段、公开登录脚本、QR 登录入口、验证码/可信设备 UI。
- [QR 登录脚本](https://pass.neu.edu.cn/tpass/comm/js/login-qrcode.js)显示二维码会话、约 3 秒轮询和约 180 秒期限；这些数值属于匿名观察，不是永久协议常量。
- [学校服务页](https://xwb.neu.edu.cn/10053/list.htm)提供 HTTPS 校园网入口线索。
- 当前公开登录脚本表现为 RSA 加密 `username + password`；实现必须动态发现并在异常时返回 `protocol_changed`，不得把快照公钥或页面标题写死。

## 证据等级

第三方代码和匿名页面只能证明 `observed_anonymous`。合成 fixture 通过后可标记 `synthetic_covered`；真实校园网完成并按 [`EVID-REDACT-001`](../evidence/README.md)记录后才可标记 `live_verified` 或 `supported`。

## PROTO-RESEARCH-001：legacy 实现排障入口

当前实现遇到需要多轮复杂测试、无法快速定位的协议或兼容性问题时，应在扩大真实网络请求、采集面或授权窗口之前，先检查固定的 legacy tag/commit。旧实现曾经可运行，因此适合回答“过去通过了哪些入口、redirect 或页面形状”并形成最小匿名判别假设；这一步是本地只读研究，不等于旧行为仍正确。

从 legacy 获得的线索必须重新经过当前 `docs/` 安全边界、封闭合成测试和必要的脱敏匿名诊断验证，再进入实现。不得直接恢复硬编码网络常量、网卡类型 fallback、明文凭据、无界 redirect、TLS 绕过、敏感日志、原始 fixture 或弱成功判断；legacy 与当前规范冲突时始终以当前规范为准。
