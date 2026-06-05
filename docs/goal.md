# observex 完整可执行 Goal Prompt v1.0

> 文件名：`observex_goal_executable_prompt_v1_0.md`  
> 目标模块：`github.com/ZoneCNH/observex`  
> 模块定位：日志 / 指标 / Trace / 上下文观测字段的独立公共基础库  
> 分层定位：L1 运行时基础能力层  
> 上游依赖：`github.com/ZoneCNH/foundationx`  
> 模板参考：`https://github.com/ZoneCNH/xlib-standard`
> 标准基线：`xlib-standard` `v0.4.19`（`4463a60`）
> 下游调用方：x.go、postgresx、kafkax、redisx、taosx、configx、ossx、业务服务启动层  
> 执行方法：Goal Runtime Prompt v3.1 + Harness + AutoResearch + Self-improving + Evidence Protocol  
> 生成日期：2026-06-01  

---

# 0. 使用方式

将本文完整交给 Agent Teams / Codex / Claude Code / Cursor Agent / GitHub Copilot Workspace 执行。

执行前必须确认：

```text
1. 当前目标是创建或完善独立 Go module：github.com/ZoneCNH/observex
2. observex 是 L1 可观测性契约库，不是 x.go 业务监控模块
3. observex 可以依赖 foundationx
4. observex 不允许依赖 x.go
5. observex 核心包不允许强依赖 Zap / Logrus / Prometheus / OpenTelemetry 外部实现
6. observex 不允许依赖 PostgreSQL / Kafka / Redis / TDengine / OSS driver
7. observex 不允许包含 x.go 业务指标、业务日志字段、业务 trace span 名称
8. observex 默认必须提供 Noop 实现，避免基础库没有注入观测组件时 panic
9. observex 必须内置字段脱敏和 label 安全约束，防止 secret 与高基数标签污染
10. 所有完成声明必须使用 DONE with evidence:
11. observex 只对齐 `xlib-standard` 的 L1 基础库模板、渲染和 release gate 约束；`v0.4.19` 新增的 L2 adapter contract pack、`templates/l2` 和 L2 readiness evidence 不并入本仓库
```

---

# 1. Master Goal

```text
GOAL-20260601-OBSERVEX-001

建立 observex 独立公共可观测性基础库，为 x.go 与基础库体系提供统一、轻量、稳定、可测试、可脱敏、可发布的 Logger / Metrics / Tracer / Context Fields 契约。

observex 必须提供日志接口、指标接口、Trace 接口、Noop 实现、标准库 slog 适配、字段构造器、字段脱敏、上下文 CorrelationID / TraceID / RequestID 辅助、指标命名规范、label 安全约束、测试工具、Examples、CI/Harness/Evidence/Release 流程。

observex 必须不依赖 x.go，不理解业务语义，不强绑定外部可观测性实现，不在核心包中引入 Prometheus / OpenTelemetry / Zap 等重依赖。Prometheus / OpenTelemetry adapter 可作为 v0.2 或独立 observex-prometheus / observex-otel 模块处理。
```

---

# 2. 问题底层本质

observex 不是“封装一个 logger”。

observex 的底层本质是：

```text
把日志、指标、Trace 从各基础库的隐式实现依赖，变成统一、可注入、可替换、可测试、可脱敏的运行时观测契约。
```

它解决的是：

```text
1. postgresx / kafkax / redisx / taosx 各自依赖不同 logger，导致基础库变重
2. 指标命名、label、错误统计口径不统一
3. Trace span 与 context 传播不一致
4. 基础库默认无观测时容易 nil panic
5. 日志字段可能泄露 password/token/secret/dsn
6. metrics label 可能出现 symbol/order_id/user_id 等高基数字段，拖垮 Prometheus
7. x.go 的 observability 逻辑和基础库契约混在一起
8. Agent Teams 无法通过 Evidence 判断观测接口是否可用、是否安全
```

observex 的核心价值：

```text
统一观测契约 + 默认安全 Noop + 脱敏字段 + 可替换 adapter + 可治理 metrics label。
```

---

# 3. 不可再拆解的基本真理

## 3.1 observex 是 L1，不是观测平台

observex 可以知道：

```text
Logger
Field
Metrics
Counter
Gauge
Histogram
Tracer
Span
Context
CorrelationID
Redactor
Noop
Slog adapter
```

observex 不应该知道：

```text
BTCUSDT
Kline
MacroRegime
M1-M7
S1-S7
TradingSignal
Order
Position
Kafka business topic
Redis business key
TDengine business table
```

## 3.2 observex 核心包必须轻量

核心包优先只依赖：

```text
Go 标准库
foundationx
```

标准库允许：

```text
context
time
errors
fmt
log/slog
sync
strings
regexp
```

核心包禁止强依赖：

```text
go.uber.org/zap
github.com/sirupsen/logrus
github.com/prometheus/client_golang
go.opentelemetry.io/otel
```

这些可以后续做 adapter 包或独立模块。

## 3.3 Noop 是一级能力

所有接口必须有 Noop 实现：

```text
NoopLogger
NoopMetrics
NoopTracer
NoopSpan
```

原因：

```text
基础库可在未接入观测系统时安全运行。
测试环境不会产生噪声。
```

## 3.4 Field 和 Label 必须安全

禁止将以下内容直接输出：

```text
password
passwd
secret
token
access_key
secret_key
private_key
dsn 明文
authorization header
cookie
```

metrics label 禁止高基数：

```text
user_id
order_id
trace_id
request_id
timestamp
raw_error
sql
payload
```

## 3.5 没有 Evidence 不得声称完成

完成声明必须是：

```text
DONE with evidence:
- go test ./...
- go test -race ./...
- make ci
- boundary gate passed
- secret gate passed
- examples passed
- release manifest generated
```

---

# 4. 被误认为真理的常见假设

| 常见假设 | 为什么错 | 正确口径 |
|---|---|---|
| observex 就是 logger 包 | 过窄 | 必须统一 Logger / Metrics / Tracer |
| 基础库直接依赖 zap 最省事 | 让基础库变重并绑定实现 | observex 定义接口，adapter 另做 |
| metrics label 越多越好 | 高基数会拖垮系统 | label 必须有治理规则 |
| Trace 必须立刻接 OpenTelemetry | v0.1 会引入重依赖 | 核心定义 Tracer 接口，OTel adapter 后续做 |
| logger nil 时可以不管 | 基础库会 panic | 必须 Noop 默认 |
| 日志脱敏交给业务 | 太晚，基础库已经可能泄露 | Field 层内置 Redactor |
| observex 可以定义 x.go 指标名 | 业务污染 | observex 只定义命名规范和接口 |
| request_id/trace_id 可作为 metric label | 高基数错误 | 只可用于 logs/traces，不用于 metrics label |


# 5. Scope

## 5.1 In Scope

```text
Logger interface
Field model
Field helper constructors
NoopLogger
SlogLogger adapter based on standard log/slog
Redactor
Secret field detection
Metrics interface
NoopMetrics
Metric label validation
Metric name validation
Tracer interface
Span interface
NoopTracer
Context field helpers
CorrelationID / TraceID / RequestID helpers
Error field helper
Duration field helper
TestKit
Examples
Harness scripts
Release manifest
Docs / ADR
```

## 5.2 Optional in v0.1

```text
slog adapter
in-memory metrics recorder for tests
in-memory logger recorder for tests
```

说明：

```text
log/slog 属于 Go 标准库，可以作为 v0.1 adapter。
```

## 5.3 Deferred / Out of Scope

```text
Zap adapter
Logrus adapter
Prometheus adapter
OpenTelemetry adapter
Grafana dashboard
Metrics HTTP endpoint
Trace exporter
Log shipping
Business metric registry
x.go metric names
x.go dashboard
Alert rules
```

推荐后续模块：

```text
github.com/ZoneCNH/observex-prometheus
github.com/ZoneCNH/observex-otel
github.com/ZoneCNH/observex-zap
```

---

# 6. 目标仓库与模块

```text
github.com/ZoneCNH/observex
```

go.mod：

```go
module github.com/ZoneCNH/observex

go 1.23
```

必须依赖：

```text
github.com/ZoneCNH/foundationx
```

v0.1 优先标准库：

```text
context
time
fmt
errors
strings
regexp
sync
log/slog
```

可选依赖必须通过 ADR：

```text
Prometheus client
OpenTelemetry SDK
Zap
Logrus
```

默认裁决：

```text
v0.1 核心包不引入 Prometheus / OpenTelemetry / Zap / Logrus。
```

---

# 7. 标准目录结构

```text
observex/
├── go.mod
├── go.sum
├── README.md
├── CHANGELOG.md
├── LICENSE
├── Makefile
├── .gitignore
├── .golangci.yml
├── pkg/
│   └── observex/
│       ├── doc.go
│       ├── field.go
│       ├── redactor.go
│       ├── logger.go
│       ├── logger_noop.go
│       ├── logger_slog.go
│       ├── metrics.go
│       ├── metrics_noop.go
│       ├── metric_name.go
│       ├── labels.go
│       ├── tracer.go
│       ├── tracer_noop.go
│       ├── context.go
│       ├── errors.go
│       ├── version.go
│       └── *_test.go
├── internal/
│   ├── sanitize/
│   ├── validation/
│   └── testutil/
├── testkit/
│   ├── logger.go
│   ├── metrics.go
│   ├── tracer.go
│   └── assert.go
├── examples/
│   ├── logger/
│   ├── metrics/
│   ├── tracer/
│   ├── slog/
│   └── redaction/
├── contracts/
│   ├── logger.schema.json
│   ├── metrics.schema.json
│   ├── tracer.schema.json
│   ├── field.schema.json
│   ├── public_api.md
│   └── metric_naming.md
├── docs/
│   ├── spec.md
│   ├── design.md
│   ├── api.md
│   ├── logger.md
│   ├── metrics.md
│   ├── tracer.md
│   ├── context.md
│   ├── redaction.md
│   ├── label-policy.md
│   ├── xgo-integration.md
│   ├── testing.md
│   ├── release.md
│   └── adr/
│       ├── ADR-20260601-001-core-no-heavy-deps.md
│       ├── ADR-20260601-002-noop-defaults.md
│       ├── ADR-20260601-003-label-cardinality-policy.md
│       └── ADR-20260601-004-adapters-deferred.md
├── scripts/
│   ├── check_boundary.sh
│   ├── check_secrets.sh
│   ├── check_contracts.sh
│   └── generate_manifest.sh
├── release/
│   └── manifest/
│       └── v0.1.0.json
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── security.yml
│       └── release.yml
└── .agent/
    ├── goal.md
    ├── spec.md
    ├── design.md
    ├── plan.md
    ├── tasks.md
    ├── harness.md
    ├── gates.md
    ├── evidence.md
    ├── review.md
    ├── release.md
    └── retrospective.md
```


# 8. Public API 设计

## 8.1 Field

文件：

```text
pkg/observex/field.go
```

目标 API：

```go
package observex

import "time"

type Field struct {
	Key    string
	Value  any
	Secret bool
}

func String(key, value string) Field
func Int(key string, value int) Field
func Int64(key string, value int64) Field
func Float64(key string, value float64) Field
func Bool(key string, value bool) Field
func Duration(key string, value time.Duration) Field
func Time(key string, value time.Time) Field
func Any(key string, value any) Field
func Error(err error) Field
func Secret(key string, value any) Field
```

要求：

```text
1. Field 是 logs/traces 的通用结构
2. Field.Secret=true 时必须脱敏
3. Error(err) 的 key 固定为 "error"
4. Duration 输出使用 time.Duration 或 string，策略文档化
```

## 8.2 Redactor

文件：

```text
pkg/observex/redactor.go
```

目标 API：

```go
type Redactor interface {
	RedactField(field Field) Field
	RedactFields(fields []Field) []Field
}

type DefaultRedactor struct{}

func NewDefaultRedactor() DefaultRedactor
func IsSecretKey(key string) bool
```

Secret key 自动识别：

```text
password
passwd
secret
token
access_token
refresh_token
api_key
access_key
secret_key
private_key
authorization
cookie
dsn
database_url
```

要求：

```text
1. secret field 输出 "***"
2. key 匹配不区分大小写
3. 不匹配单独 "key"，避免误伤
4. DSN 必须脱敏
5. tests 覆盖 fmt.Sprint 不泄露
```

## 8.3 Logger

文件：

```text
pkg/observex/logger.go
```

目标 API：

```go
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
}

func NewNoopLogger() Logger
```

设计要求：

```text
1. Logger 接口不返回 error
2. Logger 不 panic
3. Logger 必须接受 context
4. Logger 实现应自动 redaction
5. NoopLogger 是默认安全实现
```

## 8.4 Slog Adapter

文件：

```text
pkg/observex/logger_slog.go
```

目标 API：

```go
func NewSlogLogger(logger *slog.Logger, opts ...LoggerOption) Logger

type LoggerOption func(*loggerOptions)

func WithRedactor(redactor Redactor) LoggerOption
```

要求：

```text
1. 使用标准库 log/slog
2. 不引入 zap/logrus
3. nil slog.Logger 时 fallback 到 noop 或 slog.Default，策略文档化
4. fields 转 slog.Attr
5. Secret fields 必须脱敏
```

## 8.5 Metrics

文件：

```text
pkg/observex/metrics.go
```

目标 API：

```go
type Labels map[string]string

type Metrics interface {
	IncCounter(name string, labels Labels)
	AddCounter(name string, value float64, labels Labels)
	SetGauge(name string, value float64, labels Labels)
	ObserveHistogram(name string, value float64, labels Labels)
}

func NewNoopMetrics() Metrics
```

要求：

```text
1. Metrics 接口不依赖 Prometheus
2. 默认 Noop
3. Metric name 必须可校验
4. Labels 必须可校验
5. 不允许 label value 包含 secret
6. 不允许高基数 label
```

## 8.6 Metric Name / Label Policy

文件：

```text
pkg/observex/metric_name.go
pkg/observex/labels.go
```

目标 API：

```go
func ValidateMetricName(name string) error
func ValidateLabels(labels Labels) error
func SanitizeLabels(labels Labels) Labels
```

Metric name 规则：

```text
^[a-z_:][a-z0-9_:]*$
```

Label key 规则：

```text
^[a-zA-Z_][a-zA-Z0-9_]*$
```

禁止 label keys：

```text
password
secret
token
access_token
refresh_token
authorization
cookie
dsn
sql
payload
raw_error
trace_id
request_id
user_id
order_id
timestamp
```

说明：

```text
trace_id / request_id 可用于 logs/traces，但默认不允许进入 metrics labels。
```

## 8.7 Tracer / Span

文件：

```text
pkg/observex/tracer.go
```

目标 API：

```go
type Tracer interface {
	Start(ctx context.Context, name string, fields ...Field) (context.Context, Span)
}

type Span interface {
	End()
	RecordError(err error)
	SetField(field Field)
	SetFields(fields ...Field)
}

func NewNoopTracer() Tracer
```

要求：

```text
1. Tracer 不依赖 OpenTelemetry
2. 默认 NoopTracer
3. Span.End 幂等
4. SetField 自动脱敏
5. span name 不包含业务高基数 ID
```

## 8.8 Context Helpers

文件：

```text
pkg/observex/context.go
```

目标 API：

```go
func WithCorrelationID(ctx context.Context, id string) context.Context
func CorrelationIDFromContext(ctx context.Context) (string, bool)

func WithTraceID(ctx context.Context, id string) context.Context
func TraceIDFromContext(ctx context.Context) (string, bool)

func WithRequestID(ctx context.Context, id string) context.Context
func RequestIDFromContext(ctx context.Context) (string, bool)

func FieldsFromContext(ctx context.Context) []Field
```

要求：

```text
1. context key 使用私有类型
2. 不污染 context with map[string]any
3. FieldsFromContext 可用于 Logger 自动附加字段
4. 这些字段不得用于 metrics labels
```

## 8.9 Error Mapping

文件：

```text
pkg/observex/errors.go
```

目标 API：

```go
func MapError(op string, err error) error
```

映射原则：

```text
invalid metric name -> ErrorKindValidation
invalid label -> ErrorKindValidation
secret label -> ErrorKindValidation
context canceled -> ErrorKindCanceled
context deadline exceeded -> ErrorKindTimeout
```


# 9. Spec

```text
SPEC-observex-v1.0
```

## REQ-OBSERVEX-001：独立 Go module

Acceptance Criteria：

```text
AC-REQ-OBSERVEX-001-001: go.mod module 为 github.com/ZoneCNH/observex
AC-REQ-OBSERVEX-001-002: go test ./... 通过
AC-REQ-OBSERVEX-001-003: go list -deps ./... 不包含 github.com/ZoneCNH/x.go
AC-REQ-OBSERVEX-001-004: README 明确模块定位和非目标
```

## REQ-OBSERVEX-002：依赖边界

Acceptance Criteria：

```text
AC-REQ-OBSERVEX-002-001: 允许依赖 foundationx
AC-REQ-OBSERVEX-002-002: 核心包不依赖 Zap/Logrus/Prometheus/OpenTelemetry
AC-REQ-OBSERVEX-002-003: 不依赖 PostgreSQL/Kafka/Redis/TDengine/OSS driver
AC-REQ-OBSERVEX-002-004: 不依赖 x.go
AC-REQ-OBSERVEX-002-005: 不出现业务模型和业务指标名
```

## REQ-OBSERVEX-003：Field

Acceptance Criteria：

```text
AC-REQ-OBSERVEX-003-001: 定义 Field
AC-REQ-OBSERVEX-003-002: 支持 String/Int/Int64/Float64/Bool/Duration/Time/Any/Error/Secret helper
AC-REQ-OBSERVEX-003-003: Secret helper 设置 Secret=true
AC-REQ-OBSERVEX-003-004: Error helper key 为 error
```

## REQ-OBSERVEX-004：Redactor

Acceptance Criteria：

```text
AC-REQ-OBSERVEX-004-001: DefaultRedactor 可脱敏 Field
AC-REQ-OBSERVEX-004-002: Secret=true 输出 ***
AC-REQ-OBSERVEX-004-003: secret key 自动识别大小写不敏感
AC-REQ-OBSERVEX-004-004: DSN / authorization / cookie 可脱敏
AC-REQ-OBSERVEX-004-005: tests 证明 secret 不泄露
```

## REQ-OBSERVEX-005：Logger

Acceptance Criteria：

```text
AC-REQ-OBSERVEX-005-001: 定义 Logger interface
AC-REQ-OBSERVEX-005-002: NoopLogger 不 panic
AC-REQ-OBSERVEX-005-003: SlogLogger 使用标准库 log/slog
AC-REQ-OBSERVEX-005-004: Logger 自动脱敏 fields
AC-REQ-OBSERVEX-005-005: Logger 支持 context
```

## REQ-OBSERVEX-006：Metrics

Acceptance Criteria：

```text
AC-REQ-OBSERVEX-006-001: 定义 Metrics interface
AC-REQ-OBSERVEX-006-002: NoopMetrics 不 panic
AC-REQ-OBSERVEX-006-003: 支持 Counter/Gauge/Histogram 语义
AC-REQ-OBSERVEX-006-004: ValidateMetricName 可校验名称
AC-REQ-OBSERVEX-006-005: ValidateLabels 可拒绝 secret/high-cardinality labels
```

## REQ-OBSERVEX-007：Tracer

Acceptance Criteria：

```text
AC-REQ-OBSERVEX-007-001: 定义 Tracer interface
AC-REQ-OBSERVEX-007-002: 定义 Span interface
AC-REQ-OBSERVEX-007-003: NoopTracer 不 panic
AC-REQ-OBSERVEX-007-004: Span.End 幂等
AC-REQ-OBSERVEX-007-005: RecordError 不 panic
```

## REQ-OBSERVEX-008：Context Helpers

Acceptance Criteria：

```text
AC-REQ-OBSERVEX-008-001: 支持 CorrelationID
AC-REQ-OBSERVEX-008-002: 支持 TraceID
AC-REQ-OBSERVEX-008-003: 支持 RequestID
AC-REQ-OBSERVEX-008-004: context key 私有
AC-REQ-OBSERVEX-008-005: FieldsFromContext 返回可用于日志的字段
```

## REQ-OBSERVEX-009：TestKit

Acceptance Criteria：

```text
AC-REQ-OBSERVEX-009-001: 提供 RecordingLogger
AC-REQ-OBSERVEX-009-002: 提供 RecordingMetrics
AC-REQ-OBSERVEX-009-003: 提供 RecordingTracer
AC-REQ-OBSERVEX-009-004: 提供 AssertNoSecretLeak
```

## REQ-OBSERVEX-010：Harness

Acceptance Criteria：

```text
AC-REQ-OBSERVEX-010-001: make ci 通过
AC-REQ-OBSERVEX-010-002: boundary gate 通过
AC-REQ-OBSERVEX-010-003: secret gate 通过
AC-REQ-OBSERVEX-010-004: contract gate 通过
AC-REQ-OBSERVEX-010-005: examples gate 通过
AC-REQ-OBSERVEX-010-006: release manifest 生成
```


# 10. Plan

```text
PLAN-GOAL-20260601-OBSERVEX-001-v1.0
```

## Phase 0：Context Recovery

目标：

```text
确认 observex 在基础库体系中的位置、依赖边界、x.go 集成方式。
```

输出：

```text
.agent/context.md
```

必须记录：

```text
observex 是 L1
foundationx 是 L0
observex 核心包不强绑定 Prometheus / OpenTelemetry / Zap / Logrus
x.go 负责选择具体观测实现
基础库只依赖 observex 接口
```

## Phase 1：Skeleton

创建：

```text
go.mod
README.md
CHANGELOG.md
Makefile
pkg/observex/*
docs/*
scripts/*
.agent/*
```

## Phase 2：Field + Redactor

实现：

```text
Field helpers
DefaultRedactor
secret key detection
```

## Phase 3：Logger

实现：

```text
Logger interface
NoopLogger
SlogLogger
```

## Phase 4：Metrics

实现：

```text
Metrics interface
NoopMetrics
metric name validation
label validation
```

## Phase 5：Tracer

实现：

```text
Tracer
Span
NoopTracer
NoopSpan
```

## Phase 6：Context Helpers

实现：

```text
CorrelationID
TraceID
RequestID
FieldsFromContext
```

## Phase 7：TestKit + Examples

实现：

```text
RecordingLogger
RecordingMetrics
RecordingTracer
AssertNoSecretLeak
examples
```

## Phase 8：Harness + CI

实现：

```text
boundary
secret
contract
examples
evidence gates
```

## Phase 9：Docs + ADR

补齐：

```text
README
docs
ADR
contracts
```

## Phase 10：Release + Retrospective

输出：

```text
release manifest
retrospective patches
```


# 11. Task Breakdown

## TASK-OBSERVEX-001：创建模块骨架

```bash
mkdir -p observex
cd observex
go mod init github.com/ZoneCNH/observex
mkdir -p pkg/observex internal/sanitize internal/validation internal/testutil testkit examples/logger examples/metrics examples/tracer examples/slog examples/redaction contracts docs/adr scripts release/manifest .agent .github/workflows
touch README.md CHANGELOG.md Makefile .gitignore .golangci.yml
```

证据：

```text
EVID-TASK-OBSERVEX-001-20260601-001: tree output
EVID-TASK-OBSERVEX-001-20260601-002: go env GOMOD
```

## TASK-OBSERVEX-002：接入 foundationx

```bash
go get github.com/ZoneCNH/foundationx
```

要求：

```text
不接入 x.go
不接入 driver
不接入 Prometheus/OpenTelemetry/Zap/Logrus
```

## TASK-OBSERVEX-003：实现 Field

文件：

```text
pkg/observex/field.go
pkg/observex/field_test.go
```

测试：

```text
TestFieldHelpers
TestSecretField
TestErrorField
TestDurationField
```

## TASK-OBSERVEX-004：实现 Redactor

文件：

```text
pkg/observex/redactor.go
pkg/observex/redactor_test.go
```

测试：

```text
TestRedactorMasksSecretFlag
TestRedactorDetectsPassword
TestRedactorDetectsToken
TestRedactorDetectsAuthorization
TestRedactorDetectsCookie
TestRedactorDetectsDSN
TestRedactorDoesNotOvermatchKey
TestRedactorNoSecretLeak
```

## TASK-OBSERVEX-005：实现 Logger / NoopLogger

文件：

```text
pkg/observex/logger.go
pkg/observex/logger_noop.go
pkg/observex/logger_test.go
```

测试：

```text
TestNoopLoggerDoesNotPanic
TestLoggerInterfaceCompile
TestLoggerAcceptsContext
```

## TASK-OBSERVEX-006：实现 SlogLogger

文件：

```text
pkg/observex/logger_slog.go
pkg/observex/logger_slog_test.go
```

测试：

```text
TestSlogLoggerWritesFields
TestSlogLoggerRedactsSecrets
TestSlogLoggerNilFallback
TestSlogLoggerContextFields
```

## TASK-OBSERVEX-007：实现 Metrics / NoopMetrics

文件：

```text
pkg/observex/metrics.go
pkg/observex/metrics_noop.go
pkg/observex/metrics_test.go
```

测试：

```text
TestNoopMetricsDoesNotPanic
TestMetricsInterfaceCompile
TestCounterGaugeHistogramMethods
```

## TASK-OBSERVEX-008：实现 Metric Name / Labels

文件：

```text
pkg/observex/metric_name.go
pkg/observex/labels.go
pkg/observex/labels_test.go
```

测试：

```text
TestValidateMetricNameValid
TestValidateMetricNameInvalid
TestValidateLabelsValid
TestValidateLabelsRejectSecret
TestValidateLabelsRejectHighCardinality
TestSanitizeLabelsMasksSecrets
```

## TASK-OBSERVEX-009：实现 Tracer / NoopTracer

文件：

```text
pkg/observex/tracer.go
pkg/observex/tracer_noop.go
pkg/observex/tracer_test.go
```

测试：

```text
TestNoopTracerStart
TestNoopSpanEndIdempotent
TestNoopSpanRecordError
TestNoopSpanSetField
TestTracerInterfaceCompile
```

## TASK-OBSERVEX-010：实现 Context Helpers

文件：

```text
pkg/observex/context.go
pkg/observex/context_test.go
```

测试：

```text
TestWithCorrelationID
TestWithTraceID
TestWithRequestID
TestFieldsFromContext
TestContextKeysPrivate
```

## TASK-OBSERVEX-011：实现 Error Mapping

文件：

```text
pkg/observex/errors.go
pkg/observex/errors_test.go
```

测试：

```text
TestMapErrorInvalidMetricName
TestMapErrorInvalidLabel
TestMapErrorContextCanceled
TestMapErrorDeadlineExceeded
```

## TASK-OBSERVEX-012：实现 TestKit

文件：

```text
testkit/logger.go
testkit/metrics.go
testkit/tracer.go
testkit/assert.go
```

能力：

```text
RecordingLogger
RecordingMetrics
RecordingTracer
AssertNoSecretLeak
AssertMetricRecorded
AssertSpanRecorded
```

## TASK-OBSERVEX-013：编写 Examples

目录：

```text
examples/logger
examples/metrics
examples/tracer
examples/slog
examples/redaction
```

要求：

```text
1. examples 不包含真实密钥
2. examples 可以 go run
3. examples 展示 secret redaction
4. examples 不包含 x.go 业务语义
```

## TASK-OBSERVEX-014：编写 Harness Scripts

文件：

```text
scripts/check_boundary.sh
scripts/check_secrets.sh
scripts/check_contracts.sh
scripts/generate_manifest.sh
```

## TASK-OBSERVEX-015：编写 Makefile

必须包含：

```text
fmt
vet
lint
test
race
boundary
security
contracts
examples
evidence
ci
release-check
```

## TASK-OBSERVEX-016：编写 GitHub Actions

文件：

```text
.github/workflows/ci.yml
.github/workflows/security.yml
.github/workflows/release.yml
```

## TASK-OBSERVEX-017：编写文档与 ADR

必须完成：

```text
README.md
docs/spec.md
docs/design.md
docs/api.md
docs/logger.md
docs/metrics.md
docs/tracer.md
docs/context.md
docs/redaction.md
docs/label-policy.md
docs/xgo-integration.md
docs/testing.md
docs/release.md
docs/adr/ADR-20260601-001-core-no-heavy-deps.md
docs/adr/ADR-20260601-002-noop-defaults.md
docs/adr/ADR-20260601-003-label-cardinality-policy.md
docs/adr/ADR-20260601-004-adapters-deferred.md
```

## TASK-OBSERVEX-018：生成 Release Manifest

命令：

```bash
make evidence
```

输出：

```text
release/manifest/v0.1.0.json
```

## TASK-OBSERVEX-019：x.go 集成示例文档

文件：

```text
docs/xgo-integration.md
```

必须说明：

```text
1. x.go 启动层选择具体 logger/metrics/tracer
2. postgresx/kafkax/redisx/taosx 只接收 observex.Logger / Metrics / Tracer
3. x.go 业务指标名保留在 x.go，不进入 observex
4. trace_id/request_id 可进日志和 trace，不进 metrics label
```

## TASK-OBSERVEX-020：Retrospective

输出：

```text
.agent/retrospective.md
.agent/patch_prompt.md
.agent/patch_harness.md
.agent/patch_rule.md
```


# 12. Harness Gates

## Gate 1：Format

```bash
go fmt ./...
```

## Gate 2：Vet

```bash
go vet ./...
```

## Gate 3：Unit Test

```bash
go test ./...
```

## Gate 4：Race Test

```bash
go test -race ./...
```

## Gate 5：Boundary

```bash
./scripts/check_boundary.sh
```

必须检查：

```text
不依赖 github.com/ZoneCNH/x.go
不依赖 PostgreSQL/Kafka/Redis/TDengine/OSS driver
核心包不依赖 Prometheus/OpenTelemetry/Zap/Logrus
不出现业务术语
```

## Gate 6：Secret

```bash
./scripts/check_secrets.sh
```

必须检查：

```text
源码、examples、docs、release manifest 不包含真实 secret
Redactor tests 证明 secret 不泄露
```

## Gate 7：Contract

```bash
./scripts/check_contracts.sh
```

检查：

```text
contracts/logger.schema.json
contracts/metrics.schema.json
contracts/tracer.schema.json
contracts/field.schema.json
contracts/public_api.md
contracts/metric_naming.md
docs/api.md
```

## Gate 8：Examples

```bash
go run ./examples/logger
go run ./examples/metrics
go run ./examples/tracer
go run ./examples/slog
go run ./examples/redaction
```

## Gate 9：Evidence

```bash
./scripts/generate_manifest.sh
```

生成：

```text
release/manifest/v0.1.0.json
```

---

# 13. Boundary Gate 脚本模板

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "checking observex boundary..."

FORBIDDEN_DEPS=(
  "github.com/ZoneCNH/x.go"
  "github.com/ZoneCNH/x.go/internal"
  "database/sql"
  "github.com/jackc/pgx"
  "github.com/segmentio/kafka-go"
  "github.com/IBM/sarama"
  "github.com/confluentinc/confluent-kafka-go"
  "github.com/redis/go-redis"
  "github.com/taosdata"
  "github.com/prometheus/client_golang"
  "go.opentelemetry.io/otel"
  "go.uber.org/zap"
  "github.com/sirupsen/logrus"
)

DEPS="$(go list -deps ./...)"

for dep in "${FORBIDDEN_DEPS[@]}"; do
  if echo "$DEPS" | grep -q "$dep"; then
    echo "ERROR: forbidden dependency found: $dep"
    exit 1
  fi
done

FORBIDDEN_TERMS=(
  "BTCUSDT"
  "ETHUSDT"
  "Kline"
  "OrderBook"
  "MarketData"
  "MacroData"
  "MacroRegime"
  "MarketRegime"
  "TradingSignal"
  "Position"
  "RiskGate"
)

for term in "${FORBIDDEN_TERMS[@]}"; do
  if grep -R "$term" ./pkg ./internal ./testkit --exclude-dir=.git; then
    echo "ERROR: forbidden business term found: $term"
    exit 1
  fi
done

echo "observex boundary check passed"
```

---

# 14. Secret Gate 脚本模板

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "checking secrets..."

PATTERNS=(
  "AKIA[0-9A-Z]{16}"
  "BEGIN RSA PRIVATE KEY"
  "BEGIN OPENSSH PRIVATE KEY"
  "BEGIN PRIVATE KEY"
  "xoxb-[0-9A-Za-z-]+"
  "ghp_[0-9A-Za-z_]+"
)

for pattern in "${PATTERNS[@]}"; do
  if grep -R -E "$pattern" .     --exclude-dir=.git     --exclude-dir=vendor     --exclude="*.sum"     --exclude="go.sum"; then
    echo "ERROR: possible secret found: $pattern"
    exit 1
  fi
done

echo "secret check passed"
```

---

# 15. Makefile 模板

```makefile
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test ./...

.PHONY: race
race:
	go test -race ./...

.PHONY: boundary
boundary:
	chmod +x scripts/*.sh
	./scripts/check_boundary.sh

.PHONY: security
security:
	chmod +x scripts/*.sh
	./scripts/check_secrets.sh

.PHONY: contracts
contracts:
	chmod +x scripts/*.sh
	./scripts/check_contracts.sh

.PHONY: examples
examples:
	go run ./examples/logger
	go run ./examples/metrics
	go run ./examples/tracer
	go run ./examples/slog
	go run ./examples/redaction

.PHONY: evidence
evidence:
	chmod +x scripts/*.sh
	./scripts/generate_manifest.sh

.PHONY: ci
ci: fmt vet test race boundary security contracts examples

.PHONY: release-check
release-check: ci evidence
```


# 16. GitHub Actions 模板

```yaml
name: observex-ci

on:
  pull_request:
  push:
    branches:
      - main

jobs:
  ci:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.23"

      - name: Cache Go
        uses: actions/cache@v4
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

      - name: Make scripts executable
        run: chmod +x scripts/*.sh

      - name: CI
        run: make ci

      - name: Generate evidence
        run: make evidence

      - name: Upload release manifest
        uses: actions/upload-artifact@v4
        with:
          name: observex-release-manifest
          path: release/manifest/*.json
```

---

# 17. Release Manifest 模板

```json
{
  "module": "github.com/ZoneCNH/observex",
  "version": "v0.1.0",
  "commit": "COMMIT_SHA",
  "go_version": "go1.23.x",
  "generated_at": "2026-06-01T00:00:00Z",
  "dependencies": {
    "foundationx": "version-from-go-mod"
  },
  "checks": {
    "fmt": "passed",
    "vet": "passed",
    "unit_test": "passed",
    "race_test": "passed",
    "boundary": "passed",
    "secret_scan": "passed",
    "contract": "passed",
    "examples": "passed"
  },
  "features": {
    "logger_interface": "enabled",
    "noop_logger": "enabled",
    "slog_adapter": "enabled",
    "metrics_interface": "enabled",
    "noop_metrics": "enabled",
    "tracer_interface": "enabled",
    "noop_tracer": "enabled",
    "redactor": "enabled",
    "context_helpers": "enabled",
    "prometheus_adapter": "deferred",
    "opentelemetry_adapter": "deferred",
    "zap_adapter": "deferred"
  },
  "security": {
    "field_redaction": "verified",
    "label_policy": "verified",
    "secret_scan": "passed"
  },
  "artifacts": [
    "coverage.out",
    "contract-report.json"
  ],
  "notes": {
    "breaking_changes": "none",
    "known_risks": []
  }
}
```

---

# 18. Traceability Matrix

| Requirement | Acceptance Criteria | Design | Task | Test | Evidence | Status |
|---|---|---|---|---|---|---|
| REQ-OBSERVEX-001 | AC-001-* | Module Design | TASK-001 | go test ./... | EVID-001 | TODO |
| REQ-OBSERVEX-002 | AC-002-* | Boundary | TASK-014 | boundary gate | EVID-014 | TODO |
| REQ-OBSERVEX-003 | AC-003-* | Field | TASK-003 | field_test.go | EVID-003 | TODO |
| REQ-OBSERVEX-004 | AC-004-* | Redactor | TASK-004 | redactor_test.go | EVID-004 | TODO |
| REQ-OBSERVEX-005 | AC-005-* | Logger | TASK-005/006 | logger_test.go | EVID-005 | TODO |
| REQ-OBSERVEX-006 | AC-006-* | Metrics | TASK-007/008 | metrics_test.go | EVID-007 | TODO |
| REQ-OBSERVEX-007 | AC-007-* | Tracer | TASK-009 | tracer_test.go | EVID-009 | TODO |
| REQ-OBSERVEX-008 | AC-008-* | Context | TASK-010 | context_test.go | EVID-010 | TODO |
| REQ-OBSERVEX-009 | AC-009-* | TestKit | TASK-012 | testkit tests | EVID-012 | TODO |
| REQ-OBSERVEX-010 | AC-010-* | Harness | TASK-014/015/018 | make release-check | EVID-018 | TODO |
```

---

# 19. Risk Register

## RISK-OBSERVEX-001：核心包变重

风险：

```text
observex 直接引入 Prometheus、OpenTelemetry、Zap 等依赖，导致所有基础库被动变重。
```

缓解：

```text
核心包只定义接口和 Noop/Slog。
外部 adapter 独立模块或 v0.2。
Boundary Gate 检查 forbidden deps。
```

## RISK-OBSERVEX-002：Secret 泄露

风险：

```text
日志 Field、Trace Field、Metric Label 泄露 password/token/dsn。
```

缓解：

```text
DefaultRedactor
Secret Field
Label policy
Secret Gate
NoSecretLeak tests
```

## RISK-OBSERVEX-003：高基数 label 污染 metrics

风险：

```text
trace_id/request_id/user_id/order_id 被用于 metric label。
```

缓解：

```text
ValidateLabels 拒绝高基数字段。
docs/label-policy.md 固化规则。
```

## RISK-OBSERVEX-004：业务语义污染

风险：

```text
observex 内置 x.go 指标名、业务 span 名称。
```

缓解：

```text
observex 只定义接口和命名规则。
x.go 定义业务指标。
Boundary Gate 检查业务词汇。
```

## RISK-OBSERVEX-005：接口过度抽象

风险：

```text
Logger/Metrics/Tracer 设计过大，基础库接入困难。
```

缓解：

```text
v0.1 只保留最小方法集合。
新增方法必须证明至少两个基础库需要。
```

---

# 20. Decision Log

## DEC-20260601-001：核心包不引入重依赖

决策：

```text
observex core 不引入 Prometheus / OpenTelemetry / Zap / Logrus。
```

原因：

```text
避免基础库体系根依赖膨胀。
```

## DEC-20260601-002：Noop 默认

决策：

```text
Logger / Metrics / Tracer 都提供 Noop 实现。
```

原因：

```text
保证基础库无观测注入时仍可运行。
```

## DEC-20260601-003：Metrics label 严格治理

决策：

```text
trace_id/request_id/user_id/order_id 等高基数字段默认禁止作为 label。
```

原因：

```text
保护 Prometheus 类系统不被高基数拖垮。
```

## DEC-20260601-004：外部 adapter 延后

决策：

```text
Prometheus / OpenTelemetry / Zap adapter 不进入 core v0.1。
```

原因：

```text
保持 observex 核心稳定、轻量、可复用。
```

---

# 21. AutoResearch Protocol

触发条件：

```text
1. 是否将 slog adapter 纳入 v0.1
2. log/slog Attr 转换行为不确定
3. metric name 正则与 Prometheus 兼容性不确定
4. label key 规则不确定
5. OpenTelemetry adapter 是否单独建库
6. Prometheus adapter 是否单独建库
7. GitHub Actions action 版本不确定
```

输出必须写入：

```text
docs/adr/ADR-YYYYMMDD-NNN-<topic>.md
```

禁止：

```text
1. 不经 ADR 引入 Prometheus / OpenTelemetry / Zap / Logrus
2. 不经 Review 扩大 Logger/Metrics/Tracer 接口
3. 不经 Gate 放宽 label policy
```


# 22. x.go 集成规范

x.go 启动层正确方式：

```go
logger := observex.NewSlogLogger(slog.Default())
metrics := observex.NewNoopMetrics()
tracer := observex.NewNoopTracer()

pgClient, err := postgresx.New(
	ctx,
	pgCfg,
	postgresx.WithLogger(logger),
	postgresx.WithMetrics(metrics),
)
```

基础库正确方式：

```go
type Client struct {
	logger observex.Logger
	metrics observex.Metrics
	tracer observex.Tracer
}

func New(ctx context.Context, cfg Config, opts ...Option) (*Client, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if options.logger == nil {
		options.logger = observex.NewNoopLogger()
	}
	if options.metrics == nil {
		options.metrics = observex.NewNoopMetrics()
	}
	if options.tracer == nil {
		options.tracer = observex.NewNoopTracer()
	}
}
```

业务指标留在 x.go：

```text
xgo_market_kline_ingested_total
xgo_macro_regime_detected_total
xgo_regime_transition_total
```

这些不进入 observex。

规则：

```text
1. request_id / trace_id 可进日志字段
2. request_id / trace_id 默认不进 metrics label
3. password / dsn / token 永远不得明文输出
4. 基础库只使用通用指标名或由调用方注入 metric prefix
```

---

# 23. Release Protocol

## 23.1 v0.1.0 发布前

执行：

```bash
make release-check
```

必须通过：

```text
fmt
vet
test
race
boundary
security
contracts
examples
evidence
```

## 23.2 CHANGELOG

```markdown
## v0.1.0 - 2026-06-01

### Added
- Added Field model and field helpers.
- Added DefaultRedactor and secret key detection.
- Added Logger interface.
- Added NoopLogger.
- Added standard library slog adapter.
- Added Metrics interface.
- Added NoopMetrics.
- Added metric name and label validation.
- Added Tracer and Span interfaces.
- Added NoopTracer and NoopSpan.
- Added context helpers for correlation_id, trace_id, and request_id.
- Added TestKit and examples.
- Added boundary, secret, contract, example, and evidence gates.

### Security
- Secret fields are masked by default.
- Metrics labels reject secret and high-cardinality keys.
- Secret Gate added.

### Deferred
- Prometheus adapter.
- OpenTelemetry adapter.
- Zap adapter.
- Logrus adapter.

### Breaking Changes
- None.
```

## 23.3 Release 声明

```text
DONE with evidence:
- make release-check passed
- go test ./... passed
- go test -race ./... passed
- boundary gate passed
- secret gate passed
- examples passed
- release/manifest/v0.1.0.json generated
```

---

# 24. Retrospective Protocol

输出：

```text
.agent/retrospective.md
```

模板：

```markdown
# observex Retrospective

## Release
- Version:
- Commit:
- Date:

## What worked
-

## What failed
-

## API stability concerns
-

## Boundary risks
-

## Security findings
-

## Redaction findings
-

## Label policy findings
-

## Harness improvements
-

## Adapter candidates
- observex-prometheus:
- observex-otel:
- observex-zap:

## Reusable patterns for other base libs
- postgresx:
- redisx:
- kafkax:
- taosx:
- configx:

## Next issue candidates
-

## Patch outputs
- PATCH-PROMPT:
- PATCH-HARNESS:
- PATCH-RULE:
```

---

# 25. Final DoD

## Task DoD

```text
代码实现完成
单元测试完成
无业务语义污染
无 x.go 依赖
无重型观测依赖
无 driver 依赖
无密钥泄露
go fmt / go vet / go test / go test -race 通过
```

## Module DoD

```text
Field 完整
Redactor 完整
Logger 完整
NoopLogger 完整
SlogLogger 完整
Metrics 完整
NoopMetrics 完整
MetricName/LabelPolicy 完整
Tracer 完整
NoopTracer 完整
Context Helpers 完整
Error Mapping 完整
TestKit 完整
Examples 完整
Docs 完整
ADR 完整
Harness 完整
Release Manifest 完整
```

## Goal DoD

```text
observex 可作为 x.go 和基础库体系的可观测性契约库使用
observex 不依赖 x.go
observex core 不依赖 Prometheus/OpenTelemetry/Zap/Logrus
observex 不依赖 driver
observex 不包含业务指标
observex 不泄露 secret
observex v0.1.0 release evidence 完整
retrospective patch 生成
```

完成声明必须是：

```text
DONE with evidence:
- go test ./... passed
- go test -race ./... passed
- make ci passed
- make release-check passed
- boundary gate passed
- secret gate passed
- examples passed
- release/manifest/v0.1.0.json generated
```

---

# 26. 最小可行执行顺序

Agent 执行时按以下顺序，不要跳步：

```text
1. 创建 go module 和目录结构
2. 接入 foundationx
3. 编写 core-no-heavy-deps ADR
4. 实现 Field
5. 实现 Redactor
6. 实现 Logger / NoopLogger
7. 实现 SlogLogger
8. 实现 Metrics / NoopMetrics
9. 实现 MetricName / LabelPolicy
10. 实现 Tracer / NoopTracer
11. 实现 Context Helpers
12. 实现 Error Mapping
13. 实现 TestKit
14. 编写 Examples
15. 编写 scripts
16. 编写 Makefile
17. 编写 GitHub Actions
18. 编写 docs/contracts
19. 运行 make ci
20. 运行 make release-check
21. 生成 release manifest
22. 编写 retrospective
23. 输出 DONE with evidence
```

---

# 27. 给 Agent 的最终执行指令

```text
你现在要执行 GOAL-20260601-OBSERVEX-001。

请严格按 Goal Runtime Prompt v3.1 执行：
Goal → Context Recovery → Spec → Design → Plan → Tasks → Execution → Verification → Evidence → Review → Release → Retrospective → Self-improving。

你必须创建或完善 github.com/ZoneCNH/observex。

硬性约束：
1. observex 是 L1 可观测性契约库。
2. observex 必须依赖 foundationx。
3. observex 不允许依赖 github.com/ZoneCNH/x.go。
4. observex core 不允许依赖 Prometheus/OpenTelemetry/Zap/Logrus。
5. observex 不允许依赖 PostgreSQL/Kafka/Redis/TDengine/OSS driver。
6. observex 不允许包含 x.go 业务语义和业务指标。
7. observex 必须提供 Noop Logger/Metrics/Tracer。
8. observex 必须内置字段脱敏与 label policy。
9. observex 不允许在日志、错误、Evidence 中输出 secret 原值。
10. 不允许没有 Evidence 就声称 DONE。

必须实现：
1. Field model and helpers
2. DefaultRedactor
3. Logger interface
4. NoopLogger
5. SlogLogger
6. Metrics interface
7. NoopMetrics
8. Metric name validation
9. Label validation and sanitization
10. Tracer and Span interfaces
11. NoopTracer
12. Context helpers
13. Error Mapping
14. TestKit
15. Examples
16. Harness scripts
17. Makefile
18. GitHub Actions
19. Docs / ADR
20. Release Manifest
21. Retrospective patches

执行完成后输出：

DONE with evidence:
- 具体命令
- 具体测试结果
- 具体文件路径
- release manifest 路径
- known risks
- next recommended issue
```

---

# 28. 最终推荐路径

observex v0.1.0 必须先做“轻核心、强契约、安全脱敏”：

```text
Field
Redactor
Logger
Metrics
Tracer
Noop
Slog
Context
LabelPolicy
Evidence
```

暂不做：

```text
Prometheus adapter
OpenTelemetry adapter
Zap adapter
Logrus adapter
业务指标库
Dashboard
Alert rules
Exporter
```

最重要的三条红线：

```text
1. core 不引入重型观测依赖
2. 不承载业务指标
3. 不泄露 secret / 不污染高基数 label
```

最小交付：

```text
v0.1.0 = Logger + Metrics + Tracer 统一契约 + Noop + Slog + Redactor + LabelPolicy + Harness + Release Evidence
```
