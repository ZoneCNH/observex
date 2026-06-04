package testkit

import (
	"strings"
	"testing"
)

func RequireNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func AssertNoSecretLeak(t testing.TB, text string, rawSecrets ...string) {
	t.Helper()
	for _, raw := range rawSecrets {
		if raw == "" {
			continue
		}
		if strings.Contains(text, raw) {
			t.Fatalf("expected text not to contain raw secret %q", raw)
		}
	}

	indicators := []string{
		"password" + "=",
		"passwd" + "=",
		"secret" + "=",
		"token" + "=",
		"access_key" + "=",
		"secret_key" + "=",
	}
	lower := strings.ToLower(text)
	for _, indicator := range indicators {
		if strings.Contains(lower, indicator) {
			t.Fatalf("expected text not to contain secret indicator %q", indicator)
		}
	}
}
