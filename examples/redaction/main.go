package main

import (
	"fmt"

	"github.com/ZoneCNH/observex/pkg/observex"
)

func main() {
	redactor := observex.NewDefaultRedactor()
	field := redactor.RedactField(observex.Secret("api_key", "raw-value-123"))
	fmt.Println(field.Value)
}
