package users

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	mauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	mmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1/mock"
)

func TestProvideUsersService(T *testing.T) {

	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		mockUserCount := uint64(0)
		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUserCount", mock.Anything, mock.Anything).
			Return(mockUserCount, nil)

		uc := &mmetrics.UnitCounter{}
		uc.On("IncrementBy", mockUserCount).Return()

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, nil
		}

		service, err := ProvideUsersService(
			config.AuthSettings{},
			noop.ProvideNoopLogger(),
			mockDB,
			&mauth.Authenticator{},
			func(req *http.Request) uint64 { return 0 },
			&mencoding.EncoderDecoder{},
			ucp,
			nil,
		)
		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

}
