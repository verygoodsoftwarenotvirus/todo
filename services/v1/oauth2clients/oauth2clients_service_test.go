package oauth2clients

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	mauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1/mock"
	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	mmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"

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

func TestProvideOAuth2ClientsService(t *testing.T) {
	t.Helper()

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
		noop.ProvideNoopLogger(),
		mockDB,
		&mauth.Authenticator{},
		func(req *http.Request) uint64 { return 0 },
		&mencoding.EncoderDecoder{},
		ucp,
	)
	assert.NoError(t, err)
	assert.NotNil(t, service)
}
