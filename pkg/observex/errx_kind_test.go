package observex

import "testing"

func TestErrorKindToLabel(t *testing.T) {
	tests := []struct {
		name string
		kind ErrorKind
		want string
	}{
		{"config", ErrorKindConfig, "config"},
		{"validation", ErrorKindValidation, "validation"},
		{"connection", ErrorKindConnection, "connection"},
		{"unavailable", ErrorKindUnavailable, "unavailable"},
		{"timeout", ErrorKindTimeout, "timeout"},
		{"auth", ErrorKindAuth, "auth"},
		{"conflict", ErrorKindConflict, "conflict"},
		{"rate_limit", ErrorKindRateLimit, "rate_limit"},
		{"canceled", ErrorKindCanceled, "canceled"},
		{"not_found", ErrorKindNotFound, "not_found"},
		{"already_exists", ErrorKindAlreadyExists, "already_exists"},
		{"internal", ErrorKindInternal, "internal"},
		{"unknown_kind", ErrorKind("custom"), "unknown"},
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
