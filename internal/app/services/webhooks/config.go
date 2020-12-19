package webhooks

// Config represents our database configuration.
type Config struct {
	// Debug determines if debug logging or other development conditions are active.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// Enabled determines if we should migrate the database.
	Enabled bool `json:"enabled" mapstructure:"enabled" toml:"enabled,omitempty"`
}
