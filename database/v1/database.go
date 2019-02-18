package database

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
	"github.com/opentracing/opentracing-go"
)

var (
	// Providers represents what we provide to dependency injectors
	Providers = wire.NewSet(
		ProvidePostgresDatabase,
		ProvideDatabaseClient,
		ProvideTracer,
	)
)

// ProvideTracer provides a tracer
func ProvideTracer() (Tracer, error) {
	return tracing.ProvideTracer("database-client")
}

type (
	// SchemaDirectory is an arbitrary string type
	SchemaDirectory string

	// ConnectionDetails is an arbitrary string type
	ConnectionDetails string

	// Database describes anything that stores data for our services
	Database interface {
		Migrate(ctx context.Context) error
		IsReady(ctx context.Context) (ready bool)

		AdminUserExists(ctx context.Context) (bool, error)

		models.ItemHandler
		models.UserHandler
		models.OAuth2ClientHandler
	}

	// Tracer is an opentracing.Tracer alias
	Tracer opentracing.Tracer

	// SecretGenerator generates secrets
	SecretGenerator interface {
		GenerateSecret(length uint) string
	}

	// ClientIDExtractor extracts client IDs from an interface
	ClientIDExtractor interface {
		ExtractClientID(req *http.Request) string
	}

	// UserIDExtractor extracts user IDs from a request
	UserIDExtractor interface {
		ExtractUserID(req *http.Request) (string, error)
	}
)
