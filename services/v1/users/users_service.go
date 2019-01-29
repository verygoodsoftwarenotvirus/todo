package users

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/sirupsen/logrus"
)

// MiddlewareCtxKey is the context key we search for when interacting with user-related requests
const MiddlewareCtxKey models.ContextKey = "user_input"

type (
	// RequestValidator validates request
	RequestValidator interface {
		Validate(req *http.Request) (bool, error)
	}

	// Service handles our users
	Service struct {
		cookieName      CookieName
		database        database.Database
		authenticator   auth.Enticator
		logger          *logrus.Logger
		usernameFetcher func(*http.Request) string
	}
)

// UsernameFetcher fetches usernames from requests
type UsernameFetcher func(*http.Request) string

// CookieName is an arbitrary type alias
type CookieName string

// ProvideUsersService builds a new UsersService
func ProvideUsersService(
	cookieName CookieName,
	logger *logrus.Logger,
	database database.Database,
	authenticator auth.Enticator,
	usernameFetcher UsernameFetcher,
) *Service {
	if usernameFetcher == nil {
		panic("usernameFetcher must be provided")
	}
	us := &Service{
		cookieName:      cookieName,
		database:        database,
		authenticator:   authenticator,
		logger:          logger,
		usernameFetcher: usernameFetcher,
	}
	return us
}
