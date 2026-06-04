# 指标契约

`observex` 的 metrics 设计目标是稳定命名、低基数 label 和后端无关。核心包只提供接口、常量和校验工具，不直接依赖 Prometheus、OpenTelemetry Metrics 或其他采集 SDK。

## 接口

`Metrics` 包含四类写入方法：

- `IncCounter`
- `AddCounter`
- `ObserveHistogram`
- `SetGauge`

默认实现是 `NoopMetrics`。生成的基础库应通过 `WithMetrics` 注入实际实现。

## 内置指标

- `client_created_total`
- `client_closed_total`
- `client_errors_total`
- `client_health_status`
- `client_health_latency_ms`
- `client_requests_total`
- `client_request_duration_seconds`
- `client_retries_total`
- `client_inflight`

`New`、`Close` 和 `HealthCheck` 会记录生命周期和健康检查指标。请求耗时、重试和 inflight 指标作为派生库扩展点保留。

## Label 规则

指标名和 label key 必须是 lower snake case 且以字母开头。`Labels` 必须通过 `ValidateLabels` 或 `SanitizeLabels` 处理，避免高基数和敏感字段进入时序系统。

禁止作为 label 的字段包括：

- `trace_id`
- `request_id`
- `correlation_id`
- `user_id`
- `order_id`
- `timestamp`
- `raw_error`
- `sql`
- `payload`

需要排障的高基数字段应进入日志或 trace event，而不是 metrics label。
