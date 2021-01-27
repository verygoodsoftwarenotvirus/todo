package querier

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	name        = "db_client"
	loggerName  = name
	tracingName = name
)

var _ database.DataManager = (*Client)(nil)

/*
	NOTE: the primary purpose of this client is to allow convenient
	wrapping of actual query execution.
*/

// Client is a wrapper around a database querier. Client is where all logging and trace propagation should happen,
// the querier is where the actual database querying is performed.
type Client struct {
	config          *dbconfig.Config
	db              *sql.DB
	sqlQueryBuilder database.SQLQueryBuilder
	timeFunc        func() uint64
	logger          logging.Logger
	tracer          tracing.Tracer
	migrateOnce     sync.Once
}

// ProvideDatabaseClient provides a new DataManager client.
func ProvideDatabaseClient(
	ctx context.Context,
	logger logging.Logger,
	db *sql.DB,
	cfg *dbconfig.Config,
	sqlQueryBuilder database.SQLQueryBuilder,
) (database.DataManager, error) {
	tracer := tracing.NewTracer(tracingName)

	ctx, span := tracer.StartSpan(ctx)
	defer span.End()

	c := &Client{
		db:              db,
		config:          cfg,
		tracer:          tracer,
		timeFunc:        defaultTimeFunc,
		logger:          logger.WithName(loggerName),
		sqlQueryBuilder: sqlQueryBuilder,
	}

	if cfg.Debug {
		c.logger.SetLevel(logging.DebugLevel)
	}

	if cfg.RunMigrations {
		c.logger.Debug("migrating querier")

		if err := c.Migrate(ctx, cfg.CreateTestUser); err != nil {
			return nil, fmt.Errorf("migrating database: %w", err)
		}

		c.logger.Debug("querier migrated!")
	}

	return c, nil
}

// Migrate is a simple wrapper around the core querier Migrate call.
func (c *Client) Migrate(ctx context.Context, testUserConfig *types.TestUserCreationConfig) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.Info("migrating db")

	if !c.IsReady(ctx, 50) {
		return database.ErrDBUnready
	}

	c.migrateOnce.Do(c.sqlQueryBuilder.BuildMigrationFunc(c.db))

	if testUserConfig != nil {
		c.logger.Debug("creating test user")

		query, args := c.sqlQueryBuilder.BuildTestUserCreationQuery(testUserConfig)

		var id uint64
		userCreationRow := c.db.QueryRowContext(ctx, query, args...)
		userCreationRow.Scan(&id)

		if _, accountCreationErr := c.CreateAccount(ctx,
			&types.AccountCreationInput{
				Name:          fmt.Sprintf("%s_default", testUserConfig.Username),
				BelongsToUser: id,
			},
		); accountCreationErr != nil {
			c.logger.Error(accountCreationErr, "creating test user")
			return fmt.Errorf("creating test user: %w", accountCreationErr)
		}

		c.logger.WithValue(keys.UsernameKey, testUserConfig.Username).Debug("created test user and account")
	}

	return nil
}

// IsReady is a simple wrapper around the core querier IsReady call.
func (c *Client) IsReady(ctx context.Context, maxAttempts uint8) (ready bool) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	attemptCount := 0

	logger := c.logger.WithValues(map[string]interface{}{
		"interval":     time.Second,
		"max_attempts": maxAttempts,
	})
	logger.Debug("IsReady called")

	for !ready {
		err := c.db.PingContext(ctx)
		if err != nil {
			logger.WithValue("attempt_count", attemptCount).Debug("ping failed, waiting for db")
			time.Sleep(time.Second)

			attemptCount++
			if attemptCount >= int(maxAttempts) {
				return false
			}
		} else {
			ready = true
			return ready
		}
	}

	return false
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

func (c *Client) handleRows(rows database.ResultIterator) error {
	if rowErr := rows.Err(); rowErr != nil {
		c.logger.Error(rowErr, "row error")
		return rowErr
	}

	if closeErr := rows.Close(); closeErr != nil {
		c.logger.Error(closeErr, "closing database rows")
		return closeErr
	}

	return nil
}

func (c *Client) execContext(ctx context.Context, queryDescription, query string, args []interface{}) error {
	_, err := c.execContextAndReturnResult(ctx, queryDescription, query, args)

	return err
}

func (c *Client) execContextAndReturnResult(ctx context.Context, queryDescription, query string, args []interface{}) (sql.Result, error) {
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
