# 脱敏契约

`observex` 的日志、trace 和 Evidence 输出不得泄露原始凭据。脱敏由 `Field`、`Secret`、`Redactor` 和 `foundationx.SecretString` 协作完成。

## 核心类型

- `RedactedValue`：固定输出 `***`。
- `Secret(key, value)`：把字段标记为敏感字段。
- `Redactor`：字段脱敏接口。
- `DefaultRedactor`：默认脱敏实现。
- `IsSecretKey`：判断 key 是否属于敏感字段名。

默认脱敏器会处理三类输入：

- `Secret` 标记的字段。
- 命中敏感 key 规则的字段。
- 实现 `foundationx.Sanitizer` 的字段值。

## 输出规则

脱敏后的字段可以进入日志、trace、测试输出和 release Evidence。需要保留原始值的调用方必须在自己的安全边界内管理，不得通过 `observex` 的日志或 manifest 输出原文。

新增字段时优先使用 `Secret` 显式标记，而不是依赖 key 名称猜测。
