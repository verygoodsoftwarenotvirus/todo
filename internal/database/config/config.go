package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding/mysql"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/lib/pq"
	"github.com/luna-duclos/instrumentedsql"
)

const (
	// PostgresProvider is the string used to refer to postgres.
	PostgresProvider = "postgres"
	// MySQLProvider is the string used to refer to MySQL.
	MySQLProvider = "mysql"

	// DefaultMetricsCollectionInterval is the default amount of time we wait between database metrics queries.
	DefaultMetricsCollectionInterval = 2 * time.Second
)

var (
	errInvalidDatabase = errors.New("invalid database")
	errNilDBProvided   = errors.New("invalid DB connection provided")
)

type (
	// Config represents our database configuration.
	Config struct {
		CreateTestUser            *types.TestUserCreationConfig `json:"create_test_user" mapstructure:"create_test_user" toml:"create_test_user,omitempty"`
		Provider                  string                        `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		ConnectionDetails         database.ConnectionDetails    `json:"connection_details" mapstructure:"connection_details" toml:"connection_details,omitempty"`
		MetricsCollectionInterval time.Duration                 `json:"metrics_collection_interval" mapstructure:"metrics_collection_interval" toml:"metrics_collection_interval,omitempty"`
		Debug                     bool                          `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		RunMigrations             bool                          `json:"run_migrations" mapstructure:"run_migrations" toml:"run_migrations,omitempty"`
		MaxPingAttempts           uint8                         `json:"max_ping_attempts" mapstructure:"max_ping_attempts" toml:"max_ping_attempts,omitempty"`
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

var instrumentedDriverRegistration sync.Once

// ProvideDatabaseConnection provides a database implementation dependent on the configuration.
func ProvideDatabaseConnection(logger logging.Logger, cfg *Config) (*sql.DB, error) {
	switch cfg.Provider {
	case PostgresProvider:
		logger.WithValue(keys.ConnectionDetailsKey, cfg.ConnectionDetails).Debug("Establishing connection to postgres")

		instrumentedDriverRegistration.Do(func() {
			sql.Register(
				"instrumented-postgres",
				instrumentedsql.WrapDriver(
					&pq.Driver{},
					instrumentedsql.WithOmitArgs(),
					instrumentedsql.WithTracer(tracing.NewInstrumentedSQLTracer("postgres_connection")),
					instrumentedsql.WithLogger(tracing.NewInstrumentedSQLLogger(logger)),
				),
			)
		})

		db, err := sql.Open("instrumented-postgres", string(cfg.ConnectionDetails))
		if err != nil {
			return nil, err
		}

		return db, nil
	case MySQLProvider:
		return mysql.ProvideMySQLConnection(logger, cfg.ConnectionDetails)
	default:
		return nil, fmt.Errorf("%w: %q", errInvalidDatabase, cfg.Provider)
	}
}

// ProvideSessionManager provides a session manager based on some settings.
// There's not a great place to put this function. I don't think it belongs in Auth because it accepts a DB connection,
// but it obviously doesn't belong in the database package, or maybe it does.
func ProvideSessionManager(cookieConfig authservice.CookieConfig, dbConf Config, db *sql.DB) (*scs.SessionManager, error) {
	sessionManager := scs.New()

	if db == nil {
		return nil, errNilDBProvided
	}

	switch dbConf.Provider {
	case PostgresProvider:
		sessionManager.Store = postgresstore.New(db)
	case MySQLProvider:
		sessionManager.Store = mysqlstore.New(db)
	default:
		return nil, fmt.Errorf("%w: %q", errInvalidDatabase, dbConf.Provider)
	}

	sessionManager.Lifetime = cookieConfig.Lifetime
	sessionManager.Cookie.Name = cookieConfig.Name
	sessionManager.Cookie.Domain = cookieConfig.Domain
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Path = "/"
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode
	sessionManager.Cookie.Secure = cookieConfig.SecureOnly

	return sessionManager, nil
}
