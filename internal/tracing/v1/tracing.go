package tracing

import (
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
