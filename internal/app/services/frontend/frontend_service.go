package frontend

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName = "frontend_service"
)

type (
	// service is responsible for serving HTML (and other static resources).
	service struct {
		logger logging.Logger
		config Config
	}
)

// ProvideService provides the frontend service to dependency injection.
func ProvideService(logger logging.Logger, cfg Config) types.FrontendService {
	svc := &service{
		config: cfg,
		logger: logger.WithName(serviceName),
	}

	return svc
}
