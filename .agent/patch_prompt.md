# Patch Prompt

执行补丁时遵循以下约束：

- 优先复用现有 `pkg/observex` API 和 `foundationx` 原语。
- 不新增 Prometheus、OpenTelemetry、Zap、Logrus、数据库、消息队列或对象存储 SDK 依赖。
- 新增文档默认使用中文；代码标识符、命令、路径和协议固定短语保留原文。
- 新增日志、trace 或 Evidence 字段必须经过 `Field` 与 `Redactor` 规则。
- 新增 metrics 必须更新 `contracts/metrics.md`、`contracts/metrics.schema.json` 和相关测试。
- 修改发布流程后必须同步 `release/manifest/template.json`、`internal/tools/releasemanifest` 和 workflow artifact 路径。
