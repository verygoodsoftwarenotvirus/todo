package oauth2clients

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

var (
	// Providers are what we provide for dependency injection.
	Providers = wire.NewSet(
		ProvideOAuth2ClientsService,
		ProvideOAuth2ClientDataServer,
		ProvideOAuth2ClientsServiceClientIDFetcher,
	)
)

// ProvideOAuth2ClientDataServer is an arbitrary function for dependency injection's sake.
func ProvideOAuth2ClientDataServer(s *Service) types.OAuth2ClientDataServer {
	return s
}

// ProvideOAuth2ClientsServiceClientIDFetcher provides a ClientIDFetcher.
func ProvideOAuth2ClientsServiceClientIDFetcher(logger logging.Logger) ClientIDFetcher {
	return routeparams.BuildRouteParamIDFetcher(logger, OAuth2ClientIDURIParamKey, "oauth2 client")
}
