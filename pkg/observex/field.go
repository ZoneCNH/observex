package observex

import "time"

type Attr = Field

type Field struct {
	Key    string
	Value  any
	Secret bool
}

func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}

func Secret(key string, value any) Field {
	return Field{Key: key, Value: value, Secret: true}
}

func Err(err error) Field {
	return ErrorField(err)
}

func ErrorField(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: ""}
	}
	return Field{Key: "error", Value: err.Error()}
}
