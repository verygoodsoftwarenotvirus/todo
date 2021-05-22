package postgres

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
			Description: "create audit log table",
			Script: `
			CREATE TABLE IF NOT EXISTS audit_log (
				id BIGSERIAL NOT NULL PRIMARY KEY,
				external_id TEXT NOT NULL,
				event_type TEXT NOT NULL,
				context JSONB NOT NULL,
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW())
			);`,
		},
		{
			Version:     0.03,
			Description: "create users table",
			Script: `
			CREATE TABLE IF NOT EXISTS users (
				id BIGSERIAL NOT NULL PRIMARY KEY,
				external_id TEXT NOT NULL,
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
			Version:     0.04,
			Description: "create accounts table",
			Script: `
			CREATE TABLE IF NOT EXISTS accounts (
				id BIGSERIAL NOT NULL PRIMARY KEY,
				external_id TEXT NOT NULL,
				name TEXT NOT NULL,
				billing_status TEXT NOT NULL DEFAULT 'unpaid',
				contact_email TEXT NOT NULL DEFAULT '',
				contact_phone TEXT NOT NULL DEFAULT '',
				payment_processor_customer_id TEXT NOT NULL DEFAULT '',
				subscription_plan_id TEXT,
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				belongs_to_user BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				UNIQUE("belongs_to_user", "name")
			);`,
		},
		{
			Version:     0.05,
			Description: "create account user memberships table",
			Script: `
			CREATE TABLE IF NOT EXISTS account_user_memberships (
				id BIGSERIAL NOT NULL PRIMARY KEY,
				belongs_to_account BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
				belongs_to_user BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				default_account BOOLEAN NOT NULL DEFAULT 'false',
				account_roles TEXT NOT NULL DEFAULT 'account_user',
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				UNIQUE("belongs_to_account", "belongs_to_user")
			);`,
		},
		{
			Version:     0.06,
			Description: "create API clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS api_clients (
				id BIGSERIAL NOT NULL PRIMARY KEY,
				external_id TEXT NOT NULL,
				name TEXT DEFAULT '',
				client_id TEXT NOT NULL,
				secret_key BYTEA NOT NULL,
				permissions BIGINT NOT NULL DEFAULT 0,
				admin_permissions BIGINT NOT NULL DEFAULT 0,
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				belongs_to_user BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE
			);`,
		},
		{
			Version:     0.07,
			Description: "create webhooks table",
			Script: `
			CREATE TABLE IF NOT EXISTS webhooks (
				id BIGSERIAL NOT NULL PRIMARY KEY,
				external_id TEXT NOT NULL,
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
				belongs_to_account INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE
			);`,
		},
		{
			Version:     0.8,
			Description: "create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				id BIGSERIAL NOT NULL PRIMARY KEY,
				external_id TEXT NOT NULL,
				name TEXT NOT NULL,
				details TEXT NOT NULL DEFAULT '',
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				belongs_to_account INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE
			);`,
		},
	}
)

// BuildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a postgres database.
func (b *Postgres) BuildMigrationFunc(db *sql.DB) func() {
	return func() {
		driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})
		if err := darwin.Migrate(driver, migrations, nil); err != nil {
			panic(fmt.Errorf("migrating database: %w", err))
		}
	}
}
