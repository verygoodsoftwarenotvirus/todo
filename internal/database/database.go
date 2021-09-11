package database

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var (
	// ErrDatabaseNotReady indicates the given database is not ready.
	ErrDatabaseNotReady = errors.New("database is not ready yet")
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

	// SQLQueryExecutor is a subset interface for sql.{DB|Tx} objects.
	SQLQueryExecutor interface {
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
		Migrate(ctx context.Context, maxAttempts uint8, testUserConfig *types.TestUserCreationConfig) error
		IsReady(ctx context.Context, maxAttempts uint8) (ready bool)

		types.AdminUserDataManager
		types.AccountDataManager
		types.AccountUserMembershipDataManager
		types.UserDataManager
		types.AuditLogEntryDataManager
		types.APIClientDataManager
		types.WebhookDataManager
		types.ItemDataManager
		types.AdminAuditManager
		types.AuthAuditManager
	}
)
