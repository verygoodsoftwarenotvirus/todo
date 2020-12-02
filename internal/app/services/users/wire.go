package users

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

// Providers is what we provide for dependency injectors.
var Providers = wire.NewSet(
	ProvideUsersService,
	ProvideUserDataServer,
	ProvideUsersServiceUserIDFetcher,
	ProvideUsersServiceSessionInfoFetcher,
)

// ProvideUserDataServer is an arbitrary function for dependency injection's sake.
func ProvideUserDataServer(s *Service) types.UserDataService {
	return s
}

// ProvideUsersServiceUserIDFetcher provides a UsernameFetcher.
func ProvideUsersServiceUserIDFetcher(logger logging.Logger) UserIDFetcher {
	return routeparams.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user")
}

// ProvideUsersServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideUsersServiceSessionInfoFetcher() SessionInfoFetcher {
	return routeparams.SessionInfoFetcherFromRequestContext
}
