// +build !migrations

package sqlite

import (
	"context"

	"github.com/GuiaBolso/darwin"
	"github.com/pkg/errors"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Create user table",
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
			Description: "Add OAuth2 Clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS oauth_clients (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
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
			Description: "Create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
				"name" TEXT NOT NULL,
				"details" TEXT,
				"created_on" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
				"updated_on" INTEGER,
				"completed_on" INTEGER DEFAULT NULL,
				"belongs_to" INTEGER NOT NULL,
				FOREIGN KEY(belongs_to) REFERENCES users(id)
			);`,
		},
	}
)

// Migrate migrates a given Sqlite database.
func (s *Sqlite) Migrate(context.Context) error {
	driver := darwin.NewGenericDriver(s.database, darwin.SqliteDialect{})

	d := darwin.New(driver, migrations, nil)
	err := d.Migrate()

	if err != nil {
		return errors.Wrap(err, "migrating sqlite database")
	}
	return nil
}
