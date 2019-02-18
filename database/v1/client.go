package database

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"

	"github.com/opentracing/opentracing-go"
)

// Client is a wrapper around a database
type Client struct {
	database Database

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
	database Database,
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
