package database

import (
	"context"
	"database/sql"
	"errors"
	"io"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/auth"
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

	// UserCreationConfig is a helper struct because of cyclical imports.
	UserCreationConfig struct {
		Username string
		Password string
		IsAdmin  bool
	}

	// Querier is a subset interface for sql.{DB|Tx} objects.
	Querier interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}

	// ConnectionDetails is a string alias for dependency injection.
	ConnectionDetails string

	// DataManager describes anything that stores data for our services.
	DataManager interface {
		Migrate(ctx context.Context, authenticator auth.Authenticator, testUserConfig *UserCreationConfig) error
		IsReady(ctx context.Context) (ready bool)
		BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

		types.AuditLogDataManager
		types.ItemDataManager
		types.UserDataManager
		types.AdminUserDataManager
		types.OAuth2ClientDataManager
		types.WebhookDataManager

		types.ItemAuditManager
		types.WebhookAuditManager
		types.AuthAuditManager
		types.UserAuditManager
		types.OAuth2ClientAuditManager
	}
)
