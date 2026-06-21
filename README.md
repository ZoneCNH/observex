# observex

`observex` 是 `github.com/ZoneCNH/observex` 的 L1 vendor-neutral 可观测性契约库。它把日志、指标、追踪、健康检查、上下文字段、脱敏、label policy、测试记录器和 release evidence 固化为轻量 Go 契约，供基础库和业务启动层显式注入适配器，而不是在核心包内绑定任何具体观测后端。

本仓库不是通用基础库模板，也不是业务监控模块。核心包只允许表达稳定运行时契约；具体平台适配、业务字段和下游 provider 行为必须留在 adapter、生成库或业务层。当前实现可复用 `github.com/ZoneCNH/foundationx` 的稳定错误、健康状态和 secret 脱敏原语，但 observex 本身仍作为 L1 observability owner 维护契约、文档、examples、contracts 和 evidence gates。

## 目标

本模块的目标是让基础库从第一天就具备可验证的最小生产观测语义：

- 显式 `Config`，带稳定 validation error 和可脱敏视图。
- `New`、`Close` 和 `HealthCheck` 具备上下文处理、幂等关闭、日志、trace span 和生命周期 metrics。
- `Logger`、`Metrics`、`Tracer`、`Span`、`Field` 和 `Redactor` 是轻量接口，不绑定 Prometheus、OpenTelemetry、Zap、Logrus 或业务 provider。
- `NoopLogger`、`NoopMetrics` 和 `NoopTracer` 保证未注入观测组件时安全运行；public `MemoryLogger`、`MemoryMetrics`、`MemoryTracer` 是 canonical recording model，`testkit` 提供基于 Memory 的 deterministic、race-safe 断言包装用于下游 smoke 和 contract tests。
- `ErrorKind`、health status、field schema、logger schema、tracer schema、metrics schema 和指标命名由 `contracts/` 锁定。
- `make ci`、`make release-check` 和 release manifest 生成可追溯 Evidence。

## 非目标


- 不依赖 `x.go`。
- 不包含 `x.go` 业务模型。
- 不直接依赖 Prometheus、OpenTelemetry、Zap、Logrus 或外部业务 SDK。
- 不隐式读取生产密钥。
- 不创建隐藏全局客户端。

## 标准结构

- `pkg/observex`：公共包 API。
- `internal/`：脱敏、校验和运行时说明等内部辅助代码。
- `testkit/`：可复用测试夹具和断言。
- `examples/`：最小使用示例。
- `contracts/`：JSON schema 和 metrics contract。
- `docs/`：规格、设计、API、配置、测试和发布模板。
- `scripts/`：Harness Gate 脚本。
- `.agent/`：Goal Runtime 工件、Evidence、评审、发布和复盘模板。
- `release/manifest/`：release manifest 模板；`v<version>.json` / `latest.json` 和 sha256 sidecar 由 release gate 生成并作为 Evidence artifact 保存，不作为手写源码提交。

## 文档入口

- [规格](docs/spec.md)：模板能力、验收标准和可追踪性。
- [设计](docs/design.md)：模块边界、公共 API、错误、健康检查和指标设计。
- [API](docs/api.md)：`Config`、`Client`、typed error、logger、metrics、tracer、health JSON 和 redaction contract。
- [配置](docs/config.md)：显式配置、validation 和脱敏规则。
- [日志](docs/logger.md)：`Logger`、`SlogLogger`、context 字段和脱敏字段规则。
- [指标](docs/metrics.md)：`Metrics` 接口、指标常量和 label policy。
- [追踪](docs/tracer.md)：`Tracer`、`Span` 和生命周期 span 约定。
- [Context](docs/context.md)：trace、request、correlation id 与上下文字段。
- [脱敏](docs/redaction.md)：`Redactor`、`Secret` field 和 `foundationx.SecretString` 协作。
- [Label Policy](docs/label-policy.md)：低基数、非敏感、lower snake case 约束。
- [x.go 集成边界](docs/xgo-integration.md)：独立基础库和上层业务仓库的集成边界。
- [生成](docs/generation.md)：从模板渲染 `configx` 等具体下游基础库。
- [错误模型](docs/errors.md)：`ErrorKind`、`NewError`、`WrapError` 和重试语义。
- [可观测性](docs/observability.md)：指标名、健康状态和 JSON 字段。
- [测试策略母版](docs/test-strategy.md)：Required、Extended 和 profile-specific gates。
- [测试](docs/testing.md)：单元、race、contracts、boundary 和 release 验证要求。
- [供应链](docs/supply-chain.md)：可校验 release Evidence、源码摘要、contract 指纹、依赖清单和 CI artifact。
- [发布](docs/release.md)：`release-check`、manifest 字段和 Evidence 规则。
- [Evidence](docs/evidence.md)：本轮文档/证据审计、下游 smoke 状态、known gaps 和复盘候选。
- [下游 Evidence](docs/downstream-evidence.md)：真实下游 adoption 记录、synthetic smoke 与 blocker 区分。
- [结构分析](docs/structural-analysis-2026-06-04.md)：本地结构分、已关闭问题和仍不可声明的外部下游边界。
- [深度分析](docs/deep-analysis-2026-06-04.md)：工程质量评分、结构性问题、改进路线图。
- [ADR-20260604-001](docs/adr/ADR-20260604-001-l1-observability-contract-owner.md)：确认 observex 作为 L1 可观测性契约 owner。

## 命令

本地运行完整 gate 前需要安装 `golangci-lint` 和 `govulncheck`；CI 会显式安装这两个工具。缺少任一工具时，`make lint` 或 `make security` 必须失败，不允许把必需 gate 记录为跳过。

```bash
make ci
make ci-extended
VERSION=v0.3.6 make release-version
VERSION=v0.3.6 make release-check
make release-preflight VERSION=v0.3.6
VERSION=v0.3.6 make evidence
```

如果当前目录被父级 `go.work` 包含，建议使用 `GOWORK=off` 验证本模板的独立性：

```bash
GOWORK=off VERSION=v0.3.6 make release-check
```

## Evidence

完成需要 release manifest 和 CI Evidence。`VERSION=vX.Y.Z` 是 release evidence 入口的必需参数；`make release-version` 会先校验该值与 `pkg/observex/version.go` 一致。默认 artifact 由 `VERSION` 生成，例如 `release/manifest/v0.3.6.json`，也可以通过 `RELEASE_MANIFEST=...` 覆盖路径。release workflow 还必须发布版本化 manifest、版本化 sha256 sidecar、`release/manifest/latest.json` 和 `release/manifest/latest.json.sha256`。manifest 文件是生成产物，不提交到源码历史。manifest 会记录 module、commit、tree SHA、源码摘要、contract 指纹（包含 public API signature snapshot）、依赖清单、工具版本、downstream adoption/blocker、生成时间、工作区状态和 gate 结果，并由 CI 上传为 artifact。`make release-evidence-check` 会先校验 `release/downstream/adoption.json`，再验证 manifest 与当前仓库事实一致，`make release-final-check` 会额外要求工作区为 `clean`。最终完成声明必须包含 `DONE with evidence:`。

## Smoke 覆盖

`go test ./...` 必须覆盖公共包、`internal/`、`contracts/`、`testkit/` 和 `examples/`。当前示例 smoke 测试会验证 `examples/basic`、`examples/config`、`examples/health`、`examples/logger`、`examples/slog`、`examples/metrics`、`examples/tracer`、`examples/noop` 和 `examples/redaction`，防止文档示例和模板行为漂移。
