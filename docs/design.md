# DESIGN-observex-v1.0

## 架构

生成的库是独立 Go module。公共 API 位于 `pkg/observex`，内部辅助代码位于 `internal/`，contracts 位于 `contracts/`，运行 Evidence 位于 `release/manifest/`。`scripts/render_template.sh` 是模板到具体基础库的唯一内置渲染入口。

## 公共 API

模板暴露 `Config`、`SanitizedConfig`、`Client`、`New`、`Close`、`Option`、`HealthCheck`、`ErrorKind`、`NewError`、`WrapError`、`MapError`、`IsKind`、`Field`、`Logger`、`SlogLogger`、`Metrics`、`Tracer`、`Span`、`Redactor`、`Labels`、context helper、指标常量、`ModuleName` 和 `Version`。

## 配置

调用方必须显式传入配置。生成的库不得隐式读取 `x.go` 生产密钥路径。`Validate` 使用稳定 validation error 表达缺失字段和负数 timeout，`Sanitize` 只返回可安全记录的脱敏视图。`contracts/config.schema.json` 使用外部字段 `timeout_ms`，并通过 contract 回归测试锁定到 `Config.Timeout`。

## 错误模型

错误使用 `foundationx` 的稳定 `ErrorKind` 枚举，并通过 `Unwrap` 支持错误包装。上下文超时归类为 `timeout` 且可重试；上下文取消归类为 `canceled` 且不可重试。公共判断使用 `IsKind`，不要依赖错误字符串。

## 日志、字段与脱敏

`Field` 是公共可观测性字段载体，支持字符串、数字、布尔、时间、持续时间、错误和任意值。`Secret` 字段和命中敏感 key 的字段必须被 `Redactor` 处理为 `***`；实现 `foundationx.Sanitizer` 的值由其自身 `Sanitize` 输出。`SlogLogger` 是标准库 `log/slog` 适配器，默认使用 no-op 输出，避免隐藏全局 logger。

## Context

context helper 使用私有 key 保存 `trace_id`、`request_id`、`correlation_id` 和附加字段。`FieldsFromContext` 会把这些值合并到日志字段中，避免每个调用点手工拼接。

## 健康检查

持有资源的客户端暴露 `HealthCheck(context.Context)`，并返回 `healthy`、`degraded` 或 `unhealthy`。返回结构使用 `name`、`status`、`message`、`checked_at`、`latency_ms` 和 `metadata` JSON 字段；nil client、零值 client、已关闭 client、nil context 和已取消 context 都必须返回 `unhealthy`。已初始化且未关闭的 client 如果本次检查的 context deadline 预算短于 `Config.Timeout`，必须返回 `degraded`，并在 `metadata` 中记录降级原因。

## 指标

指标通过 `Metrics` 接口注入，默认使用无操作实现。模板锁定 client 生命周期、错误、健康检查、请求、重试和 inflight 指标名称，具体列表以 `contracts/metrics.md` 和 `pkg/observex` 指标常量为准。指标名和 label 必须通过 `ValidateMetricName`、`ValidateLabels` 或 `SanitizeLabels` 保护，禁止高基数和敏感 label。

## 追踪

`Tracer` 和 `Span` 是轻量抽象，`New` 与 `Close` 会创建生命周期 span。下游库可以在 adapter 层把该接口接到 OpenTelemetry，但模板本身只提供 `NoopTracer`，不引入 OTel 依赖。

## 测试

模板要求为配置校验、脱敏、字段、logger、metrics、tracer、context、label policy、客户端生命周期、健康检查和内部辅助代码提供单元测试与竞态测试。

## 发布

发布前必须通过 Harness Gate，并显式传入 `VERSION=vX.Y.Z` 生成 `release/manifest/<VERSION>.json`。该文件是 release Evidence artifact，不提交到源码历史；仓库只提交 `release/manifest/template.json`。`make release-check` 会先运行 CI 和 integration gate，再以 `CHECK_STATUS=passed` 生成 manifest；manifest 记录实际执行 gate 的 `commit`、`generated_by`、`go_version` 和 `tree_state`。integration gate 会渲染临时 `configx` 和 `corekit` 并运行测试，防止模板替换链路回归。
