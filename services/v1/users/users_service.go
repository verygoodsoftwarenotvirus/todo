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

	UserServiceConfig struct {
		CookieName    string
		Logger        *logrus.Logger
		Database      database.Database
		Authenticator auth.Enticator
		LoginMonitor  BruteForceLoginDetector
	}

	UsersService struct {
		cookieName    string
		database      database.Database
		authenticator auth.Enticator
		loginMonitor  BruteForceLoginDetector
		logger        *logrus.Logger
	}
)

func NewUsersService(cfg UserServiceConfig) *UsersService {
	us := &UsersService{
		cookieName:    cfg.CookieName,
		database:      cfg.Database,
		loginMonitor:  cfg.LoginMonitor,
		authenticator: cfg.Authenticator,
		logger:        cfg.Logger,
	}
	return us
}
