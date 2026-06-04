# 规格

## 需求

创建独立 Go 可观测性基础库模板，包含 logger、metrics、tracer、context 字段、redaction、label policy、错误、健康检查、测试、文档、contracts、CI、Harness、manifest、Evidence、评审和复盘模板。

## 非目标

- 不依赖 `x.go`。
- 不包含 `x.go` 业务模型。
- 不直接依赖 Prometheus、OpenTelemetry、Zap 或 Logrus。
- 不隐式加载生产密钥。
