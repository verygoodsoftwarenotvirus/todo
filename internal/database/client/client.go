package dbclient

import (
	"context"
	"database/sql"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

var _ database.DataManager = (*Client)(nil)

/*
	NOTE: the primary purpose of this client is to allow convenient
	wrapping of actual query execution.
*/

// Client is a wrapper around a database querier. Client is where all
// logging and trace propagation should happen, the querier is where
// the actual database querying is performed.
type Client struct {
	db      *sql.DB
	querier database.DataManager
	debug   bool
	logger  logging.Logger
}

// Migrate is a simple wrapper around the core querier Migrate call.
func (c *Client) Migrate(ctx context.Context, authenticator auth.Authenticator, testUserConfig *database.UserCreationConfig) error {
	ctx, span := tracing.StartSpan(ctx, "Migrate")
	defer span.End()

	return c.querier.Migrate(ctx, authenticator, testUserConfig)
}

// IsReady is a simple wrapper around the core querier IsReady call.
func (c *Client) IsReady(ctx context.Context) (ready bool) {
	ctx, span := tracing.StartSpan(ctx, "IsReady")
	defer span.End()

	return c.querier.IsReady(ctx)
}

// BeginTx is a simple wrapper around the core querier BeginTx call.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	ctx, span := tracing.StartSpan(ctx, "BeginTx")
	defer span.End()

	return c.querier.BeginTx(ctx, opts)
}

// ProvideDatabaseClient provides a new DataManager client.
func ProvideDatabaseClient(
	ctx context.Context,
	logger logging.Logger,
	querier database.DataManager,
	db *sql.DB,
	authenticator auth.Authenticator,
	testUserConfig *database.UserCreationConfig,
	shouldMigrate,
	debug bool,
) (database.DataManager, error) {
	c := &Client{
		db:      db,
		querier: querier,
		debug:   debug,
		logger:  logger.WithName("db_client"),
	}

	if debug {
		c.logger.SetLevel(logging.DebugLevel)
	}

	if shouldMigrate {
		c.logger.Debug("migrating querier")

		if err := c.querier.Migrate(ctx, authenticator, testUserConfig); err != nil {
			return nil, fmt.Errorf("error migrating database: %w", err)
		}

		c.logger.Debug("querier migrated!")
	}

	return c, nil
}
