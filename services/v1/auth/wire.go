package auth

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

var (
	// Providers is our collection of what we provide to other services
	Providers = wire.NewSet(
		ProvideAuthService,
		ProvideWebsocketAuthFunc,
		ProvideOAuth2ClientValidator,
	)
)

// ProvideWebsocketAuthFunc provides a WebsocketAuthFunc
func ProvideWebsocketAuthFunc(svc *Service) newsman.WebsocketAuthFunc {
	return svc.WebsocketAuthFunction
}

// ProvideOAuth2ClientValidator converts an oauth2clients.Service to an OAuth2ClientValidator
func ProvideOAuth2ClientValidator(s *oauth2clients.Service) OAuth2ClientValidator {
	return s
}
