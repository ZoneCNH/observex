package observex

import (
	"testing"

	"github.com/ZoneCNH/foundationx/pkg/foundationx"
)

func TestErrorKindToLabel(t *testing.T) {
	tests := []struct {
		name string
		kind foundationx.ErrorKind
		want string
	}{
		{"config", foundationx.ErrorKindConfig, "config"},
		{"validation", foundationx.ErrorKindValidation, "validation"},
		{"connection", foundationx.ErrorKindConnection, "connection"},
		{"unavailable", foundationx.ErrorKindUnavailable, "unavailable"},
		{"timeout", foundationx.ErrorKindTimeout, "timeout"},
		{"auth", foundationx.ErrorKindAuth, "auth"},
		{"conflict", foundationx.ErrorKindConflict, "conflict"},
		{"rate_limit", foundationx.ErrorKindRateLimit, "rate_limit"},
		{"canceled", foundationx.ErrorKindCanceled, "canceled"},
		{"not_found", foundationx.ErrorKindNotFound, "not_found"},
		{"already_exists", foundationx.ErrorKindAlreadyExist, "already_exists"},
		{"internal", foundationx.ErrorKindInternal, "internal"},
		{"unknown_kind", foundationx.ErrorKind("custom"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorKindToLabel(tt.kind)
			if got != tt.want {
				t.Errorf("ErrorKindToLabel(%q) = %q, want %q", tt.kind, got, tt.want)
			}
			// Verify all labels are valid snake_case.
			if !labelKeyRE.MatchString(got) {
				t.Errorf("ErrorKindToLabel(%q) = %q is not valid snake_case", tt.kind, got)
			}
		})
	}
}
