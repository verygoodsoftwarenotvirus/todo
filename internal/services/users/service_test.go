package users

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/chi"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/images"
	mockuploads "gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	expectedUserCount := uint64(123)

	uc := &mockmetrics.UnitCounter{}
	mockDB := database.BuildMockDatabase()
	mockDB.UserDataManager.On(
		"GetAllUsersCount",
		testutil.ContextMatcher,
	).Return(expectedUserCount, nil)

	s := ProvideUsersService(
		&auth.Config{},
		logging.NewNonOperationalLogger(),
		&mocktypes.UserDataManager{},
		&mocktypes.AccountDataManager{},
		&passwords.MockAuthenticator{},
		mockencoding.NewMockEncoderDecoder(),
		func(counterName, description string) metrics.UnitCounter {
			return uc
		},
		&images.MockImageUploadProcessor{},
		&mockuploads.UploadManager{},
		chi.NewRouteParamManager(),
	)

	mock.AssertExpectationsForObjects(t, mockDB, uc)

	return s.(*service)
}

func TestProvideUsersService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.IsType(logging.NewNonOperationalLogger()), UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

		s := ProvideUsersService(
			&auth.Config{},
			logging.NewNonOperationalLogger(),
			&mocktypes.UserDataManager{},
			&mocktypes.AccountDataManager{},
			&passwords.MockAuthenticator{},
			mockencoding.NewMockEncoderDecoder(),
			func(counterName, description string) metrics.UnitCounter {
				return &mockmetrics.UnitCounter{}
			},
			&images.MockImageUploadProcessor{},
			&mockuploads.UploadManager{},
			rpm,
		)

		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}