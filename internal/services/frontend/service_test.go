package frontend

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/stretchr/testify/mock"
)

func dummyIDFetcher(*http.Request) uint64 {
	return 0
}

func buildTestService(t *testing.T) *service {
	t.Helper()

	cfg := &Config{}
	logger := logging.NewNoopLogger()
	authService := &mocktypes.AuthService{}
	usersService := &mocktypes.UsersService{}
	dataManager := database.BuildMockDatabase()
	rpm := mockrouting.NewRouteParamManager()

	rpm.On("BuildRouteParamIDFetcher", logger, apiClientIDURLParamKey, "API client").Return(dummyIDFetcher)
	rpm.On("BuildRouteParamIDFetcher", logger, accountIDURLParamKey, "account").Return(dummyIDFetcher)
	rpm.On("BuildRouteParamIDFetcher", logger, webhookIDURLParamKey, "webhook").Return(dummyIDFetcher)
	rpm.On("BuildRouteParamIDFetcher", logger, itemIDURLParamKey, "item").Return(dummyIDFetcher)

	s := ProvideService(
		cfg,
		logger,
		authService,
		usersService,
		dataManager,
		rpm,
		capitalism.NewMockPaymentManager(),
	)

	mock.AssertExpectationsForObjects(t, authService, usersService, dataManager, rpm)

	return s.(*service)
}
