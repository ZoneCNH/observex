package observex

import "time"

// Attr aliases Field for callers that prefer attribute terminology.
type Attr = Field

// Field carries a structured observability value.
type Field struct {
	// Key identifies the field in logs, spans, and related records.
	Key string
	// Value contains the field payload.
	Value any
	// Secret marks Value for redaction before it is recorded or emitted.
	Secret bool
}

// String returns a string-valued Field.
func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

// Int returns an int-valued Field.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 returns an int64-valued Field.
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 returns a float64-valued Field.
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool returns a bool-valued Field.
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration returns a time.Duration-valued Field.
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Time returns a time.Time-valued Field.
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Any returns a Field for values without a narrower helper.
func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// Secret returns a Field whose value must be redacted before emission.
func Secret(key string, value any) Field {
	return Field{Key: key, Value: value, Secret: true}
}

// Err returns an error Field with the conventional "error" key.
func Err(err error) Field {
	return ErrorField(err)
}

// ErrorField returns an error Field with an empty value for nil errors.
func ErrorField(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: ""}
	}
	return Field{Key: "error", Value: err.Error()}
}
