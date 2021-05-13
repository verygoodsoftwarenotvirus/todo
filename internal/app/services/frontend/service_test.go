package frontend

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/mock"
)

func buildTestService(t *testing.T) *Service {
	t.Helper()

	cfg := &Config{}
	logger := logging.NewNonOperationalLogger()
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
	)

	mock.AssertExpectationsForObjects(t, authService, usersService, dataManager, rpm)

	return s
}
