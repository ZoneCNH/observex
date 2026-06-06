# 测试模板

## 占位符

- `observex`
- `observex`

## 测试策略

本模板遵循 [测试策略母版](test-strategy.md)。默认强制 SDD、ATDD、TDD、Contract、Boundary、Security、Integration Smoke 和 Evidence；默认增强 Property、Fuzz Smoke、Golden、Compatibility 和 Observability；Chaos、Mutation、Long Soak 和 Full E2E 只由派生库按 profile 启用。

## 测试模式矩阵

| 模式 | 是否默认强制 | Gate | 说明 |
|---|---:|---|---|
| SDD | 是 | `docs/spec.md` | 规格先行 |
| ATDD | 是 | `docs/testing.md` | 验收标准先行 |
| TDD / Unit | 是 | `make test` | 核心逻辑测试 |
| Race | 是 | `make race` | 并发安全 |
| Contract | 是 | `make contracts` | schema、metrics、errors、public API signature snapshot |
| Boundary | 是 | `make boundary` | 模块边界 |
| Security | 是 | `make security` | `govulncheck` 和 secret scan |
| Integration Smoke | 是 | `make integration` | 模板渲染后可运行 |
| Evidence | 是 | `make evidence` / `make release-check` | release manifest 与 gate 结果 |
| Property | 推荐 | `make property` | 不变量测试 |
| Fuzz Smoke | 推荐 | `make fuzz-smoke` | 边界输入测试 |
| Golden | 推荐 | `make golden` | 稳定输出回归 |
| Compatibility | 推荐 | `make contracts` | 公共契约兼容性 |
| Observability | 推荐 | `make contracts` / `make test` | metrics、health、logs |
| Chaos | 按库启用 | profile-specific | 存储和消息库 |
| Mutation | 按库启用 | critical-only | 高风险逻辑 |
| Full BDD | 不默认 | docs only | 基础库不强制 |
| Full DDD | 不作为测试模式 | boundary rule | 只保留边界思想 |

## 必需 Gate

本地执行 gate 前必须可用：

- `golangci-lint`
- `govulncheck`

缺少上述工具时，`make lint` 或 `make security` 必须失败。

- `make fmt`
- `make vet`
- `make lint`
- `make test`
- `make race`
- `make boundary`
- `make security`
- `make contracts`
- `make integration`
- `make evidence`

## 扩展 Gate

扩展 gate 推荐在发布前、公共 API 变更、contract 变更、schema 变更、metrics 变更和安全敏感变更时运行：

- `make property`
- `make fuzz-smoke`
- `make golden`
- `make ci-extended`
- `make release-check-extended`

`make ci` 必须保持轻量，扩展 gate 不进入默认 `make ci`。

## 必需覆盖范围

- `go test ./...` 必须覆盖公共包、`internal/`、`contracts/`、`testkit/` 和 `examples/`。
- 配置校验。
- 配置脱敏。
- typed error kind 和 wrapped cause。
- 客户端创建、取消 context、过期 context。
- 幂等关闭、zero-value client、取消 context。
- 健康与非健康状态检查。
- 健康检查 JSON 字段 contract。
- 生命周期 metrics 和健康 metrics。
- `Logger` context 字段合并和脱敏。
- `SlogLogger` 输出不得泄露 secret。
- `Tracer` / `Span` 生命周期事件。
- `ValidateMetricName`、`ValidateLabels` 和 `SanitizeLabels` 的命名、高基数和敏感值规则。
- `Redactor` 与 `foundationx.Sanitizer` 的协作。
- `contracts/` 与公共常量同步。
- `contracts/public_api.snapshot` 必须与 `pkg/observex` 导出 API 同步；有意变更需运行 `GOWORK=off go run ./internal/tools/apisnapshot ./pkg/observex > contracts/public_api.snapshot`，再执行 `GOWORK=off make contracts` 并评审兼容性。
- `contracts/config.schema.json` 与 `Config` 字段映射同步。
- `scripts/render_template.sh` 生成的临时 `configx` 可以通过 `GOWORK=off go test ./...`。
- `Config.Sanitize` 的 secret 不变量必须由 property test 覆盖。
- `Config` 边界输入必须由 fuzz-smoke 覆盖。
- `HealthStatus` JSON 公共输出必须由 golden test 锁定。

## 示例与 testkit Smoke

- `examples/basic` 必须输出当前 module name。
- `examples/config` 必须输出脱敏后的 secret 值。
- `examples/health` 必须输出 `healthy`。
- `examples/noop` 必须输出 `noop healthy`。
- `examples/logger` 必须输出 context 字段和脱敏字段。
- `examples/slog` 必须输出 `slog` JSON 且不泄露原始 secret。
- `examples/metrics` 必须输出已排序 label 的 counter 示例。
- `examples/tracer` 必须输出 span start/event/end 生命周期。
- `examples/redaction` 必须输出 `***`。
- `testkit` 必须验证 `Config("fixture")` 生成可通过 `Validate` 的测试配置。
- Memory-canonical testkit 必须验证 `testkit.RecordingLogger`、`RecordingMetrics` 和 `RecordingTracer` 包装 public `observex.Memory*`，记录 shape、序列号和 reset 行为不得漂移。
- `testkit.RequireNoError` 必须接受 `nil`，作为生成库测试断言的最小契约。
- `testkit.RequireGolden` 必须比较稳定公共输出，并在 mismatch 时输出 expected 和 actual 上下文。
- `testkit.AssertNoSecretLeak` 必须检测输出中不存在原始敏感值。
- `testkit.RecordingLogger`、`RecordingMetrics` 和 `RecordingTracer` 必须支持派生库验证注入式可观测性，并且必须包装 public Memory 记录器而不是维护平行记录模型。

生成的基础库必须保持测试独立于 `x.go`。


## Downstream Evidence

`make integration` 必须渲染并验证 `configx` 与 `corekit` 两个 fixture，并把该 smoke 的 durable reference 保存在 `release/downstream/adoption.json`。该 JSON 必须分离 `fixture_smoke` 与 `real_adoption`：fixture smoke 记录合成命令状态和退出码；真实下游不可用时，`real_adoption` 必须保留 `external_real_downstream` blocker。`scripts/check_downstream_evidence.sh` 是 release evidence 的前置门禁：缺少任一分支、fixture 命令状态/退出码或真实下游 blocker 时，`make release-evidence-check` 必须失败。
