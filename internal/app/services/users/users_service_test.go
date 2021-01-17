package users

import (
	"errors"
	"testing"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"
	mockuploads "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	expectedUserCount := uint64(123)

	uc := &mockmetrics.UnitCounter{}
	mockDB := database.BuildMockDatabase()
	mockDB.UserDataManager.On("GetAllUsersCount", mock.MatchedBy(testutil.ContextMatcher())).Return(expectedUserCount, nil)

	s, err := ProvideUsersService(
		authservice.Config{},
		noop.NewLogger(),
		&mocktypes.UserDataManager{},
		&mocktypes.AccountDataManager{},
		&mocktypes.AuditLogEntryDataManager{},
		&mockauth.Authenticator{},
		&mockencoding.EncoderDecoder{},
		func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return uc, nil
		},
		&images.MockImageUploadProcessor{},
		&mockuploads.UploadManager{},
	)
	require.NoError(t, err)

	mock.AssertExpectationsForObjects(t, mockDB, uc)

	return s.(*service)
}

func TestProvideUsersService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s, err := ProvideUsersService(
			authservice.Config{},
			noop.NewLogger(),
			&mocktypes.UserDataManager{},
			&mocktypes.AccountDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mockauth.Authenticator{},
			&mockencoding.EncoderDecoder{},
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return &mockmetrics.UnitCounter{}, nil
			},
			&images.MockImageUploadProcessor{},
			&mockuploads.UploadManager{},
		)
		assert.NoError(t, err)
		assert.NotNil(t, s)
	})

	T.Run("with error initializing counter", func(t *testing.T) {
		t.Parallel()
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, errors.New("blah")
		}

		s, err := ProvideUsersService(
			authservice.Config{},
			noop.NewLogger(),
			&mocktypes.UserDataManager{},
			&mocktypes.AccountDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mockauth.Authenticator{},
			&mockencoding.EncoderDecoder{},
			ucp,
			&images.MockImageUploadProcessor{},
			&mockuploads.UploadManager{},
		)
		assert.Error(t, err)
		assert.Nil(t, s)
	})
}
