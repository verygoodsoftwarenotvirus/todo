package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
)

var (
	// Providers is our collection of what we provide to other services.
	Providers = wire.NewSet(
		ProvideAuditService,
		ProvideAuditLogEntryDataServer,
	)
)

// ProvideAuditLogEntryDataServer is an arbitrary function for dependency injection's sake.
func ProvideAuditLogEntryDataServer(s *Service) types.AuditLogDataServer {
	return s
}
