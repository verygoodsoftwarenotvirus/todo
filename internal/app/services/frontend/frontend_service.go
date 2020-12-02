package frontend

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName = "frontend_service"
)

type (
	// Service is responsible for serving HTML (and other static resources).
	Service struct {
		logger         logging.Logger
		logStaticFiles bool
		config         config.FrontendSettings
	}
)

// ProvideService provides the frontend service to dependency injection.
func ProvideService(logger logging.Logger, cfg config.FrontendSettings) *Service {
	svc := &Service{
		config:         cfg,
		logStaticFiles: cfg.LogStaticFiles,
		logger:         logger.WithName(serviceName),
	}

	return svc
}
