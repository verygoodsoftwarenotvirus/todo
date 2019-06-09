package oauth2clients

import (
	"context"
	"crypto/rand"
	"database/sql"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"

	"github.com/pkg/errors"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/manage"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

const (
	// MiddlewareCtxKey is a string alias for referring to OAuth2 clients in contexts
	MiddlewareCtxKey models.ContextKey   = "oauth2_client"
	counterName      metrics.CounterName = "oauth2_clients"
	serviceName                          = "oauth2_clients_service"
)

type (
	oauth2Handler interface {
		SetAllowGetAccessRequest(bool)
		SetClientAuthorizedHandler(handler oauth2server.ClientAuthorizedHandler)
		SetClientScopeHandler(handler oauth2server.ClientScopeHandler)
		SetClientInfoHandler(handler oauth2server.ClientInfoHandler)
		SetUserAuthorizationHandler(handler oauth2server.UserAuthorizationHandler)
		SetAuthorizeScopeHandler(handler oauth2server.AuthorizeScopeHandler)
		SetResponseErrorHandler(handler oauth2server.ResponseErrorHandler)
		SetInternalErrorHandler(handler oauth2server.InternalErrorHandler)
		ValidationBearerToken(*http.Request) (oauth2.TokenInfo, error)
		HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error
		HandleTokenRequest(res http.ResponseWriter, req *http.Request) error
	}

	// ClientIDFetcher is a function for fetching client IDs out of requests
	ClientIDFetcher func(req *http.Request) uint64

	// Service manages our OAuth2 clients via HTTP
	Service struct {
		logger               logging.Logger
		database             database.Database
		authenticator        auth.Authenticator
		encoderDecoder       encoding.EncoderDecoder
		urlClientIDExtractor func(req *http.Request) uint64

		tokenStore          oauth2.TokenStore
		oauth2Handler       oauth2Handler
		oauth2ClientCounter metrics.UnitCounter
	}
)

var _ oauth2.ClientStore = (*clientStore)(nil)

type clientStore struct {
	database database.Database
}

// according to the ID for the client information
func (s *clientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	client, err := s.database.GetOAuth2ClientByClientID(context.Background(), id)

	if err == sql.ErrNoRows {
		return nil, errors.New("invalid client")
	} else if err != nil {
		return nil, err
	}

	return client, nil
}

func newClientStore(database database.Database) *clientStore {
	cs := &clientStore{
		database: database,
	}
	return cs
}

// ProvideOAuth2ClientsService builds a new OAuth2ClientsService
func ProvideOAuth2ClientsService(
	ctx context.Context,
	logger logging.Logger,
	database database.Database,
	authenticator auth.Authenticator,
	clientIDFetcher ClientIDFetcher,
	encoderDecoder encoding.EncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
) (*Service, error) {
	counter, err := counterProvider(counterName, "number of oauth2 clients managed by the oauth2 client service")
	if err != nil {
		return nil, errors.Wrap(err, "error initializing counter")
	}

	manager := manage.NewDefaultManager()
	clientStore := newClientStore(database)
	tokenStore, err := oauth2store.NewMemoryTokenStore()
	manager.MapClientStorage(clientStore)
	manager.MustTokenStorage(tokenStore, err)
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)
	oHandler := oauth2server.NewDefaultServer(manager)
	oHandler.SetAllowGetAccessRequest(true)

	s := &Service{
		database:             database,
		logger:               logger.WithName(serviceName),
		encoderDecoder:       encoderDecoder,
		authenticator:        authenticator,
		urlClientIDExtractor: clientIDFetcher,
		oauth2ClientCounter:  counter,
		tokenStore:           tokenStore,
		oauth2Handler:        oHandler,
	}

	clients, err := s.database.GetAllOAuth2Clients(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "fetching oauth2 clients")
	}
	counter.IncrementBy(ctx, uint64(len(clients)))

	s.initializeOAuth2Handler()

	return s, nil
}

func (s *Service) initializeOAuth2Handler() {
	s.oauth2Handler.SetAllowGetAccessRequest(true)
	s.oauth2Handler.SetClientAuthorizedHandler(s.ClientAuthorizedHandler)
	s.oauth2Handler.SetClientScopeHandler(s.ClientScopeHandler)
	s.oauth2Handler.SetClientInfoHandler(oauth2server.ClientFormHandler)
	s.oauth2Handler.SetAuthorizeScopeHandler(s.AuthorizeScopeHandler)
	s.oauth2Handler.SetResponseErrorHandler(s.OAuth2ResponseErrorHandler)
	s.oauth2Handler.SetInternalErrorHandler(s.OAuth2InternalErrorHandler)
	s.oauth2Handler.SetUserAuthorizationHandler(s.UserAuthorizationHandler)

	// this sad type cast is here because I have an arbitrary
	// test-only interface for OAuth2 interactions.
	if x, ok := s.oauth2Handler.(*oauth2server.Server); ok {
		x.Config.AllowedGrantTypes = []oauth2.GrantType{
			// oauth2.AuthorizationCode,
			oauth2.ClientCredentials,
			// oauth2.Refreshing,
			// oauth2.Implicit
		}
	}
}

// HandleAuthorizeRequest is a simple wrapper around the internal server's HandleAuthorizeRequest
func (s *Service) HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error {
	s.logger.Debug("HandleAuthorizeRequest called")
	err := s.oauth2Handler.HandleAuthorizeRequest(res, req)
	return err
}

// HandleTokenRequest is a simple wrapper around the internal server's HandleTokenRequest
func (s *Service) HandleTokenRequest(res http.ResponseWriter, req *http.Request) error {
	s.logger.Debug("HandleTokenRequest called")
	err := s.oauth2Handler.HandleTokenRequest(res, req)
	return err
}
