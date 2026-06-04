# Tracing 契约

`observex` 只定义轻量 `Tracer` 和 `Span` 接口，不直接引入 OpenTelemetry。这样基础库可以在无 tracing 后端时保持可用，也能由上层 adapter 按需桥接到实际系统。

## 接口

- `Tracer.Start(ctx, name, fields...)`：开始 span，返回派生 context 和 `Span`。
- `Span.SetField(field)`：设置或覆盖 span 字段。
- `Span.AddEvent(name, fields...)`：记录事件。
- `Span.End(fields...)`：结束 span 并附加收尾字段。
- `NoopTracer` / `NoopSpan`：默认空实现。

## 生命周期 span

核心包固定记录以下生命周期 span：

- `observex.New`
- `observex.Close`

生成后的具体基础库可以继续沿用该命名风格，例如 `<package>.Request`、`<package>.Retry` 或 `<package>.HealthCheck`。业务服务 span 名称不属于本模板，应由上层业务仓库定义。

## 字段规则

trace 字段使用 `Field`。敏感字段必须用 `Secret` 或实现 `foundationx.Sanitizer`，由 adapter 在写出前复用 `Redactor` 规则。高基数字段可以进入 trace event，但不能复制到 metrics label。
