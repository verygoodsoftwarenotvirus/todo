package auth

import (
	"testing"

	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
)

func TestProvideWebsocketAuthFunc(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ProvideWebsocketAuthFunc(buildTestService(t))
	})
}

func TestProvideOAuth2ClientValidator(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ProvideOAuth2ClientValidator(&oauth2clientsservice.Service{})
	})
}
