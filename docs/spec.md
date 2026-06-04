# SPEC-observex-v1.0

## 需求

- 为可复用基础库提供独立 Go module。
- 提供 `Config`、`Validate`、`Sanitize`、`Client`、`New`、`Option`、`HealthCheck`、错误模型、logger、metrics、tracer、redactor、label policy、context 字段和版本元数据。
- `Validate`、`New`、`Close` 和 `HealthCheck` 必须返回或记录可分类的生产语义，包括 typed error、幂等关闭、上下文取消、日志、trace span、metrics 和健康状态。
- 可观测性接口必须是轻量 contract：默认 no-op，可由上层适配到 Prometheus、OpenTelemetry、Zap、Logrus 或自研系统，但本模块不直接依赖这些实现。
- 提供 Harness Gate 脚本、生成脚本、CI 工作流、contracts、examples、Evidence artifact、release 和复盘模板。

## 验收标准

- `GOWORK=off go test ./...` 和 `GOWORK=off go test -race ./...` 通过。
- `GOWORK=off make release-check` 通过，并以 `CHECK_STATUS=passed` 生成未提交的 `release/manifest/v0.1.0.json` Evidence artifact。
- `contracts/config.schema.json` 与 `Config` 字段映射保持一致，`timeout_ms` 映射到 `Config.Timeout`。
- `contracts/error.schema.json`、`contracts/health.schema.json`、`contracts/field.schema.json`、`contracts/logger.schema.json`、`contracts/tracer.schema.json`、`contracts/metrics.schema.json`、`contracts/metric_naming.md` 和 `contracts/metrics.md` 与公共常量保持一致。
- `ValidateMetricName`、`ValidateLabels` 和 `SanitizeLabels` 必须拒绝高基数、保留字段和疑似敏感 label。
- logger、metrics、tracer、redaction 和 context 示例必须有 smoke 测试。
- `scripts/render_template.sh` 可以生成 `configx` 形态并通过 `GOWORK=off go test ./...`。
- 模块不得依赖 `github.com/ZoneCNH/x.go`。
- 模块不得隐式读取生产密钥。

## 非目标

- 不包含业务模型、生产连接默认值和隐藏全局客户端。

## 可追踪性

- 目标：`GOAL-20260601-001`
- 模板占位符：`observex`、`github.com/ZoneCNH/observex`、`observex`
- 模板参考：`https://github.com/ZoneCNH/baselib-template`
- 共享基础依赖：`https://github.com/ZoneCNH/foundationx`
