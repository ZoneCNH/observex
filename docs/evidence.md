# Evidence 审计与下游 smoke 记录

本页记录 observex L1 可观测性契约库的证据口径。它不是 release manifest 的替代品；正式发布仍以 `make release-check` / `make release-check-extended` 生成的 manifest、sha256、contract hashes 和命令退出码为准。

## 必须保留的证据

每次声称完成或发布前，至少记录：

| 证据 | 命令或 artifact | 期望 |
|---|---|---|
| 依赖边界 | `GOWORK=off go list -deps ./...` 后过滤 forbidden providers | 核心和依赖图中无 Prometheus、OpenTelemetry、Zap、Logrus、Loki、Tempo、Jaeger、Redis、Kafka、Postgres、TDengine、OSS、ClickHouse、x.go、Binance、FRED 或业务 provider |
| Examples smoke | `GOWORK=off go test ./examples/...` | noop、recording/memory、health、metrics、tracing、redaction 示例可运行 |
| 默认 CI | `GOWORK=off make ci` | fmt、vet、lint、test、race、examples、boundary、security、contracts 全部通过 |
| 扩展 gate | `GOWORK=off make release-check-extended`，或记录精确 blocker | property、golden、fuzz smoke、integration、evidence 校验通过 |
| Release manifest | `release/manifest/latest.json`、版本化 manifest 和 sha256 sidecar | manifest 与当前 HEAD、contract hashes、依赖清单、工具版本和 gate 状态一致 |
| Contract hashes | `sha256sum contracts/...` | public API、logs、metrics、traces、health、redaction/schema hash 可复现 |
| 下游 smoke | `GOWORK=off make integration` 或等效 downstream runbook | 下游临时模块可独立 test、contracts、boundary、evidence |
| Public API signature snapshot | `contracts/public_api.snapshot` + `GOWORK=off make contracts` | 导出的 `pkg/observex` 签名与快照一致；有意变更必须显式更新快照 |
| Memory-canonical testkit | `GOWORK=off go test ./pkg/observex ./testkit` | `testkit.Recording*` 包装 public `observex.Memory*`，不维护并行记录模型 |
| 持久下游 blocker | `docs/downstream-evidence.md` | 真实下游证据存在，或以 `external_downstream_unavailable` 明确标记 non-final blocker |

## v0.3.0 发布状态

- v0.3.0 已于 2026-06-04 发布，tag 已推送，GitHub Release 已创建。
- 工程质量深度分析见 [deep-analysis-2026-06-04.md](deep-analysis-2026-06-04.md)，综合评分 82/100。
- P0 问题已修复：`.gitignore` 阻止 agent 运行时状态入库，CI 锁定 `govulncheck@v1.3.0`，commit 历史已 squash。

## 本轮审计结论

- Public API 已由 `contracts/public_api.md` 和 `contracts/public_api.snapshot` 双重锚定；Public API signature snapshot 会捕获导出类型、接口、函数、方法和公共字段的签名漂移。
- release tooling 会生成 versioned manifest、versioned sha256 sidecar、`latest.json` 与 `latest.json.sha256`；`release-evidence-check` 会校验 manifest 状态和 sidecar hash。
- downstream adoption smoke 由 `make integration` 驱动临时模块；真实持久下游证据缺失时，最终 evidence 必须在 `docs/downstream-evidence.md` 记录精确 blocker、exit code 或 `external_downstream_unavailable`，不得把合成 smoke 宣称为真实采用。
- Memory-canonical testkit：`testkit.RecordingLogger`、`RecordingMetrics` 和 `RecordingTracer` 只是 public Memory 记录器的测试辅助包装层。

## Retrospective 候选项

1. 为 readiness 单独扩展更丰富的 status/schema，前提是 L2 需要独立策略。
2. 扩展 downstream smoke runbook，串联 noop、memory、health、metrics、tracing 与 redaction。
3. 将 release evidence 模板固定 DONE 所需的命令、exit code、contract hash 与 blocker 字段。
4. 随 provider 禁用清单变化持续更新 boundary denylist。


## Downstream Adoption Evidence

Release evidence 现在包含 `downstream_adoption` 字段，并要求 `release/downstream/adoption.json` 作为 durable source record。该记录必须包含：

- `fixtures`：当前固定为 `configx` 与 `corekit` 两个渲染下游。
- `commands`：至少记录 `GOWORK=off make integration` 及其期望退出码。
- `blockers`：若没有真实外部下游仓库证据，必须保留 `external_real_downstream` blocker。

`make release-evidence-check` 会先运行 `scripts/check_downstream_evidence.sh`，再校验 manifest、latest manifest 和 sha256 sidecar。最终 release 仍必须使用 `GOWORK=off make release-final-check`，以证明 manifest 与当前 HEAD、source digest、contract fingerprints、dependencies 和 clean tree 一致。
