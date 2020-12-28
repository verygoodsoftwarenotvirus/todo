package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/GuiaBolso/darwin"
	"github.com/Masterminds/squirrel"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     0.00,
			Description: "create plans table and default plan",
			Script: `
			CREATE TABLE IF NOT EXISTS plans (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"name" TEXT NOT NULL,
				"price" INTEGER NOT NULL,
				"period" INTEGER NOT NULL,
				"created_on" INTEGER NOT NULL DEFAULT extract(epoch FROM NOW()),
				"last_updated_on" INTEGER,
				"archived_on" INTEGER DEFAULT NULL
			);

			INSERT INTO plans (name,price,period) VALUES ('free', 0, 0);
		`,
		},
		{
			Version:     0.01,
			Description: "create users table",
			Script: `
			CREATE TABLE IF NOT EXISTS users (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"username" TEXT NOT NULL,
				"hashed_password" TEXT NOT NULL,
				"salt" BYTEA NOT NULL,
				"password_last_changed_on" INTEGER,
				"requires_password_change" BOOLEAN NOT NULL DEFAULT 'false',
				"two_factor_secret" TEXT NOT NULL,
				"two_factor_secret_verified_on" BIGINT DEFAULT NULL,
				"is_admin" BOOLEAN NOT NULL DEFAULT 'false',
				"admin_permissions" BIGINT NOT NULL DEFAULT 0,
				"account_status" TEXT NOT NULL DEFAULT 'created',
				"status_explanation" TEXT NOT NULL DEFAULT '',
				"created_on" BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				"last_updated_on" BIGINT DEFAULT NULL,
				"archived_on" BIGINT DEFAULT NULL,
				UNIQUE ("username")
			);`,
		},
		{
			Version:     0.02,
			Description: "create sessions table for session manager",
			Script: `
			CREATE TABLE sessions (
				token TEXT PRIMARY KEY,
				data BYTEA NOT NULL,
				expiry TIMESTAMPTZ NOT NULL
			);

			CREATE INDEX sessions_expiry_idx ON sessions (expiry);
		`,
		},
		{
			Version:     0.03,
			Description: "create oauth2_clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS oauth2_clients (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"name" TEXT DEFAULT '',
				"client_id" TEXT NOT NULL,
				"client_secret" TEXT NOT NULL,
				"redirect_uri" TEXT DEFAULT '',
				"scopes" TEXT NOT NULL,
				"implicit_allowed" BOOLEAN NOT NULL DEFAULT 'false',
				"created_on" BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				"last_updated_on" BIGINT DEFAULT NULL,
				"archived_on" BIGINT DEFAULT NULL,
				"belongs_to_user" BIGINT NOT NULL,
				FOREIGN KEY("belongs_to_user") REFERENCES users(id)
			);`,
		},
		{
			Version:     0.04,
			Description: "create webhooks table",
			Script: `
			CREATE TABLE IF NOT EXISTS webhooks (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"name" TEXT NOT NULL,
				"content_type" TEXT NOT NULL,
				"url" TEXT NOT NULL,
				"method" TEXT NOT NULL,
				"events" TEXT NOT NULL,
				"data_types" TEXT NOT NULL,
				"topics" TEXT NOT NULL,
				"created_on" BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				"last_updated_on" BIGINT DEFAULT NULL,
				"archived_on" BIGINT DEFAULT NULL,
				"belongs_to_user" BIGINT NOT NULL,
				FOREIGN KEY("belongs_to_user") REFERENCES users(id)
			);`,
		},
		{
			Version:     0.05,
			Description: "create audit log table",
			Script: `
			CREATE TABLE IF NOT EXISTS audit_log (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"event_type" TEXT NOT NULL,
				"context" JSONB NOT NULL,
				"created_on" BIGINT NOT NULL DEFAULT extract(epoch FROM NOW())
			);`,
		},
		{
			Version:     0.06,
			Description: "create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"name" CHARACTER VARYING NOT NULL,
				"details" CHARACTER VARYING NOT NULL DEFAULT '',
				"created_on" BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				"last_updated_on" BIGINT DEFAULT NULL,
				"archived_on" BIGINT DEFAULT NULL,
				"belongs_to_user" BIGINT NOT NULL,
				FOREIGN KEY("belongs_to_user") REFERENCES users(id)
			);`,
		},
	}
)

// buildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a postgres database.
func buildMigrationFunc(db *sql.DB) func() {
	return func() {
		driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})
		if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
			panic(fmt.Errorf("error migrating database: %w", err))
		}
	}
}

// Migrate migrates the database. It does so by invoking the migrateOnce function via sync.Once, so it should be
// safe (as in idempotent, though not necessarily recommended) to call this function multiple times.
func (q *Postgres) Migrate(ctx context.Context, testUserConfig *types.TestUserCreationConfig) error {
	q.logger.Info("migrating db")

	if !q.IsReady(ctx) {
		return database.ErrDBUnready
	}

	q.migrateOnce.Do(buildMigrationFunc(q.db))

	if testUserConfig != nil {
		q.logger.Debug("creating test user")

		query, args, err := q.sqlBuilder.
			Insert(queriers.UsersTableName).
			Columns(
				queriers.UsersTableUsernameColumn,
				queriers.UsersTableHashedPasswordColumn,
				queriers.UsersTableSaltColumn,
				queriers.UsersTableTwoFactorColumn,
				queriers.UsersTableIsAdminColumn,
				queriers.UsersTableAccountStatusColumn,
				queriers.UsersTableAdminPermissionsColumn,
				queriers.UsersTableTwoFactorVerifiedOnColumn,
			).
			Values(
				testUserConfig.Username,
				testUserConfig.HashedPassword,
				[]byte("aaaaaaaaaaaaaaaa"),
				// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
				"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				testUserConfig.IsAdmin,
				types.GoodStandingAccountStatus,
				math.MaxUint32,
				squirrel.Expr(currentUnixTimeQuery),
			).
			ToSql()
		q.logQueryBuildingError(err)

		if _, dbErr := q.db.ExecContext(ctx, query, args...); dbErr != nil {
			q.logger.Error(err, "creating user")
			return fmt.Errorf("error creating test user: %w", dbErr)
		}

		q.logger.WithValue(keys.UsernameKey, testUserConfig.Username).Debug("created test user")
	}

	return nil
}
