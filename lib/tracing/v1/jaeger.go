package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/expvar"
)

// ProvideTracer provides a tracer
func ProvideTracer(service string) (opentracing.Tracer, error) {
	cfg, err := config.FromEnv()
	if err != nil {
		// return nil, err
		return opentracing.NoopTracer{}, nil
	}
	cfg.ServiceName = service
	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1

	metricsFactory := expvar.NewFactory(10).Namespace(cfg.ServiceName, nil)
	tracer, _, err := cfg.NewTracer(
		config.Metrics(metricsFactory),
	)
	if err != nil {
		// return nil, err
		return opentracing.NoopTracer{}, nil
	}

	return tracer, nil
}

// ProvideNoopTracer provides a noop tracer
func ProvideNoopTracer() opentracing.Tracer {
	return &opentracing.NoopTracer{}
}

// FetchSpanFromContext extracts a span from a context
func FetchSpanFromContext(ctx context.Context, tracer opentracing.Tracer, operationName string) (span opentracing.Span) {
	var parentCtx opentracing.SpanContext
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		parentCtx = parentSpan.Context()
	}

	if parentCtx != nil {
		span = tracer.StartSpan(operationName, opentracing.ChildOf(parentCtx))
	} else {
		span = tracer.StartSpan(operationName)
	}

	ctx = opentracing.ContextWithSpan(ctx, span) // make the Span current in the context

	return span
}
