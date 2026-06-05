package observex

import (
	"regexp"
	"strings"
)

// Labels carries low-cardinality metric dimensions.
type Labels map[string]string

var (
	metricNameRE = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)
	labelKeyRE   = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)
)

// ValidateMetricName checks that name is safe for observex metrics.
func ValidateMetricName(name string) error {
	if !metricNameRE.MatchString(strings.TrimSpace(name)) {
		return validationError("ValidateMetricName", "metric name must be lower snake case and start with a letter", nil)
	}
	return nil
}

// ValidateLabels checks label keys and values for cardinality and secret risks.
func ValidateLabels(labels Labels) error {
	for key, value := range labels {
		if !labelKeyRE.MatchString(strings.TrimSpace(key)) {
			return validationError("ValidateLabels", "label key must be lower snake case and start with a letter", nil)
		}
		if isReservedLabelKey(key) {
			return validationError("ValidateLabels", "label key is reserved for high-cardinality data", nil)
		}
		if IsSecretKey(key) {
			return validationError("ValidateLabels", "label key may expose a secret", nil)
		}
		if valueLooksSecret(value) {
			return validationError("ValidateLabels", "label value may expose a secret", nil)
		}
	}
	return nil
}

// SanitizeLabels returns a copy with unsafe labels removed or redacted.
func SanitizeLabels(labels Labels) Labels {
	if len(labels) == 0 {
		return nil
	}
	sanitized := make(Labels, len(labels))
	for key, value := range labels {
		key = strings.TrimSpace(key)
		if !labelKeyRE.MatchString(key) || isReservedLabelKey(key) || IsSecretKey(key) {
			continue
		}
		if valueLooksSecret(value) {
			value = RedactedValue
		}
		sanitized[key] = value
	}
	if len(sanitized) == 0 {
		return nil
	}
	return sanitized
}

// CloneLabels returns a shallow copy of labels.
func CloneLabels(labels Labels) Labels {
	if len(labels) == 0 {
		return nil
	}
	copied := make(Labels, len(labels))
	for key, value := range labels {
		copied[key] = value
	}
	return copied
}

func isReservedLabelKey(key string) bool {
	switch normalizeSecretKey(key) {
	case "trace_id", "request_id", "correlation_id", "user_id", "order_id",
		"timestamp", "raw_error", "sql", "payload":
		return true
	default:
		return false
	}
}

func valueLooksSecret(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	indicators := []string{
		"password" + "=",
		"passwd" + "=",
		"secret" + "=",
		"token" + "=",
		"authorization:",
		"bearer ",
		"access_key" + "=",
		"secret_key" + "=",
	}
	for _, indicator := range indicators {
		if strings.Contains(normalized, indicator) {
			return true
		}
	}
	return false
}
