package admin

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
)

// Providers is our collection of what we provide to other services.
var Providers = wire.NewSet(
	ProvideAdminService,
	ProvideAdminServer,
)

// ProvideAdminServer does the job I wish wire would do for itself.
func ProvideAdminServer(s *Service) types.AdminServer {
	return s
}
