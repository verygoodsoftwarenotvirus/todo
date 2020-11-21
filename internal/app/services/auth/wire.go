package auth

import (
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"

	"github.com/google/wire"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideAuthService,
	ProvideOAuth2ClientValidator,
	ProvideAuthServiceSessionInfoFetcher,
)

// ProvideOAuth2ClientValidator converts an oauth2clients.Service to an OAuth2ClientValidator.
func ProvideOAuth2ClientValidator(s *oauth2clientsservice.Service) OAuth2ClientValidator {
	return s
}

// ProvideAuthServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideAuthServiceSessionInfoFetcher() SessionInfoFetcher {
	return routeparams.SessionInfoFetcherFromRequestContext
}
