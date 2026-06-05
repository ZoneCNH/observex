package testkit

import (
	"time"

	"github.com/ZoneCNH/observex/pkg/observex"
)

// Config returns a valid observex configuration for tests.
func Config(name string) observex.Config {
	return observex.Config{
		Name:    name,
		Timeout: time.Second,
	}
}
