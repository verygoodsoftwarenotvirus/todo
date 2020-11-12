package users

import (
	"errors"

	"net/http"
	"testing"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func buildTestService(t *testing.T) *Service {
	t.Helper()

	expectedUserCount := uint64(123)

	uc := &mockmetrics.UnitCounter{}
	mockDB := database.BuildMockDatabase()
	mockDB.UserDataManager.On("GetAllUsersCount", mock.Anything).Return(expectedUserCount, nil)

	service, err := ProvideUsersService(
		config.AuthSettings{},
		noop.NewLogger(),
		&mockmodels.UserDataManager{},
		&mockmodels.AuditLogDataManager{},
		&mockauth.Authenticator{},
		func(req *http.Request) uint64 { return 0 },
		func(req *http.Request) (*models.SessionInfo, error) { return nil, nil },
		&mockencoding.EncoderDecoder{},
		func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return uc, nil
		},
	)
	require.NoError(t, err)

	mock.AssertExpectationsForObjects(t, mockDB, uc)

	return service
}

func TestProvideUsersService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		service, err := ProvideUsersService(
			config.AuthSettings{},
			noop.NewLogger(),
			&mockmodels.UserDataManager{},
			&mockmodels.AuditLogDataManager{},
			&mockauth.Authenticator{},
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) (*models.SessionInfo, error) { return nil, nil },
			&mockencoding.EncoderDecoder{},
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return &mockmetrics.UnitCounter{}, nil
			},
		)
		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

	T.Run("with nil userIDFetcher", func(t *testing.T) {
		t.Parallel()
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, nil
		}

		service, err := ProvideUsersService(
			config.AuthSettings{},
			noop.NewLogger(),
			&mockmodels.UserDataManager{},
			&mockmodels.AuditLogDataManager{},
			&mockauth.Authenticator{},
			nil,
			nil,
			&mockencoding.EncoderDecoder{},
			ucp,
		)
		assert.Error(t, err)
		assert.Nil(t, service)
	})

	T.Run("with error initializing counter", func(t *testing.T) {
		t.Parallel()
		var ucp metrics.UnitCounterProvider = func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return &mockmetrics.UnitCounter{}, errors.New("blah")
		}

		service, err := ProvideUsersService(
			config.AuthSettings{},
			noop.NewLogger(),
			&mockmodels.UserDataManager{},
			&mockmodels.AuditLogDataManager{},
			&mockauth.Authenticator{},
			func(req *http.Request) uint64 { return 0 },
			func(req *http.Request) (*models.SessionInfo, error) { return nil, nil },
			&mockencoding.EncoderDecoder{},
			ucp,
		)
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}
