# 指标命名与类型规范

本文档定义 `observex` 及其下游模块的指标命名、label 使用和指标类型选择规范。

## 命名格式

所有指标名遵循统一前缀格式：

```text
foundationx_{module}_{operation}_{measure}
```

| 段 | 说明 | 示例 |
|---|---|---|
| `foundationx` | 固定前缀，标识 FoundationX 体系 | `foundationx` |
| `{module}` | 模块名，snake_case | `market_data`, `risk_engine`, `kernel` |
| `{operation}` | 操作名，snake_case | `fetch`, `evaluate`, `submit` |
| `{measure}` | 度量名，snake_case | `total`, `latency_ms`, `errors_total` |

### 命名规则

- 全部使用 `lower_snake_case`，首字符必须是字母
- 匹配正则：`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`
- 生命周期计数器必须使用 `_total` 后缀
- 耗时 histogram 必须在名称中体现单位（`_seconds` 或 `_ms`）
- 禁止使用点号、连字符、空格或大写字母

### 代码生成

使用 `MetricName(module, operation, measure)` helper 自动生成合规指标名：

```go
name := MetricName("market-data", "fetch", "latency_ms")
// → "foundationx_market_data_fetch_latency_ms"
```

## Label 规范

### 推荐 Label 名

| Label | 用途 | 典型值 |
|---|---|---|
| `error_kind` | 错误分类 | `timeout`, `validation`, `auth` |
| `source` | 数据来源或调用方 | `binance`, `risk_engine` |
| `status` | 操作结果 | `success`, `failure` |
| `operation` | 操作类型 | `fetch`, `evaluate` |
| `kind` | 子分类 | `counter`, `gauge` |
| `name` | 组件名称 | `observex`, `market_data` |

### 禁止 Label 名

以下字段不得作为 metrics label：

- `error`, `err` — 使用 `error_kind` 代替
- `msg`, `level` — 日志语义，非 metric 维度
- `trace_id`, `request_id`, `correlation_id` — 链路字段，高基数
- `user_id`, `order_id` — 业务实体，高基数
- `timestamp`, `raw_error` — 时间和原始错误
- `sql`, `payload` — 原始载荷
- 任何敏感 key 或疑似凭据值

### Label 校验

代码中应使用以下工具校验 label：

- `ValidateLabels(labels Labels) error` — 校验 label key 和 value
- `SanitizeLabels(labels Labels) Labels` — 自动移除或脱敏不安全 label
- `LabelPolicy.ValidateLabel(name string) error` — 单个 label 名策略检查
- `LabelPolicy.ValidateLabels(names []string) []error` — 批量 label 名策略检查

## 指标类型选择

| 类型 | 场景 | 示例 |
|---|---|---|
| **Counter** | 单调递增的累计值 | 请求总数、错误总数、重试次数 |
| **Gauge** | 可增可减的瞬时值 | 当前连接数、健康状态（0/1）、队列长度 |
| **Histogram** | 值的分布统计 | 请求延迟、响应大小、批处理耗时 |

### Counter 规范

- 名称以 `_total` 结尾
- 使用 `IncCounter` 或 `AddCounter`
- 不可减少，重启后从零开始

```go
metrics.IncCounter("foundationx_kernel_start_total", Labels{"status": "success"})
```

### Gauge 规范

- 名称不加 `_total` 后缀
- 使用 `SetGauge`
- 可以任意设置值

```go
metrics.SetGauge("foundationx_observex_inflight", float64(count), Labels{})
```

### Histogram 规范

- 名称必须包含单位：`_seconds`, `_ms`, `_bytes`
- 使用 `ObserveHistogram`
- 记录单次观测值，聚合由后端完成

```go
metrics.ObserveHistogram("foundationx_market_data_fetch_latency_ms", 42.5, Labels{"source": "binance"})
```

## CI 执行

指标命名合规性由 `scripts/check-label-policy.sh` 在 CI 中自动检查。该脚本扫描所有 `.go` 文件，验证 label key 是否符合 snake_case 规范且不在禁止列表中。
