package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"

	"github.com/GuiaBolso/darwin"
	"github.com/Masterminds/squirrel"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "create users table",
			Script: `
			CREATE TABLE IF NOT EXISTS users (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"username" TEXT NOT NULL,
				"hashed_password" TEXT NOT NULL,
				"salt" TINYBLOB NOT NULL,
				"password_last_changed_on" INTEGER,
				"requires_password_change" BOOLEAN NOT NULL DEFAULT 'false',
				"two_factor_secret" TEXT NOT NULL,
				"two_factor_secret_verified_on" INTEGER DEFAULT NULL,
				"is_admin" BOOLEAN NOT NULL DEFAULT 'false',
				"admin_permissions" INTEGER NOT NULL DEFAULT 0,
				"account_status" TEXT NOT NULL DEFAULT 'created',
				"status_explanation" TEXT NOT NULL DEFAULT '',
				"created_on" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				"last_updated_on" INTEGER,
				"archived_on" INTEGER DEFAULT NULL,
				CONSTRAINT username_unique UNIQUE (username)
			);`,
		},
		{
			Version:     2,
			Description: "create sessions table for session manager",
			Script: `
			CREATE TABLE sessions (
				token TEXT PRIMARY KEY,
				data BLOB NOT NULL,
				expiry REAL NOT NULL
			);

			CREATE INDEX sessions_expiry_idx ON sessions(expiry);
			`,
		},
		{
			Version:     3,
			Description: "create oauth2_clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS oauth2_clients (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"name" TEXT DEFAULT '',
				"client_id" TEXT NOT NULL,
				"client_secret" TEXT NOT NULL,
				"redirect_uri" TEXT DEFAULT '',
				"scopes" TEXT NOT NULL,
				"implicit_allowed" BOOLEAN NOT NULL DEFAULT 'false',
				"created_on" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				"last_updated_on" INTEGER,
				"archived_on" INTEGER DEFAULT NULL,
				"belongs_to_user" INTEGER NOT NULL,
				FOREIGN KEY(belongs_to_user) REFERENCES users(id)
			);`,
		},
		{
			Version:     4,
			Description: "create webhooks table",
			Script: `
			CREATE TABLE IF NOT EXISTS webhooks (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"name" TEXT NOT NULL,
				"content_type" TEXT NOT NULL,
				"url" TEXT NOT NULL,
				"method" TEXT NOT NULL,
				"events" TEXT NOT NULL,
				"data_types" TEXT NOT NULL,
				"topics" TEXT NOT NULL,
				"created_on" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				"last_updated_on" INTEGER,
				"archived_on" INTEGER DEFAULT NULL,
				"belongs_to_user" INTEGER NOT NULL,
				FOREIGN KEY(belongs_to_user) REFERENCES users(id)
			);`,
		},
		{
			Version:     5,
			Description: "create audit log table",
			Script: `
			CREATE TABLE IF NOT EXISTS audit_log (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"event_type" TEXT NOT NULL,
				"context" JSON NOT NULL,
				"created_on" BIGINT NOT NULL DEFAULT (strftime('%s','now'))
			);`,
		},
		{
			Version:     6,
			Description: "create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"name" CHARACTER VARYING NOT NULL,
				"details" CHARACTER VARYING NOT NULL DEFAULT '',
				"created_on" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				"last_updated_on" INTEGER DEFAULT NULL,
				"archived_on" INTEGER DEFAULT NULL,
				"belongs_to_user" INTEGER NOT NULL,
				FOREIGN KEY(belongs_to_user) REFERENCES users(id)
			);`,
		},
	}
)

// buildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a sqlite database.
func buildMigrationFunc(db *sql.DB) func() {
	return func() {
		driver := darwin.NewGenericDriver(db, darwin.SqliteDialect{})
		if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
			panic(fmt.Errorf("error migrating database: %w", err))
		}
	}
}

// Migrate migrates the database. It does so by invoking the migrateOnce function via sync.Once, so it should be
// safe (as in idempotent, though not necessarily recommended) to call this function multiple times.
func (q *Sqlite) Migrate(ctx context.Context, authenticator password.Authenticator, testUserConfig *database.UserCreationConfig) error {
	q.logger.Info("migrating db")

	if !q.IsReady(ctx) {
		return database.ErrDBUnready
	}

	q.migrateOnce.Do(buildMigrationFunc(q.db))

	if testUserConfig != nil {
		hp, err := authenticator.HashPassword(ctx, testUserConfig.Password)
		if err != nil {
			return err
		}

		query, args, err := q.sqlBuilder.
			Insert(queriers.UsersTableName).
			Columns(
				queriers.UsersTableUsernameColumn,
				queriers.UsersTableHashedPasswordColumn,
				queriers.UsersTableSaltColumn,
				queriers.UsersTableTwoFactorColumn,
				queriers.UsersTableIsAdminColumn,
				queriers.UsersTableAdminPermissionsColumn,
				queriers.UsersTableTwoFactorVerifiedOnColumn,
			).
			Values(
				testUserConfig.Username,
				hp,
				[]byte("aaaaaaaaaaaaaaaa"),
				// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
				"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				testUserConfig.IsAdmin,
				math.MaxUint32,
				squirrel.Expr(currentUnixTimeQuery),
			).
			ToSql()
		q.logQueryBuildingError(err)

		if _, dbErr := q.db.ExecContext(ctx, query, args...); dbErr != nil {
			return dbErr
		}

		q.logger.WithValue("username", testUserConfig.Username).Debug("created user")
	}

	return nil
}
