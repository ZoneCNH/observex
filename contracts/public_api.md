# Public API Contract

`pkg/observex` 的公开 API 是稳定集成面。新增能力优先以接口和轻量值对象表达，不能引入具体后端依赖。

必须保留的核心类型与函数：

- `Config`
- `Client`
- `New`
- `Logger`
- `NoopLogger`
- `MemoryLogger`
- `LogRecord`
- `LogLevel`
- `SlogLogger`
- `Metrics`
- `NoopMetrics`
- `MemoryMetrics`
- `MetricRecord`
- `MetricKind`
- `Tracer`
- `NoopTracer`
- `MemoryTracer`
- `Span`
- `SpanRecord`
- `SpanEvent`
- `Field`
- `Attr`
- `Secret`
- `Redactor`
- `DefaultRedactor`
- `Labels`
- `ValidateMetricName`
- `ValidateLabels`
- `SanitizeLabels`
- `HealthStatus`
- `HealthReporter`
- `ReadinessReporter`
- `HealthCheck`
- `ReadinessCheck`
- `NoopHealthReporter`
- `MemoryHealthReporter`
- `Error`
- `ErrorKind`
- `ErrorKindCanceled`
- `ErrorKindNotFound`
- `ErrorKindAlreadyExists`
- `MapError`
- `WithLogger`
- `WithMetrics`
- `WithTracer`
- `WithTraceID`
- `TraceID`
- `WithRequestID`
- `RequestID`
- `WithCorrelationID`
- `CorrelationID`


## 签名快照 Gate

`contracts/public_api.snapshot` 是公开导出 API 的签名级快照，覆盖导出类型、接口、函数、常量、变量、字段和方法。`make contracts` 会同时校验符号清单和签名快照；公共 API 发生有意变更时，先评审兼容性，再运行：

```bash
GOWORK=off go run ./internal/tools/apisnapshot ./pkg/observex > contracts/public_api.snapshot
```

`scripts/check_public_api_snapshot.sh` 和 `make contracts` 必须保持通过。禁止只更新文档清单而不更新签名快照。签名快照必须保持 provider-neutral，不能暴露具体后端或业务模型。

内联更新命令：`GOWORK=off go run ./internal/tools/apisnapshot ./pkg/observex > contracts/public_api.snapshot`。
