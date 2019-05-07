package auth

import (
	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

var (
	// Providers is our collection of what we provide to other services
	Providers = wire.NewSet(
		ProvideAuthService,
		ProvideWebsocketAuthFunc,
	)
)

// ProvideWebsocketAuthFunc provides a WebsocketAuthFunc
func ProvideWebsocketAuthFunc(svc *Service) newsman.WebsocketAuthFunc {
	return svc.WebsocketAuthFunction
}
