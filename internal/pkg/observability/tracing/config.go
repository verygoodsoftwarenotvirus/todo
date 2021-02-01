package tracing

import (
	"fmt"
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"go.opencensus.io/trace"
)

const (
	// Jaeger represents the popular distributed tracing server.
	Jaeger = "jaeger"
)

type (
	// Config contains settings related to tracing.
	Config struct {
		// Provider indicates what exporter should handle our traces.
		Provider string `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		// SpanCollectionProbability indicates the probability that a collected span will be reported.
		SpanCollectionProbability float64 `json:"span_collection_probability" mapstructure:"span_collection_probability" toml:"span_collection_probability,omitempty"`
	}
)

// Initialize provides an instrumentation handler.
func (cfg *Config) Initialize(l logging.Logger) error {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(cfg.SpanCollectionProbability)})

	logger := l.WithValue("tracing_provider", cfg.Provider)
	logger.Info("setting tracing provider")

	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case Jaeger:
		agentHost := os.Getenv("JAEGER_AGENT_HOST")
		agentPort := os.Getenv("JAEGER_AGENT_PORT")
		serviceName := os.Getenv("JAEGER_SERVICE_NAME")

		if agentHost != "" && agentPort != "" && serviceName != "" {
			_, err := SetupJaeger(fmt.Sprintf("http://%s:%s", agentHost, agentPort), serviceName)
			if err != nil {
				return fmt.Errorf("initializing jaeger tracer: %w", err)
			}
		}
	}

	return nil
}
