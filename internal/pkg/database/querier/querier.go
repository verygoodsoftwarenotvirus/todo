package querier

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

type idRetrievalStrategy int

const (
	name        = "db_client"
	loggerName  = name
	tracingName = name

	defaultBatchSize = 1000

	// DefaultIDRetrievalStrategy uses the standard library .LastInsertId method to retrieve the ID.
	DefaultIDRetrievalStrategy idRetrievalStrategy = iota
	// ReturningStatementIDRetrievalStrategy scans IDs to results for database drivers that support queries with RETURNING statements.
	ReturningStatementIDRetrievalStrategy
)

var _ database.DataManager = (*SQLQuerier)(nil)

// SQLQuerier is a wrapper around a database querier. SQLQuerier is where all logging and trace propagation should happen,
// the querier is where the actual database querying is performed.
type SQLQuerier struct {
	config          *dbconfig.Config
	db              *sql.DB
	sqlQueryBuilder querybuilding.SQLQueryBuilder
	timeFunc        func() uint64
	logger          logging.Logger
	tracer          tracing.Tracer
	migrateOnce     sync.Once
	idStrategy      idRetrievalStrategy
}

// ProvideDatabaseClient provides a new DataManager client.
func ProvideDatabaseClient(
	ctx context.Context,
	logger logging.Logger,
	db *sql.DB,
	cfg *dbconfig.Config,
	sqlQueryBuilder querybuilding.SQLQueryBuilder,
	shouldCreateTestUser bool,
) (database.DataManager, error) {
	tracer := tracing.NewTracer(tracingName)

	ctx, span := tracer.StartSpan(ctx)
	defer span.End()

	c := &SQLQuerier{
		db:              db,
		config:          cfg,
		tracer:          tracer,
		timeFunc:        defaultTimeFunc,
		logger:          logging.EnsureLogger(logger).WithName(loggerName),
		sqlQueryBuilder: sqlQueryBuilder,
		idStrategy:      DefaultIDRetrievalStrategy,
	}

	if cfg.Provider == dbconfig.PostgresProviderKey {
		c.idStrategy = ReturningStatementIDRetrievalStrategy
	}

	if cfg.Debug {
		c.logger.SetLevel(logging.DebugLevel)
	}

	if cfg.RunMigrations {
		c.logger.Debug("migrating querier")

		var testUser *types.TestUserCreationConfig
		if shouldCreateTestUser {
			testUser = cfg.CreateTestUser
		}

		if err := c.Migrate(ctx, cfg.MaxPingAttempts, testUser); err != nil {
			return nil, observability.PrepareError(err, logger, span, "migrating database")
		}

		c.logger.Debug("querier migrated!")
	}

	return c, nil
}

// Migrate is a simple wrapper around the core querier Migrate call.
func (q *SQLQuerier) Migrate(ctx context.Context, maxAttempts uint8, testUserConfig *types.TestUserCreationConfig) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	q.logger.Info("migrating db")

	if !q.IsReady(ctx, maxAttempts) {
		return database.ErrDBUnready
	}

	q.migrateOnce.Do(q.sqlQueryBuilder.BuildMigrationFunc(q.db))

	if testUserConfig != nil {
		q.logger.Debug("creating test user")

		query, args := q.sqlQueryBuilder.BuildTestUserCreationQuery(ctx, testUserConfig)

		// these structs will be fleshed out by createUser
		user := &types.User{Username: testUserConfig.Username}
		account := &types.Account{DefaultNewMemberPermissions: permissions.ServiceUserPermission(math.MaxInt64)}

		if err := q.createUser(ctx, user, account, query, args); err != nil {
			q.logger.Error(err, "creating test user")
			return observability.PrepareError(err, q.logger, span, "creating test user")
		}

		q.logger.WithValue(keys.UsernameKey, testUserConfig.Username).Debug("created test user and account")
	}

	return nil
}

// IsReady is a simple wrapper around the core querier IsReady call.
func (q *SQLQuerier) IsReady(ctx context.Context, maxAttempts uint8) (ready bool) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	attemptCount := 0

	logger := q.logger.WithValues(map[string]interface{}{
		"interval":     time.Second.String(),
		"max_attempts": maxAttempts,
	})

	for !ready {
		err := q.db.PingContext(ctx)
		if err != nil {
			logger.WithValue("attempt_count", attemptCount).Debug("ping failed, waiting for db")
			time.Sleep(time.Second)

			attemptCount++
			if attemptCount >= int(maxAttempts) {
				break
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

func (q *SQLQuerier) currentTime() uint64 {
	if q == nil || q.timeFunc == nil {
		return defaultTimeFunc()
	}

	return q.timeFunc()
}

func (q *SQLQuerier) checkRowsForErrorAndClose(ctx context.Context, rows database.ResultIterator) error {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger

	if err := rows.Err(); err != nil {
		q.logger.Error(err, "row error")
		return observability.PrepareError(err, logger, span, "row error")
	}

	if err := rows.Close(); err != nil {
		q.logger.Error(err, "closing database rows")
		return observability.PrepareError(err, logger, span, "closing database rows")
	}

	return nil
}

func (q *SQLQuerier) rollbackTransaction(ctx context.Context, tx *sql.Tx) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if err := tx.Rollback(); err != nil {
		observability.AcknowledgeError(err, q.logger, span, "rolling back transaction")
	}
}

func (q *SQLQuerier) getIDFromResult(ctx context.Context, res sql.Result) uint64 {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	id, err := res.LastInsertId()
	if err != nil {
		observability.AcknowledgeError(err, q.logger.WithValue(keys.RowIDErrorKey, true), span, "fetching row ID")
	}

	return uint64(id)
}

func (q *SQLQuerier) getOneRow(ctx context.Context, querier database.Querier, queryDescription, query string, args ...interface{}) *sql.Row {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachDatabaseQueryToSpan(span, fmt.Sprintf("%s single row fetch query", queryDescription), query, args)

	return querier.QueryRowContext(ctx, query, args...)
}

func (q *SQLQuerier) performReadQuery(ctx context.Context, querier database.Querier, queryDescription, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachDatabaseQueryToSpan(span, fmt.Sprintf("%s fetch query", queryDescription), query, args)

	rows, err := querier.QueryContext(ctx, query, args...)
	if err != nil {
		logger := q.logger
		return nil, observability.PrepareError(err, logger, span, "scanning user")
	}

	return rows, nil
}

func (q *SQLQuerier) performCountQuery(ctx context.Context, querier database.Querier, query, queryDesc string) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachDatabaseQueryToSpan(span, fmt.Sprintf("%s count query", queryDesc), query, nil)

	var count uint64
	if err := q.getOneRow(ctx, querier, queryDesc, query).Scan(&count); err != nil {
		return 0, observability.PrepareError(err, q.logger, span, "executing count query")
	}

	return count, nil
}

func (q *SQLQuerier) performBooleanQuery(ctx context.Context, querier database.Querier, query string, args []interface{}) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	var exists bool

	logger := q.logger.WithValue(keys.DatabaseQueryKey, query)
	tracing.AttachDatabaseQueryToSpan(span, "boolean query", query, args)

	err := querier.QueryRowContext(ctx, query, args...).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, observability.PrepareError(err, logger, span, "executing boolean query")
	}

	return exists, nil
}

func (q *SQLQuerier) performWriteQueryIgnoringReturn(ctx context.Context, querier database.Querier, queryDescription, query string, args []interface{}) error {
	_, err := q.performWriteQuery(ctx, querier, true, queryDescription, query, args)

	return err
}

func (q *SQLQuerier) performWriteQuery(ctx context.Context, querier database.Querier, ignoreReturn bool, queryDescription, query string, args []interface{}) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("query", query).WithValue("description", queryDescription).WithValue("args", args)
	tracing.AttachDatabaseQueryToSpan(span, queryDescription, query, args)

	if q.idStrategy == ReturningStatementIDRetrievalStrategy && !ignoreReturn {
		var id uint64

		if err := querier.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
			return 0, observability.PrepareError(err, logger, span, "executing %s query", queryDescription)
		}

		logger.Debug("query executed successfully")

		return id, nil
	} else if q.idStrategy == ReturningStatementIDRetrievalStrategy {
		res, err := querier.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, observability.PrepareError(err, logger, span, "executing %s query", queryDescription)
		}

		if count, err := res.RowsAffected(); count == 0 || err != nil {
			logger.Debug("no rows modified by query")
			span.AddEvent("no_rows_modified")

			return 0, sql.ErrNoRows
		}

		logger.Debug("query executed successfully")

		return 0, nil
	}

	res, err := querier.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, observability.PrepareError(err, logger, span, "executing query")
	}

	if res != nil {
		if rowCount, err := res.RowsAffected(); err == nil && rowCount == 0 {
			logger.Debug("no rows modified by query")
			span.AddEvent("no_rows_modified")

			return 0, sql.ErrNoRows
		}
	}

	logger.Debug("query executed successfully")

	return q.getIDFromResult(ctx, res), nil
}
