# ADR-20260604-001：observex 拥有 L1 observability contract

## 状态

已接受。

## 背景

`observex` 必须作为 base libraries 和 service bootstrap code 的 vendor-neutral L1 contract。它不能回落为通用 Go template，也不能引入 concrete backend、business domain model、hidden global、listener、goroutine 或 secret read。

## 决策

- Public contracts 覆盖 logger、metrics、tracer/span、health/readiness、field/attr、redaction、label policy、examples、确定性且 race-safe 的 Memory implementations，以及 release evidence。
- Core code 可以依赖 Go stdlib 和稳定的 L0 primitives，但不得 import concrete observability backend 或 provider。
- Adapter package 和 downstream services 拥有 integrations。Business metrics、business spans、provider details、hidden goroutines、listeners 与 secret reads 均不属于本库范围。
- Completion evidence 必须包含 tests、provider dependency scan、examples smoke、CI/release gates、manifest、sha256、contract hashes 和 known gaps。

## 后果

- README、docs 与 examples 必须把项目描述为 L1 observability contract library，而不是 template。
- 未来任何 API rename 或 compatibility bridge，包括 `HealthReporter` / `ReadinessReporter`、`HealthCheck` / `ReadinessCheck`、`Attr` / `Field`，都必须先更新 contracts 与 examples，再声明 release-ready。
- Heavy vendor integrations 可以作为显式 adapter 存在，但不得进入 root contract package。
