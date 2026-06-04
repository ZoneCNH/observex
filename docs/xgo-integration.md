# x.go 集成边界

`observex` 是独立基础库，不是 `x.go` 的业务观测模块。它可以被 `x.go`、业务服务或其他基础库使用，但核心包不得反向依赖 `x.go`。

## 允许

- 暴露后端无关的 `Logger`、`Metrics`、`Tracer`、`Field`、`Labels` 和 `Redactor`。
- 复用 `foundationx` 的错误、健康状态和脱敏原语。
- 在上层 adapter 中桥接 Prometheus、OpenTelemetry、Zap、Logrus 或业务规范。
- 在 `x.go` 中组合 observex contract，并注入具体实现。

## 禁止

- 从核心包 import `github.com/ZoneCNH/x.go`。
- 在核心包中定义业务指标、业务 span 名称或业务日志字段集合。
- 在核心包中直接依赖 Prometheus、OpenTelemetry、Zap、Logrus、数据库、消息队列或对象存储 SDK。
- 隐式读取生产配置或凭据。

## 验证

边界由 `make boundary` 和 `scripts/check_boundary.sh` 验证。新增依赖、包结构或生成脚本变更后必须重新运行该 gate。
