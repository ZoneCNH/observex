package observex

import (
	"regexp"
	"strings"
)

var (
	// nonSnakePattern matches characters that are not lowercase letters, digits, or underscores.
	nonSnakePattern = regexp.MustCompile(`[^a-z0-9_]+`)
	// multiUnderscore matches consecutive underscores.
	multiUnderscore = regexp.MustCompile(`_+`)
)

// MetricName builds a metric name with the foundationx naming convention:
//
//	foundationx_{module}_{operation}_{measure}
//
// All inputs are converted to snake_case automatically. Empty segments are
// skipped. Leading/trailing underscores are trimmed.
func MetricName(module, operation, measure string) string {
	parts := []string{"foundationx"}
	if module != "" {
		parts = append(parts, toSnakeCase(module))
	}
	if operation != "" {
		parts = append(parts, toSnakeCase(operation))
	}
	if measure != "" {
		parts = append(parts, toSnakeCase(measure))
	}
	return strings.Join(parts, "_")
}

// toSnakeCase converts an arbitrary string to lower snake_case.
// Handles CamelCase, PascalCase, acronyms (HTTPRequest → http_request),
// hyphens, dots, and spaces.
func toSnakeCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	var buf strings.Builder
	runes := []rune(s)
	n := len(runes)

	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := runes[i-1]
				// Insert underscore between lowercase → uppercase.
				if prev >= 'a' && prev <= 'z' {
					buf.WriteRune('_')
				}
				// Insert underscore between digit → uppercase.
				if prev >= '0' && prev <= '9' {
					buf.WriteRune('_')
				}
				// Insert underscore at acronym boundary: uppercase → uppercase + lowercase
				// e.g., "HTTPRequest": the 'R' is followed by 'e', so insert '_' before 'R'.
				if prev >= 'A' && prev <= 'Z' && i+1 < n && runes[i+1] >= 'a' && runes[i+1] <= 'z' {
					buf.WriteRune('_')
				}
			}
			buf.WriteRune(r + ('a' - 'A'))
		} else {
			buf.WriteRune(r)
		}
	}

	result := buf.String()

	// Replace common separators with underscores.
	replacer := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	result = replacer.Replace(result)

	// Clean up non-snake characters and collapse multiple underscores.
	result = nonSnakePattern.ReplaceAllString(result, "_")
	result = multiUnderscore.ReplaceAllString(result, "_")
	result = strings.Trim(result, "_")

	return result
}
