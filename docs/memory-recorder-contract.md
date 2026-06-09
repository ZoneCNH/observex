# MemoryRecorder 行为契约

本文档定义 `observex` 测试用内存记录器（`MemoryLogger`、`MemoryMetrics`、`MemoryTracer`）的行为规范。这些记录器是 `Logger`、`Metrics`、`Tracer` 接口的内存实现，专为测试和示例设计。

## MemoryLogger

### 接口实现

`MemoryLogger` 实现 `Logger` 接口：

```go
type Logger interface {
    Debug(ctx context.Context, msg string, fields ...Field)
    Info(ctx context.Context, msg string, fields ...Field)
    Warn(ctx context.Context, msg string, fields ...Field)
    Error(ctx context.Context, msg string, fields ...Field)
}
```

### 行为规范

| 属性 | 规范 |
|---|---|
| 记录顺序 | `Sequence` 从 1 开始，单调递增 |
| 字段脱敏 | 所有字段经过 `Redactor` 处理后存储 |
| 上下文字段 | `FieldsFromContext(ctx)` 自动合并到日志记录 |
| 线程安全 | 所有方法可从多个 goroutine 并发调用 |
| nil 安全 | nil 接收器的所有方法为 no-op |
| 重置 | `Reset()` 清除所有记录并重置 Sequence |

### 查询方法

- `Records() []LogRecord` — 返回当前所有记录的快照（深拷贝）
- 每条 `LogRecord` 包含 `Sequence`、`Level`、`Message`、`Fields`

### Level 常量

| Level | 说明 |
|---|---|
| `LogLevelDebug` | 调试级别 |
| `LogLevelInfo` | 信息级别 |
| `LogLevelWarn` | 警告级别 |
| `LogLevelError` | 错误级别 |

### 使用示例

```go
logger := observex.NewMemoryLogger()
logger.Info(ctx, "server started", observex.String("port", "8080"))

records := logger.Records()
assert len(records) == 1
assert records[0].Level == observex.LogLevelInfo
assert records[0].Message == "server started"
```

## MemoryMetrics

### 接口实现

`MemoryMetrics` 实现 `Metrics` 接口：

```go
type Metrics interface {
    IncCounter(name string, labels Labels)
    AddCounter(name string, delta float64, labels Labels)
    ObserveHistogram(name string, value float64, labels Labels)
    SetGauge(name string, value float64, labels Labels)
}
```

### 行为规范

| 属性 | 规范 |
|---|---|
| 记录顺序 | `Sequence` 从 1 开始，单调递增 |
| 指标名校验 | 名称经过 `sanitizeMetricName` 处理 |
| Label 校验 | Labels 经过 `SanitizeLabels` 处理 |
| Counter 累加 | 同名同 label 的 counter 值累加 |
| Gauge 覆盖 | 同名同 label 的 gauge 值直接覆盖 |
| Histogram 不聚合 | 记录原始观测值，不计算分位数 |
| 线程安全 | 所有方法可从多个 goroutine 并发调用 |
| nil 安全 | nil 接收器的所有方法为 no-op |
| 重置 | `Reset()` 清除所有记录和聚合状态 |

### 查询方法

- `Records() []MetricRecord` — 返回所有记录的快照
- `Counters() map[string]float64` — 返回 counter 聚合值
- `Gauges() map[string]float64` — 返回当前 gauge 值
- 每条 `MetricRecord` 包含 `Sequence`、`Kind`、`Name`、`Value`、`Labels`

### MetricKind 常量

| Kind | 对应操作 |
|---|---|
| `MetricKindCounter` | `IncCounter` / `AddCounter` |
| `MetricKindHistogram` | `ObserveHistogram` |
| `MetricKindGauge` | `SetGauge` |

### 使用示例

```go
metrics := observex.NewMemoryMetrics()
metrics.IncCounter("requests_total", observex.Labels{"status": "ok"})
metrics.SetGauge("inflight", 5, nil)

counters := metrics.Counters()
assert counters["requests_total|status=ok"] == 1

gauges := metrics.Gauges()
assert gauges["inflight"] == 5
```

## MemoryTracer

### 接口实现

`MemoryTracer` 实现 `Tracer` 接口：

```go
type Tracer interface {
    Start(ctx context.Context, name string, fields ...Field) (context.Context, Span)
}
```

`memorySpan` 实现 `Span` 接口：

```go
type Span interface {
    SetField(field Field)
    AddEvent(name string, fields ...Field)
    End(fields ...Field)
}
```

### 行为规范

| 属性 | 规范 |
|---|---|
| 记录顺序 | `Sequence` 从 1 开始，单调递增 |
| 字段脱敏 | 所有字段经过 `Redactor` 处理 |
| Span 生命周期 | `Start` 创建未结束 span，`End` 标记完成 |
| Event 记录 | `AddEvent` 追加事件到 span 的 Events 列表 |
| 重复 End | 多次调用 `End` 只记录第一次 |
| 线程安全 | 所有方法可从多个 goroutine 并发调用 |
| nil 安全 | nil 接收器的所有方法为 no-op |
| 重置 | `Reset()` 清除所有 span |

### 查询方法

- `Spans() []SpanRecord` — 返回所有 span 的快照（深拷贝）
- 每条 `SpanRecord` 包含 `Sequence`、`Name`、`Fields`、`Events`、`Ended`、`EndFields`
- `SpanEvent` 包含 `Name` 和 `Fields`

### 使用示例

```go
tracer := observex.NewMemoryTracer()
ctx, span := tracer.Start(ctx, "fetch-data", observex.String("source", "binance"))
span.SetField(observex.Int("count", 100))
span.AddEvent("cache-miss")
span.End(observex.String("status", "ok"))

spans := tracer.Spans()
assert len(spans) == 1
assert spans[0].Name == "fetch-data"
assert spans[0].Ended == true
assert len(spans[0].Events) == 1
```

## 构造选项

`MemoryLogger` 和 `MemoryTracer` 支持通过 `WithRedactor` 自定义脱敏器：

```go
redactor := observex.NewDefaultRedactor("custom_secret")
logger := observex.NewMemoryLogger(observex.WithRedactor(redactor))
tracer := observex.NewMemoryTracer(observex.WithRedactor(redactor))
```

不指定时使用 `NewDefaultRedactor()` 作为默认脱敏器。
