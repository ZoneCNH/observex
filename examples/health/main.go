package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ZoneCNH/observex/pkg/observex"
)

func main() {
	client, err := observex.New(context.Background(), observex.Config{Name: "observex"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create client: %v\n", err)
		return
	}

	status := client.HealthCheck(context.Background())
	fmt.Println(status.Status)
}
