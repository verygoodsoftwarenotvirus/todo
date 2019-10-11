package dbclient

import (
	"context"
	"database/sql"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"go.opencensus.io/trace"
)

var _ database.Database = (*Client)(nil)

/*
	NOTE: the primary purpose of this client is to allow convenient
	wrapping of actual query execution.
*/

// Client is a wrapper around a database querier. Client is where all
// logging and trace propagation should happen, the querier is where
// the actual database querying is performed.
type Client struct {
	db      *sql.DB
	querier database.Database
	debug   bool
	logger  logging.Logger
}

// Migrate is a simple wrapper around the core querier Migrate call
func (c *Client) Migrate(ctx context.Context) error {
	return c.querier.Migrate(ctx)
}

// IsReady is a simple wrapper around the core querier IsReady call
func (c *Client) IsReady(ctx context.Context) (ready bool) {
	return c.querier.IsReady(ctx)
}

// ProvideDatabaseClient provides a new Database client
func ProvideDatabaseClient(
	ctx context.Context,
	db *sql.DB,
	querier database.Database,
	debug bool,
	logger logging.Logger,
) (database.Database, error) {
	c := &Client{
		db:      db,
		querier: querier,
		debug:   debug,
		logger:  logger.WithName("db_client"),
	}

	if debug {
		c.logger.SetLevel(logging.DebugLevel)
	}

	c.logger.Debug("migrating querier")
	if err := c.querier.Migrate(ctx); err != nil {
		return nil, err
	}
	c.logger.Debug("querier migrated!")

	return c, nil
}

// attachUserIDToSpan provides a consistent way to attach a user's ID to a span
func attachUserIDToSpan(span *trace.Span, userID uint64) {
	if span != nil {
		span.AddAttributes(
			trace.StringAttribute("user_id", strconv.FormatUint(userID, 10)),
		)
	}
}

// attachFilterToSpan provides a consistent way to attach a filter's info to a span
func attachFilterToSpan(span *trace.Span, filter *models.QueryFilter) {
	if filter != nil && span != nil {
		span.AddAttributes(
			trace.StringAttribute("filter_page", strconv.FormatUint(filter.QueryPage(), 10)),
			trace.StringAttribute("filter_limit", strconv.FormatUint(filter.Limit, 10)),
		)
	}
}
