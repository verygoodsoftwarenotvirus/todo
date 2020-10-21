package dbclient

import (
	"context"
	"database/sql"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"

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
func (c *Client) Migrate(ctx context.Context, createTestUser bool) error {
	ctx, span := tracing.StartSpan(ctx, "Migrate")
	defer span.End()

	return c.querier.Migrate(ctx, createTestUser)
}

// IsReady is a simple wrapper around the core querier IsReady call.
func (c *Client) IsReady(ctx context.Context) (ready bool) {
	ctx, span := tracing.StartSpan(ctx, "IsReady")
	defer span.End()

	return c.querier.IsReady(ctx)
}

// ProvideDatabaseClient provides a new DataManager client.
func ProvideDatabaseClient(
	ctx context.Context,
	logger logging.Logger,
	querier database.DataManager,
	db *sql.DB,
	shouldMigrate,
	createTestUser,
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
		if err := c.querier.Migrate(ctx, createTestUser); err != nil {
			return nil, err
		}
		c.logger.Debug("querier migrated!")
	}

	return c, nil
}
