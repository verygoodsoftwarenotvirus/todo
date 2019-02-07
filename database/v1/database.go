package database

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// SchemaDirectory is an arbitrary string type
type SchemaDirectory string

// ConnectionDetails is an arbitrary string type
type ConnectionDetails string

// Database describes anything that stores data for our services
type Database interface {
	Migrate(ctx context.Context, schemaDir SchemaDirectory) error
	IsReady(ctx context.Context) (ready bool)

	models.ItemHandler
	models.UserHandler
	models.OAuth2ClientHandler
}

// SecretGenerator generates secrets
type SecretGenerator interface {
	GenerateSecret(length uint) string
}

// ClientIDExtractor extracts client IDs from an interface
type ClientIDExtractor interface {
	ExtractClientID(req *http.Request) string
}

// UserIDExtractor extracts user IDs from a request
type UserIDExtractor interface {
	ExtractUserID(req *http.Request) (string, error)
}

// Scannable represents any database response (i.e. either a transaction or a regular execution response)
type Scannable interface {
	Scan(dest ...interface{}) error
}
