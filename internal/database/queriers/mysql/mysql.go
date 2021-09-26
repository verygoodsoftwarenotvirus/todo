package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/luna-duclos/instrumentedsql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	name        = "db_client"
	loggerName  = name
	tracingName = name
)

var _ database.DataManager = (*SQLQuerier)(nil)

// SQLQuerier is the primary database querying client. All tracing/logging/query execution happens here. Query building generally happens elsewhere.
type SQLQuerier struct {
	config      *dbconfig.Config
	db          *sql.DB
	timeFunc    func() uint64
	sqlBuilder  squirrel.StatementBuilderType
	logger      logging.Logger
	tracer      tracing.Tracer
	migrateOnce sync.Once
}

var instrumentedDriverRegistration sync.Once

// ProvideDatabaseClient provides a new DataManager client.
func ProvideDatabaseClient(
	ctx context.Context,
	logger logging.Logger,
	cfg *dbconfig.Config,
	shouldCreateTestUser bool,
) (database.DataManager, error) {
	tracer := tracing.NewTracer(tracingName)

	ctx, span := tracer.StartSpan(ctx)
	defer span.End()

	logger.WithValue(keys.ConnectionDetailsKey, cfg.ConnectionDetails).Debug("Establishing connection to maria DB")

	instrumentedDriverRegistration.Do(func() {
		sql.Register(
			driverName,
			instrumentedsql.WrapDriver(
				&mysql.MySQLDriver{},
				instrumentedsql.WithOmitArgs(),
				instrumentedsql.WithTracer(tracing.NewInstrumentedSQLTracer("mysql_connection")),
				instrumentedsql.WithLogger(tracing.NewInstrumentedSQLLogger(logger)),
			),
		)
	})

	db, err := sql.Open(driverName, string(cfg.ConnectionDetails))
	if err != nil {
		return nil, fmt.Errorf("opening connection to database: %w", err)
	}

	c := &SQLQuerier{
		db:         db,
		config:     cfg,
		tracer:     tracer,
		timeFunc:   defaultTimeFunc,
		logger:     logging.EnsureLogger(logger).WithName(loggerName),
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question),
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

		if err = c.Migrate(ctx, cfg.MaxPingAttempts, testUser); err != nil {
			return nil, observability.PrepareError(err, logger, span, "migrating database")
		}

		c.logger.Debug("querier migrated!")
	}

	return c, nil
}

// ProvideSessionStore provides the scs Store for MySQL.
func (q *SQLQuerier) ProvideSessionStore() scs.Store {
	return mysqlstore.New(q.db)
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

func (q *SQLQuerier) getOneRow(ctx context.Context, querier database.SQLQueryExecutor, queryDescription, query string, args []interface{}) *sql.Row {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachDatabaseQueryToSpan(span, fmt.Sprintf("%s single row fetch query", queryDescription), query, args)

	row := querier.QueryRowContext(ctx, query, args...)

	return row
}

func (q *SQLQuerier) performReadQuery(ctx context.Context, querier database.SQLQueryExecutor, queryDescription, query string, args []interface{}) (*sql.Rows, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("query", query).WithValue("args", args)

	tracing.AttachDatabaseQueryToSpan(span, fmt.Sprintf("%s fetch query", queryDescription), query, args)

	rows, err := querier.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "executing query")
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, observability.PrepareError(rowsErr, logger, span, "scanning results")
	}

	return rows, nil
}

func (q *SQLQuerier) performCountQuery(ctx context.Context, querier database.SQLQueryExecutor, query, queryDesc string) (uint64, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachDatabaseQueryToSpan(span, fmt.Sprintf("%s count query", queryDesc), query, nil)

	var count uint64
	if err := q.getOneRow(ctx, querier, queryDesc, query, []interface{}{}).Scan(&count); err != nil {
		return 0, observability.PrepareError(err, q.logger, span, "executing count query")
	}

	return count, nil
}

func (q *SQLQuerier) performBooleanQuery(ctx context.Context, querier database.SQLQueryExecutor, query string, args []interface{}) (bool, error) {
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

func (q *SQLQuerier) performWriteQuery(ctx context.Context, querier database.SQLQueryExecutor, queryDescription, query string, args []interface{}) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue("query", query).WithValue("description", queryDescription).WithValue("args", args)
	tracing.AttachDatabaseQueryToSpan(span, queryDescription, query, args)

	res, err := querier.ExecContext(ctx, query, args...)
	if err != nil {
		return observability.PrepareError(err, logger, span, "executing %s query", queryDescription)
	}

	var affectedRowCount int64
	if affectedRowCount, err = res.RowsAffected(); affectedRowCount == 0 || err != nil {
		// the only errors returned by the currently supported drivers are either
		// always nil or simply indicate that no rows were affected by the query.

		logger.Debug("no rows modified by query")
		span.AddEvent("no_rows_modified")

		return sql.ErrNoRows
	}

	return nil
}
