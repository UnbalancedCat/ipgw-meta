---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
adr: ADR-0019
status: accepted
---

# ADR-0019：产品等价匿名提交上下文

维护者于 2026-10-30 确认 [P3 匿名上下文范围](../../operations/development-validation.md#p3-anonymous-context-proposal)，批准连续离线实施、两端合成回归、独立打包和自检。真实匿名窗口仍需新的单次授权。产品 SDK/CLI、选键语义、默认系统 PKI 和正式认证判据不变；ADR-0018 的 password 观察禁止读取正文/header/query 的规则不变。

复用 ADR-0017 的 SDK 匿名流程及已建立的内存响应观察，仅对 SDK 实际完整消费的成功页面/脚本做有界分析，不额外获取被短路的脚本。保持 provider 永远拒绝、零真实凭据/profile/POST/activation/logout。来源分类、表单目标关系及当前 parser 首次文本匹配的局部映射采用独立 auth-context-dev 身份和 schema 1，不改写旧回执或旧包。

parser_match 只描述当前 parser 对相同输入的文本匹配，不能称为正确键或浏览器实际选择。它可以指向注释、字符串、子串或受支持的赋值；映射或来源角色无法证明时输出 unknown。现有分析器的 use_relation 继续固定 unknown，不加入 JavaScript 执行器或使用关系推断。此明确限定的匹配投影不使 ADR-0016 的产品选键提案成为 accepted。

操作预算和内存/语法/输出预算以专题规范为准；不完整、超限、非成功响应和角色不确定不提升为有效来源。通过独立闭集校验和源码/产物绑定后才交付新包。合成等价测试必须覆盖请求顺序、Cookie、Read/Close/error、provider 边界、各类来源及部分覆盖，不以通过为理由放宽产品契约。

仅证明匿名分类完成不代表 HTTP 500 根因已查明、真实密码正确或目标账号登录成功。无法区分假设时集中报告缺口，不自动扩大观察或执行真实登录。
