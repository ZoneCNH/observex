package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ZoneCNH/observex/pkg/observex"
)

func main() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := observex.NewSlogLogger(slog.New(handler))
	ctx := observex.WithTraceID(context.Background(), "trace-example")
	logger.Info(ctx, "client ready", observex.Secret("api_key", "raw-value-123"))
}
