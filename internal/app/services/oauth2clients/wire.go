package oauth2clients

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
)

var (
	// Providers are what we provide for dependency injection.
	Providers = wire.NewSet(
		ProvideOAuth2ClientsService,
		ProvideOAuth2ClientDataServer,
	)
)

// ProvideOAuth2ClientDataServer is an arbitrary function for dependency injection's sake.
func ProvideOAuth2ClientDataServer(s *Service) types.OAuth2ClientDataServer {
	return s
}
