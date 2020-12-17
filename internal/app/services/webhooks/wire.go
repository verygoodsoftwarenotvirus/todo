package webhooks

import (
	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

var (
	// Providers is our collection of what we provide to other services.
	Providers = wire.NewSet(
		ProvideWebhooksService,
		ProvideWebhooksServiceWebhookIDFetcher,
		ProvideWebhooksServiceUserIDFetcher,
		ProvideWebhooksServiceSessionInfoFetcher,
	)
)

// ProvideWebhooksServiceWebhookIDFetcher provides an WebhookIDFetcher.
func ProvideWebhooksServiceWebhookIDFetcher(logger logging.Logger) WebhookIDFetcher {
	return routeparams.BuildRouteParamIDFetcher(logger, WebhookIDURIParamKey, "webhook")
}

// ProvideWebhooksServiceUserIDFetcher provides a UserIDFetcher.
func ProvideWebhooksServiceUserIDFetcher() UserIDFetcher {
	return routeparams.UserIDFetcherFromRequestContext
}

// ProvideWebhooksServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideWebhooksServiceSessionInfoFetcher() SessionInfoFetcher {
	return routeparams.SessionInfoFetcherFromRequestContext
}
