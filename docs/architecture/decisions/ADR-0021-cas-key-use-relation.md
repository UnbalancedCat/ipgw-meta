---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
adr: ADR-0021
status: accepted
---

# ADR-0021：按有限登录使用关系动态取键

维护者于 2026-10-30 批准 P3 实施方案。匿名观察证明旧文本匹配可能命中注释；维护者提供的脱敏脚本显示不同的活动登录关系。这支持修复选择器，不证明历史 HTTP 500 的唯一根因，也不证明浏览器实际执行路径。

共享 SDK 使用固定 Goja 静态 AST 与现有有界派生 parser，只解析内存，不执行 JavaScript 或读取 source map。接受同一顶层普通函数声明、同一函数体语句块内连续的五步关系：局部 const 字符串 → const RSAEncryptor 实例 → encrypt 结果局部变量 → `$("#rsa").val(result)` → `$("#loginForm")[0].submit()`。变量和函数名不限；encrypt 输入为局部用户名与密码读取值相加，分别来自 `#un` 与 `#pd` 的 val()，可经局部 selector 变量。输入声明须在链之前；绑定重复、写入、遮蔽、跨函数、条件链或未解释的登录 rsa 写入拒绝。普通函数体外的链暂不支持。此为有限协议结构识别，不是完整 JavaScript 数据流或执行证明。

检查页面所有可执行经典内联脚本及既有同源 HTTPS、路径含 login/rsa 的外部引用，最多八项，不去重改变计数，不遇首键短路。非脚本 HTML、template/noscript 内容及非执行数据脚本不参与。module/nomodule 等不受支持执行形式及 base[href] 拒绝，不猜测浏览器的替代 URL 解析。全部允许来源读取和分析完整后，只有一个 RSA 模数/指数身份才接受；多个相同键可接受。没有关系的裸赋值不再兼容；注释、字符串与子串不构成候选。不得固定键值、文件名、次序或历史 fallback。

解析沿用每源 1 MiB、累计 4 MiB、十万 token、16 Mi 次字符工作量、深度 64；AST 总遍历限十万节点，分析累计协作期限十秒且服从父上下文。HTTP 保留原单响应上限；网络等待不占分析时间预算。资源限制、取消、解析不完整、歧义或来源错误均在凭据提供器之前停止。网络错误保留原分类，选择失败使用 protocol_changed。RSA 结构和位数校验前移，密码容量检查仍在私有输入后；ADR-0014 的加密边界不变。

固定静态依赖可进入产品闭包，JS VM、诊断包、Windows Job 和进程终止不可进入。新开发包只保留 password_core、self-test、preflight；旧匿名真实入口在新构建中拒绝，旧包与回执保持原义，诊断参考正则与产品解耦。没有新增真实运行授权或正式证据。

验收覆盖生成 RSA 的注释诱饵、来源顺序/相等/冲突、拒绝时零凭据/POST、实际合成 POST 解密、预算/取消、Windows/Linux 回归及六平台闭包构建。最终真实闭环独立授权，正式 password_core 仍为 blocked/live_unverified。
