package config

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding/mariadb"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"
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

	randStringSize = 32
	randReadSize   = 24
)

var errNilDatabaseConnection = errors.New("nil DB connection provided")

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

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

	return ioutil.WriteFile(path, byteSlice, 0600)
}

// Validate validates a ServerConfig struct.
func (cfg *ServerConfig) Validate(ctx context.Context) error {
	if err := cfg.Auth.Validate(ctx); err != nil {
		return fmt.Errorf("validating the Auth portion of config: %w", err)
	}

	if err := cfg.Database.Validate(ctx); err != nil {
		return fmt.Errorf("validating the Database portion of config: %w", err)
	}

	if err := cfg.Observability.Validate(ctx); err != nil {
		return fmt.Errorf("validating the Observability portion of config: %w", err)
	}

	if err := cfg.Meta.Validate(ctx); err != nil {
		return fmt.Errorf("validating the Meta portion of config: %w", err)
	}

	if err := cfg.Frontend.Validate(ctx); err != nil {
		return fmt.Errorf("validating the Frontend portion of config: %w", err)
	}

	if err := cfg.Uploads.Validate(ctx); err != nil {
		return fmt.Errorf("validating the Uploads portion of config: %w", err)
	}

	if err := cfg.Search.Validate(ctx); err != nil {
		return fmt.Errorf("validating the Search portion of config: %w", err)
	}

	if err := cfg.Routing.Validate(ctx); err != nil {
		return fmt.Errorf("validating the Routing portion of config: %w", err)
	}

	if err := cfg.Server.Validate(ctx); err != nil {
		return fmt.Errorf("validating the Server portion of config: %w", err)
	}

	if err := cfg.Webhooks.Validate(ctx); err != nil {
		return fmt.Errorf("validating the Webhooks portion of config: %w", err)
	}

	if err := cfg.AuditLog.Validate(ctx); err != nil {
		return fmt.Errorf("validating the AuditLog portion of config: %w", err)
	}

	return nil
}

// RandString produces a random string.
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func RandString() string {
	b := make([]byte, randReadSize)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(b)
}

// ProvideDatabaseClient provides a database implementation dependent on the configuration.
func (cfg *ServerConfig) ProvideDatabaseClient(
	ctx context.Context,
	logger logging.Logger,
	rawDB *sql.DB,
) (database.DataManager, error) {
	if rawDB == nil {
		return nil, errNilDatabaseConnection
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Database.Provider)) {
	case "sqlite":
		return querier.ProvideDatabaseClient(ctx, logger, rawDB, &cfg.Database, sqlite.ProvideSqlite(logger))
	case "mariadb":
		return querier.ProvideDatabaseClient(ctx, logger, rawDB, &cfg.Database, mariadb.ProvideMariaDB(logger))
	case "postgres":
		return querier.ProvideDatabaseClient(ctx, logger, rawDB, &cfg.Database, postgres.ProvidePostgres(logger))
	default:
		return nil, fmt.Errorf("invalid provider: %q", cfg.Database.Provider)
	}
}
