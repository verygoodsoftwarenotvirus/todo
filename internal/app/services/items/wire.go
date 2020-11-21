package items

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideItemsService,
	ProvideItemDataServer,
	ProvideItemsServiceSearchIndex,
	ProvideItemsServiceItemIDFetcher,
	ProvideItemsServiceSessionInfoFetcher,
)

// ProvideItemDataServer is an arbitrary function for dependency injection's sake.
func ProvideItemDataServer(s *Service) types.ItemDataServer {
	return s
}

// ProvideItemsServiceItemIDFetcher provides an ItemIDFetcher.
func ProvideItemsServiceItemIDFetcher(logger logging.Logger) ItemIDFetcher {
	return routeparams.BuildRouteParamIDFetcher(logger, ItemIDURIParamKey, "item")
}

// ProvideItemsServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideItemsServiceSessionInfoFetcher() SessionInfoFetcher {
	return routeparams.SessionInfoFetcherFromRequestContext
}
