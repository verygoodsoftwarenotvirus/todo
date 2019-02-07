package oauth2clients

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
	"github.com/opentracing/opentracing-go"
	"gopkg.in/oauth2.v3"
	oauth2store "gopkg.in/oauth2.v3/store"
)

const (
	// MiddlewareCtxKey is a string alias for referring to OAuth2 clients in contexts
	MiddlewareCtxKey models.ContextKey = "oauth2_client"
)

type (
	// Tracer is a type alias we use for dependency injection
	Tracer opentracing.Tracer

	// Service manages our OAuth2 clients via HTTP
	Service struct {
		database      database.Database
		authenticator auth.Enticator
		logger        logging.Logger
		tracer        opentracing.Tracer
		clientStore   *oauth2store.ClientStore
		tokenStore    oauth2.TokenStore
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
	database database.Database,
	authenticator auth.Enticator,
	logger logging.Logger,
	clientStore *oauth2store.ClientStore,
	tokenStore oauth2.TokenStore,
	tracer Tracer,
) *Service {
	us := &Service{
		database:      database,
		authenticator: authenticator,
		logger:        logger,
		clientStore:   clientStore,
		tokenStore:    tokenStore,
		tracer:        tracer,
	}
	return us
}
