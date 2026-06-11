# foundationx 兼容边界

`observex` 通过 `internal/foundationx` 中的本地替换 module 依赖 `github.com/ZoneCNH/foundationx`。这让可观测库保持可构建，同时保留调用方使用的已测试 foundation 兼容面。

## 兼容范围

本地 module 只刻意镜像 `observex` 当前测试锁定的 foundation API：

- **错误类型系统**：`ErrorKind` (11 种：AlreadyExist/Auth/Canceled/Config/Conflict/Connection/Internal/NotFound/RateLimit/Timeout/Unavailable/Validation)、`IsKind`、`AsFoundationError`
- **密钥脱敏**：`NewSecretString`、`Sanitizer`

`observex` 公共 API 主要在错误分类和敏感数据脱敏路径上依赖 foundationx。由于 `go.mod` 使用 local replace，contract tests 也会锁定上述支撑 helpers，避免调用方在当前 module 边界内遇到 foundationx 漂移。未列出的 foundationx API 不属于兼容范围，除非先补充 contract tests。

## 不可变边界

- `docs/goal.md` 保持权威契约，模板应用工作不得重写它。
- 可观测输出保持 vendor-neutral：日志、指标、追踪均通过接口抽象，不得直接绑定 Prometheus/Otel/Zap 等具体后端。
- 模块不得包含生成的 `x.go` files，也不得导入 `x.go` 或 Redis、Kafka、PostgreSQL、TDengine、object-storage SDKs 等 infrastructure driver packages。
- 验证 evidence、examples、release manifests 与 documentation 只能使用脱敏后的 secret 输出。
- label policy 始终保持低基数：禁止将 request ID、user ID 等高基数字段作为 metric label。

## 升级规则

将本地 module 替换为 upstream foundationx release 前，必须先用 `GOWORK=off go test ./...` 以及 boundary、contract、secret scanners 证明 `ErrorKind` 分类、`IsKind` 语义、`AsFoundationError` 转换、`Sanitizer` 脱敏和 `NewSecretString` 行为保持兼容。

迁移截止：`observex` v0.4 前完成，迁移完成后删除 `internal/foundationx/`。
