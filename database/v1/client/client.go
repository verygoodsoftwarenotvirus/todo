package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"

	"github.com/opentracing/opentracing-go"
)

var _ database.Database = (*Client)(nil)

/*
	NOTE: the original purpose of this client is to allow convenient wrapping of actual query execution.

	I didn't want to neglect one database implementation while another flourished, so I created this wrapper.
	In reality, a better abstraction is needed, but I need some time to think about what that should look like,
	but it probably will look more like this and less like it did.
*/

// Client is a wrapper around a database
type Client struct {
	database database.Database

	debug  bool
	logger logging.Logger
	tracer opentracing.Tracer
}

// Migrate is a simple wrapper around the core database Migrate call
func (c *Client) Migrate(ctx context.Context) error {
	return c.database.Migrate(ctx)
}

// IsReady is a simple wrapper around the core database IsReady call
func (c *Client) IsReady(ctx context.Context) (ready bool) {
	return c.database.IsReady(ctx)
}

// ProvideDatabaseClient provides a database client
func ProvideDatabaseClient(
	database database.Database,
	debug bool,
	logger logging.Logger,
	tracer Tracer,
) (*Client, error) {
	c := &Client{
		database: database,
		debug:    debug,
		logger:   logger,
		tracer:   tracer,
	}

	return c, nil
}
