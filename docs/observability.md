# 可观测性模板

## 占位符

- `observex`
- `observex`

## 指标

使用 `contracts/metrics.md`、`contracts/metrics.schema.json` 和 `contracts/metric_naming.md` 中的 metrics contract。模板内置的最小指标包括：

- `client_created_total`
- `client_closed_total`
- `client_errors_total`
- `client_health_status`
- `client_health_latency_ms`
- `client_requests_total`
- `client_request_duration_seconds`
- `client_retries_total`
- `client_inflight`

生命周期指标由 `New`、`Close` 和 `HealthCheck` 直接记录；请求、耗时、重试和 inflight 指标作为生成具体库后的扩展 contract。所有指标名必须是 lower snake case 且以字母开头；label 必须是低基数、非敏感字段。`trace_id`、`request_id`、`correlation_id`、`user_id`、`order_id`、`timestamp`、`raw_error`、`sql` 和 `payload` 不能作为 metrics label。

## 健康检查

持有资源的客户端必须暴露 `HealthCheck(context.Context)`。返回值必须使用 `contracts/health.schema.json` 中的字段名：

- `name`
- `status`
- `message`
- `checked_at`
- `latency_ms`
- `metadata`

`status` 只能是 `healthy`、`degraded` 或 `unhealthy`。未初始化、已关闭、`nil` context、canceled context 都必须返回 `unhealthy`。已初始化且未关闭的 client 如果本次检查的 context deadline 预算短于 `Config.Timeout`，必须返回 `degraded`，并继续记录 `client_health_status` 和 `client_health_latency_ms`，其中 `status` label 为 `degraded`。

## 日志

日志通过 `Logger` 接口注入，默认 `NoopLogger`。`SlogLogger` 将 `FieldsFromContext(ctx)` 与调用点字段合并后写入 `log/slog`。字段进入日志前必须通过 `Redactor`，`Secret` 字段、敏感 key 和实现 `foundationx.Sanitizer` 的值都不得泄露原文。

## Tracing

追踪通过 `Tracer` 与 `Span` 接口注入，默认 `NoopTracer`。`New` 和 `Close` 会创建生命周期 span；派生库可以在请求、重试、外部 I/O 和健康检查中继续使用同一接口。模板只定义 contract，不直接引入 OpenTelemetry。

## Context 字段

`WithTraceID`、`WithRequestID`、`WithCorrelationID` 和 `WithContextField` 用于在调用链上传递可观测性字段。context 字段适合进入日志或 trace event，不适合进入 metrics label。

本模板不得依赖 `x.go`。
