package users

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/sirupsen/logrus"
)

const MiddlewareCtxKey models.ContextKey = "user_input"

type (
	RequestValidator interface {
		Validate(req *http.Request) (bool, error)
	}

	UsersServiceConfig struct {
		CookieName    string
		Logger        *logrus.Logger
		Database      database.Database
		Authenticator auth.Enticator
	}

	UsersService struct {
		cookieName    string
		database      database.Database
		authenticator auth.Enticator
		logger        *logrus.Logger
	}
)

func NewUsersService(cfg UsersServiceConfig) *UsersService {
	us := &UsersService{
		cookieName:    cfg.CookieName,
		database:      cfg.Database,
		authenticator: cfg.Authenticator,
		logger:        cfg.Logger,
	}
	return us
}
