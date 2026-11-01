---
plan_id: IPGW-META-V1
revision: 2026-08-28-r2
---

# IPGW-Meta 文档中心

`docs/` 是 IPGW-Meta 产品行为、架构、安全策略、公共接口与发布门禁的唯一规范源。实现、测试和 `agent/` 工作包都必须引用这里的稳定 ID；不得在其他目录维护第二份规范。

## 导航

- [当前认证收敛 Roadmap（2026-10-29，位于唯一批准计划）](upgrade/plan.md#roadmap-20261029)
- [实施状态与当前证据摘要](upgrade/status.md#status-20261029)
- [迁移矩阵](upgrade/migration-matrix.md)
- [稳定 ID 索引](stable-ids.md)
- [总体架构](architecture/overview.md)
- [协议正确性](architecture/protocol-correctness.md)
- [安全边界](architecture/security.md)
- [架构决策记录](architecture/decisions/README.md)
- [认证能力矩阵](compatibility/auth-capabilities.md)
- [无 GUI / 无人值守认证手册](runbooks/headless-auth.md)
- [校园网实验室运行手册](runbooks/campus-lab.md)
- [CLI 参考](reference/cli.md)
- [Go SDK 参考](reference/go-sdk.md)
- [JSON CLI 参考](reference/json-cli.md)
- [配置迁移操作手册](operations/config-migration.md)
- [BHKDesktop 本机开发与合成测试](operations/development.md)
- [产品等价开发认证验证（P2）](operations/development-validation.md)
- [CAS 来源诊断与 D4–D15 历史专项（非默认主线）](operations/cas-provenance-diagnostic.md)
- [发布操作手册](operations/release.md)
- [BHK 正式验收 v2 与辅助门禁（P4）](operations/bhk-acceptance.md)
- [离线安装规范](operations/offline-install.md)
- [真实校园网验收规范](operations/live-validation.md)
- [协议研究快照](research/active-implementations.md)
- [P1 旧版与 v1 认证行为对照（2026-10-29）](research/auth-baseline-2026-10-29.md)
- [脱敏证据规范](evidence/README.md)

## 文档治理

2026-10-29 当前路线先做旧版/v1 行为对照，再推进产品等价验证、最小修复和正式验收。roadmap 中“已确认目标”不表示现有工具已具备能力；当前正式门禁继续生效，直到对应规范、ADR、实现与回归完成同步切换。日期明确的旧状态、交接和诊断记录保留为历史，不沿用其中的“下一步”或一次性授权。

- 所有规范文件必须包含相同的 `plan_id` 与当前 `revision`。
- 稳定标识采用 `ARCH-*`、`SEC-*`、`PROTO-*`、`CLI-*`、`SDK-*`、`JSON-*`、`MIG-*`、`REL-*`、`AUTH-*`、`EVID-*` 或 `ADR-*` 前缀。
- Evidence 的日期/序号实例标识采用[证据字段规范](evidence/README.md#evid-auth-001认证证据字段)的 `EVID-YYYYMMDD-NNN` 格式，不是需要标题声明的规范稳定 ID。文档检查区分两者，但不据引用本身证明私有 evidence 存在或验收通过。
- 行为改变必须同时更新批准计划相关引用、对应专题规范、状态与测试；重大取舍新增 ADR。
- `agent/` 只可链接这里的规范，不得复制正文。
- `docs/stable-ids.md` 由 `go run ./cmd/doccheck` 根据 `docs/` 中的权威标题确定性生成，不得手工编辑；CI 使用 `--check` 检查漂移。
- 生成站点只能写入 `docs/_site/`，该目录不进入版本控制。
