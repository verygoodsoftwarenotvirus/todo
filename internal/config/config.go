package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"os"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/webhooks"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querier"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding/mariadb"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// DevelopmentRunMode is the run mode for a development environment.
	DevelopmentRunMode runMode = "development"
	// TestingRunMode is the run mode for a testing environment.
	TestingRunMode runMode = "testing"
	// ProductionRunMode is the run mode for a production environment.
	ProductionRunMode runMode = "production"
	// DefaultRunMode is the default run mode.
	DefaultRunMode = DevelopmentRunMode
	// DefaultStartupDeadline is the default amount of time we allow for server startup.
	DefaultStartupDeadline = time.Minute
)

var (
	errNilDatabaseConnection   = errors.New("nil DB connection provided")
	errInvalidDatabaseProvider = errors.New("invalid database provider")
)

type (
	// runMode describes what method of operation the server is under.
	runMode string

	// ServerConfig is our server configuration struct. It is composed of all the other setting structs.
	ServerConfig struct {
		Search        search.Config         `json:"search" mapstructure:"search" toml:"search,omitempty"`
		Encoding      encoding.Config       `json:"encoding" mapstructure:"encoding" toml:"meta,omitempty"`
		Capitalism    capitalism.Config     `json:"capitalism" mapstructure:"capitalism" toml:"capitalism"`
		Uploads       uploads.Config        `json:"uploads" mapstructure:"uploads" toml:"uploads,omitempty"`
		Observability observability.Config  `json:"observability" mapstructure:"observability" toml:"observability,omitempty"`
		Routing       routing.Config        `json:"routing" mapstructure:"routing" toml:"routing,omitempty"`
		Meta          MetaSettings          `json:"meta" mapstructure:"meta" toml:"meta,omitempty"`
		Database      dbconfig.Config       `json:"database" mapstructure:"database" toml:"database,omitempty"`
		Auth          authentication.Config `json:"auth" mapstructure:"auth" toml:"auth,omitempty"`
		Server        server.Config         `json:"server" mapstructure:"server" toml:"server,omitempty"`
		AuditLog      audit.Config          `json:"audit_log" mapstructure:"audit_log" toml:"audit_log,omitempty"`
		Webhooks      webhooks.Config       `json:"webhooks" mapstructure:"webhooks" toml:"webhooks,omitempty"`
		Frontend      frontend.Config       `json:"frontend" mapstructure:"frontend" toml:"frontend,omitempty"`
	}
)

// EncodeToFile renders your config to a file given your favorite encoder.
func (cfg *ServerConfig) EncodeToFile(path string, marshaller func(v interface{}) ([]byte, error)) error {
	byteSlice, err := marshaller(*cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, byteSlice, 0600)
}

var _ validation.ValidatableWithContext = (*ServerConfig)(nil)

// ValidateWithContext validates a ServerConfig struct.
func (cfg *ServerConfig) ValidateWithContext(ctx context.Context) error {
	if err := cfg.Search.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Search portion of config: %w", err)
	}

	if err := cfg.Uploads.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Uploads portion of config: %w", err)
	}

	if err := cfg.Routing.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Routing portion of config: %w", err)
	}

	if err := cfg.Meta.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Meta portion of config: %w", err)
	}

	if err := cfg.Capitalism.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Capitalism portion of config: %w", err)
	}

	if err := cfg.Encoding.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Encoding portion of config: %w", err)
	}

	if err := cfg.Encoding.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Encoding portion of config: %w", err)
	}

	if err := cfg.Observability.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Observability portion of config: %w", err)
	}

	if err := cfg.Database.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Database portion of config: %w", err)
	}

	if err := cfg.Auth.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Auth portion of config: %w", err)
	}

	if err := cfg.Server.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating HTTPServer portion of config: %w", err)
	}

	if err := cfg.Webhooks.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Webhooks portion of config: %w", err)
	}

	if err := cfg.AuditLog.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating AuditLog portion of config: %w", err)
	}

	return nil
}

// ProvideDatabaseClient provides a database implementation dependent on the configuration.
// NOTE: you may be tempted to move this to the database/config package. This is a fool's errand.
func (cfg *ServerConfig) ProvideDatabaseClient(ctx context.Context, logger logging.Logger, rawDB *sql.DB) (database.DataManager, error) {
	if rawDB == nil {
		return nil, errNilDatabaseConnection
	}

	var qb querybuilding.SQLQueryBuilder
	shouldCreateTestUser := cfg.Meta.RunMode != ProductionRunMode

	switch strings.ToLower(strings.TrimSpace(cfg.Database.Provider)) {
	case "sqlite":
		qb = sqlite.ProvideSqlite(logger)
	case "mariadb":
		qb = mariadb.ProvideMariaDB(logger)
	case "postgres":
		qb = postgres.ProvidePostgres(logger)
	default:
		return nil, fmt.Errorf("%w: %q", errInvalidDatabaseProvider, cfg.Database.Provider)
	}

	return querier.ProvideDatabaseClient(ctx, logger, rawDB, &cfg.Database, qb, shouldCreateTestUser)
}
