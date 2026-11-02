---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
snapshot_date: 2026-11-01
---

# 三网络手动测试与 ac_id 实现对照

本页记录维护者观察及固定版本的外部源码，不增加产品规则。正式验证状态见[双 SSID 开发记录](wifi-bound-progress-2026-11-01.md)与[当前状态](../upgrade/status.md)。本轮只读取公开 GitHub 源码并更新文档，未运行外部实现、访问校园门户或重放认证。

## 手动开发测试与浏览器观察

2026-11-01，维护者按本次提供的 PowerShell `ipgw-dev` 临时函数，手动尝试普通开发产品的 login、status、logout、status，反馈有线、NEU、NEU-2.4G 三者均无问题。所提供入口是既有 wifi-dev 包的 `payload/ipgw-meta.exe`，包源码摘要为 `718583a7b888cbf49a2acc5717baaa551907daf53815c032b7ba44571964d6ce`，构建记录中的产品 SHA-256 为 `700c25ac6718c9b4668f709a455c82e0b3a21c7d87a325b1bd42bff51a02b202`。

这次额外手动试用未回传逐命令输出、退出码或独立绑定回执。它是维护者报告的三网络使用结果，不能补写为七步套件、拓扑监测通过或正式 Candidate evidence；不从此前受控窗口推定本次每项网络设置均重新核验。此前两个 Wi-Fi 受控开发窗口的结构化 pass 回执单独保留。

维护者还报告，浏览器访问 HTTPS 网关首页后进入 `/srun_portal_pc`，观察到以下查询字段值；只保存字段和值，不保存带查询参数的完整认证 URL：

| 本次接入方式 | 浏览器观察到的 ac_id | theme | 手动登录/注销反馈 |
|---|---|---|---|
| 有线 | 1 | pro | 无问题 |
| NEU | 15 | pro | 无问题 |
| NEU-2.4G | 15 | pro | 无问题 |

这证明的是维护者本次环境中观察到的对应关系，不证明全校所有位置、控制器或未来配置不变，也不证明 CLI 在这次手动运行中实际发现了同一数值（没有该项观察回执）。SSID 不充当 ac_id 查表键；`/index_15.html` 自动门户观察与 Wi-Fi 的值一致，但不据路径数字直接新增产品发现来源。自有路由器/桥接/NAT 情况仍是后续独立需求。

## 外部版本与源码发现

2026-11-01 通过 GitHub 只读接口查询默认分支 head，并再次按完整提交读取下列文件。结论限于这些版本与路径；没有验证它们在当前校园网的实际可用性。

| 仓库 | 查询分支 | 固定提交 |
|---|---|---|
| neucn/ipgw | master | `8a52d793f38a8dad176790882e88d2c6aa77ccaf` |
| Neboer/ipgw-py-manager | master | `3583017ab207c7618effd455caabd2a8b9783de6` |
| 7aaauTools/neu-login | main | `558e78957425d789550dc75fe334c2be63bdbebf` |

### neucn/ipgw

`requestLoginApi` 访问 HTTPS 网关首页，使用 HTTP 客户端跟随跳转后的 URL 查询参数构造 CAS service；它传递完整 RawQuery，而非按网卡或 SSID 固定编号。`Logout` 则使用固定含 ac_id=1 的 Referer，且没有将动态 ac_id 作为注销查询字段。参考 [pkg/handler/ipgw.go](https://github.com/neucn/ipgw/blob/8a52d793f38a8dad176790882e88d2c6aa77ccaf/pkg/handler/ipgw.go#L120)。

可借鉴的是让当前门户提供接入上下文；不能据此复制完整查询参数、自动重定向或固定 Referer，也不能推定它的注销在所有网络上失败或成功。

### Neboer/ipgw-py-manager

`get_ipgw_session_acid` 访问 HTTP 网关首页并跟随重定向，解析最终 URL 的 ac_id。初始化时保存在实例，随后传入激活和注销函数。参考 [portal.py](https://github.com/Neboer/ipgw-py-manager/blob/3583017ab207c7618effd455caabd2a8b9783de6/ipgw/core/api/portal.py#L70) 与[调用链](https://github.com/Neboer/ipgw-py-manager/blob/3583017ab207c7618effd455caabd2a8b9783de6/ipgw/core/ipgw.py#L17)。

但 `SSO_prepare` 使用的初始 target 中 service 的 ac_id 固定为 1；动态 ac_id 在 `SSO_login` 的邮箱提示分支中才用于再次构造地址。因此不能把它描述为所有 CAS 阶段均动态一致。参考 [SSO.py](https://github.com/Neboer/ipgw-py-manager/blob/3583017ab207c7618effd455caabd2a8b9783de6/ipgw/core/api/SSO.py#L13)。其 HTTP 激活/注销方式不进入本项目的认证路径。

### 7aaauTools/neu-login

Python 与 ESP32 两个入口都设置默认 AC_ID=16；运行时 `detect_ac_id` 请求一个外部 HTTP 连通性地址，从响应正文或 Location 中匹配 ac_id 或 index_ 后面的数字。检测值优先，否则 `_cur_acid` 使用默认值；`fix_service_url` 在 service 缺少 ac_id 时补入当前值，已有字段则保留。参考 [Python 入口](https://github.com/7aaauTools/neu-login/blob/558e78957425d789550dc75fe334c2be63bdbebf/neu_login.py#L439)与 [ESP32 入口](https://github.com/7aaauTools/neu-login/blob/558e78957425d789550dc75fe334c2be63bdbebf/neu_login_esp32.py#L463)。

其默认 16 是实现常量，不是当前校园配置的实测证明；与本次观察到的 15 不同，说明照搬默认值并不能代替本轮发现。该实现同时存在额外外网探测、自动会话重建及 TLS 验证降级等与本项目约束不同的行为，本轮不采纳、不运行。源码注释对错误编号导致放行异常的解释只作待验证假设，不升级为本项目已证实根因。

## 与本项目的关系

本地核对基线为分支 `codex/bhk-v1-acceptance`、HEAD `c21094ab3e7b2a4064b66ae82cf2fbc2173f92cd` 加现有未提交工作树；本次未改产品代码。

- [auth.go](../../auth.go) 的 discoverACID 有界读取网关响应，校验 Location 的来源和形状，解析唯一数字 ac_id；登录将发现值放入 CAS service，并验证返回 ticket 的接入上下文。没有 SSID→1/15 的固定映射。
- [client.go](../../client.go) 的 networkKey 返回空值，持久协议缓存回退实际禁用；不能把 discoverACID 中保留的缓存代码分支称为当前已启用能力。
- 当前注销契约按 [ADR-0024](../architecture/decisions/ADR-0024-logout-explicit-callback.md)，没有因本次对照而新增 ac_id、IP、浏览器 Cookie 或 Referer；三网络手动反馈不要求顺手改变已通过的注销请求。
- 当前发现边界以 [PROTO-DISCOVERY-001](../architecture/protocol-correctness.md#proto-discovery-001发现优先) 为准。保留动态发现及 HTTPS 认证，不新增默认编号、外部连通性探测或基于 SSID 的推断。需要改变发现语义时，先以独立合成复现与规范/ADR 决策确定范围。

## 后续更新时如何参考

外部实现对照作为开发更新的常规检查项，执行方式见[开发说明](../operations/development.md#外部实现对照)。本次三份固定提交是下次比较起点，不能把“已看过”无限沿用为最新结论。对照产出应说明相同、差异、未知和是否采纳；没有相关变化时记录即可，不为凑对照而改变产品或新增真实测试。
