# Context 字段

`observex` 使用 typed private context key 传递观测字段，避免与调用方或其他库的 context value 冲突。context 字段是请求链路上的辅助信息，不是配置来源，也不是隐式全局状态。

## 标准字段

- `WithTraceID` / `TraceID`
- `WithRequestID` / `RequestID`
- `WithCorrelationID` / `CorrelationID`
- `WithContextField`
- `WithContextFields`
- `ContextFields`
- `FieldsFromContext`

`FieldsFromContext` 会返回显式字段，并追加存在的 `trace_id`、`request_id` 和 `correlation_id`。

## 使用边界

context 字段可以进入日志和 trace event。它们不应进入 metrics label，因为这些字段通常具有高基数，会显著增加时序系统负担。

`nil` context 会被核心包归一化为 `context.Background()`，但公开 API 仍应优先要求调用方显式传入 context。
