package oauth2clients

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
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
	// CreationMiddlewareCtxKey is a string alias for referring to OAuth2 client creation data
	CreationMiddlewareCtxKey models.ContextKey = "create_oauth2_client"

	counterName        metrics.CounterName = "oauth2_clients"
	counterDescription string              = "number of oauth2 clients managed by the oauth2 client service"
	serviceName        string              = "oauth2_clients_service"
)

var (
	_ models.OAuth2ClientDataServer = (*Service)(nil)
	_ oauth2.ClientStore            = (*clientStore)(nil)
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

	clientStore struct {
		database database.Database
	}
)

func newClientStore(db database.Database) *clientStore {
	cs := &clientStore{
		database: db,
	}
	return cs
}

// GetByID implements oauth2.ClientStorage
func (s *clientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	client, err := s.database.GetOAuth2ClientByClientID(context.Background(), id)

	if err == sql.ErrNoRows {
		return nil, errors.New("invalid client")
	} else if err != nil {
		return nil, fmt.Errorf("querying for client: %w", err)
	}

	return client, nil
}

// ProvideOAuth2ClientsService builds a new OAuth2ClientsService
func ProvideOAuth2ClientsService(
	ctx context.Context,
	logger logging.Logger,
	db database.Database,
	authenticator auth.Authenticator,
	clientIDFetcher ClientIDFetcher,
	encoderDecoder encoding.EncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
) (*Service, error) {
	counter, err := counterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	manager := manage.NewDefaultManager()
	clientStore := newClientStore(db)
	tokenStore, err := oauth2store.NewMemoryTokenStore()
	manager.MapClientStorage(clientStore)
	manager.MustTokenStorage(tokenStore, err)
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)
	oHandler := oauth2server.NewDefaultServer(manager)
	oHandler.SetAllowGetAccessRequest(true)

	s := &Service{
		database:             db,
		logger:               logger.WithName(serviceName),
		encoderDecoder:       encoderDecoder,
		authenticator:        authenticator,
		urlClientIDExtractor: clientIDFetcher,
		oauth2ClientCounter:  counter,
		tokenStore:           tokenStore,
		oauth2Handler:        oHandler,
	}

	initializeOAuth2Handler(s.oauth2Handler, s)
	count, err := s.database.GetAllOAuth2ClientCount(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("fetching oauth2 clients: %w", err)
	}
	counter.IncrementBy(ctx, count)

	return s, nil
}

// initializeOAuth2Handler
func initializeOAuth2Handler(handler oauth2Handler, s *Service) {
	handler.SetAllowGetAccessRequest(true)
	handler.SetClientAuthorizedHandler(s.ClientAuthorizedHandler)
	handler.SetClientScopeHandler(s.ClientScopeHandler)
	handler.SetClientInfoHandler(oauth2server.ClientFormHandler)
	handler.SetAuthorizeScopeHandler(s.AuthorizeScopeHandler)
	handler.SetResponseErrorHandler(s.OAuth2ResponseErrorHandler)
	handler.SetInternalErrorHandler(s.OAuth2InternalErrorHandler)
	handler.SetUserAuthorizationHandler(s.UserAuthorizationHandler)

	// this sad type cast is here because I have an arbitrary
	// test-only interface for OAuth2 interactions.
	if x, ok := handler.(*oauth2server.Server); ok {
		x.Config.AllowedGrantTypes = []oauth2.GrantType{
			oauth2.ClientCredentials,
			// oauth2.AuthorizationCode,
			// oauth2.Refreshing,
			// oauth2.Implicit,
		}
	}
}

// HandleAuthorizeRequest is a simple wrapper around the internal server's HandleAuthorizeRequest
func (s *Service) HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error {
	return s.oauth2Handler.HandleAuthorizeRequest(res, req)
}

// HandleTokenRequest is a simple wrapper around the internal server's HandleTokenRequest
func (s *Service) HandleTokenRequest(res http.ResponseWriter, req *http.Request) error {
	return s.oauth2Handler.HandleTokenRequest(res, req)
}
