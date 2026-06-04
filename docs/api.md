# API

## 公共 API

- `Config`：由用户显式提供的配置。
- `Validate`：拒绝无效配置，并返回 `ErrorKindValidation`。
- `Sanitize`：在日志或 Evidence 采集前屏蔽敏感值。
- `New`：基于显式配置创建客户端；拒绝 `nil`、canceled 和 expired context；成功时记录 `client_created_total`、日志和 `observex.New` span。
- `Close`：释放资源，并且必须幂等；成功首次关闭时记录 `client_closed_total`、日志和 `observex.Close` span。
- `HealthCheck`：报告客户端健康状态，JSON 字段必须匹配 `contracts/health.schema.json`；当本次检查的 context deadline 预算短于 `Config.Timeout` 时返回 `degraded`。
- `ReadinessCheck`：与 `HealthCheck` 使用相同状态结构；`HealthReporter` / `ReadinessReporter` 是可注入契约，内置 `NoopHealthReporter` 与 `MemoryHealthReporter`。
- `ErrorKind`：复用 `foundationx` 稳定错误枚举，包含 `config`、`validation`、`connection`、`unavailable`、`timeout`、`auth`、`conflict`、`rate_limit`、`canceled`、`not_found`、`already_exists` 和 `internal`。
- `NewError` / `WrapError` / `MapError` / `IsKind`：创建、包装、映射和判断稳定错误，包装时必须保留 cause。
- `Field` / `Attr`：日志、trace 和事件字段载体，支持 `Secret` 标记和 `ErrorField`；`Attr` 是兼容别名。
- `Redactor` / `DefaultRedactor`：脱敏字段和实现 `foundationx.Sanitizer` 的值。
- `Logger` / `NoopLogger` / `MemoryLogger` / `SlogLogger`：注入式日志接口和标准库 `slog` 适配器；`MemoryLogger` 是 canonical recording logger，`testkit.RecordingLogger` 只包装它。
- `Metrics` / `NoopMetrics` / `MemoryMetrics`：注入式指标钩子；`MemoryMetrics` 是 canonical recording metrics，指标名必须匹配 `contracts/metrics.md` 和 `contracts/metric_naming.md`。
- `Labels` / `ValidateMetricName` / `ValidateLabels` / `SanitizeLabels`：指标 label 约束，拒绝高基数字段和敏感字段。
- `Tracer` / `Span` / `NoopTracer` / `MemoryTracer`：注入式 tracing contract，不直接依赖 OpenTelemetry；`MemoryTracer` 是 canonical recording tracer，`testkit.RecordingTracer` 只包装它。
- `WithTraceID`、`WithRequestID`、`WithCorrelationID`、`WithContextField`、`FieldsFromContext`：context 字段工具。
- `Version`：发布版本。

生成的基础库不得依赖 `x.go`、业务模型、Prometheus、OpenTelemetry、Zap 或 Logrus。上述实现应由上层 adapter 或派生库显式接入。

## 证据对齐

公共 API 变更必须同步更新 `contracts/public_api.md`、`contracts/public_api.snapshot`、相关 schema、examples smoke 和 `docs/evidence.md`。当前 public API 已对齐 goal 要求中的 `HealthReporter`、`ReadinessReporter`、`HealthCheck`、`ReadinessCheck`、`Field` / `Attr`、Noop 与 Memory recorder surface。发布前任何命名或兼容桥变更都必须先用 `internal/tools/apisnapshot` 重新生成 signature snapshot，并更新 contracts、examples 与 evidence。
