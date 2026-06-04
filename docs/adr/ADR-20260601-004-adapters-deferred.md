# ADR-20260601-004 Adapter 延后到独立仓库

## 状态

Accepted

## 背景

Prometheus、OpenTelemetry、Zap 和 Logrus adapter 都有独立生命周期和依赖升级节奏。把它们放入模板核心会扩大基础库的依赖边界。

## 决策

`observex` 当前只交付接口、Noop 实现、`slog` 标准库适配器和测试夹具。后端 adapter 由独立仓库或上层应用实现。

## 后果

核心库保持稳定和轻量。后续如果需要官方 adapter，应以独立模块发布，并通过 contract 测试证明兼容 `observex` 接口。
