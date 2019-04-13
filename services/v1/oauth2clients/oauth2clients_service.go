package oauth2clients

import (
	"context"
	"database/sql"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/manage"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

const (
	// MiddlewareCtxKey is a string alias for referring to OAuth2 clients in contexts
	MiddlewareCtxKey models.ContextKey = "oauth2_client"

	serviceName = "oauth2_clients_service"
)

type (
	// Tracer is a type alias we use for dependency injection
	Tracer opentracing.Tracer

	// ClientIDFetcher is a function for fetching client IDs out of requests
	ClientIDFetcher func(req *http.Request) uint64

	// Service manages our OAuth2 clients via HTTP
	Service struct {
		database             database.Database
		authenticator        auth.Enticator
		logger               logging.Logger
		tracer               opentracing.Tracer
		encoder              encoding.ServerEncoderDecoder
		urlClientIDExtractor func(req *http.Request) uint64

		tokenStore        oauth2.TokenStore
		oauth2Handler     *oauth2server.Server
		oauth2ClientStore *oauth2store.ClientStore
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
	logger logging.Logger,
	database database.Database,
	authenticator auth.Enticator,
	clientIDFetcher ClientIDFetcher,
	encoder encoding.ServerEncoderDecoder,
) *Service {

	manager := manage.NewDefaultManager()
	clientStore := newClientStore(database)
	manager.MapClientStorage(clientStore)
	tokenStore, err := oauth2store.NewMemoryTokenStore()
	manager.MustTokenStorage(tokenStore, err)
	server := oauth2server.NewDefaultServer(manager)

	s := &Service{
		database:             database,
		logger:               logger.WithName(serviceName),
		tracer:               tracing.ProvideTracer(serviceName),
		encoder:              encoder,
		authenticator:        authenticator,
		urlClientIDExtractor: clientIDFetcher,

		tokenStore:    tokenStore,
		oauth2Handler: server,
	}

	s.oauth2Handler.SetAllowGetAccessRequest(true)
	s.oauth2Handler.SetClientAuthorizedHandler(s.ClientAuthorizedHandler)
	s.oauth2Handler.SetClientScopeHandler(s.ClientScopeHandler)
	s.oauth2Handler.SetClientInfoHandler(oauth2server.ClientFormHandler)
	s.oauth2Handler.SetUserAuthorizationHandler(s.UserAuthorizationHandler)
	s.oauth2Handler.SetAuthorizeScopeHandler(s.AuthorizeScopeHandler)
	s.oauth2Handler.SetResponseErrorHandler(s.OAuth2ResponseErrorHandler)
	s.oauth2Handler.SetInternalErrorHandler(s.OAuth2InternalErrorHandler)
	s.oauth2Handler.Config.AllowedGrantTypes = []oauth2.GrantType{
		oauth2.AuthorizationCode,
		oauth2.ClientCredentials,
		oauth2.Refreshing,
		oauth2.Implicit,
	}

	return s
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
