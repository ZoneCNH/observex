package observex

import (
	"context"
	"io"
	"log/slog"
)

// Logger records structured messages with optional fields.
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
}

// NoopLogger drops all log records.
type NoopLogger struct{}

// NewNoopLogger returns a logger that intentionally drops all records.
func NewNoopLogger() NoopLogger {
	return NoopLogger{}
}

// Debug drops a debug log record.
func (NoopLogger) Debug(ctx context.Context, msg string, fields ...Field) {}

// Info drops an informational log record.
func (NoopLogger) Info(ctx context.Context, msg string, fields ...Field) {}

// Warn drops a warning log record.
func (NoopLogger) Warn(ctx context.Context, msg string, fields ...Field) {}

// Error drops an error log record.
func (NoopLogger) Error(ctx context.Context, msg string, fields ...Field) {}

// SlogLogger adapts log/slog to the observex Logger interface.
type SlogLogger struct {
	logger   *slog.Logger
	redactor Redactor
}

// LoggerOption customizes loggers that support redaction options.
type LoggerOption func(*loggerOptions)

type loggerOptions struct {
	redactor Redactor
}

// NewSlogLogger returns a Logger backed by logger.
func NewSlogLogger(logger *slog.Logger, opts ...LoggerOption) *SlogLogger {
	options := loggerOptions{redactor: NewDefaultRedactor()}
	for _, opt := range opts {
		opt(&options)
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if options.redactor == nil {
		options.redactor = NewDefaultRedactor()
	}
	return &SlogLogger{logger: logger, redactor: options.redactor}
}

// WithRedactor configures a logger or tracer redactor.
func WithRedactor(redactor Redactor) LoggerOption {
	return func(o *loggerOptions) {
		if redactor != nil {
			o.redactor = redactor
		}
	}
}

// Debug records a debug log entry.
func (l *SlogLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slog.LevelDebug, msg, fields...)
}

// Info records an informational log entry.
func (l *SlogLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slog.LevelInfo, msg, fields...)
}

// Warn records a warning log entry.
func (l *SlogLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slog.LevelWarn, msg, fields...)
}

// Error records an error log entry.
func (l *SlogLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slog.LevelError, msg, fields...)
}

func (l *SlogLogger) log(ctx context.Context, level slog.Level, msg string, fields ...Field) {
	if l == nil || l.logger == nil {
		return
	}
	ctx = nonNilContext(ctx)
	if !l.logger.Enabled(ctx, level) {
		return
	}
	allFields := append(FieldsFromContext(ctx), fields...)
	if l.redactor != nil {
		allFields = l.redactor.RedactFields(allFields)
	}
	attrs := make([]slog.Attr, 0, len(allFields))
	for _, field := range allFields {
		if field.Key == "" {
			continue
		}
		attrs = append(attrs, slog.Any(field.Key, field.Value))
	}
	l.logger.LogAttrs(ctx, level, msg, attrs...)
}
