package tracing

import (
	"context"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// Jaeger represents the popular distributed tracing server.
	Jaeger = "jaeger"
)

type (
	// Config contains settings related to tracing.
	Config struct {
		// Jaeger configures the Jaeger tracer.
		Jaeger   *JaegerConfig `json:"jaeger" mapstructure:"jaeger" toml:"jaeger,omitempty"`
		Provider string        `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		// SpanCollectionProbability indicates the probability that a collected span will be reported.
		SpanCollectionProbability float64 `json:"span_collection_probability" mapstructure:"span_collection_probability" toml:"span_collection_probability,omitempty"`
	}

	// JaegerConfig contains settings related to tracing with Jaeger.
	JaegerConfig struct {
		CollectorEndpoint string `json:"collector_endpoint" mapstructure:"collector_endpoint" toml:"collector_endpoint,omitempty"`
		ServiceName       string `json:"service_name" mapstructure:"service_name" toml:"service_name,omitempty"`
	}
)

// Initialize provides an instrumentation handler.
func (c *Config) Initialize(l logging.Logger) (flushFunc func(), err error) {
	logger := l.WithValue("tracing_provider", c.Provider)
	logger.Info("setting tracing provider")

	switch strings.TrimSpace(strings.ToLower(c.Provider)) {
	case Jaeger:
		logger.Debug("setting up jaeger")
		return c.SetupJaeger()
	default:
		logger.Debug("invalid tracing config")
		return nil, nil
	}
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates the config struct.
func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		// validation.Field(&c.Provider, validation.In(Jaeger)),
		validation.Field(&c.Jaeger, validation.When(c.Provider == Jaeger, validation.Required).Else(validation.Nil)),
	)
}

var _ validation.ValidatableWithContext = (*JaegerConfig)(nil)

// ValidateWithContext validates the config struct.
func (c *JaegerConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.CollectorEndpoint, validation.Required),
		validation.Field(&c.ServiceName, validation.Required),
	)
}
