package main

import (
	"fmt"
	"time"

	"github.com/ZoneCNH/observex/pkg/observex"
)

func main() {
	cfg := observex.Config{
		Name:    "observex",
		Timeout: time.Second,
		Secret:  "example",
	}

	fmt.Println(cfg.Sanitize().Secret)
}
