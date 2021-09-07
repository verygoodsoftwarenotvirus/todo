package frontend

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func dummyIDFetcher(*http.Request) string {
	return ""
}

func TestProvideService(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	logger := logging.NewNoopLogger()
	authService := &mocktypes.AuthService{}
	usersService := &mocktypes.UsersService{}
	dataManager := database.BuildMockDatabase()

	rpm := mockrouting.NewRouteParamManager()
	rpm.On("BuildRouteParamStringIDFetcher", apiClientIDURLParamKey).Return(dummyIDFetcher)
	rpm.On("BuildRouteParamStringIDFetcher", accountIDURLParamKey).Return(dummyIDFetcher)
	rpm.On("BuildRouteParamStringIDFetcher", webhookIDURLParamKey).Return(dummyIDFetcher)
	rpm.On("BuildRouteParamStringIDFetcher", itemIDURLParamKey).Return(dummyIDFetcher)

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
	assert.NotNil(t, s)
}
