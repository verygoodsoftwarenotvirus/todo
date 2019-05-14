package frontend

import (
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

const (
	serviceName = "frontend_service"
)

type (
	// LoginRoute is a string alias for dependency injection's sake
	LoginRoute string

	// Service is responsible for serving HTML (and relevant resources)
	Service struct {
		logger    logging.Logger
		loginPage []byte
	}
)

// ProvideFrontendService provides the frontend service to dependency injection
func ProvideFrontendService(
	logger logging.Logger,
	loginRoute LoginRoute,
) *Service {
	svc := &Service{
		logger:    logger.WithName(serviceName),
		loginPage: buildLoginPage(loginRoute),
	}
	return svc
}
