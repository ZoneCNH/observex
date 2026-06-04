package observex

import (
	"context"
	"sync"
)

type Client struct {
	cfg         Config
	metrics     Metrics
	logger      Logger
	tracer      Tracer
	mu          sync.Mutex
	initialized bool
	closed      bool
}

func New(ctx context.Context, cfg Config, opts ...Option) (*Client, error) {
	const op = "observex.New"
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	if ctx == nil {
		err := validationError(op, "context is required", nil)
		recordErrorMetric(options.metrics, "new", err)
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		wrapped := contextError(op, err)
		recordErrorMetric(options.metrics, "new", wrapped)
		return nil, wrapped
	}
	if err := cfg.Validate(); err != nil {
		recordErrorMetric(options.metrics, "new", err)
		return nil, err
	}

	ctx, span := options.tracer.Start(ctx, "observex.New", String("name", cfg.Name))
	defer span.End()

	options.metrics.IncCounter(MetricClientCreatedTotal, Labels{"name": cfg.Name})
	options.logger.Info(ctx, "observex client created", String("name", cfg.Name))
	return &Client{
		cfg:         cfg,
		metrics:     options.metrics,
		logger:      options.logger,
		tracer:      options.tracer,
		initialized: true,
	}, nil
}

func (c *Client) Close(ctx context.Context) error {
	const op = "observex.Close"
	if c == nil {
		return validationError(op, "client is nil", nil)
	}
	if ctx == nil {
		err := validationError(op, "context is required", nil)
		recordErrorMetric(c.metrics, "close", err)
		return err
	}
	if err := ctx.Err(); err != nil {
		wrapped := contextError(op, err)
		recordErrorMetric(c.metrics, "close", wrapped)
		return wrapped
	}

	c.mu.Lock()
	if !c.initialized {
		c.mu.Unlock()
		err := validationError(op, "client is not initialized", nil)
		recordErrorMetric(c.metrics, "close", err)
		return err
	}
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	name := c.cfg.Name
	metrics := c.metrics
	logger := c.logger
	tracer := c.tracer
	c.mu.Unlock()

	if tracer == nil {
		tracer = NoopTracer{}
	}
	ctx, span := tracer.Start(ctx, "observex.Close")
	defer span.End()

	if metrics != nil {
		metrics.IncCounter(MetricClientClosedTotal, Labels{"name": name})
	}
	if logger != nil {
		logger.Info(ctx, "observex client closed", String("name", name))
	}
	return nil
}

func recordErrorMetric(metrics Metrics, op string, err error) {
	if metrics == nil {
		return
	}
	metrics.IncCounter(MetricClientErrorsTotal, Labels{
		"op":   op,
		"kind": string(errorKind(err)),
	})
}
