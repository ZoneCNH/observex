# 设计

## 结构

- `pkg/observex` 包含可编译的公共 API 模板。
- `pkg/observex` 中的观测能力通过接口注入，默认提供 Noop 实现。
- `internal/` 包含可复用内部辅助示例。
- `scripts/` 包含 Harness Gate。
- `.github/workflows/` 包含 CI、集成、安全和发布检查。
- `release/manifest/` 包含发布 Evidence 模板和生成输出。

## 边界

核心包只依赖 `foundationx` 的错误、健康状态和脱敏原语，不绑定具体观测后端。
