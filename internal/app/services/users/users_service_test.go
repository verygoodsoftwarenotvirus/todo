package users

import (
	"errors"

	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics/mock"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

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
	mockDB.UserDataManager.On("GetAllUsersCount", mock.Anything).Return(expectedUserCount, nil)

	s, err := ProvideUsersService(
		config.AuthSettings{},
		noop.NewLogger(),
		&mockmodels.UserDataManager{},
		&mockmodels.AuditLogDataManager{},
		&mockauth.Authenticator{},
		func(req *http.Request) uint64 { return 0 },
		func(req *http.Request) (*types.SessionInfo, error) { return nil, nil },
		&mockencoding.EncoderDecoder{},
		func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
			return uc, nil
		},
	)
	require.NoError(t, err)

	mock.AssertExpectationsForObjects(t, mockDB, uc)

	return s.(*service)
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
			func(req *http.Request) (*types.SessionInfo, error) { return nil, nil },
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
			func(req *http.Request) (*types.SessionInfo, error) { return nil, nil },
			&mockencoding.EncoderDecoder{},
			ucp,
		)
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}
