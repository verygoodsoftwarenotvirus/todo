package oauth2clients

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/go-oauth2/oauth2/v4/manage"
	oauth2server "github.com/go-oauth2/oauth2/v4/server"
	oauth2store "github.com/go-oauth2/oauth2/v4/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	tokenStore, err := oauth2store.NewMemoryTokenStore()
	require.NoError(t, err)

	manager := manage.NewDefaultManager()
	manager.MustTokenStorage(tokenStore, err)

	return &service{
		clientDataManager:    database.BuildMockDatabase(),
		logger:               logging.NewNonOperationalLogger(),
		encoderDecoder:       &mockencoding.EncoderDecoder{},
		authenticator:        &mockauth.Authenticator{},
		urlClientIDExtractor: func(req *http.Request) uint64 { return 0 },
		oauth2ClientCounter:  &mockmetrics.UnitCounter{},
		oauth2Handler:        oauth2server.NewDefaultServer(manager),
		tracer:               tracing.NewTracer(serviceName),
	}
}

func TestProvideOAuth2ClientsService(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		mockOAuth2ClientDataManager := &mocktypes.OAuth2ClientDataManager{}

		s, err := ProvideOAuth2ClientsService(
			logging.NewNonOperationalLogger(),
			mockOAuth2ClientDataManager,
			&mocktypes.UserDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mockauth.Authenticator{},
			&mockencoding.EncoderDecoder{},
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return nil, nil
			},
		)
		assert.NoError(t, err)
		assert.NotNil(t, s)

		mock.AssertExpectationsForObjects(t, mockOAuth2ClientDataManager)
	})

	T.Run("with error providing counter", func(t *testing.T) {
		t.Parallel()
		mockOAuth2ClientDataManager := &mocktypes.OAuth2ClientDataManager{}

		s, err := ProvideOAuth2ClientsService(
			logging.NewNonOperationalLogger(),
			mockOAuth2ClientDataManager,
			&mocktypes.UserDataManager{},
			&mocktypes.AuditLogEntryDataManager{},
			&mockauth.Authenticator{},
			&mockencoding.EncoderDecoder{},
			func(counterName metrics.CounterName, description string) (metrics.UnitCounter, error) {
				return nil, errors.New("blah")
			},
		)

		assert.Error(t, err)
		assert.Nil(t, s)

		mock.AssertExpectationsForObjects(t, mockOAuth2ClientDataManager)
	})
}

func Test_clientStore_GetByID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleOAuth2Client.ClientID,
		).Return(exampleOAuth2Client, nil)

		c := &clientStore{
			dataManager: mockDB,
			tracer:      tracing.NewTracer("test"),
		}
		actual, err := c.GetByID(ctx, exampleOAuth2Client.ClientID)

		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client.ClientID, actual.GetID())

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with no rows", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleID := "blah"

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleID,
		).Return((*types.OAuth2Client)(nil), sql.ErrNoRows)

		c := &clientStore{
			dataManager: mockDB,
			tracer:      tracing.NewTracer("test"),
		}
		_, err := c.GetByID(ctx, exampleID)

		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error reading from clientDataManager", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleID := "blah"

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleID,
		).Return((*types.OAuth2Client)(nil), errors.New(exampleID))

		c := &clientStore{
			dataManager: mockDB,
			tracer:      tracing.NewTracer("test"),
		}
		_, err := c.GetByID(ctx, exampleID)

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
