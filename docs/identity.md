# observex 身份

## 我是谁

`observex` 是 FoundationX 的 **L1 Vendor-Neutral 可观测性契约库**。定义日志/指标/追踪/健康检查的抽象接口，不绑定具体后端。

> ⚠️ **身份声明**：observex 是 concrete L1 runtime library，不是模板源。模板生成属于 xlib-standard。

## 我做什么

| 能力 | 职责 |
|------|------|
| logger | 日志接口定义 |
| metrics | 指标接口 + P0 指标命名规范 |
| tracer | 追踪接口定义 |
| redactor | 敏感数据脱敏 |
| label policy | 低基数 label 规则 |
| health schema | 健康检查 JSON schema |
| noop | 零依赖默认实现 |
| memory recorder | 测试用内存记录器 |

## 我不做什么

| 不是 | 原因 |
|------|------|
| **不是 Prometheus/Zap/OTel SDK** | observex 只定义接口，不绑定后端实现 |
| **不是告警路由** | 告警属于 alertx |
| **不是业务监控规则** | 监控规则由各域定义 |
| **不是模板源** | 模板生成属于 xlib-standard |

## 宪法合规

| 条款 | 遵循方式 |
|------|----------|
| §3.3 | L1 运行时，仅依赖 kernel |
| §6.1 | foundationx_<module>_<operation>_<measure> 命名规范 |
| §6.3 | 低基数 label policy（禁止高基数字段） |
| §6.4 | Redactor 脱敏处理 |
