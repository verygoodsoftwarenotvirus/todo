package config

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
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

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

type (
	// runMode describes what method of operation the server is under.
	runMode string

	// ServerSettings describes the settings pertinent to the HTTP serving portion of the service.
	ServerSettings struct {
		// Debug determines if debug logging or other development conditions are active.
		Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		// HTTPPort indicates which port to serve HTTP traffic on.
		HTTPPort uint16 `json:"http_port" mapstructure:"http_port" toml:"http_port,omitempty"`
	}

	// ServerConfig is our server configuration struct. It is comprised of all the other setting structs
	// For information on this structs fields, refer to their definitions.
	ServerConfig struct {
		Auth     AuthSettings     `json:"auth" mapstructure:"auth" toml:"auth,omitempty"`
		Database DatabaseSettings `json:"database" mapstructure:"database" toml:"database,omitempty"`
		Metrics  MetricsSettings  `json:"metrics" mapstructure:"metrics" toml:"metrics,omitempty"`
		Meta     MetaSettings     `json:"meta" mapstructure:"meta" toml:"meta,omitempty"`
		Frontend FrontendSettings `json:"frontend" mapstructure:"frontend" toml:"frontend,omitempty"`
		Upload   UploadSettings   `json:"uploads" mapstructure:"uploads" toml:"uploads,omitempty"`
		Search   SearchSettings   `json:"search" mapstructure:"search" toml:"search,omitempty"`
		Server   ServerSettings   `json:"server" mapstructure:"server" toml:"server,omitempty"`
		Webhooks WebhooksSettings `json:"webhooks" mapstructure:"webhooks" toml:"webhooks,omitempty"`
		AuditLog AuditLogSettings `json:"audit_log" mapstructure:"audit_log" toml:"audit_log,omitempty"`
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
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	if err := cfg.Auth.Validate(ctx); err != nil {
		return fmt.Errorf("error validating the Auth portion of config: %w", err)
	}

	if err := cfg.Database.Validate(ctx); err != nil {
		return fmt.Errorf("error validating the Database portion of config: %w", err)
	}

	if err := cfg.Frontend.Validate(ctx); err != nil {
		return fmt.Errorf("error validating the Frontend portion of config: %w", err)
	}

	if err := cfg.Metrics.Validate(ctx); err != nil {
		return fmt.Errorf("error validating the Metrics portion of config: %w", err)
	}

	if err := cfg.Meta.Validate(ctx); err != nil {
		return fmt.Errorf("error validating the Meta portion of config: %w", err)
	}

	if err := cfg.Search.Validate(ctx); err != nil {
		return fmt.Errorf("error validating the Search portion of config: %w", err)
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
