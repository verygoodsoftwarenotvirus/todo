package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

// prepareError standardizes our error handling by logging, tracing, and formatting an error consistently.
func prepareError(err error, logger logging.Logger, span trace.Span, descriptionFmt string, descriptionArgs ...interface{}) error {
	desc := fmt.Sprintf(descriptionFmt, descriptionArgs...)

	logging.EnsureLogger(logger).Error(err, desc)
	tracing.AttachErrorToSpan(span, err)

	return fmt.Errorf("%s: %w", desc, err)
}

func errorOnPurpose(ctx context.Context, tracer tracing.Tracer) {
	_, span := tracer.StartSpan(ctx)
	defer span.End()

	logger := logging.NewNonOperationalLogger()

	if e := prepareError(errors.New("blah 1 blah"), logger, span, "testing this from errorOnPurpose at %d", time.Now().UnixNano()); e == nil {
		println("buh?")
	}

	if e := observability.PrepareError(errors.New("blah 2 blah"), logger, span, "testing this from errorOnPurpose at %d", time.Now().UnixNano()); e == nil {
		println("buh?")
	}

	observability.AcknowledgeError(errors.New("blah 3 blah"), logger, span, "testing this from errorOnPurpose at %d", time.Now().UnixNano())

}

func main() {
	ctx := context.Background()
	logger := zerolog.NewLogger()

	cfg := &tracing.Config{
		Jaeger: &tracing.JaegerConfig{
			CollectorEndpoint: "http://localhost:14268/api/traces",
			ServiceName:       "tracing-experiment",
		},
		Provider:                  "jaeger",
		SpanCollectionProbability: 1,
	}

	ff, err := cfg.Initialize(logger)
	if err != nil {
		panic(err)
	}
	defer ff()

	tracer := tracing.NewTracer("experiment")

	_, span := tracer.StartSpan(ctx)
	defer span.End()

	err = errors.New("blah blah blah")

	observability.AcknowledgeError(errors.New("blah 1 blah"), logger, span, "testing this at %d", time.Now().UnixNano())

	if e := prepareError(errors.New("blah 2 blah"), logger, span, "testing this at %d", time.Now().UnixNano()); e == nil {
		println("buh?")
	}

	errorOnPurpose(ctx, tracer)

	if e := observability.PrepareError(errors.New("blah 3 blah"), logger, span, "testing this at %d", time.Now().UnixNano()); e == nil {
		println("buh?")
	}

	observability.AcknowledgeError(errors.New("blah 4 blah"), logger, span, "testing this at %d", time.Now().UnixNano())
}
