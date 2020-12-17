package auth

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"

	"github.com/google/wire"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideService,
	ProvideAuthServiceSessionInfoFetcher,
)

// ProvideAuthServiceSessionInfoFetcher provides a SessionInfoFetcher.
func ProvideAuthServiceSessionInfoFetcher() SessionInfoFetcher {
	return routeparams.SessionInfoFetcherFromRequestContext
}
