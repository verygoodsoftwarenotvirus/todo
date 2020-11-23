package config

import (
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	"time"
)

const (
	// DevelopmentRunMode is the run mode for a development environment.
	DevelopmentRunMode RunMode = "development"
	// TestingRunMode is the run mode for a testing environment.
	TestingRunMode RunMode = "testing"
	// ProductionRunMode is the run mode for a production environment.
	ProductionRunMode RunMode = "production"
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
	b := make([]byte, randReadSize)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(b)
}
