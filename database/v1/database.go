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
		ExportData(ctx context.Context, user *models.User) (*models.DataExport, error)

		models.ItemDataManager
		models.UserDataManager
		models.OAuth2ClientDataManager
		models.WebhookDataManager
	}

	// ConnectionDetails is a string alias for a Postgres url
	ConnectionDetails string

	// Scanner represents any database response (i.e. either a transaction or a regular execution response)
	Scanner interface {
		Scan(dest ...interface{}) error
	}

	// Querier is a subset interface for sql.{DB|Tx|Stmt} objects
	Querier interface {
		ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
	}
)
