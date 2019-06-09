package oauth2clients

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	mauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1/mock"
	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/noop"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	mmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/oauth2.v3/manage"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

func buildTestService(t *testing.T) *Service {
	t.Helper()

	manager := manage.NewDefaultManager()
	tokenStore, err := oauth2store.NewMemoryTokenStore()
	require.NoError(t, err)
	manager.MustTokenStorage(tokenStore, err)
	server := oauth2server.NewDefaultServer(manager)

	service := &Service{
		database:             database.BuildMockDatabase(),
		logger:               noop.ProvideNoopLogger(),
		encoderDecoder:       &mencoding.EncoderDecoder{},
		authenticator:        &mauth.Authenticator{},
		urlClientIDExtractor: func(req *http.Request) uint64 { return 0 },

		oauth2ClientCounter: &mmetrics.UnitCounter{},
		tokenStore:          tokenStore,
		oauth2Handler:       server,
	}

	return service
}

func TestProvideOAuth2ClientsService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2Clients", mock.Anything).
			Return([]*models.OAuth2Client{}, nil)
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2ClientCount", mock.Anything).
			Return(0)

		uc := &mmetrics.UnitCounter{}
		uc.On("IncrementBy", uint64(0)).Return()

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, nil
		}

		service, err := ProvideOAuth2ClientsService(
			context.Background(),
			noop.ProvideNoopLogger(),
			mockDB,
			&mauth.Authenticator{},
			func(req *http.Request) uint64 { return 0 },
			&mencoding.EncoderDecoder{},
			ucp,
		)
		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

	T.Run("with error providing counter", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2Clients", mock.Anything).
			Return([]*models.OAuth2Client{}, nil)
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2ClientCount", mock.Anything).
			Return(0)

		uc := &mmetrics.UnitCounter{}
		uc.On("IncrementBy", uint64(0)).Return()

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return nil, errors.New("blah")
		}

		service, err := ProvideOAuth2ClientsService(
			context.Background(),
			noop.ProvideNoopLogger(),
			mockDB,
			&mauth.Authenticator{},
			func(req *http.Request) uint64 { return 0 },
			&mencoding.EncoderDecoder{},
			ucp,
		)
		assert.Error(t, err)
		assert.Nil(t, service)
	})

	T.Run("with error fetching oauth2 clients", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2Clients", mock.Anything).
			Return([]*models.OAuth2Client{}, errors.New("blah"))
		mockDB.OAuth2ClientDataManager.On("GetAllOAuth2ClientCount", mock.Anything).
			Return(0)

		uc := &mmetrics.UnitCounter{}
		uc.On("IncrementBy", uint64(0)).Return()

		var ucp metrics.UnitCounterProvider = func(
			counterName metrics.CounterName,
			description string,
		) (metrics.UnitCounter, error) {
			return uc, nil
		}

		service, err := ProvideOAuth2ClientsService(
			context.Background(),
			noop.ProvideNoopLogger(),
			mockDB,
			&mauth.Authenticator{},
			func(req *http.Request) uint64 { return 0 },
			&mencoding.EncoderDecoder{},
			ucp,
		)
		assert.Error(t, err)
		assert.Nil(t, service)
	})

}

func Test_clientStore_GetByID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleID := "blah"

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.
			On("GetOAuth2ClientByClientID", mock.Anything, exampleID).
			Return(&models.OAuth2Client{ClientID: exampleID}, nil)

		c := &clientStore{database: mockDB}

		actual, err := c.GetByID(exampleID)
		assert.NoError(t, err)
		assert.Equal(t, exampleID, actual.GetID())
	})

	T.Run("with no rows", func(t *testing.T) {
		exampleID := "blah"

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.
			On("GetOAuth2ClientByClientID", mock.Anything, exampleID).
			Return((*models.OAuth2Client)(nil), sql.ErrNoRows)

		c := &clientStore{database: mockDB}

		_, err := c.GetByID(exampleID)
		assert.Error(t, err)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		exampleID := "blah"

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.
			On("GetOAuth2ClientByClientID", mock.Anything, exampleID).
			Return((*models.OAuth2Client)(nil), errors.New(exampleID))

		c := &clientStore{database: mockDB}

		_, err := c.GetByID(exampleID)
		assert.Error(t, err)
	})
}

func TestService_HandleAuthorizeRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		moah := &mockOauth2Handler{}
		moah.On("HandleAuthorizeRequest", mock.Anything, mock.Anything).
			Return(nil)
		s.oauth2Handler = moah

		req, res := buildRequest(t), httptest.NewRecorder()

		assert.NoError(t, s.HandleAuthorizeRequest(res, req))
	})
}

func TestService_HandleTokenRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		moah := &mockOauth2Handler{}
		moah.On("HandleTokenRequest", mock.Anything, mock.Anything).
			Return(nil)
		s.oauth2Handler = moah

		req, res := buildRequest(t), httptest.NewRecorder()

		assert.NoError(t, s.HandleTokenRequest(res, req))
	})
}
