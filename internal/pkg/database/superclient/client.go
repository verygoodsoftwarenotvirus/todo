package superclient

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	maximumConnectionAttempts = 50
)

var _ database.DataManager = (*Client)(nil)

/*
	NOTE: the primary purpose of this client is to allow convenient
	wrapping of actual query execution.
*/

// Client is a wrapper around a database querier. Client is where all logging and trace propagation should happen,
// the querier is where the actual database querying is performed.
type Client struct {
	// config          *dbconfig.Config
	db              *sql.DB
	sqlQueryBuilder database.SQLQueryBuilder
	timeFunc        func() uint64
	debug           bool
	logger          logging.Logger
	tracer          tracing.Tracer
}

// Migrate is a simple wrapper around the core querier Migrate call.
func (c *Client) Migrate(ctx context.Context, testUserConfig *types.TestUserCreationConfig) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Debug("Migrate called")

	return nil
}

// IsReady is a simple wrapper around the core querier IsReady call.
func (c *Client) IsReady(ctx context.Context) (ready bool) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	attemptCount := 0

	logger := c.logger.WithValues(map[string]interface{}{
		"interval":     time.Second,
		"max_attempts": maximumConnectionAttempts,
	})
	logger.Debug("IsReady called")

	for !ready {
		err := c.db.PingContext(ctx)
		if err != nil {
			logger.WithValue("attempt_count", attemptCount).Debug("ping failed, waiting for db")
			time.Sleep(time.Second)

			attemptCount++
			if attemptCount >= maximumConnectionAttempts {
				return false
			}
		} else {
			ready = true
			return ready
		}
	}

	return false
}

// ProvideDatabaseClient provides a new DataManager client.
func ProvideDatabaseClient(
	ctx context.Context,
	logger logging.Logger,
	querier database.DataManager,
	db *sql.DB,
	testUserConfig *types.TestUserCreationConfig,
	shouldMigrate,
	debug bool,
) (database.DataManager, error) {
	c := &Client{
		db:              db,
		sqlQueryBuilder: nil,
		timeFunc:        defaultTimeFunc,
		debug:           debug,
		logger:          logger.WithName("db_client"),
		tracer:          tracing.NewTracer("db_client"),
	}

	if debug {
		c.logger.SetLevel(logging.DebugLevel)
	}

	if shouldMigrate {
		c.logger.Debug("migrating querier")

		if err := querier.Migrate(ctx, testUserConfig); err != nil {
			return nil, fmt.Errorf("migrating database: %w", err)
		}

		c.logger.Debug("querier migrated!")
	}

	return c, nil
}

func defaultTimeFunc() uint64 {
	return uint64(time.Now().Unix())
}

func (c *Client) currentTime() uint64 {
	if c == nil || c.timeFunc == nil {
		return defaultTimeFunc()
	}

	return c.timeFunc()
}

func (c *Client) getIDFromResult(res sql.Result) uint64 {
	id, err := res.LastInsertId()
	if err != nil {
		c.logger.WithValue(keys.RowIDErrorKey, true).Error(err, "fetching row ID")
	}

	return uint64(id)
}

func (c *Client) execContext(ctx context.Context, queryDescription, query string, args ...interface{}) error {
	_, err := c.execContextAndReturnResult(ctx, queryDescription, query, args...)

	return err
}

func (c *Client) execContextAndReturnResult(ctx context.Context, queryDescription, query string, args ...interface{}) (sql.Result, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	res, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing %s query: %w", queryDescription, err)
	}

	if res != nil {
		if rowCount, rowCountErr := res.RowsAffected(); rowCountErr == nil && rowCount == 0 {
			return nil, sql.ErrNoRows
		}
	}

	return res, nil
}
