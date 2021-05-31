package frontend

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"

	"github.com/stretchr/testify/mock"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	cfg := &Config{}
	logger := logging.NewNoopLogger()
	authService := &mocktypes.AuthService{}
	usersService := &mocktypes.UsersService{}
	dataManager := database.BuildMockDatabase()
	rpm := mockrouting.NewRouteParamManager()

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
