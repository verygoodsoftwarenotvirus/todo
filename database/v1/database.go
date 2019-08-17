package database

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

var (
	_ Scanner = (*sql.Row)(nil)
	_ Querier = (*sql.DB)(nil)
	_ Querier = (*sql.Tx)(nil)
)

type (
	// Scanner represents any database response (i.e. sql.Row[s])
	Scanner interface {
		Scan(dest ...interface{}) error
	}

	// Querier is a subset interface for sql.{DB|Tx} objects
	Querier interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}

	// ConnectionDetails is a string alias for dependency injection
	ConnectionDetails string

	// Database describes anything that stores data for our services
	Database interface {
		Migrate(ctx context.Context) error
		IsReady(ctx context.Context) (ready bool)

		models.ItemDataManager
		models.UserDataManager
		models.OAuth2ClientDataManager
		models.WebhookDataManager
	}
)
