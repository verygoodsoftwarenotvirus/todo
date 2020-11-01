package auth

import (
	"testing"

	oauth2clientsservice "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"

	"github.com/stretchr/testify/assert"
)

func TestProvideOAuth2ClientValidator(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		assert.NotNil(t, ProvideOAuth2ClientValidator(&oauth2clientsservice.Service{}))
	})
}
