# Rule

## 硬约束

- 核心包不得 import `github.com/ZoneCNH/x.go`。
- 核心包不得直接依赖 Prometheus、OpenTelemetry、Zap、Logrus 或业务 SDK。
- 不得隐式读取生产凭据。
- `release/manifest/v<version>.json` 是生成产物，不提交到源码历史。
- `make lint` 缺少 `golangci-lint` 时必须失败。
- `make security` 缺少 `govulncheck` 时必须失败。

## 完成口径

最终报告必须列出已运行验证命令、结果和未验证项。只有 release gate 真实通过后，才能把对应 manifest check status 标记为 `passed`。
