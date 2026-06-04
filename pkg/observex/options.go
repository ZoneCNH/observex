package observex

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

func WithMetrics(metrics Metrics) Option {
	return func(o *options) {
		if metrics != nil {
			o.metrics = metrics
		}
	}
}

func WithLogger(logger Logger) Option {
	return func(o *options) {
		if logger != nil {
			o.logger = logger
		}
	}
}

func WithTracer(tracer Tracer) Option {
	return func(o *options) {
		if tracer != nil {
			o.tracer = tracer
		}
	}
}
