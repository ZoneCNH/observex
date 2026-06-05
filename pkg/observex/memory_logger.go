package observex

import (
	"context"
	"sync"
)

// LogLevel identifies the severity of an in-memory log record.
type LogLevel string

const (
	// LogLevelDebug represents diagnostic log records.
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo represents informational log records.
	LogLevelInfo LogLevel = "info"
	// LogLevelWarn represents warning log records.
	LogLevelWarn LogLevel = "warn"
	// LogLevelError represents error log records.
	LogLevelError LogLevel = "error"
)

// LogRecord captures a single call made to MemoryLogger.
type LogRecord struct {
	// Sequence is the one-based order in which the record was captured.
	Sequence uint64
	// Level is the severity supplied by the log method.
	Level LogLevel
	// Message is the log message.
	Message string
	// Fields are the redacted context and call fields captured with the record.
	Fields []Field
}

// MemoryLogger records log entries in memory for tests and examples.
type MemoryLogger struct {
	mu       sync.Mutex
	redactor Redactor
	next     uint64
	records  []LogRecord
}

// NewMemoryLogger returns a logger that stores redacted records in memory.
func NewMemoryLogger(opts ...LoggerOption) *MemoryLogger {
	options := loggerOptions{redactor: NewDefaultRedactor()}
	for _, opt := range opts {
		opt(&options)
	}
	if options.redactor == nil {
		options.redactor = NewDefaultRedactor()
	}
	return &MemoryLogger{redactor: options.redactor}
}

// Debug records a debug log entry.
func (l *MemoryLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LogLevelDebug, msg, fields...)
}

// Info records an informational log entry.
func (l *MemoryLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LogLevelInfo, msg, fields...)
}

// Warn records a warning log entry.
func (l *MemoryLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LogLevelWarn, msg, fields...)
}

// Error records an error log entry.
func (l *MemoryLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LogLevelError, msg, fields...)
}

// Records returns a snapshot of the recorded log entries.
func (l *MemoryLogger) Records() []LogRecord {
	if l == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	out := make([]LogRecord, len(l.records))
	for i, record := range l.records {
		out[i] = record
		out[i].Fields = cloneFields(record.Fields)
	}
	return out
}

// Reset removes all recorded log entries.
func (l *MemoryLogger) Reset() {
	if l == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.records = nil
	l.next = 0
}

func (l *MemoryLogger) log(ctx context.Context, level LogLevel, msg string, fields ...Field) {
	if l == nil {
		return
	}
	allFields := append(FieldsFromContext(ctx), fields...)
	allFields = redactFields(l.redactor, allFields)

	l.mu.Lock()
	defer l.mu.Unlock()

	l.next++
	l.records = append(l.records, LogRecord{
		Sequence: l.next,
		Level:    level,
		Message:  msg,
		Fields:   cloneFields(allFields),
	})
}
