package oauth2clients

import (
	"context"
	"database/sql"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/encoding/v1"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/manage"
	oauth2models "gopkg.in/oauth2.v3/models"
	oauth2server "gopkg.in/oauth2.v3/server"
	oauth2store "gopkg.in/oauth2.v3/store"
)

const (
	// MiddlewareCtxKey is a string alias for referring to OAuth2 clients in contexts
	MiddlewareCtxKey models.ContextKey = "oauth2_client"
)

type (
	// Tracer is a type alias we use for dependency injection
	Tracer opentracing.Tracer

	// ClientIDFetcher is a function for fetching client IDs out of requests
	ClientIDFetcher func(req *http.Request) string

	// UserIDFetcher is a function for fetching user IDs out of requests
	UserIDFetcher func(req *http.Request) uint64

	// Service manages our OAuth2 clients via HTTP
	Service struct {
		database          database.Database
		authenticator     auth.Enticator
		logger            logging.Logger
		tracer            opentracing.Tracer
		encoder           encoding.ResponseEncoder
		tokenStore        oauth2.TokenStore
		clientIDFetcher   func(req *http.Request) string
		userIDFetcher     func(req *http.Request) uint64
		oauth2Handler     *oauth2server.Server
		oauth2ClientStore *oauth2store.ClientStore
	}
)

var (
	// Providers are what we provide for dependency injection
	Providers = wire.NewSet(
		ProvideOAuth2ClientsServiceTracer,
		ProvideOAuth2ClientsService,
	)
)

// ProvideOAuth2ClientsServiceTracer is an obligatory Tracer wrapper
func ProvideOAuth2ClientsServiceTracer() (Tracer, error) {
	return tracing.ProvideTracer("oauth2-clients-service")
}

// ProvideOAuth2ClientsService builds a new OAuth2ClientsService
func ProvideOAuth2ClientsService(
	logger logging.Logger,
	database database.Database,
	authenticator auth.Enticator,
	clientIDFetcher ClientIDFetcher,
	userIDFetcher UserIDFetcher,
	tracer Tracer,
	encoder encoding.ResponseEncoder,
) *Service {

	manager := manage.NewDefaultManager()
	clientStore := oauth2store.NewClientStore()
	manager.MapClientStorage(clientStore)
	tokenStore, err := oauth2store.NewMemoryTokenStore()
	manager.MustTokenStorage(tokenStore, err)
	server := oauth2server.NewDefaultServer(manager)

	us := &Service{
		database:          database,
		authenticator:     authenticator,
		logger:            logger,
		tokenStore:        tokenStore,
		tracer:            tracer,
		encoder:           encoder,
		clientIDFetcher:   clientIDFetcher,
		userIDFetcher:     userIDFetcher,
		oauth2Handler:     server,
		oauth2ClientStore: clientStore,
	}

	us.oauth2Handler.SetAllowGetAccessRequest(true)
	us.oauth2Handler.SetClientAuthorizedHandler(us.ClientAuthorizedHandler)
	us.oauth2Handler.SetClientScopeHandler(us.ClientScopeHandler)
	us.oauth2Handler.SetClientInfoHandler(oauth2server.ClientFormHandler)
	us.oauth2Handler.SetUserAuthorizationHandler(us.UserAuthorizationHandler)
	us.oauth2Handler.SetAuthorizeScopeHandler(us.AuthorizeScopeHandler)
	us.oauth2Handler.SetResponseErrorHandler(us.OAuth2ResponseErrorHandler)
	us.oauth2Handler.SetInternalErrorHandler(us.OAuth2InternalErrorHandler)
	us.oauth2Handler.Config.AllowedGrantTypes = []oauth2.GrantType{
		oauth2.AuthorizationCode,
		oauth2.ClientCredentials,
		oauth2.Refreshing,
		oauth2.Implicit,
	}

	return us
}

// InitializeOAuth2Clients initializes an OAuth2 client
func (s *Service) InitializeOAuth2Clients() (clientCount uint) {
	clientList, err := s.database.GetAllOAuth2Clients(context.Background())
	if err == sql.ErrNoRows {
		return
	} else if err != nil {
		s.logger.Fatal(errors.Wrap(err, "querying oauth clients to add to the clientStore"))
	}

	clientCount = uint(len(clientList))
	s.logger.WithValues(map[string]interface{}{
		"client_count": clientCount,
	}).Debug("loading OAuth2 clients")

	for _, client := range clientList {
		s.logger.WithValue("client_id", client.ClientID).Debug("loading client")

		if err = s.oauth2ClientStore.Set(client.ClientID, &oauth2models.Client{
			ID:     client.ClientID,
			Secret: client.ClientSecret,
			Domain: client.RedirectURI,
			UserID: strconv.FormatUint(client.BelongsTo, 10),
		}); err != nil {
			s.logger.Fatal(errors.Wrap(err, "error encountered loading oauth clients to the clientStore"))
		}
	}

	return
}

// HandleAuthorizeRequest is a simple wrapper around the internal server's HandleAuthorizeRequest
func (s *Service) HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error {
	s.logger.Debug("HandleAuthorizeRequest called")
	err := s.oauth2Handler.HandleAuthorizeRequest(res, req)
	s.logger.Debug("returning from HandleAuthorizeRequest")
	return err
}

// HandleTokenRequest is a simple wrapper around the internal server's HandleTokenRequest
func (s *Service) HandleTokenRequest(res http.ResponseWriter, req *http.Request) error {
	s.logger.Debug("HandleTokenRequest called")
	err := s.oauth2Handler.HandleTokenRequest(res, req)
	s.logger.Debug("returning from HandleTokenRequest")
	return err
}
