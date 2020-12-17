package frontend

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName = "frontend_service"
)

type (
	// service is responsible for serving HTML (and other static resources).
	service struct {
		logger         logging.Logger
		logStaticFiles bool
		config         config.FrontendSettings
	}
)

// ProvideService provides the frontend service to dependency injection.
func ProvideService(logger logging.Logger, cfg config.FrontendSettings) types.FrontendService {
	svc := &service{
		config:         cfg,
		logStaticFiles: cfg.LogStaticFiles,
		logger:         logger.WithName(serviceName),
	}

	return svc
}
