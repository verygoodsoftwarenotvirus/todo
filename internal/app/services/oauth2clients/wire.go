package oauth2clients

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

var (
	// Providers are what we provide for dependency injection.
	Providers = wire.NewSet(
		ProvideOAuth2ClientsService,
		ProvideOAuth2ClientsServiceClientIDFetcher,
	)
)

// ProvideOAuth2ClientsServiceClientIDFetcher provides a ClientIDFetcher.
func ProvideOAuth2ClientsServiceClientIDFetcher(logger logging.Logger) ClientIDFetcher {
	return routeparams.BuildRouteParamIDFetcher(logger, OAuth2ClientIDURIParamKey, "oauth2 client")
}
