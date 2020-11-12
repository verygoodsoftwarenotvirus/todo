package oauth2clients

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/auth/mock"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
	"gopkg.in/oauth2.v3/manage"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

func buildTestService(t *testing.T) *Service {
	t.Helper()

	tokenStore, err := oauth2store.NewMemoryTokenStore()
	require.NoError(t, err)
	manager := manage.NewDefaultManager()
	manager.MustTokenStorage(tokenStore, err)

	service := &Service{
		clientDataManager:    database.BuildMockDatabase(),
		logger:               noop.NewLogger(),
		encoderDecoder:       &mockencoding.EncoderDecoder{},
		authenticator:        &mockauth.Authenticator{},
		urlClientIDExtractor: func(req *http.Request) uint64 { return 0 },
		oauth2ClientCounter:  &mockmetrics.UnitCounter{},
		oauth2Handler:        oauth2server.NewDefaultServer(manager),
	}

	return service
}

func TestProvideOAuth2ClientsService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		mockOAuth2ClientDataManager := &mockmodels.OAuth2ClientDataManager{}

		service, err := ProvideOAuth2ClientsService(
			noop.NewLogger(),
			mockOAuth2ClientDataManager,
			&mockmodels.UserDataManager{},
			&mockmodels.AuditLogDataManager{},
			&mockauth.Authenticator{},
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return nil, nil
			},
		)
		assert.NoError(t, err)
		assert.NotNil(t, service)

		mock.AssertExpectationsForObjects(t, mockOAuth2ClientDataManager)
	})

	T.Run("with error providing counter", func(t *testing.T) {
		t.Parallel()
		mockOAuth2ClientDataManager := &mockmodels.OAuth2ClientDataManager{}

		service, err := ProvideOAuth2ClientsService(
			noop.NewLogger(),
			mockOAuth2ClientDataManager,
			&mockmodels.UserDataManager{},
			&mockmodels.AuditLogDataManager{},
			&mockauth.Authenticator{},
			func(req *http.Request) uint64 { return 0 },
			&mockencoding.EncoderDecoder{},
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return nil, errors.New("blah")
			},
		)

		assert.Error(t, err)
		assert.Nil(t, service)

		mock.AssertExpectationsForObjects(t, mockOAuth2ClientDataManager)
	})
}

func Test_clientStore_GetByID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleOAuth2Client.ClientID,
		).Return(exampleOAuth2Client, nil)

		c := &clientStore{dataManager: mockDB}
		actual, err := c.GetByID(exampleOAuth2Client.ClientID)

		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client.ClientID, actual.GetID())

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with no rows", func(t *testing.T) {
		t.Parallel()
		exampleID := "blah"

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleID,
		).Return((*types.OAuth2Client)(nil), sql.ErrNoRows)

		c := &clientStore{dataManager: mockDB}
		_, err := c.GetByID(exampleID)

		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error reading from clientDataManager", func(t *testing.T) {
		t.Parallel()
		exampleID := "blah"

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleID,
		).Return((*types.OAuth2Client)(nil), errors.New(exampleID))

		c := &clientStore{dataManager: mockDB}
		_, err := c.GetByID(exampleID)

		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_HandleAuthorizeRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		req, res := buildRequest(t), httptest.NewRecorder()

		moah := &mockOAuth2Handler{}
		moah.On(
			"HandleAuthorizeRequest",
			mock.Anything,
			mock.Anything,
		).Return(nil)
		s.oauth2Handler = moah

		assert.NoError(t, s.HandleAuthorizeRequest(res, req))

		mock.AssertExpectationsForObjects(t, moah)
	})
}

func TestService_HandleTokenRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		req, res := buildRequest(t), httptest.NewRecorder()

		moah := &mockOAuth2Handler{}
		moah.On(
			"HandleTokenRequest",
			mock.Anything,
			mock.Anything,
		).Return(nil)
		s.oauth2Handler = moah

		assert.NoError(t, s.HandleTokenRequest(res, req))

		mock.AssertExpectationsForObjects(t, moah)
	})
}
