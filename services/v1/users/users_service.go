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
		CookieName      string
		Logger          *logrus.Logger
		Database        database.Database
		Authenticator   auth.Enticator
		UsernameFetcher func(*http.Request) string
	}

	UsersService struct {
		cookieName      string
		database        database.Database
		authenticator   auth.Enticator
		logger          *logrus.Logger
		usernameFetcher func(*http.Request) string
	}
)

func NewUsersService(cfg UsersServiceConfig) *UsersService {
	if cfg.UsernameFetcher == nil {
		panic("usernameFetcher must be provided")
	}
	us := &UsersService{
		cookieName:      cfg.CookieName,
		database:        cfg.Database,
		authenticator:   cfg.Authenticator,
		logger:          cfg.Logger,
		usernameFetcher: cfg.UsernameFetcher,
	}
	return us
}
