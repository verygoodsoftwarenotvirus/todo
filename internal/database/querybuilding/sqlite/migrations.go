package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/GuiaBolso/darwin"
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
				external_id TEXT NOT NULL,
				event_type TEXT NOT NULL,
				context JSON NOT NULL,
				created_on BIGINT NOT NULL DEFAULT (strftime('%s','now'))
			);`,
		},
		{
			Version:     0.03,
			Description: "create users table",
			Script: `
			CREATE TABLE IF NOT EXISTS users (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				external_id TEXT NOT NULL,
				username TEXT NOT NULL,
				avatar_src TEXT,
				hashed_password TEXT NOT NULL,
				password_last_changed_on INTEGER DEFAULT NULL,
				requires_password_change BOOLEAN NOT NULL DEFAULT 'false',
				two_factor_secret TEXT NOT NULL,
				two_factor_secret_verified_on INTEGER DEFAULT NULL,
				service_roles TEXT NOT NULL DEFAULT 'service_user',
				reputation TEXT NOT NULL DEFAULT 'unverified',
				reputation_explanation TEXT NOT NULL DEFAULT '',
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER DEFAULT NULL,
				archived_on INTEGER DEFAULT NULL,
				CONSTRAINT username_unique UNIQUE (username)
			);`,
		},
		{
			Version:     0.04,
			Description: "create accounts table",
			Script: `
			CREATE TABLE IF NOT EXISTS accounts (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				external_id TEXT NOT NULL,
				name TEXT NOT NULL,
				billing_status TEXT NOT NULL DEFAULT 'unpaid',
				contact_email TEXT NOT NULL DEFAULT '',
				contact_phone TEXT NOT NULL DEFAULT '',
				payment_processor_customer_id TEXT NOT NULL DEFAULT '',
				subscription_plan_id TEXT,
				belongs_to_user INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER DEFAULT NULL,
				archived_on INTEGER DEFAULT NULL,
				CONSTRAINT account_name_unique UNIQUE (name, belongs_to_user)
			);`,
		},
		{
			Version:     0.05,
			Description: "create account user memberships table",
			Script: `
			CREATE TABLE IF NOT EXISTS account_user_memberships (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				belongs_to_account INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
				belongs_to_user INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				account_roles TEXT NOT NULL DEFAULT 'account_user',
				default_account BOOLEAN NOT NULL DEFAULT 'false',
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER DEFAULT NULL,
				archived_on INTEGER DEFAULT NULL,
				CONSTRAINT plan_name_unique UNIQUE (belongs_to_account, belongs_to_user)
			);`,
		},
		{
			Version:     0.06,
			Description: "create API clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS api_clients (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				external_id TEXT NOT NULL,
				name TEXT DEFAULT '',
				client_id TEXT NOT NULL,
				secret_key TEXT NOT NULL,
				permissions INTEGER NOT NULL DEFAULT 0,
				admin_permissions INTEGER NOT NULL DEFAULT 0,
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER DEFAULT NULL,
				archived_on INTEGER DEFAULT NULL,
				belongs_to_user INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
			);`,
		},
		{
			Version:     0.07,
			Description: "create webhooks table",
			Script: `
			CREATE TABLE IF NOT EXISTS webhooks (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				external_id TEXT NOT NULL,
				name TEXT NOT NULL,
				content_type TEXT NOT NULL,
				url TEXT NOT NULL,
				method TEXT NOT NULL,
				events TEXT NOT NULL,
				data_types TEXT NOT NULL,
				topics TEXT NOT NULL,
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER DEFAULT NULL,
				archived_on INTEGER DEFAULT NULL,
				belongs_to_account INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE
			);`,
		},
		{
			Version:     0.08,
			Description: "create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				external_id TEXT NOT NULL,
				name TEXT NOT NULL,
				details TEXT NOT NULL DEFAULT '',
				created_on INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				last_updated_on INTEGER DEFAULT NULL,
				archived_on INTEGER DEFAULT NULL,
				belongs_to_account INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE
			);`,
		},
	}
)

// BuildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a sqlite database.
func (b *Sqlite) BuildMigrationFunc(db *sql.DB) func() {
	return func() {
		d := darwin.NewGenericDriver(db, darwin.SqliteDialect{})
		if err := darwin.Migrate(d, migrations, nil); err != nil {
			panic(fmt.Errorf("migrating database: %w", err))
		}
	}
}
