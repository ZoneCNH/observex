package observex

// ErrorKindToLabel converts an ErrorKind to a metric label value.
// The returned string is always valid snake_case suitable for use as a metric label.
func ErrorKindToLabel(kind ErrorKind) string {
	switch kind {
	case ErrorKindConfig:
		return "config"
	case ErrorKindValidation:
		return "validation"
	case ErrorKindConnection:
		return "connection"
	case ErrorKindUnavailable:
		return "unavailable"
	case ErrorKindTimeout:
		return "timeout"
	case ErrorKindAuth:
		return "auth"
	case ErrorKindConflict:
		return "conflict"
	case ErrorKindRateLimit:
		return "rate_limit"
	case ErrorKindCanceled:
		return "canceled"
	case ErrorKindNotFound:
		return "not_found"
	case ErrorKindAlreadyExists:
		return "already_exists"
	case ErrorKindInternal:
		return "internal"
	default:
		return "unknown"
	}
}
