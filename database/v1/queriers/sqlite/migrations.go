package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GuiaBolso/darwin"
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
				"password_last_changed_on" INTEGER,
				"two_factor_secret" TEXT NOT NULL,
				"is_admin" BOOLEAN NOT NULL DEFAULT 'false',
				"created_on" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				"updated_on" INTEGER,
				"archived_on" INTEGER DEFAULT NULL,
				CONSTRAINT username_unique UNIQUE (username)
			);`,
		},
		{
			Version:     2,
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
				"updated_on" INTEGER,
				"archived_on" INTEGER DEFAULT NULL,
				"belongs_to" INTEGER NOT NULL,
				FOREIGN KEY(belongs_to) REFERENCES users(id)
			);`,
		},
		{
			Version:     3,
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
				"updated_on" INTEGER,
				"archived_on" INTEGER DEFAULT NULL,
				"belongs_to" INTEGER NOT NULL,
				FOREIGN KEY(belongs_to) REFERENCES users(id)
			);`,
		},
		{
			Version:     4,
			Description: "create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"name" CHARACTER VARYING NOT NULL,
				"details" CHARACTER VARYING NOT NULL DEFAULT '',
				"created_on" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				"updated_on" INTEGER DEFAULT NULL,
				"archived_on" INTEGER DEFAULT NULL,
				"belongs_to" INTEGER NOT NULL,
				FOREIGN KEY(belongs_to) REFERENCES users(id)
			);`,
		},
	}
)

// buildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a sqlite database
func buildMigrationFunc(db *sql.DB) func() {
	return func() {
		driver := darwin.NewGenericDriver(db, darwin.SqliteDialect{})
		if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
			panic(err)
		}
	}
}

// Migrate migrates the database. It does so by invoking the migrateOnce function via sync.Once, so it should be
// safe (as in idempotent, though not necessarily recommended) to call this function multiple times.
func (s *Sqlite) Migrate(ctx context.Context) error {
	s.logger.Info("migrating db")
	if !s.IsReady(ctx) {
		return errors.New("db is not ready yet")
	}

	s.migrateOnce.Do(buildMigrationFunc(s.db))

	return nil
}
