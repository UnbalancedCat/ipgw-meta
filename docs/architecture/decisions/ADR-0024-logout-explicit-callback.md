---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
adr: ADR-0024
status: accepted
---

# ADR-0024：注销请求显式指定 JSONP callback

2026-10-31 维护者批准callback单变量候选修复、离线回归与开发包交付。真实password_core仍单独确认现场及单次窗口，密码由维护者私有终端输入；不操作Chrome，不复制其会话，不创建Candidate或发布。

## 依据与证据限制

[网页对照](../../research/logout-browser-context-2026-10-31.md)确认网页注销请求携带callback，展示响应为单个JSONP调用，内层整数ecode为0。使用独立合成动态值的离线回归证明现有严格解析器能够接受该结构。SDK先前请求没有callback，观察结果为封装other；其原始正文未知，不能把网页响应当作SDK历史响应。

callback是优先验证的表示请求差异，不是已证明的唯一根因。网页还带ip、ac_id等参数且处于不同会话。本决策只改变callback，避免同时引入多个变量；不从HTTP200、网页提示或旧版直接返回成功推导新成功判据。

## 决策

- 共享SDK的在线注销HTTPS GET保留既有action=logout和来自初始状态的username，仅增加固定 `callback=ipgw_logout`。值是内部合法标识符，不含时间、身份或用户输入，不暴露新公共配置。
- 不新增ip、ac_id、缓存破坏参数、发现请求、重试或CAS退出；不导入浏览器Cookie/header，不修改默认传输、TLS、绑定、重定向和原有Cookie Jar规则。
- 响应继续使用原有严格JSON/JSONP解码与ecode整数/可选success一致性判据，再独立复核offline。原本已离线仍直接AlreadyOffline，不发注销请求。
- callback只请求表示格式，不执行JavaScript；不要求响应回调名与固定请求值相等，不取消既有裸JSON兼容，也不扩大现有合法回调语法。这些行为均沿用既有解析器，本次不新增解析契约。
- 保持公共SDK、CLI、产品JSON、退出码与登录成功不变量；显式switch经过同一共享注销路径，失败仍在凭据读取前停止。

## 验收与退出

使用假传输验证方法、HTTPS、精确三项查询键、特殊字符username安全编码及无额外请求；覆盖JSONP成功后offline、业务拒绝、畸形/截断/重复/冲突、HTTP失败、状态仍online和初始offline。新样例只用合成值，不保存真实材料。观测入口与普通产品复用同一修复，观察不改变请求或状态判据。

完成Windows/Linux相关测试与vet、Linux race、既有观察器有界fuzz、六平台产品构建/依赖隔离、文档和脱敏检查；以源码快照分别绑定普通产品和独立observed CLI，运行整链self-test与构建审计。现有Windows安装/junction缺口单列，不因专项通过而关闭。

新包真实验证需独立单次授权；本决策不自动登录、注销或读profile。若验证失败，保留闭集结果，不自动再添加其他参数或放宽响应接受条件。若通过，仅记为该开发包本次闭环，正式Candidate与M3状态不能据此升级。
