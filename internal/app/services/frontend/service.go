package frontend

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/panicking"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	serviceName string = "frontends_service"
)

type (
	// SearchIndex is a type alias for dependency injection's sake.
	SearchIndex search.IndexManager

	// Service handles to-do list items.
	Service struct {
		logger      logging.Logger
		tracer      tracing.Tracer
		panicker    panicking.Panicker
		authService types.AuthService
		useFakes    bool
	}
)

// ProvideService builds a new ItemsService.
func ProvideService(logger logging.Logger, authService types.AuthService) *Service {
	svc := &Service{
		useFakes:    true,
		logger:      logging.EnsureLogger(logger).WithName(serviceName),
		tracer:      tracing.NewTracer(serviceName),
		panicker:    panicking.NewProductionPanicker(),
		authService: authService,
	}

	return svc
}
