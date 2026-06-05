package observex

import (
	"errors"
	"time"

	"github.com/ZoneCNH/observex/internal/sanitize"
	"github.com/ZoneCNH/observex/internal/validation"
)

// Config contains the required settings for constructing a Client.
type Config struct {
	// Name identifies the client instance in logs, metrics, and health output.
	Name string
	// Timeout is the caller-defined operation timeout budget.
	Timeout time.Duration
	// Secret is an example sensitive field used to exercise sanitization behavior.
	Secret string
}

// SanitizedConfig is the public, redacted representation of Config.
type SanitizedConfig struct {
	// Name is copied from Config.Name.
	Name string
	// Timeout is copied from Config.Timeout.
	Timeout time.Duration
	// Secret is redacted when present.
	Secret string
}

// Validate checks whether the configuration is usable by New.
func (c Config) Validate() error {
	if err := validation.RequireNonEmpty("name", c.Name); err != nil {
		return validationError("Config.Validate", err.Error(), err)
	}
	if c.Timeout < 0 {
		err := errors.New("timeout must not be negative")
		return validationError("Config.Validate", err.Error(), err)
	}
	return nil
}

// Sanitize returns a copy of Config with sensitive values redacted.
func (c Config) Sanitize() SanitizedConfig {
	return SanitizedConfig{
		Name:    c.Name,
		Timeout: c.Timeout,
		Secret:  sanitize.Secret(c.Secret),
	}
}
