package config

import (
	"github.com/google/wire"
)

var (
	// Providers represents this package's offering to the dependency manager.
	Providers = wire.NewSet(
		wire.FieldsOf(new(*ServerConfig),
			"Auth",
			"Database",
			"Observability",
			"Meta",
			"Encoding",
			"Frontend",
			"Uploads",
			"Search",
			"Server",
			"Webhooks",
			"AuditLog",
		),
	)
)