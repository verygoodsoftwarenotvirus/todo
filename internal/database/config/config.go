package config

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/alexedwards/scs/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// PostgresProvider is the string used to refer to postgres.
	PostgresProvider = "postgres"
	// MySQLProvider is the string used to refer to MySQL.
	MySQLProvider = "mysql"
)

type (
	// Config represents our database configuration.
	Config struct {
		CreateTestUser    *types.TestUserCreationConfig `json:"create_test_user" mapstructure:"create_test_user" toml:"create_test_user,omitempty"`
		Provider          string                        `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		ConnectionDetails database.ConnectionDetails    `json:"connection_details" mapstructure:"connection_details" toml:"connection_details,omitempty"`
		Debug             bool                          `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		RunMigrations     bool                          `json:"run_migrations" mapstructure:"run_migrations" toml:"run_migrations,omitempty"`
		MaxPingAttempts   uint8                         `json:"max_ping_attempts" mapstructure:"max_ping_attempts" toml:"max_ping_attempts,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates an DatabaseSettings struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.ConnectionDetails, validation.Required),
		validation.Field(&cfg.Provider, validation.In(PostgresProvider, MySQLProvider)),
		validation.Field(&cfg.CreateTestUser, validation.When(cfg.CreateTestUser != nil, validation.Required).Else(validation.Nil)),
	)
}

// ProvideSessionManager provides a session manager based on some settings.
// There's not a great place to put this function. I don't think it belongs in Auth because it accepts a DB connection,
// but it obviously doesn't belong in the database package, or maybe it does.
func ProvideSessionManager(cookieConfig authservice.CookieConfig, dm database.DataManager) (*scs.SessionManager, error) {
	sessionManager := scs.New()

	sessionManager.Lifetime = cookieConfig.Lifetime
	sessionManager.Cookie.Name = cookieConfig.Name
	sessionManager.Cookie.Domain = cookieConfig.Domain
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Path = "/"
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode
	sessionManager.Cookie.Secure = cookieConfig.SecureOnly

	sessionManager.Store = dm.ProvideSessionStore()

	return sessionManager, nil
}
