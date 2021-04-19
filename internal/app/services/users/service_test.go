package users

import (
	"net/http"
	"testing"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/passwords"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/chi"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/mock"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"
	mockuploads "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

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
		&authservice.Config{},
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
			&authservice.Config{},
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
