package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

var (
	// Providers is our collection of what we provide to other services.
	Providers = wire.NewSet(
		ProvideService,
		ProvideAuditLogEntryDataServer,
		ProvideAuditServiceItemIDFetcher,
		ProvideAuditServiceSessionInfoFetcher,
	)
)

// ProvideAuditLogEntryDataServer is an arbitrary function for dependency injection's sake.
func ProvideAuditLogEntryDataServer(s *Service) types.AuditLogDataService {
	return s
}

// ProvideAuditServiceItemIDFetcher provides an EntryIDFetcher.
func ProvideAuditServiceItemIDFetcher(logger logging.Logger) EntryIDFetcher {
	return routeparams.BuildRouteParamIDFetcher(logger, LogEntryURIParamKey, "log entry")
}

// ProvideAuditServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideAuditServiceSessionInfoFetcher() SessionInfoFetcher {
	return routeparams.SessionInfoFetcherFromRequestContext
}
