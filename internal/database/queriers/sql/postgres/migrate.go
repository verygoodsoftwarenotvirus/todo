package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/GuiaBolso/darwin"
	postgres "github.com/lib/pq"
	"github.com/luna-duclos/instrumentedsql"
	"github.com/segmentio/ksuid"
)

const (
	driverName = "instrumented-postgres"

	// defaultTestUserTwoFactorSecret is the default TwoFactorSecret we give to test users when we initialize them.
	// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
	defaultTestUserTwoFactorSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
)

var instrumentedDriverRegistration sync.Once

// ProvidePostgresDB provides an instrumented postgres db.
func ProvidePostgresDB(logger logging.Logger, connectionDetails database.ConnectionDetails) (*sql.DB, error) {
	logger.WithValue(keys.ConnectionDetailsKey, connectionDetails).Debug("Establishing connection to postgres")

	instrumentedDriverRegistration.Do(func() {
		sql.Register(
			driverName,
			instrumentedsql.WrapDriver(
				&postgres.Driver{},
				instrumentedsql.WithOmitArgs(),
				instrumentedsql.WithTracer(tracing.NewInstrumentedSQLTracer("postgres_connection")),
				instrumentedsql.WithLogger(tracing.NewInstrumentedSQLLogger(logger)),
			),
		)
	})

	db, err := sql.Open(driverName, string(connectionDetails))
	if err != nil {
		return nil, err
	}

	return db, nil
}

const testUserExistenceQuery = `
	SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.service_roles, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = $1 AND users.two_factor_secret_verified_on IS NOT NULL
`

const testUserCreationQuery = `
	INSERT INTO users (id,username,hashed_password,two_factor_secret,reputation,service_roles,two_factor_secret_verified_on) VALUES ($1,$2,$3,$4,$5,$6,extract(epoch FROM NOW()))
`

// Migrate is a simple wrapper around the core querier Migrate call.
func (q *SQLQuerier) Migrate(ctx context.Context, maxAttempts uint8, testUserConfig *types.TestUserCreationConfig) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	q.logger.Info("migrating db")

	if !q.IsReady(ctx, maxAttempts) {
		return database.ErrDatabaseNotReady
	}

	q.migrateOnce.Do(q.migrationFunc)

	if testUserConfig != nil {
		q.logger.Debug("creating test user")

		testUserExistenceArgs := []interface{}{testUserConfig.Username}

		userRow := q.getOneRow(ctx, q.db, "user", testUserExistenceQuery, testUserExistenceArgs)
		_, _, _, err := q.scanUser(ctx, userRow, false)
		if err != nil {
			if testUserConfig.ID == "" {
				testUserConfig.ID = ksuid.New().String()
			}

			testUserCreationArgs := []interface{}{
				testUserConfig.ID,
				testUserConfig.Username,
				testUserConfig.HashedPassword,
				defaultTestUserTwoFactorSecret,
				types.GoodStandingAccountStatus,
				authorization.ServiceAdminRole.String(),
			}

			// these structs will be fleshed out by createUser
			user := &types.User{
				ID:       testUserConfig.ID,
				Username: testUserConfig.Username,
			}
			account := &types.Account{
				ID: ksuid.New().String(),
			}

			if err = q.createUser(ctx, user, account, testUserCreationQuery, testUserCreationArgs); err != nil {
				return observability.PrepareError(err, q.logger, span, "creating test user")
			}
			q.logger.WithValue(keys.UsernameKey, testUserConfig.Username).Debug("created test user and account")
		}
	}

	return nil
}

var (
	migrations = []darwin.Migration{
		{
			Version:     0.0,
			Description: "create sessions table for session manager",
			Script: `
			CREATE TABLE sessions (
				token TEXT PRIMARY KEY,
				data BYTEA NOT NULL,
				expiry TIMESTAMPTZ NOT NULL,
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW())
			);`,
		},
		{
			Version:     0.01,
			Description: "create sessions table for session manager",
			Script:      `CREATE INDEX sessions_expiry_idx ON sessions (expiry);`,
		},
		{
			Version:     0.02,
			Description: "create users table",
			Script: `
			CREATE TABLE IF NOT EXISTS users (
				id CHAR(27) NOT NULL PRIMARY KEY,
				username TEXT NOT NULL,
				avatar_src TEXT,
				hashed_password TEXT NOT NULL,
				password_last_changed_on INTEGER,
				requires_password_change BOOLEAN NOT NULL DEFAULT 'false',
				two_factor_secret TEXT NOT NULL,
				two_factor_secret_verified_on BIGINT DEFAULT NULL,
				service_roles TEXT NOT NULL DEFAULT 'service_user',
				reputation TEXT NOT NULL DEFAULT 'unverified',
				reputation_explanation TEXT NOT NULL DEFAULT '',
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				UNIQUE("username")
			);`,
		},
		{
			Version:     0.03,
			Description: "create accounts table",
			Script: `
			CREATE TABLE IF NOT EXISTS accounts (
				id CHAR(27) NOT NULL PRIMARY KEY,
				name TEXT NOT NULL,
				billing_status TEXT NOT NULL DEFAULT 'unpaid',
				contact_email TEXT NOT NULL DEFAULT '',
				contact_phone TEXT NOT NULL DEFAULT '',
				payment_processor_customer_id TEXT NOT NULL DEFAULT '',
				subscription_plan_id TEXT,
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				belongs_to_user CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				UNIQUE("belongs_to_user", "name")
			);`,
		},
		{
			Version:     0.04,
			Description: "create account user memberships table",
			Script: `
			CREATE TABLE IF NOT EXISTS account_user_memberships (
				id CHAR(27) NOT NULL PRIMARY KEY,
				belongs_to_account CHAR(27) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
				belongs_to_user CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				default_account BOOLEAN NOT NULL DEFAULT 'false',
				account_roles TEXT NOT NULL DEFAULT 'account_user',
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				UNIQUE("belongs_to_account", "belongs_to_user")
			);`,
		},
		{
			Version:     0.05,
			Description: "create API clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS api_clients (
				id CHAR(27) NOT NULL PRIMARY KEY,
				name TEXT DEFAULT '',
				client_id TEXT NOT NULL,
				secret_key BYTEA NOT NULL,
				permissions BIGINT NOT NULL DEFAULT 0,
				admin_permissions BIGINT NOT NULL DEFAULT 0,
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				belongs_to_user CHAR(27) NOT NULL REFERENCES users(id) ON DELETE CASCADE
			);`,
		},
		{
			Version:     0.06,
			Description: "create webhooks table",
			Script: `
			CREATE TABLE IF NOT EXISTS webhooks (
				id CHAR(27) NOT NULL PRIMARY KEY,
				name TEXT NOT NULL,
				content_type TEXT NOT NULL,
				url TEXT NOT NULL,
				method TEXT NOT NULL,
				events TEXT NOT NULL,
				data_types TEXT NOT NULL,
				topics TEXT NOT NULL,
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				belongs_to_account CHAR(27) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE
			);`,
		},
		{
			Version:     0.07,
			Description: "create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				id CHAR(27) NOT NULL PRIMARY KEY,
				name TEXT NOT NULL,
				details TEXT NOT NULL DEFAULT '',
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				belongs_to_account CHAR(27) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE
			);`,
		},
	}
)

// BuildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a postgres database.
func (q *SQLQuerier) migrationFunc() {
	driver := darwin.NewGenericDriver(q.db, darwin.PostgresDialect{})
	if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
		panic(fmt.Errorf("migrating database: %w", err))
	}
}
