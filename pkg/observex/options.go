package observex

// Option customizes Client construction.
type Option func(*options)

type options struct {
	metrics Metrics
	logger  Logger
	tracer  Tracer
}

func defaultOptions() options {
	return options{
		metrics: NoopMetrics{},
		logger:  NoopLogger{},
		tracer:  NoopTracer{},
	}
}

// WithMetrics configures the metrics recorder used by a Client.
func WithMetrics(metrics Metrics) Option {
	return func(o *options) {
		if metrics != nil {
			o.metrics = metrics
		}
	}
}

// WithLogger configures the logger used by a Client.
func WithLogger(logger Logger) Option {
	return func(o *options) {
		if logger != nil {
			o.logger = logger
		}
	}
}

// WithTracer configures the tracer used by a Client.
func WithTracer(tracer Tracer) Option {
	return func(o *options) {
		if tracer != nil {
			o.tracer = tracer
		}
	}
}
