# 日志契约

`observex` 只定义日志接口和标准库 `slog` 适配器，不直接绑定 Zap、Logrus 或业务日志框架。生成的基础库应把日志组件作为显式依赖注入，默认使用 `NoopLogger`，避免调用方未配置日志时产生 panic。

## 公共接口

- `Logger`：提供 `Debug`、`Info`、`Warn` 和 `Error` 四个级别。
- `NoopLogger`：默认空实现，用于模板和测试中的安全默认值。
- `SlogLogger`：基于标准库 `log/slog` 的最小适配器。
- `WithRedactor`：为 `SlogLogger` 覆盖默认脱敏器。

日志字段统一使用 `Field`。调用方可以用 `String`、`Int`、`Bool`、`Duration`、`Time`、`Any`、`Secret` 和 `ErrorField` 创建字段。`Secret` 字段、敏感 key，以及实现 `foundationx.Sanitizer` 的值，写入日志前必须被脱敏。

## Context 字段

`SlogLogger` 会自动合并 `FieldsFromContext(ctx)` 和调用点传入的字段。推荐在入口层写入：

- `trace_id`
- `request_id`
- `correlation_id`

这些字段适合进入日志和 trace event，但不适合作为 metrics label。

## 后端适配

如果项目需要 Zap 或 Logrus，应在上层 adapter 仓库实现 `Logger` 接口。核心包不得为了某个日志后端引入额外依赖，也不得创建隐藏全局 logger。
