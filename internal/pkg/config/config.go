package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querier"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding/mariadb"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"

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

	// ServerConfig is our server configuration struct. It is comprised of all the other setting structs
	// For information on this structs fields, refer to their definitions.
	ServerConfig struct {
		Search        search.Config          `json:"search" mapstructure:"search" toml:"search,omitempty"`
		Uploads       uploads.Config         `json:"uploads" mapstructure:"uploads" toml:"uploads,omitempty"`
		Routing       routing.Config         `json:"routing" mapstructure:"routing" toml:"routing,omitempty"`
		Meta          MetaSettings           `json:"meta" mapstructure:"meta" toml:"meta,omitempty"`
		Encoding      encoding.Config        `json:"encoding" mapstructure:"encoding" toml:"meta,omitempty"`
		Frontend      frontendservice.Config `json:"frontend" mapstructure:"frontend" toml:"frontend,omitempty"`
		Observability observability.Config   `json:"observability" mapstructure:"observability" toml:"observability,omitempty"`
		Database      dbconfig.Config        `json:"database" mapstructure:"database" toml:"database,omitempty"`
		Auth          authservice.Config     `json:"auth" mapstructure:"auth" toml:"auth,omitempty"`
		Server        httpserver.Config      `json:"server" mapstructure:"server" toml:"server,omitempty"`
		Webhooks      webhooksservice.Config `json:"webhooks" mapstructure:"webhooks" toml:"webhooks,omitempty"`
		AuditLog      audit.Config           `json:"audit_log" mapstructure:"audit_log" toml:"audit_log,omitempty"`
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
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Search),
		validation.Field(&cfg.Uploads),
		validation.Field(&cfg.Routing),
		validation.Field(&cfg.Meta),
		validation.Field(&cfg.Encoding),
		validation.Field(&cfg.Frontend),
		validation.Field(&cfg.Observability),
		validation.Field(&cfg.Database),
		validation.Field(&cfg.Auth),
		validation.Field(&cfg.Server),
		validation.Field(&cfg.Webhooks),
		validation.Field(&cfg.AuditLog),
	)
}

// ProvideDatabaseClient provides a database implementation dependent on the configuration.
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
