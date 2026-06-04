package observex

import (
	"fmt"
	"strings"

	"github.com/ZoneCNH/foundationx/pkg/foundationx"
)

const RedactedValue = "***"

type Redactor interface {
	RedactField(field Field) Field
	RedactFields(fields []Field) []Field
}

type DefaultRedactor struct {
	extraKeys map[string]struct{}
}

func NewDefaultRedactor(extraKeys ...string) DefaultRedactor {
	keys := make(map[string]struct{}, len(extraKeys))
	for _, key := range extraKeys {
		normalized := normalizeSecretKey(key)
		if normalized != "" {
			keys[normalized] = struct{}{}
		}
	}
	return DefaultRedactor{extraKeys: keys}
}

func (r DefaultRedactor) RedactField(field Field) Field {
	if field.Key == "" {
		return field
	}
	if field.Secret || r.isSecretKey(field.Key) {
		field.Value = redactedString(field.Value)
		field.Secret = false
		return field
	}
	if sanitizer, ok := field.Value.(foundationx.Sanitizer); ok {
		field.Value = sanitizer.Sanitize()
	}
	return field
}

func (r DefaultRedactor) RedactFields(fields []Field) []Field {
	if len(fields) == 0 {
		return nil
	}
	redacted := make([]Field, 0, len(fields))
	for _, field := range fields {
		redacted = append(redacted, r.RedactField(field))
	}
	return redacted
}

func IsSecretKey(key string) bool {
	return NewDefaultRedactor().isSecretKey(key)
}

func (r DefaultRedactor) isSecretKey(key string) bool {
	normalized := normalizeSecretKey(key)
	if normalized == "" {
		return false
	}
	if _, ok := r.extraKeys[normalized]; ok {
		return true
	}

	switch normalized {
	case "password", "passwd", "passphrase", "secret", "token", "access_token",
		"refresh_token", "api_key", "access_key", "secret_key", "private_key",
		"dsn", "database_url", "authorization", "cookie":
		return true
	}

	return strings.Contains(normalized, "password") ||
		strings.Contains(normalized, "passwd") ||
		strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "private_key") ||
		strings.Contains(normalized, "api_key") ||
		strings.Contains(normalized, "access_key") ||
		strings.Contains(normalized, "database_url") ||
		strings.HasSuffix(normalized, "_dsn") ||
		strings.HasSuffix(normalized, "_authorization") ||
		strings.HasSuffix(normalized, "_cookie")
}

func normalizeSecretKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	replacer := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	return replacer.Replace(key)
}

func redactedString(value any) string {
	if value == nil {
		return RedactedValue
	}
	if masked := foundationx.NewSecretString(fmt.Sprint(value)).String(); masked != "" {
		return masked
	}
	return RedactedValue
}
