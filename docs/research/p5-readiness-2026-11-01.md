---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# P5 候选源码准备记录（2026-11-01）

本页为执行记录；路线以 [roadmap](../upgrade/plan.md#roadmap-20261029) 为准，门禁以 [BHK 正式验收](../operations/bhk-acceptance.md) 为准。本阶段终点是同一新 Candidate 的验收及 promotion 材料准备，不创建最终 tag、不公开 release。准备期间未执行校园请求、真实 profile/凭据读取或认证。

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
