package observex

import "github.com/ZoneCNH/foundationx/pkg/foundationx"

// ErrorKindToLabel converts an errx/foundationx ErrorKind to a metric label value.
// The returned string is always valid snake_case suitable for use as a metric label.
func ErrorKindToLabel(kind foundationx.ErrorKind) string {
	switch kind {
	case foundationx.ErrorKindConfig:
		return "config"
	case foundationx.ErrorKindValidation:
		return "validation"
	case foundationx.ErrorKindConnection:
		return "connection"
	case foundationx.ErrorKindUnavailable:
		return "unavailable"
	case foundationx.ErrorKindTimeout:
		return "timeout"
	case foundationx.ErrorKindAuth:
		return "auth"
	case foundationx.ErrorKindConflict:
		return "conflict"
	case foundationx.ErrorKindRateLimit:
		return "rate_limit"
	case foundationx.ErrorKindCanceled:
		return "canceled"
	case foundationx.ErrorKindNotFound:
		return "not_found"
	case foundationx.ErrorKindAlreadyExist:
		return "already_exists"
	case foundationx.ErrorKindInternal:
		return "internal"
	default:
		return "unknown"
	}
}
