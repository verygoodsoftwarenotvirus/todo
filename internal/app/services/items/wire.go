package items

import (
	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideService,
	ProvideItemsServiceSearchIndex,
	ProvideItemsServiceItemIDFetcher,
	ProvideItemsServiceSessionInfoFetcher,
)

// ProvideItemsServiceItemIDFetcher provides an ItemIDFetcher.
func ProvideItemsServiceItemIDFetcher(logger logging.Logger) ItemIDFetcher {
	return routeparams.BuildRouteParamIDFetcher(logger, ItemIDURIParamKey, "item")
}

// ProvideItemsServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideItemsServiceSessionInfoFetcher() SessionInfoFetcher {
	return routeparams.SessionInfoFetcherFromRequestContext
}
