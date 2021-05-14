package random

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"io"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

const (
	arbitrarySize uint16 = 128
)

var (
	_ Generator = (*standardGenerator)(nil)

	defaultGenerator = NewGenerator(logging.NewNonOperationalLogger())
)

func init() {
	if _, err := rand.Read(make([]byte, arbitrarySize)); err != nil {
		panic(err)
	}
}

type (
	// Generator should generate random strings securely.
	Generator interface {
		GenerateBase32EncodedString(context.Context, int) (string, error)
		GenerateBase64EncodedString(context.Context, int) (string, error)
		GenerateRawBytes(context.Context, int) ([]byte, error)
	}

	standardGenerator struct {
		logger     logging.Logger
		tracer     tracing.Tracer
		randReader io.Reader
	}
)

// NewGenerator builds a new Generator.
func NewGenerator(logger logging.Logger) Generator {
	return &standardGenerator{
		logger:     logging.EnsureLogger(logger).WithName("random_string_generator"),
		tracer:     tracing.NewTracer("secret_generator"),
		randReader: rand.Reader,
	}
}

// GenerateBase32EncodedString generates a one-off value with an anonymous Generator.
func GenerateBase32EncodedString(ctx context.Context, length int) (string, error) {
	return defaultGenerator.GenerateBase32EncodedString(ctx, length)
}

// GenerateBase64EncodedString generates a one-off value with an anonymous Generator.
func GenerateBase64EncodedString(ctx context.Context, length int) (string, error) {
	return defaultGenerator.GenerateBase64EncodedString(ctx, length)
}

// GenerateRawBytes generates a one-off value with an anonymous Generator.
func GenerateRawBytes(ctx context.Context, length int) ([]byte, error) {
	return defaultGenerator.GenerateRawBytes(ctx, length)
}

// GenerateBase32EncodedString generates a base64-encoded string of a securely random byte array of a given length.
func (g *standardGenerator) GenerateBase32EncodedString(ctx context.Context, length int) (string, error) {
	_, span := tracing.StartSpan(ctx)
	defer span.End()

	logger := g.logger.WithValue("requested_length", length)

	b := make([]byte, length)
	if _, err := g.randReader.Read(b); err != nil {
		return "", observability.PrepareError(err, logger, span, "reading from secure random source")
	}

	return base32.StdEncoding.EncodeToString(b), nil
}

// GenerateBase64EncodedString generates a base64-encoded string of a securely random byte array of a given length.
func (g *standardGenerator) GenerateBase64EncodedString(ctx context.Context, length int) (string, error) {
	_, span := tracing.StartSpan(ctx)
	defer span.End()

	logger := g.logger.WithValue("requested_length", length)

	b := make([]byte, length)
	if _, err := g.randReader.Read(b); err != nil {
		return "", observability.PrepareError(err, logger, span, "reading from secure random source")
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateRawBytes generates a securely random byte array.
func (g *standardGenerator) GenerateRawBytes(ctx context.Context, length int) ([]byte, error) {
	_, span := tracing.StartSpan(ctx)
	defer span.End()

	logger := g.logger.WithValue("requested_length", length)

	b := make([]byte, length)
	if _, err := g.randReader.Read(b); err != nil {
		return nil, observability.PrepareError(err, logger, span, "reading from secure random source")
	}

	return b, nil
}
