package observex

import (
	"context"
	"io"
	"log/slog"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
}

type NoopLogger struct{}

// NewNoopLogger returns a logger that intentionally drops all records.
func NewNoopLogger() NoopLogger {
	return NoopLogger{}
}

func (NoopLogger) Debug(ctx context.Context, msg string, fields ...Field) {}

func (NoopLogger) Info(ctx context.Context, msg string, fields ...Field) {}

func (NoopLogger) Warn(ctx context.Context, msg string, fields ...Field) {}

func (NoopLogger) Error(ctx context.Context, msg string, fields ...Field) {}

type SlogLogger struct {
	logger   *slog.Logger
	redactor Redactor
}

type LoggerOption func(*loggerOptions)

type loggerOptions struct {
	redactor Redactor
}

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

func WithRedactor(redactor Redactor) LoggerOption {
	return func(o *loggerOptions) {
		if redactor != nil {
			o.redactor = redactor
		}
	}
}

func (l *SlogLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slog.LevelDebug, msg, fields...)
}

func (l *SlogLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slog.LevelInfo, msg, fields...)
}

func (l *SlogLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, slog.LevelWarn, msg, fields...)
}

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
