package auth

import (
	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/oauth2clients"

	"github.com/google/wire"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideAuthService,
	ProvideOAuth2ClientValidator,
)

// ProvideOAuth2ClientValidator converts an oauth2clients.Service to an OAuth2ClientValidator.
func ProvideOAuth2ClientValidator(s *oauth2clientsservice.Service) OAuth2ClientValidator {
	return s
}
