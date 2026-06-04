# ADR-20260601-002 字段脱敏是公共契约

## 状态

Accepted

## 背景

日志、trace 和 release Evidence 都可能携带配置和运行时字段。仅依赖调用方自觉脱敏容易出现遗漏。

## 决策

公共字段统一使用 `Field` 表达，并提供 `Secret`、`Redactor` 和 `DefaultRedactor`。默认脱敏器复用 `foundationx.Sanitizer` 和 `foundationx.SecretString`，输出固定掩码。

## 后果

新增日志、trace 或测试夹具时可以复用同一套脱敏规则。调用方如果需要保留原始值，必须在自己的安全边界内处理。
