package config

import (
	"github.com/google/wire"
)

var (
	// Providers represents this package's offering to the dependency manager.
	Providers = wire.NewSet(
		ProvideDatabaseClient,
		wire.FieldsOf(new(*InstanceConfig),
			"Database",
			"Observability",
			"Capitalism",
			"Meta",
			"Encoding",
			"Uploads",
			"Search",
			"Server",
			"AuditLog",
			"Services",
		),
		wire.FieldsOf(new(*ServicesConfigurations),
			"Auth",
			"Items",
			"Webhooks",
			"Frontend",
		),
	)
)
