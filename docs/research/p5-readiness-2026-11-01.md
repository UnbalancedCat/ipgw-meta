---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# P5 候选源码准备记录（2026-11-01）

本页为执行记录；路线以 [roadmap](../upgrade/plan.md#roadmap-20261029) 为准，门禁以 [BHK 正式验收](../operations/bhk-acceptance.md) 为准。本阶段终点是同一新 Candidate 的验收及 promotion 材料准备，不创建最终 tag、不公开 release。准备期间未执行校园请求、真实 profile/凭据读取或认证。

## PR #29 与首轮 CI 修正

维护者批准的 344 文件清单已在新分支 `codex/bhk-v1-acceptance` 提交为 `2fc0d1644920a2aefb10a8859e4baebe84cf0501`，并推送、创建 [PR #29](https://github.com/UnbalancedCat/ipgw-meta/pull/29)。原分支及旧包保留；未合并、未启动 Candidate。下方“尚未提交”等段落是此次授权之前的准备快照。

首轮 CI（CI run `41266117777`、Native install run `41266117667`）不能判通过：Linux 普通测试/race、六平台交叉构建、资产打包、Linux 两架构及 macOS 两架构原生安装成功；文档、macOS 普通测试、Windows 普通测试及两架构原生安装失败。失败定位和修正均限于测试/文档：

- 文档检查：旧交接中的两个 ignored build 回执链接在干净检出不存在。保留路径和历史事实，改为本地产物代码文本，不把私有回执纳入仓库或放宽链接检查。
- macOS：新 livegate 测试引用只在 Linux/Windows 定义的私有文件系统夹具。把三个文件系统用例完整移入相同平台约束文件；纯 schema、进程投影和拓扑测试继续跨平台。v1 发布拒绝在文件系统操作之前验证。
- macOS：auxgate 接受路径夹具先解析临时目录的系统别名；生产路径检查不变。
- Windows：管理员进程新建目录可能默认由 Administrators 所有。仅在当前进程为管理员、初始所有者确为该组、且刚创建的测试目录为空时，将夹具所有者设为当前用户；其他情况拒绝。之后仍验证私有 ACL 和所有者，产品安装器不变。

修正后本地验证：

| 检查 | 结果与本地回执 |
|---|---|
| Windows P4 测试/vet | pass；`build/dev-runs/windows/20261101T025741Z-9c8cf4b5-p4/result.json` |
| Linux P4 测试/vet/race/有界 fuzz | pass；`build/dev-runs/linux/20261101T025800Z-ac3e3e03-p4/result.json` |
| Windows 普通进程安装回归 | pass；`build/dev-runs/windows/20261101T030037Z-21d3eec9-install/result.json`；未选择 Candidate 的 smoke 不计正式验收 |
| Windows 管理员安装回归 | 四组全部实际 pass、无 skip；`build/p5-preparation-20261101/a092bc24d9a64aba9c3b44fb2ba2132b.install.result.json`；新目录初始 owner 为 Administrators，测试程序 SHA-256 `12a83198e02ffa81136d76601ec10145fdc03395d205fbd31ec3a7ca1b91cdc1`，运行前后相同 |
| macOS 两架构受影响测试、Windows arm64 安装测试编译 | pass；`build/p5-preparation-20261101/ci-fixes-5621da3910f14ef6b407426ce4f66b45/result.json`；仅编译，不冒充原生执行 |
| 无 ignored build 的源码快照 doccheck | pass；同上 clean-source-doccheck 步骤 |

上述修正尚待补充提交/推送及新 CI；不能用本地通过覆盖首轮远端失败。没有产品行为修改，也没有正式 Candidate 或正式认证证据。合并、Candidate workflow 及真实窗口仍需各自后续授权。

## P4 未覆盖项补齐

固定 Go1.25.0 离线编译 internal/config 测试程序，SHA-256 为 `f77b81cc5871a4e5f50b4197f1e39ef9e44eb871d4ada72c422f5ec3989a9bc4`。普通 Unbal 进程的五项测试仍 skip；随后通过 Windows UAC 在同一账户原生管理员进程运行相同字节，五项均实际 run/pass，退出码 0，无超时，程序前后哈希一致：

- TestReadRestrictedPasswordRejectsFinalReparsePoint
- TestValidateRestrictedWindowsParentsRejectsDirectorySymlink
- TestStoreLoadWindowsRejectsConfigReparsePoint
- TestStoreLoadWindowsRejectsDirectoryReparsePoint
- TestStoreSaveWindowsRejectsLockReparsePoint

只创建独立合成测试目录；没有启用 Developer Mode、修改永久账户权限、真实安装或仓库 ACL。回执为 `build/p5-preparation-20261101/00991b8b8061474091e20c427f4a0b22.result.json`。首次普通进程回执的 exit_code 为 null，是 PowerShell Start-Process 退出码获取问题；已改用显式 Process 对象，保留原回执，不据其零值推断通过。五项 skip 本身真实存在，管理员回执是独立重测。

发布手册 REL-LIVE-001 正文残留的 NAS 必需矩阵已改为三组 BHK 认证加两项辅助检查，历史日期记录保留。修正不改变已经批准的 P4 契约。

准备阶段质量检查通过：`build/dev-runs/linux/20261101T021602Z-c3083ec5-quality/result.json` 覆盖 promotion、milestone、Candidate、actionlint、gitleaks 规则及历史/工作树扫描；文档检查 `build/dev-runs/windows/20261101T021551Z-0ac8e63e-docs/result.json` 通过。没有修改产品代码，前次全仓/race/cross 报告继续作为对应源码的离线依据；后续新 PR 必须重新取得 CI 结果。

## 源码和远端基线

本地分支 codex/cas-submit-field-closure，HEAD b4eeb760945f0568ca163d957ddd55df2afa86c4；P3/P4 尚未提交。原 handoff 保持历史字节。拟提交包含共享取键/注销修复、对应回归、独立开发工具与历史诊断源码、P4 正式门禁及规范；不删除历史诊断，也不把它们装入产品。已有六平台依赖检查拒绝诊断和 JS runtime 进入产品闭包；静态解析器依赖保持锁定。

逐文件清单与 SHA-256 在 `build/p5-preparation-20261101/review-manifest.json`，可读路径在同目录 review-paths.txt。排除 build/release、私有材料、Python 字节码缓存；不执行 git add 全目录。清单仅用于审阅，不表示已暂存或已提交。

GitHub 只读核验：main 仍为 dac245947b9bd11b5489e2fd87ca2823c2d2dac5，签名 verified/valid；旧 PR #28 已合并，当前无开放 PR；最终 v1.0.0 ref 返回 404。main 的 22 项 checks 当时均成功，包含 Candidate 要求的 13 项，但仅属于旧 SHA。保护由 active main-v1-protection ruleset 实施，要求 PR、签名、strict checks，不能把旧 branch protection 字段解释为没有保护。新代码必须走新 PR；合并后对新 SHA 重新核验全部条件。

## 提交和后续检查点

准备一个新的 codex/bhk-v1-acceptance 集成分支和 PR，目标 main；不重开 PR #28、不强推。提交标题拟为 `fix: converge v1 authentication and BHK acceptance gates`。提交/推送/PR 的具体清单批准后才执行，合并及 Candidate workflow 另有检查点。

新 Candidate 到位后先完成原生安装辅助记录，再分别安排有线 password_core、terminal_qr。Wi-Fi 在切网前备齐包、依赖、命令和本地停止/恢复说明；维护者断开有线后独立运行辅助检查，通过后才运行已明确授权的一次 password_core。恢复有线只恢复管理连接，不证明账号已 offline；回执与身份逐项核验后才形成公开摘要。

正式认证仍 blocked/live_unverified，M3 in_progress。当前没有新 Candidate、五份正式证据或可用于发布的 promotion lock。任何旧开发包、旧 Candidate 和本页合成测试均不能替代它们。
