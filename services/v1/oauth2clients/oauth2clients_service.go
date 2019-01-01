package oauth2clients

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/sirupsen/logrus"
	"gopkg.in/oauth2.v3"
	oauth2store "gopkg.in/oauth2.v3/store"
)

const MiddlewareCtxKey models.ContextKey = "oauth2_client"

type (
	Oauth2ClientsServiceConfig struct {
		Logger        *logrus.Logger
		Database      database.Database
		Authenticator auth.Enticator
		ClientStore   *oauth2store.ClientStore
		TokenStore    oauth2.TokenStore
	}

	Oauth2ClientsService struct {
		database      database.Database
		authenticator auth.Enticator
		logger        *logrus.Logger
		clientStore   *oauth2store.ClientStore
		tokenStore    oauth2.TokenStore
	}
)

func NewOauth2ClientsService(cfg Oauth2ClientsServiceConfig) *Oauth2ClientsService {
	us := &Oauth2ClientsService{
		database:      cfg.Database,
		authenticator: cfg.Authenticator,
		logger:        cfg.Logger,
		clientStore:   cfg.ClientStore,
		tokenStore:    cfg.TokenStore,
	}
	return us
}
