package database

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

type (
	// Database describes anything that stores data for our services
	Database interface {
		Migrate(ctx context.Context) error
		IsReady(ctx context.Context) (ready bool)

		models.ItemDataManager
		models.UserDataManager
		models.OAuth2ClientDataManager
		models.WebhookDataManager
	}

	// ConnectionDetails is a string alias for a Postgres url
	ConnectionDetails string
)

var _ Scanner = (*sql.Row)(nil)

// Scanner represents any database response (i.e. sql.Row[s])
type Scanner interface {
	Scan(dest ...interface{}) error
}

var _ Querier = (*sql.DB)(nil)
var _ Querier = (*sql.Tx)(nil)

// Querier is a subset interface for sql.{DB|Tx} objects
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
