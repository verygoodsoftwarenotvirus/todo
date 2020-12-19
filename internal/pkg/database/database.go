package database

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var (
	_ Scanner = (*sql.Row)(nil)
	_ Querier = (*sql.DB)(nil)
	_ Querier = (*sql.Tx)(nil)

	// ErrDBUnready indicates the given database is not ready.
	ErrDBUnready = errors.New("database is not ready yet")
)

type (
	// Scanner represents any database response (i.e. sql.Row[s]).
	Scanner interface {
		Scan(dest ...interface{}) error
	}

	// ResultIterator represents any iterable database response (i.e. sql.Rows).
	ResultIterator interface {
		Next() bool
		Err() error
		Scanner
		io.Closer
	}

	// Querier is a subset interface for sql.{DB|Tx} objects.
	Querier interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}

	// MetricsCollectionInterval defines the interval at which we collect database metrics.
	MetricsCollectionInterval time.Duration

	// ConnectionDetails is a string alias for dependency injection.
	ConnectionDetails string

	// DataManager describes anything that stores data for our services.
	DataManager interface {
		Migrate(ctx context.Context, testUserConfig *types.TestUserCreationConfig) error
		IsReady(ctx context.Context) (ready bool)
		BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

		types.AdminUserDataManager
		types.UserDataManager
		types.AuditLogDataManager
		types.OAuth2ClientDataManager
		types.WebhookDataManager
		types.ItemDataManager

		types.AdminAuditManager
		types.AuthAuditManager
		types.UserAuditManager
		types.OAuth2ClientAuditManager
		types.WebhookAuditManager
		types.ItemAuditManager
	}
)
