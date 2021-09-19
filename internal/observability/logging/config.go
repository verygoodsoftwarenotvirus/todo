package logging

const (
	// ProviderZerolog indicates you'd like to use the zerolog Logger.
	ProviderZerolog = "zerolog"
)

type (
	// Config configures a zerologLogger.
	Config struct {
		_ struct{}

		Name     string `json:"name"  mapstructure:"name" toml:"name"`
		Level    Level  `json:"level"  mapstructure:"level" toml:"level"`
		Provider string `json:"provider" mapstructure:"provider" toml:"provider"`
	}
)

// ProvideLogger builds a Logger according to the provided config.
func ProvideLogger(cfg Config) Logger {
	var l Logger

	switch cfg.Provider {
	case ProviderZerolog:
		l = NewZerologLogger()
	default:
		l = NewNoopLogger()
	}

	l.SetLevel(cfg.Level)

	if cfg.Name != "" {
		l = l.WithName(cfg.Name)
	}

	return l
}
