package main

import (
	"context"
	"fmt"

	"github.com/ZoneCNH/observex/pkg/observex"
)

func main() {
	tracer := observex.NewMemoryTracer()
	_, span := tracer.Start(context.Background(), "observex.example")
	span.AddEvent("checkpoint")
	span.End()
	for _, record := range tracer.Spans() {
		fmt.Println("span", record.Name)
		for _, event := range record.Events {
			fmt.Println("event", event.Name)
		}
		if record.Ended {
			fmt.Println("span_end", record.Name)
		}
	}
}
