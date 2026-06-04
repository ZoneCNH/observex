# Metric Naming Contract

`observex` 的指标名必须稳定、可读、低基数，并保持 Prometheus 风格兼容。

## 指标名

指标名必须匹配 `^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`：

```text
^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$
```

指标名使用小写蛇形命名，不能包含点号、连字符、空格或大写字母。生命周期 counter 使用 `_total` 后缀，耗时 histogram 必须在名称中体现单位，例如 `_seconds` 或 `_ms`。

## 标签

标签键使用同一命名规则，并由 `ValidateMetricName` 与 `ValidateLabels` 执行校验。标签值必须是低基数字段，例如 `operation`、`status`、`kind`、`name`。禁止把 request id、trace id、user id、session id、email、path、url、host、ip、credential 类字段放进 labels。

## 脱敏

所有 metrics 适配器在记录 labels 前必须调用 `SanitizeLabels`。当 label 键命中 `Redactor` 的敏感键策略时，值必须替换为 `***`。
