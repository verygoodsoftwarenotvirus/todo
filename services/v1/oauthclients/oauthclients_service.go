package oauthclients

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/sirupsen/logrus"
)

const MiddlewareCtxKey models.ContextKey = "oauth_client"

type (
	Oauth2ClientsServiceConfig struct {
		Logger        *logrus.Logger
		Database      database.Database
		Authenticator auth.Enticator
	}

	Oauth2ClientsService struct {
		database      database.Database
		authenticator auth.Enticator
		logger        *logrus.Logger
	}
)

func NewOauth2ClientsService(cfg Oauth2ClientsServiceConfig) *Oauth2ClientsService {
	us := &Oauth2ClientsService{
		database:      cfg.Database,
		authenticator: cfg.Authenticator,
		logger:        cfg.Logger,
	}
	return us
}
