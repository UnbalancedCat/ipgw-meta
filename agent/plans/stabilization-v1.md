---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
derived: true
---

# Stabilization v1 工作包索引

本页只定义执行顺序和边界；所有行为与验收含义由链接的 `docs/` 稳定 ID 决定。WP 状态不得写在这里。

<a id="auth-roadmap-work-packages"></a>

## 2026-10-29 当前工作包入口

2026-10-31 P4-A–P4-F 已完成离线交付；验收及环境未覆盖项见 [交付记录](../../docs/research/p4-bhk-gates-2026-10-31.md)。契约及各层边界引用 [ADR-0025](../../docs/architecture/decisions/ADR-0025-bhk-formal-acceptance.md) 和 [BHK 正式验收](../../docs/operations/bhk-acceptance.md)，状态见 [status](../../docs/upgrade/status.md#status-20261031-p4)。下一依赖入口为 P5，不授权正式或真实操作。

阶段定义、退出条件与规则切换以[当前 roadmap](../../docs/upgrade/plan.md#roadmap-20261029)为准，状态以[当前摘要](../../docs/upgrade/status.md#status-20261029)为准。以下只索引依赖、修改边界和验收入口，不复制产品契约。

| WP ID / 阶段 | 依赖 | 允许修改范围 | 权威规范 / 决策入口 | 验收入口 / 交付 |
|---|---|---|---|---|
| `WP-AUTH-ROADMAP-DOCS` / P0 | 已批准文档方案 | plan/status、派生索引、当前交接和最小文档导航 | [roadmap](../../docs/upgrade/plan.md#roadmap-20261029)、[ADR-0005](../../docs/architecture/decisions/ADR-0005-docs-authority.md) | `python scripts/dev.py docs --context tool`；链接/定向脱敏检查、历史保留检查 |
| `WP-AUTH-BASELINE` / P1 | P0 | 只读旧版/v1 源码、研究对照与当前交接 | [roadmap](../../docs/upgrade/plan.md#roadmap-20261029)、[PROTO-RESEARCH-001](../../docs/research/active-implementations.md) | 行为差异矩阵、源码依据及后续验证假设；不执行旧版认证 |
| `WP-AUTH-VALIDATION-PATH` / P2 | P1 | 先规范/ADR，再最小开发观测、绑定与合成测试 | [REL-DEV-AUTH-001](../../docs/operations/development-validation.md)、[ADR-0017](../../docs/architecture/decisions/ADR-0017-development-auth-validation.md)、[PROTO-LOGIN-001](../../docs/architecture/protocol-correctness.md)、[EVID-REDACT-001](../../docs/evidence/README.md) | `python scripts/dev.py auth-dev --context tool`；规范/实现一致性、绑定、无泄漏及请求语义回归；真实操作独立授权 |
| `WP-AUTH-MINIMAL-FIX` / P3 | P1、P2 | 证据支持的共享认证最小修复、对应规范及回归 | [roadmap](../../docs/upgrade/plan.md#roadmap-20261029)、[PROTO-DISCOVERY-001](../../docs/architecture/protocol-correctness.md)、[PROTO-LOGIN-001](../../docs/architecture/protocol-correctness.md) | 独立复现、聚焦/负面合成回归、另行授权的开发闭环回执 |
| `WP-AUTH-BHK-GATES` / P4 | P0；可与 P1–P3 离线工作并行 | 相关专题规范/ADR、evidence/schema/runner/promotion 与回归 | [roadmap](../../docs/upgrade/plan.md#roadmap-20261029)、[REL-LIVE-MATRIX-001](../../docs/operations/live-validation.md)、[REL-PROMOTION-001](../../docs/operations/release.md) | 目标矩阵、历史回执及拒绝路径的跨层一致性回归；验收前不切换门禁 |
| `WP-AUTH-CANDIDATE-ACCEPTANCE` / P5 | P3、P4 和其他现行 hard gate | 授权后的正式 CI/Candidate/evidence/晋升流程 | [roadmap](../../docs/upgrade/plan.md#roadmap-20261029)、[REL-CANDIDATE-001](../../docs/operations/release.md)、[REL-INSTALL-003](../../docs/operations/offline-install.md)、[EVID-REVIEW-001](../../docs/evidence/README.md) | 全仓门禁、六平台原生安装与正式 evidence/身份/晋升核验；不是 P0 执行范围 |

既有开发命令见[开发入口](../../docs/operations/development.md)。P3 静态语法观察的边界由 [ADR-0020](../../docs/architecture/decisions/ADR-0020-bounded-static-cas-syntax.md) 定义，离线验收入口为 `python scripts/dev.py auth-static-context --context tool`；真实窗口另行单次授权，不修改产品选键。未来入口尚未实现时记录为未具备，不能杜撰通过回执。D4–D15 对应诊断工作包保留为历史索引，重新启用条件见 roadmap，不作为当前必经依赖。

## r2 原始工作包索引

下表保留已有工作包及依赖追溯，不是当前待执行队列；其中迁移、初始化及历史构建不应重复执行。正式发布工作包仍受现行专题规范约束，目标切换须完成 P4。

| WP ID | 依赖 | 允许修改范围 | 权威规范 | 验收命令 / 证据 |
|---|---|---|---|---|
| WP-M0-PREFLIGHT | 无 | 只读 Git/GitHub/tool/worktree 核验 | [`REL-WINDOW-001`](../../docs/operations/release.md) | refs/tool hash/secret scan/test 结果；无写入 |
| LAB-DISCOVER | 无 | ZOS 只读能力检查 | [`REL-LAB-002`](../../docs/runbooks/campus-lab.md) | 隔离能力记录；不创建 VM |
| WP-R2-SPEC-A | WP-M0-PREFLIGHT | plan/status/ADR/handoff | [`ADR-0005`](../../docs/architecture/decisions/ADR-0005-docs-authority.md) | 文档 diff 与链接检查 |
| WP-R2-SPEC-B | WP-R2-SPEC-A | release/offline/live/evidence/auth docs | [`ADR-0009`](../../docs/architecture/decisions/ADR-0009-separated-live-test-plane.md) | 稳定 ID 与交叉链接检查 |
| WP-R2-SPEC-C | WP-R2-SPEC-B | revision、doccheck、Agent 派生索引 | [`EVID-REVIEW-001`](../../docs/evidence/README.md) | `go test ./internal/doccheck`; `doccheck --check` |
| WP-M0-FREEZE-AUDIT | WP-R2-SPEC-C | `.gitattributes`、index 审计 | [`SEC-HISTORY-001`](../../docs/architecture/security.md) | EOL/name-status/numstat/secret scan/tests；不提交 |
| WP-M0-FREEZE-COMMIT | WP-M0-FREEZE-AUDIT | staged tree、安全 archive、受限备份 | [`REL-APPROVAL-001`](../../docs/operations/release.md) | index 二次核对、archive scan、backup hash |
| WP-M0-REWRITE-LOCAL | WP-M0-FREEZE-COMMIT | 隔离 mirror 与 replace rules | [`SEC-HISTORY-001`](../../docs/architecture/security.md) | fsck、全 refs/reflog scan、tree/tag 对照 |
| WP-M0-REWRITE-REMOTE | WP-M0-REWRITE-LOCAL | 精确远端 commit/tag refs | [`REL-APPROVAL-001`](../../docs/operations/release.md) | atomic per-ref lease push；远端复核 |
| WP-M0-VERIFY-REHOME | WP-M0-REWRITE-REMOTE | clean clone 与旧工作区隔离记录 | [`SEC-HISTORY-001`](../../docs/architecture/security.md) | fresh clone refs/tags/scan/tests/tree |
| WP-M0-RENAME-GOVERN | WP-M0-VERIFY-REHOME | module/origin、仓库名、rulesets | [`ADR-0002`](../../docs/architecture/decisions/ADR-0002-compatibility-and-cutover.md) | fresh clone；main/tag 保护核验 |
| WP-BASELINE-VERIFY | WP-M0-RENAME-GOVERN | 只读 M1/M2 基线 | [`REL-CI-001`](../../docs/operations/release.md) | test/race/vet/doccheck/secret scan |
| WP-CAS-PROVENANCE-DESIGN | WP-BASELINE-VERIFY | 诊断设计、ADR 草案、内部合成观察器 | [`ADR-0016`](../../docs/architecture/decisions/ADR-0016-cas-provenance-before-key-selection.md)、[`REL-CAS-DIAG-001`](../../docs/operations/cas-provenance-diagnostic.md) | doccheck；cas-observe 合成结果；无产品选键修改 |
| WP-CAS-PROVENANCE-OFFLINE | WP-CAS-PROVENANCE-DESIGN | 独立分析器/闭集报告，随后绑定 verifier/runner | [`REL-CAS-DIAG-001`](../../docs/operations/cas-provenance-diagnostic.md) | 分阶段落实草案；来源/边界/脱敏/整链绑定测试；不接入真实采集 |
| WP-CAS-ACQUIRE-DEV | WP-CAS-PROVENANCE-OFFLINE | 独立匿名传输、原生可信启动器、开发快照 | [`REL-CAS-DIAG-001`](../../docs/operations/cas-provenance-diagnostic.md) | `python scripts/dev.py cas-acquire --context tool`；loopback/TLS/预算/绑定及确认门禁；真实执行仅维护者单次确认 |
| WP-M2-CONFIG-CLOSE | WP-BASELINE-VERIFY | `internal/config` 与迁移测试 | [`MIG-TRANSACTION-001`](../../docs/operations/config-migration.md) | 失败注入、Unix 权限、keyring backend |
| WP-M2-INSTALL-UNIX | WP-M2-CONFIG-CLOSE | `install.sh` 与 Unix 测试 | [`REL-INSTALL-001`](../../docs/operations/offline-install.md) | 离线/路径/权限/failpoint 测试 |
| WP-M2-INSTALL-WINDOWS | WP-M2-CONFIG-CLOSE | `install.ps1` 与 Windows 测试 | [`REL-INSTALL-002`](../../docs/operations/offline-install.md) | ACL/reparse/事务 failpoint 测试 |
| WP-M2-INSTALL-NATIVE | WP-M2-INSTALL-UNIX, WP-M2-INSTALL-WINDOWS | 原生 runner 矩阵 | [`REL-INSTALL-003`](../../docs/operations/offline-install.md) | 六平台 smoke；三代表平台完整故障矩阵 |
| WP-M3-LIVEGATE-SCHEMA | WP-M2-INSTALL-NATIVE | live-gate schema/types/tests | [`REL-LIVEGATE-001`](../../docs/operations/live-validation.md) | schema、枚举、退出码和泄漏测试 |
| WP-M3-LIVEGATE-RUNNER | WP-M3-LIVEGATE-SCHEMA | maintainer-only runner | [`REL-LIVEGATE-002`](../../docs/operations/live-validation.md) | 合成状态机、清理权、durability 测试 |
| WP-M3-CANDIDATE | WP-M3-LIVEGATE-RUNNER | candidate workflow/manifest/packaging | [`REL-ATTEST-001`](../../docs/operations/release.md) | 一次构建、digest、attestation 测试 |
| WP-M3-PROMOTION | WP-M3-CANDIDATE | promotion workflow/lock 校验 | [`REL-PROMOTION-001`](../../docs/operations/release.md) | no-build、签名、draft/re-download 测试 |
| WP-M0-NONBLOCKING-GATES | WP-M3-PROMOTION, LAB-DISCOVER | 决策/发布/安全/状态、Makefile gate 与合成测试 | [`ADR-0011`](../../docs/architecture/decisions/ADR-0011-nonblocking-m0-governance.md) | M0 状态无关性；M1/M2/M3 缺失/重复/错误状态 fail closed |
| LAB-PROVISION | LAB-DISCOVER, WP-M3-PROMOTION | 管理/测试 VM 与匿名预检 | [`REL-LAB-001`](../../docs/runbooks/campus-lab.md) | topology/status 预检；无凭据 |
| RC-BUILD | WP-M0-NONBLOCKING-GATES, LAB-DISCOVER | 受保护 main 的 candidate-set | [`REL-CANDIDATE-001`](../../docs/operations/release.md) | artifact ID/digest/hash/attestation |
| LAB-TRANSFER | RC-BUILD | 本地下载与远端私有目录 | [`REL-TRANSFER-001`](../../docs/operations/live-validation.md) | 本地/远端一致 hash；远端无重建 |
| LAB-PASSWORD-NAS/BHK | LAB-TRANSFER | NAS/BHK 私有 TTY 和私有 evidence | [`REL-LIVE-MATRIX-001`](../../docs/operations/live-validation.md) | 指定 password suites |
| LAB-QR-NAS | LAB-PASSWORD-NAS/BHK | NAS 私有 TTY 和私有 evidence | [`AUTH-QR-002`](../../docs/compatibility/auth-capabilities.md) | terminal-qr suite |
| LAB-EVIDENCE | LAB-QR-NAS | 私有 bundle 审阅、公开摘要/lock | [`EVID-BUNDLE-001`](../../docs/evidence/README.md) | schema/hash/禁止字段人工复核 |
| WP-M3-RELEASE | LAB-EVIDENCE | 签名 tag、draft 与 release metadata | [`REL-PROMOTION-001`](../../docs/operations/release.md) | 签名验证、逐资产下载/hash |
| WP-SECURE-DISPOSAL | WP-M3-RELEASE | 已批准旧工作区/敏感备份目标 | [`REL-APPROVAL-001`](../../docs/operations/release.md) | 精确目标确认与处置记录 |

## 并行与冲突约束

- M0 冻结、历史重写、远端更新、rehome 与仓库治理严格串行；每个高风险窗口重新取得维护者批准。
- Unix 与 Windows 安装器实现可在配置契约稳定后分开执行，但原生矩阵只能测试已冻结的同一实现版本。
- live-gate、candidate 和 promotion 按表中依赖串行；真实实验室不重建候选。
- NAS/BHK 的真实认证窗口由维护者执行私有交互；Agent 只处理固定白名单 evidence。

## 全局完成检查

```text
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
secret scan
go run ./cmd/doccheck --check
六平台原生安装门禁
```

命令尚未在仓库提供时，对应 WP 仍未完成；不得用手工口头确认代替门禁。


## P3 当前执行入口（2026-10-30）

2026-10-31 后续注销工作包引用 [ADR-0024](../../docs/architecture/decisions/ADR-0024-logout-explicit-callback.md) 和 [PROTO-LOGOUT-001](../../docs/architecture/protocol-correctness.md#proto-logout-001幂等注销)。依赖P1/P2及网页静态/封装对照；修改边界仅共享请求callback及相关测试，验收入口为两端auth-dev、Linux cross及文档/脱敏检查；交付状态以[当前摘要](../../docs/upgrade/status.md#status-20261031-logout-callback)为准，真实窗口独立授权。

`WP-AUTH-MINIMAL-FIX` 当前范围与退出条件引用 [ADR-0021](../../docs/architecture/decisions/ADR-0021-cas-key-use-relation.md) 及 [开发验证](../../docs/operations/development-validation.md#p3-key-use-relation)。依赖 P1/P2 已满足；离线验收使用 `scripts/dev.py key-relation`、test/vet/cross/docs 和 auth-dev 绑定检查。真实验证需独立授权，不将离线交付标为 P3 登录闭环完成。

## P4 子工作包执行索引

契约唯一来源为 [BHK 正式验收](../../docs/operations/bhk-acceptance.md)；完成状态只记 [status](../../docs/upgrade/status.md#status-20261031-p4)。

| 子包 | 依赖 | 修改边界 | 验收入口 |
|---|---|---|---|
| P4-A | WP-AUTH-ROADMAP-DOCS | [ADR-0025](../../docs/architecture/decisions/ADR-0025-bhk-formal-acceptance.md)、[REL-BHK-GATES-001](../../docs/operations/bhk-acceptance.md) / [EVID-BHK-002](../../docs/operations/bhk-acceptance.md) | doccheck |
| P4-B | P4-A | 版本解码、历史审计、helper 自绑定 | dev.py p4；test-promotion.py |
| P4-C | P4-B | [REL-BHK-PATH-001](../../docs/operations/bhk-acceptance.md)、BHK 拓扑监测、QR 透传 | 假进程/合成拓扑；Windows/Linux 回归 |
| P4-D | P4-A | [EVID-BHK-AUX-001](../../docs/operations/bhk-acceptance.md)、Git 快照、隔离安装验证器 | Windows install；来源/测试事件负面用例 |
| P4-E | P4-B、P4-C、P4-D | Wi-Fi 辅助、两摘要、[REL-BHK-PROMOTION-002](../../docs/operations/bhk-acceptance.md) | promotion 合成端到端、scope/身份负面测试 |
| P4-F | P4-B–P4-E | 集成、平台依赖隔离、交接 | test/race/vet/quality/cross/docs；定向扫描 |
| P5 进入准备 | P4-F、WP-AUTH-MINIMAL-FIX | 仅交接缺口和执行入口 | 新 Candidate/真实窗口/提交发布另行授权 |
