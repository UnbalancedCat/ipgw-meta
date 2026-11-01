---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
adr: ADR-0018
status: accepted
---

# ADR-0018：独立开发入口的认证失败分类

2026-10-30 私有开发 password_core 在初始 offline 通过后，login 返回 exit 3 / network。现有产品公共输出有意丢弃原始错误内容；该回执不能定位 CAS 请求、activation 或最终状态，也不能证明 HTTP 500。匿名窗口能到达凭据边界，不补造本次登录的阶段事实。

维护者在 2026-10-30 确认完整方案，批准离线实现、验收和打包：一次补齐有界、被动的请求角色/HTTP 状态/传输错误类别和 SDK 结果分类。具体边界、身份、验收及停止条件见[开发验证约定](../../operations/development-validation.md#p3-failure-observation-proposal)。本决策仅扩展独立开发入口；实现和绑定验收完成前不能宣称新工具就绪，确认不授权真实运行或读取真实 profile。ADR-0017 的历史包和正式门禁含义保持。

## 推荐接入

独立 observed CLI 复用 `cli.Execute`、`app.Service`、原 profile/prompt provider 及 SDK；不复制登录流程。internal app 增加显式构造入口，向受控 decorator 提供按实际 bind IP 创建的 `transport.Default`。普通 `app.New` 和产品 `cmd/` 不接入观察器，产品 API/CLI/JSON/退出码不变。SDK 的默认 UA、Cookie、重定向、加密、信任和判据均保持。

SDK 返回时可对固定 Code/Message 组合做闭集匹配，未匹配为 unknown；不输出 Message/Cause，不扩大公共 Observer。transport 观察不读取请求正文、header、query、凭据或返回正文；不增加请求/重试，不改变原请求、响应或错误。只依据方法和固定 host/path 在内存分类，不保留 URL。

不增加凭据或 RSA 内部 hook。CAS POST 被交给 transport 只说明执行已到该调用点，不证明远端收到、解密正确或认证成功。响应状态不表示正文读取完成；SDK 固定错误分类只作补充。

## 身份与接受范围

observed CLI 是新的独立开发 EXE，不是 `cmd/ipgw-meta` 的相同字节，更不是既有 Candidate。新开发包同时绑定普通产品对照和 observed CLI，报告分别标识二者哈希及源码快照。不能原位替换旧包或沿用旧 product_sha256 指代 observed CLI。

只有新开发入口的 password_core 使用 observed CLI；其结果仍是开发证据。正式 runner、Candidate 和普通产品不接受新子进程封装。现有匿名入口与历史回执不补造观测数据。

## 比较与取舍

- 仅包装 `cli.Options.NewGateway`：无需 app seam，但 CAS GET/POST 共用错误消息，具体 HTTP 状态与 typed transport cause 已被删除；单独使用仍可能需要下一轮诊断。保留该方式作为合成接入证明，不作为推荐真实工具。
- 修改公共错误/Observer/JSON：可能提供长期可观测性，但扩大产品契约；本阶段不采用。
- 重写 SDK 或传输、抓包/抓原始日志：偏离产品等价性或敏感材料边界，不采用。

决定点是接受上述独立观测编译入口及有限新输出；不是再次申请 PowerShell 或离线测试权限。确认后仍须完成合成与绑定验收；真实窗口单独批准，凭据只由维护者在私有终端输入。
