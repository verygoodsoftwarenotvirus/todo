package oauth2clients

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestProvideOAuth2ClientsServiceClientIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideOAuth2ClientsServiceClientIDFetcher(noop.NewLogger())
	})
}
