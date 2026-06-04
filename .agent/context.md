# Context

## 当前仓库

- 模块：`github.com/ZoneCNH/observex`
- 模板参考：`https://github.com/ZoneCNH/baselib-template`
- 共享基础依赖：`https://github.com/ZoneCNH/foundationx`

## 定位

`observex` 是 L1 可观测性契约库，提供 logger、metrics、tracer、context 字段、redaction、label policy、错误、健康检查和 release Evidence。核心包不得依赖 `x.go`、业务模型或具体观测后端。

## 验证约束

在父级 workspace 存在时优先使用 `GOWORK=off`。完成声明必须包含实际运行的 gate 和结果；缺少 `golangci-lint` 或 `govulncheck` 时不得把完整 CI 声明为通过。
