package oauth2clients

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	oauth2store "gopkg.in/oauth2.v3/store"
)

const (
	// MiddlewareCtxKey is a string alias for referring to OAuth2 clients in contexts
	MiddlewareCtxKey models.ContextKey = "oauth2_client"
)

type (
	// Service manages our OAuth2 clients via HTTP
	Service struct {
		database      database.Database
		authenticator auth.Enticator
		logger        *logrus.Logger
		clientStore   *oauth2store.ClientStore
		tokenStore    oauth2.TokenStore
	}
)

// ProvideOauth2ClientsService builds a new Oauth2ClientsService
func ProvideOauth2ClientsService(
	database database.Database,
	authenticator auth.Enticator,
	logger *logrus.Logger,
	clientStore *oauth2store.ClientStore,
	tokenStore oauth2.TokenStore,
) *Service {
	us := &Service{
		database:      database,
		authenticator: authenticator,
		logger:        logger,
		clientStore:   clientStore,
		tokenStore:    tokenStore,
	}
	return us
}
