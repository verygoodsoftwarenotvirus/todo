package tracing

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/expvar"
)

func init() {
	opentracing.SetGlobalTracer(ProvideTracer("_null_"))
}

// ProvideTracer provides a tracer
func ProvideTracer(service string) opentracing.Tracer {
	cfg, err := config.FromEnv()
	if err != nil {
		return opentracing.NoopTracer{}
	}
	cfg.ServiceName = service
	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1

	metricsFactory := expvar.NewFactory(10).Namespace(cfg.ServiceName, nil)
	tracer, _, err := cfg.NewTracer(
		config.Metrics(metricsFactory),
	)
	if err != nil {
		return opentracing.NoopTracer{}
	}

	return tracer
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

	if tracer == nil {
		panic(fmt.Sprintf("nil tracer from %q", operationName))
	}

	if parentCtx != nil {
		span = tracer.StartSpan(operationName, opentracing.ChildOf(parentCtx))
	} else {
		span = tracer.StartSpan(operationName)
	}

	ctx = opentracing.ContextWithSpan(ctx, span) // make the Span current in the context

	return span
}
