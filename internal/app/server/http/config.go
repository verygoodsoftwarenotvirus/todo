package httpserver

type (
	// Config describes the settings pertinent to the HTTP serving portion of the service.
	Config struct {
		// Debug determines if debug logging or other development conditions are active.
		Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		// HTTPPort indicates which port to serve HTTP traffic on.
		HTTPPort uint16 `json:"http_port" mapstructure:"http_port" toml:"http_port,omitempty"`
	}
)
