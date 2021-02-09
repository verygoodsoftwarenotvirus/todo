package oauth2clients

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	oauth2 "github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	oauth2server "github.com/go-oauth2/oauth2/v4/server"
	oauth2store "github.com/go-oauth2/oauth2/v4/store"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

const (
	// creationMiddlewareCtxKey is a string alias for referring to OAuth2 client creation data.
	creationMiddlewareCtxKey types.ContextKey = "create_oauth2_client"

	counterName        metrics.CounterName = "oauth2_clients"
	counterDescription string              = "number of oauth2 clients managed by the oauth2 client service"
	serviceName        string              = "oauth2_clients_service"
)

var _ types.OAuth2ClientDataService = (*service)(nil)

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

	// service manages our OAuth2 clients via HTTP.
	service struct {
		logger               logging.Logger
		clientDataManager    types.OAuth2ClientDataManager
		userDataManager      types.UserDataManager
		authenticator        authentication.Authenticator
		encoderDecoder       encoding.EncoderDecoder
		urlClientIDExtractor func(req *http.Request) uint64
		oauth2Handler        oauth2Handler
		oauth2ClientCounter  metrics.UnitCounter
		initialized          bool
		tracer               tracing.Tracer
	}
)

// ProvideOAuth2ClientsService builds a new OAuth2ClientsService.
func ProvideOAuth2ClientsService(
	logger logging.Logger,
	clientDataManager types.OAuth2ClientDataManager,
	userDataManager types.UserDataManager,
	authenticator authentication.Authenticator,
	encoderDecoder encoding.EncoderDecoder,
	counterProvider metrics.UnitCounterProvider,
	routeParamManager routing.RouteParamManager,
) (types.OAuth2ClientDataService, error) {
	tokenStore, tokenStoreErr := oauth2store.NewMemoryTokenStore()
	tracer := tracing.NewTracer(serviceName)

	manager := manage.NewDefaultManager()
	manager.MapClientStorage(newClientStore(clientDataManager, tracer))
	manager.MustTokenStorage(tokenStore, tokenStoreErr)
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

	oHandler := oauth2server.NewDefaultServer(manager)
	oHandler.SetAllowGetAccessRequest(true)

	svc := &service{
		clientDataManager:    clientDataManager,
		userDataManager:      userDataManager,
		logger:               logging.EnsureLogger(logger).WithName(serviceName),
		encoderDecoder:       encoderDecoder,
		authenticator:        authenticator,
		urlClientIDExtractor: routeParamManager.BuildRouteParamIDFetcher(logger, OAuth2ClientIDURIParamKey, "oauth2 client"),
		oauth2Handler:        oHandler,
		tracer:               tracer,
	}
	svc.initialize()

	var err error
	if svc.oauth2ClientCounter, err = counterProvider(counterName, counterDescription); err != nil {
		return nil, fmt.Errorf("initializing counter: %w", err)
	}

	return svc, nil
}

// initializeOAuth2Handler.
func (s *service) initialize() {
	if s.initialized {
		return
	}

	s.oauth2Handler.SetAllowGetAccessRequest(true)
	s.oauth2Handler.SetClientAuthorizedHandler(s.ClientAuthorizedHandler)
	s.oauth2Handler.SetClientScopeHandler(s.ClientScopeHandler)
	s.oauth2Handler.SetClientInfoHandler(oauth2server.ClientFormHandler)
	s.oauth2Handler.SetAuthorizeScopeHandler(s.AuthorizeScopeHandler)
	s.oauth2Handler.SetResponseErrorHandler(s.OAuth2ResponseErrorHandler)
	s.oauth2Handler.SetInternalErrorHandler(s.OAuth2InternalErrorHandler)
	s.oauth2Handler.SetUserAuthorizationHandler(s.UserAuthorizationHandler)

	// this sad type cast is here because I have an arbitrary test-only interface for OAuth2 interactions.
	if x, ok := s.oauth2Handler.(*oauth2server.Server); ok {
		x.Config.AllowedGrantTypes = []oauth2.GrantType{
			oauth2.ClientCredentials,
			// oauth2.AuthorizationCode,
			// oauth2.Refreshing,
			// oauth2.Implicit,
		}
	}

	s.initialized = true
}

// HandleAuthorizeRequest is a simple wrapper around the internal server's HandleAuthorizeRequest.
func (s *service) HandleAuthorizeRequest(res http.ResponseWriter, req *http.Request) error {
	return s.oauth2Handler.HandleAuthorizeRequest(res, req)
}

// HandleTokenRequest is a simple wrapper around the internal server's HandleTokenRequest.
func (s *service) HandleTokenRequest(res http.ResponseWriter, req *http.Request) error {
	return s.oauth2Handler.HandleTokenRequest(res, req)
}
