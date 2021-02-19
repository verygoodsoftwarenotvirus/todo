package users

import (
	"errors"
	"net/http"
	"testing"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"
	mockuploads "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	expectedUserCount := uint64(123)

	uc := &mockmetrics.UnitCounter{}
	mockDB := database.BuildMockDatabase()
	mockDB.UserDataManager.On("GetAllUsersCount", mock.MatchedBy(testutil.ContextMatcher)).Return(expectedUserCount, nil)

	rpm := mockrouting.NewRouteParamManager()
	rpm.On("BuildRouteParamIDFetcher", mock.Anything, UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

	s, err := ProvideUsersService(
		&authservice.Config{},
		logging.NewNonOperationalLogger(),
		&mocktypes.UserDataManager{},
		&mocktypes.AccountDataManager{},
		&mockauth.Authenticator{},
		mockencoding.NewMockEncoderDecoder(),
		func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return uc, nil
		},
		&images.MockImageUploadProcessor{},
		&mockuploads.UploadManager{},
		rpm,
	)
	require.NoError(t, err)

	mock.AssertExpectationsForObjects(t, mockDB, uc, rpm)

	return s.(*service)
}

func TestProvideUsersService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On("BuildRouteParamIDFetcher", mock.Anything, UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

		s, err := ProvideUsersService(
			&authservice.Config{},
			logging.NewNonOperationalLogger(),
			&mocktypes.UserDataManager{},
			&mocktypes.AccountDataManager{},
			&mockauth.Authenticator{},
			mockencoding.NewMockEncoderDecoder(),
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return &mockmetrics.UnitCounter{}, nil
			},
			&images.MockImageUploadProcessor{},
			&mockuploads.UploadManager{},
			rpm,
		)

		assert.NoError(t, err)
		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, rpm)
	})

	T.Run("with error initializing counter", func(t *testing.T) {
		t.Parallel()

		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, errors.New("blah")
		}

		rpm := mockrouting.NewRouteParamManager()

		s, err := ProvideUsersService(
			&authservice.Config{},
			logging.NewNonOperationalLogger(),
			&mocktypes.UserDataManager{},
			&mocktypes.AccountDataManager{},
			&mockauth.Authenticator{},
			mockencoding.NewMockEncoderDecoder(),
			ucp,
			&images.MockImageUploadProcessor{},
			&mockuploads.UploadManager{},
			rpm,
		)

		assert.Error(t, err)
		assert.Nil(t, s)

		mock.AssertExpectationsForObjects(t, rpm)
	})
}
