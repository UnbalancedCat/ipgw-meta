---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# 注销网页对照：事实与缺口（2026-10-31）

执行边界由 [ADR-0023](../architecture/decisions/ADR-0023-browser-logout-comparison.md) 和[浏览器约定](../operations/browser-comparison.md#browser-logout-comparison)确定。本记录不是新的产品成功契约。维护者后续提供了自行操作Chrome的截图与说明；Agent没有执行真实窗口，不能将这些材料追认为完整的三次status对照或后续真实操作授权。

## 既有证据

- 开发包 source `a4ef40dfb8dc66afa3db1b8f6832c7139008200d1ea74c62a7a9fc6a55a1cce2` 的私有终端回执：登录、online、重复登录成功；第一次logout返回200但SDK解析拒绝。完整正文未进入支持的JSON/JSONP对象入口，不能从other推断实际格式；字段unknown不代表缺失。清理仅status即offline，套件仍失败。详见 `build/logout-observation-live-20261031/`。
- 本地旧版固定提交 `5ad8c1fa05102fbcb249e7195f4801563b3d44a5` 及指定归档是源码对照，不保证唯一对应维护者最早使用的Gemini二进制。[P1核对](auth-baseline-2026-10-29.md)保留身份和历史引用缺口。
- 独立浏览器r4离线实验10项中8项通过，2项重定向阻断失败；仅为回环合成结果，不能当作校园会话隔离验收。

## 请求与处理矩阵

| 项目 | 当前SDK | 固定旧版 | Chrome网页 |
|---|---|---|---|
| 注销方法/端点 | HTTPS GET，网关 `/cgi-bin/srun_portal` | 相同 | 维护者提供的记录为GET同路径、action=logout，列表HTTP200；配置选择normal且无双栈追加 |
| 参数键 | action、username；安全编码 | 相同键，直接拼接 | 记录显示callback、action、username、ip、ac_id、_；仅保留键，不保存动态值 |
| callback/显式JSON响应请求 | 均未设置 | 均未设置 | 请求有callback；维护者提供的Response显示单个回调包裹严格JSON对象，无末尾分号 |
| Referer | 注销请求未设置 | 门户成功页，带发现的ac_id | 未确认；不收集完整header |
| UA | 产品UA | 浏览器UA | 当前实际请求未知；不据浏览器名称推造字节 |
| 前置状态/会话 | 新client先status，使用在线username | 独立命令也新建Handler/Jar，先FetchAcID，username来自调用者 | 未确认；个人Chrome已有会话不等于SDK会话 |
| 重定向 | 默认拒绝 | 默认跟随 | normal成功处理默认主动导航至CAS `/tpass/logout`；这是脚本导航，不是已观察的HTTP重定向 |
| 业务解析 | 严格对象封装，数值ecode及可选success一致性 | 只检查请求错误，正文交debug，直接成功返回 | 通用回调按code/error的undefined关系及严格值比较，见下文；没有ecode判断 |
| 最终状态 | 业务成功后复核offline | 不复核 | 所示注销函数修改本地online标志；未独立核验服务端offline；三次status尚未执行 |

源码入口：当前 [auth.go](../../auth.go) / Logout、logoutOnline，[client.go](../../client.go) / defaultEndpoints、newHTTPClient、request，[业务解析](../../internal/srun/business.go)；旧版 `pkg/handler/ipgw.go` / DoLogout、`pkg/cmd/logout.go`、`pkg/utils/http.go`。这些差异不证明服务端为何选择某种格式，不能自动复制UA/Referer/Cookie行为。

## 时间线与证据等级

| 事件 | 等级 | 状态 |
|---|---|---|
| 既有开发回执的SDK登录与注销观察 | 严格绑定的开发运行证据 | 已核验；不是Chrome运行记录 |
| 当前/旧版请求构造与解析 | 本地源码事实 | 已核对；旧版未运行 |
| 本次Chrome窗口存在 | 本地窗口元数据 | 已确认现有校园标题窗口与DevTools窗口各一；没有据标题推断账号状态 |
| 网页注销调用链 | 维护者提供的本地脚本及配置静态事实 | 已核对normal、双栈关闭；预览字段与成功分支相容，未做动态调用栈验证 |
| 已有Network记录 | 维护者提供截图和操作说明 | 已定位action=logout及HTTP200；只保存必要脱敏投影，不转存请求头或正文 |
| 新Chrome登录/注销与三次status | 维护者自行操作说明与页面截图 | 登录/注销操作已报告；三次普通产品status未提供，不认定为完整对照验收 |

## 下一步与退出条件

配置、请求键、Response封装和合成解析验证已齐备，不再要求维护者查找或重做网页登录注销。维护者随后批准callback单变量候选修复，完成离线交付并另行授权一次私有开发验证；[完整开发闭环现已通过](../upgrade/status.md#status-20261031-password-dev-pass)。本对照工作包到此结束，不继续索要浏览器材料或重复运行。请求上下文的全部因果尚未隔离，不能用网页响应补造SDK历史正文；下一主线为正式验收规则收敛。

若只得到静态成功分支，记录为源码契约候选，不声称本轮服务器实际返回。若网页端点/上下文不同，分别记证据。只有完整且足以区分成功/失败的契约依据成立后才写合成回归及最小修复方案；否则保留缺口，不改变产品parser或正式状态。

## 本地三份脚本的静态对照

维护者提供桌面上的Portal.js、main.js、lang.js，Agent只读分析，未执行脚本、取资源、读取浏览器存储或复制原件入仓库。以下行号绑定本次文件；哈希只用于本地材料身份，不证明服务端来源或当前页面执行了相同内容。相关片段使用账号/IP变量引用，未将其运行值写入记录；不据局部审阅保证整份文件均无隐私。

| 文件 | 字节数 | SHA-256 |
|---|---:|---|
| Portal.js | 125657 | `0130f589e9b31e576dc4883f0e38c6d3fb661e3f0e69739a39c778ad4491b277` |
| main.js | 17608 | `eada7f77aab2b97b0ca2e9d109288bbdfb4119ab4a01dd2ff8c344685b8fa99c` |
| lang.js | 33775 | `83fe7cfb1bb68f9936c894050c2074d28e6ab87378c37f7f18a18963c5dd28b4` |

### 按钮、分支与请求

- main.js 190–200：普通 `#logout` 点击确认后调用 `portal.logout()`，不传成功回调。Portal.js 2602–2606 根据 `config.portal.MacAuth` 选择DM或normal分支。main.js第3行从外部CONFIG创建实例，三份材料不足以确定当前开关值。
- Portal.js 462–540：normal以JSONP发出网关注销；参数为action、username、ip、ac_id。username来自userInfo的username和domain拼接；ip通常来自userInfo，双栈另一host请求为空。若portalInfo.doub为真，还请求另一IP栈host。1332–1354显示该开关由设备分类及DoubleStackPC/DoubleStackMobile决定；页面协议、配置host及实际选路仍不能从静态代码确证。
- Portal.js 382–459：DM使用另一端点，参数为ip、username、time、unbind、sign，也有双栈分支。其契约不能与normal合并。
- main.js 364–405：“断开全部连接”调用POST `/v1/batch-online-drop`，不同于普通注销；未登录处理还从输入框取账号和密码。该功能不在本次操作范围。

### 响应与后续动作

- Portal.js 156–160、545–593：jsonp助手设置GET与dataType=jsonp，传给jQuery；通用请求成功回调接收的是jQuery处理后的res，不是原始响应字节。这里没有显式指定callback名；不得猜测真实封装。
- 若 `res.code !== undefined` 且 `res.error === undefined`：严格 `res.code === 0` 调业务成功回调，否则调失败回调。若前者undefined而后者非undefined：严格 `res.error === 'ok'` 成功，否则失败。两者都有或都没有时，这段代码不调用任何业务分支。这个判断不是重复字段检测、严格对象结构校验或ecode契约；不能直接移植为SDK安全解析器。
- normal成功回调设置本地online=false；等待pending结束后，未提供obj.success时主动导航到CAS的 `/tpass/logout`，带门户返回目标。未记录完整URL。它可能影响现有统一认证会话；不是单纯回到网关首页，也不是已观察的HTTP302。
- normal中pending有三秒归零机制，所示成功导航条件是本地online标志；这不证明每个双栈响应都业务成功或服务端已经offline。后续真实对照仍需独立status。
- lang.js是展示翻译，包含LogoutOK等标签；翻译字符串不是原始业务响应，也不能据此增加文本成功白名单。

### 当前结论与剩余缺口

已有代码足以证明网页与SDK请求及判断存在差异，但无法解释SDK收到的other正文究竟是什么。特别是JSONP请求与未请求JSONP的原始GET不能视为相同响应契约。配置、字段和展示的Response封装已补齐；不同请求参数的因果作用仍未确定，不能将callback候选修复写成已验证根因，更没有依据复制浏览器header或更换业务成功判据。

未修改产品parser、编写猜测性成功夹具或重跑SDK。提供的配置含运行地址，本记录只保留必要布尔值，不保存完整CONFIG或截图。

### 后续配置核对

维护者随后提供的Sources截图和配置文本一致显示：MacAuth=false、DoubleStackPC=false、DoubleStackMobile=false，另有isIPV6=false。与上述代码组合，静态路径确定为normal，doub为false，不追加另一IP栈请求；成功处理仍默认导航至CAS退出路径。这里的isIPV6只是页面配置，不证明实际路由或当前账号状态。其余开关不能覆盖代码中明确的导航语句，也不能据此推断该语句已经执行。

材料还显示页面引用Portal.js和main.js，与本地脚本角色相符；引用名称不证明已加载字节与本地文件逐字相同。未执行Console、导航、登录或注销，未读取实时浏览器状态。

## 后续Network材料与解析器核对

维护者报告在第二个tab完成登录和注销；第一个tab没有新增记录。截图中Fetch/XHR筛选处于选中状态，前三张Keep log未勾选、末张已勾选，并存在额外筛选条件。因此这些列表不能作为完整导航链或无其他请求的证据。末张出现网关srun_portal请求，列表HTTP200，后续Payload截图确认action=logout及上表参数键。

响应预览可见：ecode为数值0；error、res为字符串ok；error_msg为空字符串；client_ip、online_ip和srun_ver为字符串。仅记录字段类型及必要业务常量，不记录地址、版本字符串、回调标识、动态查询值、请求头、完整认证URL或截图。所示字段列表未见code/success，但预览不能证明原始JSON无重复字段、无尾随内容或精确数字编码。

只读核对当前ParseLogoutSuccess及wirejson：如果原文确为受支持的单个JSONP调用、内层严格JSON对象、ecode编码为整数0，且无重复字段或冲突success，那么当前SDK已能接受该结构。字母/数字/下划线组成的jQuery型回调名也在现有语法内。源码已有JSONP成功用例；本次未运行产品测试，也未将预览重建成真实响应字节。

因此不能把先前的code/error与ecode差异直接解释为本次失败根因：该网页响应同时提供error与ecode，能够满足两种现行业务判断。SDK历史回执为封装other，网页预览则显示回调对象，两者的请求上下文不同。callback是值得优先验证的表示格式差异，但ip、ac_id、会话等差异尚未隔离；不能断言仅加callback即可修复。

预览阶段未改成功判据、复制Cookie/请求头或重放注销；后续Response核对和离线结论见下文。维护者页面显示与操作说明不替代产品状态复核，正式password_core仍未闭合。

### Response确认与离线结论

维护者后续提供Response面板及文本：单个合法标识符回调，参数为双引号字段名的JSON对象，ecode编码为整数0，无末尾分号，所示内容没有额外前后缀。该材料只用于结构核对；没有将原文、地址、回调名或版本字符串落盘。它不是SDK先前收到的响应，也不等于独立三次status验收。

新增 `TestParseLogoutSyntheticBrowserEnvelope`，用新合成回调、文档用途地址及虚构版本字符串复现字段类型和封装。无分号/单分号成功；截断、尾随内容、重复ecode、矛盾success、仅ok但缺失ecode均拒绝。Windows本地Go 1.25.0执行internal/srun和internal/wirejson全包普通测试通过，报告位于 `build/logout-browser-envelope-20261031T075440Z-000c46fc/`。这是纯解析离线检查，没有校园请求、真实profile读取或产品实现修改；未运行Linux/race或构建新包，当前改动只是测试与文档。

### 已确认并离线交付：callback单变量候选修复

推荐下一个独立工作包先更新协议说明及ADR，然后仅在共享注销请求中增加一个固定、合法且不含身份信息的callback参数。保留现有action与username、HTTPS、默认传输及原有会话规则；不增加ip/ac_id/缓存破坏参数、不导入浏览器Cookie或headers、不增加CAS会话退出、不增加请求或自动重试。现有严格JSON/JSONP解码、ecode与可选success判据、最终offline复核均不变；不执行JavaScript。

离线退出条件：假传输验证新增参数及原参数不变，合成JSONP贯穿业务成功与offline复核；业务失败、封装异常和状态仍online均不能返回注销成功。完成相关Windows/Linux回归、文档/脱敏检查，生成普通产品与现有被动观察CLI分别绑定的开发包并自检。不得为了让候选通过而扩大正文白名单或补入浏览器会话。

真实验证仍另行一次password_core授权并由维护者私有终端输入密码，不沿用此次Chrome操作。若仅callback改变后闭环通过，只能说明该修复在该次上下文有效；历史other正文和全部服务端因果仍未直接取得。若失败，按已有观察证据停止归类，不自动加ip/ac_id或其他差异。维护者已于2026-10-31确认，实施契约见 [ADR-0024](../architecture/decisions/ADR-0024-logout-explicit-callback.md)，离线结果与最终包见[状态摘要](../upgrade/status.md#status-20261031-logout-callback)。

## 后续：维护者提供的 Sources 截图

### 结果补记：callback修复后的开发窗口

新包的单次password_core七步全部通过，首次注销完整JSONP对象被现有parser接受，并通过独立offline复核，第二次注销幂等；详见[正式状态页中的开发结果](../upgrade/status.md#status-20261031-password-dev-pass)。本轮没有复制Chrome会话或新增原始材料采集。该结果足以支持当前修复的开发可用性，不把历史未知正文变成已知，也不提升Candidate正式状态。

2026-10-31，维护者按既有Sources搜索 `srun_portal`，报告只有一处匹配。截图显示该处是CAS登录HTML的 `loginForm` action，包含登录返回目标的service和ac_id上下文，并非注销处理函数。可见参数未显示密码或会话令牌；不能据局部截图推断整份HTML无敏感材料。没有保存完整action、查询URL、HTML或截图文件。

这只证明本次所展示Sources中的匹配，不证明网关页面不存在注销代码。当前直接UI读取因Computer Use无法可靠判定浏览器URL而中止；未通过其他通道绕过限制，没有导航、刷新或认证动作。下一步优先检查维护者已有的网关标签及其已加载源码，若不存在则保持缺口，不自动打开新页面。

### 附记：CAS 手机验证码前端线索

维护者随后提供的局部HTML截图可见：手机号输入提示、图形验证码输入与图片元素、短信验证码输入、“获取短信验证码”按钮和“立即登录”按钮。截图中的公开定位标识包括 `getMobileVerifyCode`、`mobile_login_btn`；前端输入maxlength分别可见5（图形验证码）与6（短信验证码）。这些只是本次局部静态HTML属性，不等于服务端长度契约，也不能证明控件当前可见/启用或接口可用。

没有提供实际手机号或验证码；未点击获取按钮、请求验证码图片、发送短信或提交验证码。这里只记录前端存在对应结构，不保存原始页面，也不将短信登录加入本次注销验证范围。

已有 [AUTH-SMS-001](../compatibility/auth-capabilities.md#auth-sms-001手机验证码) 状态仍为 observed_anonymous + detected_only；识别后返回interaction_required、不发送或提交验证码的现行边界保持。[AUTH-SMS-002](../compatibility/auth-capabilities.md#auth-sms-002不可实测分支) 的真实支持缺口未因截图而关闭。后续若另行研究，先核对相关事件处理函数、服务端挑战流程及安全边界；本轮不实施。
