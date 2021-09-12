package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querier"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding/mariadb"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/events"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"
	auditservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	frontendservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/frontend"
	itemsservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/items"
	webhooksservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jmoiron/sqlx"
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
	errNilConfig               = errors.New("nil config provided")
	errInvalidDatabaseProvider = errors.New("invalid database provider")
)

type (
	// runMode describes what method of operation the server is under.
	runMode string

	// ServicesConfigurations collects the various service configurations.
	ServicesConfigurations struct {
		Items    itemsservice.Config    `json:"items" mapstructure:"items" toml:"items,omitempty"`
		Auth     authservice.Config     `json:"auth" mapstructure:"auth" toml:"auth,omitempty"`
		Webhooks webhooksservice.Config `json:"webhooks" mapstructure:"webhooks" toml:"webhooks,omitempty"`
		AuditLog auditservice.Config    `json:"auditLog" mapstructure:"audit_log" toml:"audit_log,omitempty"`
		Frontend frontendservice.Config `json:"frontend" mapstructure:"frontend" toml:"frontend,omitempty"`
	}

	// InstanceConfig configures an instance of the service. It is composed of all the other setting structs.
	InstanceConfig struct {
		Events        events.ProducerConfig  `json:"events" mapstructure:"events" toml:"events,omitempty"`
		Search        search.Config          `json:"search" mapstructure:"search" toml:"search,omitempty"`
		Encoding      encoding.Config        `json:"encoding" mapstructure:"encoding" toml:"encoding,omitempty"`
		Uploads       uploads.Config         `json:"uploads" mapstructure:"uploads" toml:"uploads,omitempty"`
		Observability observability.Config   `json:"observability" mapstructure:"observability" toml:"observability,omitempty"`
		Routing       routing.Config         `json:"routing" mapstructure:"routing" toml:"routing,omitempty"`
		Capitalism    capitalism.Config      `json:"capitalism" mapstructure:"capitalism" toml:"capitalism,omitempty"`
		Meta          MetaSettings           `json:"meta" mapstructure:"meta" toml:"meta,omitempty"`
		Database      dbconfig.Config        `json:"database" mapstructure:"database" toml:"database,omitempty"`
		Services      ServicesConfigurations `json:"services" mapstructure:"services" toml:"services,omitempty"`
		Server        server.Config          `json:"server" mapstructure:"server" toml:"server,omitempty"`
	}
)

// EncodeToFile renders your config to a file given your favorite encoder.
func (cfg *InstanceConfig) EncodeToFile(path string, marshaller func(v interface{}) ([]byte, error)) error {
	byteSlice, err := marshaller(*cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, byteSlice, 0600)
}

var _ validation.ValidatableWithContext = (*InstanceConfig)(nil)

// ValidateWithContext validates a InstanceConfig struct.
func (cfg *InstanceConfig) ValidateWithContext(ctx context.Context) error {
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

	if err := cfg.Server.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating HTTPServer portion of config: %w", err)
	}

	if err := cfg.Services.AuditLog.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating AuditLog portion of config: %w", err)
	}

	if err := cfg.Services.Auth.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Auth service portion of config: %w", err)
	}

	if err := cfg.Services.Frontend.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Frontend service portion of config: %w", err)
	}

	if err := cfg.Services.Webhooks.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Webhooks service portion of config: %w", err)
	}

	if err := cfg.Services.Items.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("error validating Items service portion of config: %w", err)
	}

	return nil
}

// ProvideDatabaseClient provides a database implementation dependent on the configuration.
// NOTE: you may be tempted to move this to the database/config package. This is a fool's errand.
func ProvideDatabaseClient(ctx context.Context, logger logging.Logger, rawDB *sqlx.DB, cfg *InstanceConfig) (database.DataManager, error) {
	if rawDB == nil {
		return nil, errNilDatabaseConnection
	}

	if cfg == nil {
		return nil, errNilConfig
	}

	var qb querybuilding.SQLQueryBuilder
	shouldCreateTestUser := cfg.Meta.RunMode != ProductionRunMode

	switch strings.ToLower(strings.TrimSpace(cfg.Database.Provider)) {
	case "mariadb":
		qb = mariadb.ProvideMariaDB(logger)
	case "postgres":
		qb = postgres.ProvidePostgres(logger)
	default:
		return nil, fmt.Errorf("%w: %q", errInvalidDatabaseProvider, cfg.Database.Provider)
	}

	return querier.ProvideDatabaseClient(ctx, logger, rawDB, &cfg.Database, qb, shouldCreateTestUser)
}
