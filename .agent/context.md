# Context

## 当前仓库

- 模块：`github.com/ZoneCNH/observex`
- 模板参考：`https://github.com/ZoneCNH/xlib-standard`
- 标准基线：`xlib-standard` `v0.4.19`（`4463a60`）
- 共享基础依赖：`https://github.com/ZoneCNH/foundationx`

## 定位

`observex` 是 L1 可观测性契约库，提供 logger、metrics、tracer、context 字段、redaction、label policy、错误、健康检查和 release Evidence。核心包不得依赖 `x.go`、业务模型或具体观测后端。

`xlib-standard` `v0.4.19` 新增的 L2 adapter contract pack、`templates/l2` 和 L2 readiness evidence 属于标准源边界；本仓库只继承 L1 基础库模板、渲染和 release gate 约束，不内置 L2 adapter 模板或 provider gate。

## 验证约束

在父级 workspace 存在时优先使用 `GOWORK=off`。完成声明必须包含实际运行的 gate 和结果；缺少 `golangci-lint` 或 `govulncheck` 时不得把完整 CI 声明为通过。
