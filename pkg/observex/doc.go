// Package observex provides lightweight observability contracts for base libraries.
//
// The package defines Logger, Metrics, Tracer, Field, Redactor, context propagation,
// health, errors, tests, contracts, CI gates, release manifest, and agent evidence.
//
// This package must not depend on x.go, domain models, or concrete observability
// implementations such as Prometheus, OpenTelemetry, Zap, or Logrus.
package observex
