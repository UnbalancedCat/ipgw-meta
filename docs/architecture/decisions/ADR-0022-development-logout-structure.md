---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
adr: ADR-0022
status: accepted
---

# ADR-0022：独立开发入口的有界注销响应结构观察

2026-10-31 维护者已确认完整实施与离线交付。以下边界已批准；在实现、回归与包绑定完成前，不宣称工具已具备新能力。

修复包首次开发 password_core 已通过登录、online 和重复登录；首次注销返回 HTTP 200 后发生 protocol_changed。清理只发出 status 并返回 AlreadyOffline，最后状态为 offline。现有回执不能区分响应体积拒绝与业务解析拒绝，不能证明具体字段形式或第二次注销请求成功。依据见[当前状态](../../upgrade/status.md#status-20261031-password-progress)。

本地旧版固定提交 `5ad8c1fa05102fbcb249e7195f4801563b3d44a5` 的 `pkg/handler/ipgw.go` / DoLogout 与指定归档对应函数均只检查请求错误，未校验业务结果或复核 offline；[P1 行为对照](../../research/auth-baseline-2026-10-29.md) 已记录此差异。其宽松成功返回不能证明本轮真实响应结构，也不作为放宽 v1 判据的依据。

建议只在独立 observed CLI 增加注销响应的内存旁路分析，具体材料、闭集输出、预算、版本和验收见[开发验证提案](../../operations/development-validation.md#p3-logout-structure-proposal)。这是 ADR-0018“不读取返回正文”的有限例外；普通产品、公共 Observer、SDK、JSON、退出码和 PROTO-LOGOUT-001 均不改变。不把后续 offline 当作忽略业务错误的理由。

相比仅增加 SDK 错误分类，结构观察可同时区分体积、封装、重复字段和既有业务字段形态；仅分类错误仍无法解释 parser 拒绝原因。相比保存正文或任意字段，闭集投影牺牲未知格式的解释能力，保留不得持久化原始认证响应的边界。未知结构可能仍需后续决策，不承诺下一次必然修复。

维护者已批准新正文分析，连续完成合成回归、两端验收、独立打包及零网络自检。确认不授权读取真实 profile、密码或执行真实登录/注销；新包的真实 password_core 仍由维护者在私有终端单次授权。历史包与报告保持原义，不新增匿名注销入口，也不恢复 CTL 或浏览器诊断主线。
