# ADR-20260601-003 Metrics Label Policy

## 状态

Accepted

## 背景

metrics label 会影响时序数据基数和存储成本。把 trace id、请求 id、用户 id、SQL 或原始载荷放入 label 会造成高基数和潜在敏感信息泄露。

## 决策

核心包提供 `Labels`、`ValidateMetricName`、`ValidateLabels` 和 `SanitizeLabels`。指标名和 label key 必须是 lower snake case，高基数、敏感和原始载荷字段一律禁止作为 label。

## 后果

派生库新增指标时需要通过 contract 和单元测试验证 label。排障字段进入日志或 trace event，而不是 metrics label。
