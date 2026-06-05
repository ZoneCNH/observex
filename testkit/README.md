# Testkit 测试工具

为生成的基础库提供可复用测试夹具和断言。

## 契约

- `Config(name string)` 返回带 `Name` 和 `Timeout` 的最小有效配置。
- `RequireNoError(t, err)` 在 `err == nil` 时保持静默，在非空错误时终止当前测试。
- `RequireGolden(t, path, actual)` 读取 golden 文件并比较实际输出；不一致时报告 expected / actual 上下文。
- `CaptureStdout(t, fn)` 在执行 `fn` 时捕获 `os.Stdout` 输出，用于示例程序 smoke test。
- `AssertNoSecretLeak(t, text, rawSecrets...)` 验证输出中没有原始敏感值。
- `RecordingLogger` 是 public `observex.MemoryLogger` 的测试断言包装层，记录注入式日志调用和字段。
- `RecordingMetrics` 是 public `observex.MemoryMetrics` 的测试断言包装层，记录 counter、histogram 和 gauge 调用。
- `RecordingTracer` 是 public `observex.MemoryTracer` 的测试断言包装层，记录 span start、event、field 和 end 生命周期。

## 回归覆盖

`fixture_test.go` 锁定 `Config("fixture")` 的字段和 `Validate` 结果，并验证 `RequireNoError(t, nil)` 可用。`golden_test.go` 锁定 golden 断言的匹配路径。`assert_test.go` 锁定 `CaptureStdout` 对大输出的捕获行为。`observability_test.go` 锁定 recording logger、metrics 和 tracer 的基本行为。生成后的基础库需要保留这组最小测试，以防测试夹具随包名替换、配置 contract、可观测性 contract 或稳定输出漂移。

生成的库应保持此包独立于 `x.go` 和业务特定模型。


## Memory 记录模型

public `observex.MemoryLogger`、`MemoryMetrics` 和 `MemoryTracer` 是唯一 canonical recording model。`testkit.Recording*` 类型只提供测试友好的构造、reset 和断言入口，并通过类型别名暴露 public Memory 的 record shape；不要在 testkit 中重新实现一套独立记录语义。
