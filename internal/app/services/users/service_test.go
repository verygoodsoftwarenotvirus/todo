package users

import (
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
	mockDB.UserDataManager.On("GetAllUsersCount", mock.MatchedBy(testutil.ContextMatcher)).Return(expectedUserCount, nil)

	rpm := mockrouting.NewRouteParamManager()
	rpm.On("BuildRouteParamIDFetcher", mock.Anything, UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

	s := ProvideUsersService(
		&authservice.Config{},
		logging.NewNonOperationalLogger(),
		&mocktypes.UserDataManager{},
		&mocktypes.AccountDataManager{},
		&mockauth.Authenticator{},
		mockencoding.NewMockEncoderDecoder(),
		func(counterName, description string) metrics.UnitCounter {
			return uc
		},
		&images.MockImageUploadProcessor{},
		&mockuploads.UploadManager{},
		rpm,
	)

	mock.AssertExpectationsForObjects(t, mockDB, uc, rpm)

	return s.(*service)
}

func TestProvideUsersService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On("BuildRouteParamIDFetcher", mock.Anything, UserIDURIParamKey, "user").Return(func(*http.Request) uint64 { return 0 })

		s := ProvideUsersService(
			&authservice.Config{},
			logging.NewNonOperationalLogger(),
			&mocktypes.UserDataManager{},
			&mocktypes.AccountDataManager{},
			&mockauth.Authenticator{},
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
