package dbclient

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"

	"github.com/google/wire"
	"github.com/opentracing/opentracing-go"
)

var (
	// Providers represents what we provide to dependency injectors
	Providers = wire.NewSet(
		ProvideDatabaseClient,
		ProvideTracer,
	)
)

// ProvideTracer provides a tracer
func ProvideTracer() Tracer {
	return tracing.ProvideTracer("database-client")
}

type (
	// Tracer is an opentracing.Tracer alias
	Tracer opentracing.Tracer
)
