package sqlite

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
			Description: "create sessions table for session manager",
			Script: `
			CREATE TABLE sessions (
				token TEXT PRIMARY KEY,
				data BLOB NOT NULL,
				expiry REAL NOT NULL,
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now'))
			);`,
		},
		{
			Version:     0.01,
			Description: "create sessions table for session manager",
			Script:      `CREATE INDEX sessions_expiry_idx ON sessions(expiry);`,
		},
		{
			Version:     0.02,
			Description: "create audit log table",
			Script: `
			CREATE TABLE IF NOT EXISTS audit_log (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				event_type TEXT NOT NULL,
				context JSON NOT NULL,
				created_on BIGINT NOT NULL DEFAULT (strftime('%s','now'))
			);`,
		},
		{
			Version:     0.03,
			Description: "create account subscription plans table and default plan",
			Script: `
			CREATE TABLE IF NOT EXISTS account_subscription_plans (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				description TEXT NOT NULL DEFAULT '',
				price INTEGER NOT NULL,
				period TEXT NOT NULL DEFAULT '0m0s',
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER,
				archived_on INTEGER DEFAULT NULL,
				CONSTRAINT plan_name_unique UNIQUE (name, archived_on)
			);`,
		},
		{
			Version:     0.04,
			Description: "create users table",
			Script: `
			CREATE TABLE IF NOT EXISTS users (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				username TEXT NOT NULL,
				avatar_src TEXT,
				hashed_password TEXT NOT NULL,
				salt TINYBLOB NOT NULL,
				password_last_changed_on INTEGER,
				requires_password_change BOOLEAN NOT NULL DEFAULT 'false',
				two_factor_secret TEXT NOT NULL,
				two_factor_secret_verified_on INTEGER DEFAULT NULL,
				is_site_admin BOOLEAN NOT NULL DEFAULT 'false',
				site_admin_permissions INTEGER NOT NULL DEFAULT 0,
				reputation TEXT NOT NULL DEFAULT 'unverified',
				reputation_explanation TEXT NOT NULL DEFAULT '',
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER,
				archived_on INTEGER DEFAULT NULL,
				CONSTRAINT username_unique UNIQUE (username, archived_on)
			);`,
		},
		{
			Version:     0.05,
			Description: "create accounts table",
			Script: `
			CREATE TABLE IF NOT EXISTS accounts (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				name CHARACTER VARYING NOT NULL,
				plan_id BIGINT REFERENCES account_subscription_plans(id) ON DELETE RESTRICT,
				belongs_to_user INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now'))
			);`,
		},
		{
			Version:     0.06,
			Description: "create accounts membership table",
			Script: `
			CREATE TABLE IF NOT EXISTS accounts_membership (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				primary_user_account BOOLEAN NOT NULL DEFAULT 'false',
				belongs_to_account INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
				belongs_to_user INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				archived_on INTEGER DEFAULT NULL
			);`,
		},
		{
			Version:     0.07,
			Description: "create oauth2_clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS oauth2_clients (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				name TEXT DEFAULT '',
				client_id TEXT NOT NULL,
				client_secret TEXT NOT NULL,
				redirect_uri TEXT DEFAULT '',
				scopes TEXT NOT NULL,
				implicit_allowed BOOLEAN NOT NULL DEFAULT 'false',
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER,
				archived_on INTEGER DEFAULT NULL,
				belongs_to_user INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
			);`,
		},
		{
			Version:     0.08,
			Description: "create webhooks table",
			Script: `
			CREATE TABLE IF NOT EXISTS webhooks (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				content_type TEXT NOT NULL,
				url TEXT NOT NULL,
				method TEXT NOT NULL,
				events TEXT NOT NULL,
				data_types TEXT NOT NULL,
				topics TEXT NOT NULL,
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER,
				archived_on INTEGER DEFAULT NULL,
				belongs_to_user INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
			);`,
		},
		{
			Version:     0.09,
			Description: "create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				name CHARACTER VARYING NOT NULL,
				details CHARACTER VARYING NOT NULL DEFAULT '',
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER DEFAULT NULL,
				archived_on INTEGER DEFAULT NULL,
				belongs_to_user INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
			);`,
		},
	}
)

// buildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a sqlite database.
func buildMigrationFunc(db *sql.DB) func() {
	return func() {
		d := darwin.NewGenericDriver(db, darwin.SqliteDialect{})
		if err := darwin.Migrate(d, migrations, nil); err != nil {
			panic(fmt.Errorf("migrating database: %w", err))
		}
	}
}

// Migrate migrates the database. It does so by invoking the migrateOnce function via sync.Once, so it should be
// safe (as in idempotent, though not necessarily recommended) to call this function multiple times.
func (c *Sqlite) Migrate(ctx context.Context, testUserConfig *types.TestUserCreationConfig) error {
	c.logger.Info("migrating db")

	if !c.IsReady(ctx, 50) {
		return database.ErrDBUnready
	}

	c.migrateOnce.Do(buildMigrationFunc(c.db))

	if testUserConfig != nil {
		query, args, err := c.sqlBuilder.
			Insert(queriers.UsersTableName).
			Columns(
				queriers.UsersTableUsernameColumn,
				queriers.UsersTableHashedPasswordColumn,
				queriers.UsersTableSaltColumn,
				queriers.UsersTableTwoFactorColumn,
				queriers.UsersTableIsAdminColumn,
				queriers.UsersTableReputationColumn,
				queriers.UsersTableAdminPermissionsColumn,
				queriers.UsersTableTwoFactorVerifiedOnColumn,
			).
			Values(
				testUserConfig.Username,
				testUserConfig.HashedPassword,
				[]byte("aaaaaaaaaaaaaaaa"),
				// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
				"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				testUserConfig.IsSiteAdmin,
				types.GoodStandingAccountStatus,
				math.MaxUint32,
				squirrel.Expr(currentUnixTimeQuery),
			).
			ToSql()
		c.logQueryBuildingError(err)

		res, userCreateErr := c.db.ExecContext(ctx, query, args...)
		if userCreateErr != nil {
			c.logger.Error(userCreateErr, "creating test user")
			return fmt.Errorf("creating test user: %w", userCreateErr)
		}

		id, idRetErr := res.LastInsertId()
		if idRetErr != nil {
			c.logger.Error(idRetErr, "fetching insert ID")
			return fmt.Errorf("fetching insert ID: %w", idRetErr)
		}

		if _, accountCreationErr := c.CreateAccount(ctx, &types.AccountCreationInput{
			Name:          testUserConfig.Username,
			BelongsToUser: uint64(id),
		}); accountCreationErr != nil {
			c.logger.Error(accountCreationErr, "creating test user")
			return fmt.Errorf("creating test user: %w", accountCreationErr)
		}

		c.logger.WithValue(keys.UsernameKey, testUserConfig.Username).Debug("created test user and account")
	}

	return nil
}
