package config

import (
	"crypto/rand"
	"encoding/base32"
	"io/ioutil"
	"time"

	_ "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/randinit"
)

const (
	// DevelopmentRunMode is the run mode for a development environment.
	DevelopmentRunMode RunMode = "development"
	// TestingRunMode is the run mode for a testing environment.
	TestingRunMode RunMode = "testing"
	// ProductionRunMode is the run mode for a production environment.
	ProductionRunMode RunMode = "production"

	// DefaultStartupDeadline is the default amount of time we allow for server startup.
	DefaultStartupDeadline = time.Minute

	// DefaultRunMode is the default run mode.
	DefaultRunMode = DevelopmentRunMode

	randStringSize = 32
)

var (
	// ValidModes is a helper map with every valid RunMode present.
	ValidModes = map[RunMode]struct{}{
		DevelopmentRunMode: {},
		TestingRunMode:     {},
		ProductionRunMode:  {},
	}
)

type (
	// RunMode describes what method of operation the server is under.
	RunMode string

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
		Search   SearchSettings   `json:"search" mapstructure:"search" toml:"search,omitempty"`
		Server   ServerSettings   `json:"server" mapstructure:"server" toml:"server,omitempty"`
		Webhooks WebhooksSettings `json:"webhooks" mapstructure:"webhooks" toml:"webhooks,omitempty"`
		AuditLog AuditLogSettings `json:"audit_log" mapstructure:"audit_log" toml:"audit_log,omitempty"`
	}
)

// EncodeToFile renders your config to a file given your favorite encoder.
func (cfg *ServerConfig) EncodeToFile(path string, marshaler func(v interface{}) ([]byte, error)) error {
	byteSlice, err := marshaler(*cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, byteSlice, 0600)
}

// RandString produces a random string.
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func RandString() string {
	b := make([]byte, randStringSize)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}
