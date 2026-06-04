package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ZoneCNH/observex/pkg/observex"
)

type stdoutLogger struct {
	redactor observex.Redactor
}

func (l stdoutLogger) Debug(ctx context.Context, msg string, fields ...observex.Field) {
	l.write(ctx, "debug", msg, fields...)
}

func (l stdoutLogger) Info(ctx context.Context, msg string, fields ...observex.Field) {
	l.write(ctx, "info", msg, fields...)
}

func (l stdoutLogger) Warn(ctx context.Context, msg string, fields ...observex.Field) {
	l.write(ctx, "warn", msg, fields...)
}

func (l stdoutLogger) Error(ctx context.Context, msg string, fields ...observex.Field) {
	l.write(ctx, "error", msg, fields...)
}

func (l stdoutLogger) write(ctx context.Context, level string, msg string, fields ...observex.Field) {
	allFields := append(observex.FieldsFromContext(ctx), fields...)
	if l.redactor != nil {
		allFields = l.redactor.RedactFields(allFields)
	}
	parts := make([]string, 0, len(allFields))
	for _, field := range allFields {
		if field.Key == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%v", field.Key, field.Value))
	}
	sort.Strings(parts)
	fmt.Printf("%s %s %s\n", level, msg, strings.Join(parts, " "))
}

func main() {
	logger := stdoutLogger{redactor: observex.NewDefaultRedactor()}
	ctx := observex.WithCorrelationID(context.Background(), "corr-example")
	logger.Info(ctx, "worker ready",
		observex.String("component", "example"),
		observex.Secret("api_key", "raw-value-123"),
	)
}
