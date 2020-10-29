package audit

import (
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

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
func ProvideAuditLogEntryDataServer(s *Service) models.AuditLogEntryDataServer {
	return s
}
