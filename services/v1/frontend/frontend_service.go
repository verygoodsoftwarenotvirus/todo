package frontend

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
)

const (
	serviceName = "frontend_service"
)

type (
	// Service is responsible for serving HTML (and relevant resources)
	Service struct {
		logger logging.Logger
		config config.FrontendSettings
	}
)

// ProvideFrontendService provides the frontend service to dependency injection
func ProvideFrontendService(logger logging.Logger, cfg config.FrontendSettings) *Service {
	svc := &Service{
		config: cfg,
		logger: logger.WithName(serviceName),
	}
	return svc
}
