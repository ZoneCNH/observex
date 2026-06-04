# Label Policy

metrics label 是长期存储和聚合维度，必须比日志字段更严格。`observex` 的 label policy 由 `contracts/metric_naming.md`、`ValidateMetricName`、`ValidateLabels` 和 `SanitizeLabels` 共同约束。

## 命名

指标名和 label key 必须满足：

```text
^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$
```

也就是说必须是 lower snake case，并以字母开头。

## 禁止字段

以下字段不得作为 metrics label：

- 链路字段：`trace_id`、`request_id`、`correlation_id`
- 业务实体字段：`user_id`、`order_id`
- 时间和原始错误：`timestamp`、`raw_error`
- 原始载荷：`sql`、`payload`
- 任何敏感 key 或疑似凭据值

这些字段需要排障时应进入日志或 trace event。

## 使用建议

库代码在写指标前应调用 `SanitizeLabels`。测试中应覆盖非法 key、高基数字段和疑似敏感值，确保新增指标不会破坏 contract。
