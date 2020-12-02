package admin

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideService,
	ProvideAdminService,
	ProvideAdminServiceUserIDFetcher,
	ProvideAdminServiceSessionInfoFetcher,
)

// ProvideAdminService does the job I wish wire would do for itself.
func ProvideAdminService(s *Service) types.AdminService {
	return s
}

// ProvideAdminServiceUserIDFetcher provides a UsernameFetcher.
func ProvideAdminServiceUserIDFetcher(logger logging.Logger) UserIDFetcher {
	return routeparams.BuildRouteParamIDFetcher(logger, UserIDURIParamKey, "user")
}

// ProvideAdminServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideAdminServiceSessionInfoFetcher() SessionInfoFetcher {
	return routeparams.SessionInfoFetcherFromRequestContext
}
