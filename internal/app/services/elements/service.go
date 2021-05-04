package elements

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/panicking"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
)

const (
	serviceName string = "frontends_service"
)

type (
	// SearchIndex is a type alias for dependency injection's sake.
	SearchIndex search.IndexManager

	// Service handles to-do list items.
	Service struct {
		logger   logging.Logger
		tracer   tracing.Tracer
		panicker panicking.Panicker
	}
)

// ProvideService builds a new ItemsService.
func ProvideService(logger logging.Logger) *Service {
	svc := &Service{
		logger:   logging.EnsureLogger(logger).WithName(serviceName),
		tracer:   tracing.NewTracer(serviceName),
		panicker: panicking.NewProductionPanicker(),
	}

	return svc
}
