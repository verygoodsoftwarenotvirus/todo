package frontend

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

const (
	serviceName = "frontend_service"
)

type (
	// service is responsible for serving HTML (and other static resources).
	service struct {
		logger logging.Logger
		tracer tracing.Tracer
		config Config
	}
)

func buildService(logger logging.Logger, cfg Config) *service {
	return &service{
		config: cfg,
		logger: logging.EnsureLogger(logger).WithName(serviceName),
		tracer: tracing.NewTracer(serviceName),
	}
}

// ProvideService provides the frontend service to dependency injection.
func ProvideService(logger logging.Logger, cfg Config) types.FrontendService {
	return buildService(logger, cfg)
}
