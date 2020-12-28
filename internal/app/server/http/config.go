package httpserver

import (
	"time"
)

type (
	// Config describes the settings pertinent to the HTTP serving portion of the service.
	Config struct {
		// StartupDeadline indicates how long the service can take to spin up. This includes database migrations, configuring services, etc.
		StartupDeadline time.Duration `json:"startup_deadline" mapstructure:"startup_deadline" toml:"startup_deadline,omitempty"`
		// HTTPPort indicates which port to serve HTTP traffic on.
		HTTPPort uint16 `json:"http_port" mapstructure:"http_port" toml:"http_port,omitempty"`
		// Debug determines if debug logging or other development conditions are active.
		Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	}
)
