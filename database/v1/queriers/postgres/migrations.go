package postgres

import (
	"context"
	"database/sql"

	"github.com/GuiaBolso/darwin"
	"github.com/pkg/errors"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "create users table",
			Script: `
			CREATE TABLE IF NOT EXISTS users (
				"id" bigserial NOT NULL PRIMARY KEY,
				"username" text NOT NULL,
				"hashed_password" text NOT NULL,
				"password_last_changed_on" integer,
				"two_factor_secret" text NOT NULL,
				"is_admin" boolean NOT NULL DEFAULT 'false',
				"created_on" bigint NOT NULL DEFAULT extract(epoch FROM NOW()),
				"updated_on" bigint DEFAULT NULL,
				"archived_on" bigint DEFAULT NULL,
				UNIQUE ("username")
			);`,
		},
		{
			Version:     2,
			Description: "Add oauth2_clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS oauth2_clients (
				"id" bigserial NOT NULL PRIMARY KEY,
				"name" text DEFAULT '',
				"client_id" text NOT NULL,
				"client_secret" text NOT NULL,
				"redirect_uri" text DEFAULT '',
				"scopes" text NOT NULL,
				"implicit_allowed" boolean NOT NULL DEFAULT 'false',
				"created_on" bigint NOT NULL DEFAULT extract(epoch FROM NOW()),
				"updated_on" bigint DEFAULT NULL,
				"archived_on" bigint DEFAULT NULL,
				"belongs_to" bigint NOT NULL,
				FOREIGN KEY(belongs_to) REFERENCES users(id)
			);`,
		},
		{
			Version:     3,
			Description: "create webhooks table",
			Script: `
			CREATE TABLE IF NOT EXISTS webhooks (
				"id" bigserial NOT NULL PRIMARY KEY,
				"name" text NOT NULL,
				"content_type" text NOT NULL,
				"url" text NOT NULL,
				"method" text NOT NULL,
				"events" text NOT NULL,
				"data_types" text NOT NULL,
				"topics" text NOT NULL,
				"created_on" bigint NOT NULL DEFAULT extract(epoch FROM NOW()),
				"updated_on" bigint DEFAULT NULL,
				"archived_on" bigint DEFAULT NULL,
				"belongs_to" bigint NOT NULL,
				FOREIGN KEY ("belongs_to") REFERENCES "users"("id")
			);`,
		},
		{
			Version:     4,
			Description: "create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				"id" bigserial NOT NULL PRIMARY KEY,
				"name" text NOT NULL,
				"details" text NOT NULL DEFAULT '',
				"created_on" bigint NOT NULL DEFAULT extract(epoch FROM NOW()),
				"updated_on" bigint DEFAULT NULL,
				"archived_on" bigint DEFAULT NULL,
				"belongs_to" bigint NOT NULL,
				FOREIGN KEY ("belongs_to") REFERENCES "users"("id")
			);`,
		},
	}
)

// buildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a postgres database
func buildMigrationFunc(db *sql.DB) func() {
	return func() {
		driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})
		if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
			panic(err)
		}
	}
}

// Migrate migrates the database. It does so by invoking the migrateOnce function via sync.Once, so it should be
// safe (as in idempotent, though not recommended) to call this function multiple times.
func (p *Postgres) Migrate(ctx context.Context) error {
	p.logger.Info("migrating db")
	if !p.IsReady(ctx) {
		return errors.New("db is not ready yet")
	}

	p.migrateOnce.Do(buildMigrationFunc(p.db))

	return nil
}
