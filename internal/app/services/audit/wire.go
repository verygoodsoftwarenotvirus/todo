package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

var (
	// Providers is our collection of what we provide to other services.
	Providers = wire.NewSet(
		ProvideService,
		ProvideAuditServiceItemIDFetcher,
		ProvideAuditServiceSessionInfoFetcher,
	)
)

// ProvideAuditServiceItemIDFetcher provides an EntryIDFetcher.
func ProvideAuditServiceItemIDFetcher(logger logging.Logger) EntryIDFetcher {
	return routeparams.BuildRouteParamIDFetcher(logger, LogEntryURIParamKey, "log entry")
}

// ProvideAuditServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideAuditServiceSessionInfoFetcher() SessionInfoFetcher {
	return routeparams.SessionInfoFetcherFromRequestContext
}
