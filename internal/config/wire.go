package config

import (
	"github.com/google/wire"
)

var (
	// Providers represents this package's offering to the dependency injector.
	Providers = wire.NewSet(
		ProvideDatabaseClient,
		wire.FieldsOf(
			new(*InstanceConfig),
			"Database",
			"Observability",
			"Capitalism",
			"Meta",
			"Encoding",
			"Uploads",
			"Search",
			"Server",
			"Services",
		),
		wire.FieldsOf(
			new(*ServicesConfigurations),
			"AuditLog",
			"Auth",
			"Frontend",
			"Webhooks",
			"Items",
		),
	)
)
