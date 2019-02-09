// +build !migrations

package postgres

import (
	"context"

	"github.com/GuiaBolso/darwin"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Create user table",
			Script: `
			CREATE TABLE IF NOT EXISTS users (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"username" text NOT NULL,
				"hashed_password" text NOT NULL,
				"password_last_changed_on" timestamp,
				"two_factor_secret" text NOT NULL,
				"is_admin" bool NOT NULL DEFAULT 'false',
				"created_on" timestamp NOT NULL DEFAULT to_timestamp(extract(epoch FROM NOW())),
				"updated_on" timestamp,
				"archived_on" timestamp,
				UNIQUE ("username")
			);`,
		},
		{
			Version:     2,
			Description: "Add OAuth2 Clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS oauth_clients (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"client_id" TEXT NOT NULL,
				"client_secret" TEXT NOT NULL,
				"redirect_uri" TEXT DEFAULT '',
				"scopes" TEXT NOT NULL,
				"implicit_allowed" BOOLEAN NOT NULL DEFAULT 'false',
				"created_on" timestamp NOT NULL DEFAULT to_timestamp(extract(epoch FROM NOW())),
				"updated_on" timestamp DEFAULT NULL,
				"archived_on" timestamp DEFAULT NULL,
				"belongs_to" INTEGER NOT NULL,
				FOREIGN KEY(belongs_to) REFERENCES users(id)
			);`,
		},
		{
			Version:     3,
			Description: "Create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"name" text NOT NULL,
				"details" TEXT NOT NULL DEFAULT '',
				"created_on" timestamp NOT NULL DEFAULT NOW(),
				"updated_on" timestamp,
				"archived_on" timestamp,
				"belongs_to" bigint NOT NULL,
				FOREIGN KEY ("belongs_to") REFERENCES "users"("id")
			);`,
		},
	}
)

// Migrate migrates a postgres database
func (p *Postgres) Migrate(context.Context) error {
	driver := darwin.NewGenericDriver(p.database, darwin.PostgresDialect{})
	return darwin.New(driver, migrations, nil).Migrate()
}
