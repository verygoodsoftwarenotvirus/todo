package users

import (
	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

// Providers is what we provide for dependency injectors.
var Providers = wire.NewSet(
	ProvideUsersService,
	ProvideUsersServiceUserIDFetcher,
	ProvideUsersServiceSessionInfoFetcher,
)

// ProvideUsersServiceUserIDFetcher provides a UsernameFetcher.
func ProvideUsersServiceUserIDFetcher(logger logging.Logger) UserIDFetcher {
	return routeparams.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user")
}

// ProvideUsersServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideUsersServiceSessionInfoFetcher() SessionInfoFetcher {
	return routeparams.SessionInfoFetcherFromRequestContext
}
