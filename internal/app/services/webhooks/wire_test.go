package webhooks

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestProvideWebhooksServiceUserIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideWebhooksServiceUserIDFetcher()
	})
}

func TestProvideWebhooksServiceWebhookIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_ = ProvideWebhooksServiceWebhookIDFetcher(noop.NewLogger())
	})
}
