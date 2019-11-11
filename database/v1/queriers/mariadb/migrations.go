package mariadb

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/GuiaBolso/darwin"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "create users table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS users (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `username` VARCHAR(150) NOT NULL,",
				"    `hashed_password` VARCHAR(100) NOT NULL,",
				"    `password_last_changed_on` INTEGER UNSIGNED,",
				"    `two_factor_secret` VARCHAR(256) NOT NULL,",
				"    `is_admin` BOOLEAN NOT NULL DEFAULT false,",
				"    `created_on` BIGINT UNSIGNED,",
				"    `updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    PRIMARY KEY (`id`),",
				"    UNIQUE (`username`)",
				");",
			}, "\n"),
		},
		{
			Version:     2,
			Description: "create users table creation trigger",
			Script: strings.Join([]string{
				"CREATE TRIGGER IF NOT EXISTS users_creation_trigger BEFORE INSERT ON users FOR EACH ROW",
				"BEGIN",
				"  IF (new.created_on is null)",
				"  THEN",
				"    SET new.created_on = UNIX_TIMESTAMP();",
				"  END IF;",
				"END;",
			}, "\n"),
		},
		{
			Version:     3,
			Description: "create oauth2_clients table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS oauth2_clients (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `name` VARCHAR(128) DEFAULT '',",
				"    `client_id` VARCHAR(64) NOT NULL,",
				"    `client_secret` VARCHAR(64) NOT NULL,",
				"    `redirect_uri` VARCHAR(4096) DEFAULT '',",
				"    `scopes` VARCHAR(4096) NOT NULL,",
				"    `implicit_allowed` BOOLEAN NOT NULL DEFAULT false,",
				"    `created_on` BIGINT UNSIGNED,",
				"    `updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY(`belongs_to`) REFERENCES users(`id`)",
				");",
			}, "\n"),
		},
		{
			Version:     4,
			Description: "create oauth2_clients table creation trigger",
			Script: strings.Join([]string{
				"CREATE TRIGGER IF NOT EXISTS oauth2_clients_creation_trigger BEFORE INSERT ON oauth2_clients FOR EACH ROW",
				"BEGIN",
				"  IF (new.created_on is null)",
				"  THEN",
				"    SET new.created_on = UNIX_TIMESTAMP();",
				"  END IF;",
				"END;",
			}, "\n"),
		},
		{
			Version:     5,
			Description: "create webhooks table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS webhooks (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `name` VARCHAR(128) NOT NULL,",
				"    `content_type` VARCHAR(64) NOT NULL,",
				"    `url` VARCHAR(4096) NOT NULL,",
				"    `method` VARCHAR(32) NOT NULL,",
				"    `events` VARCHAR(256) NOT NULL,",
				"    `data_types` VARCHAR(256) NOT NULL,",
				"    `topics` VARCHAR(256) NOT NULL,",
				"    `created_on` BIGINT UNSIGNED,",
				"    `updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to`) REFERENCES users(`id`)",
				");",
			}, "\n"),
		},
		{
			Version:     6,
			Description: "create webhooks table creation trigger",
			Script: strings.Join([]string{
				"CREATE TRIGGER IF NOT EXISTS webhooks_creation_trigger BEFORE INSERT ON webhooks FOR EACH ROW",
				"BEGIN",
				"  IF (new.created_on is null)",
				"  THEN",
				"    SET new.created_on = UNIX_TIMESTAMP();",
				"  END IF;",
				"END;",
			}, "\n"),
		},
		{
			Version:     7,
			Description: "create items table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS items (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `name` LONGTEXT NOT NULL,",
				"    `details` LONGTEXT NOT NULL DEFAULT '',",
				"    `created_on` BIGINT UNSIGNED,",
				"    `updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to`) REFERENCES users(`id`)",
				");",
			}, "\n"),
		},
		{
			Version:     8,
			Description: "create items table creation trigger",
			Script: strings.Join([]string{
				"CREATE TRIGGER IF NOT EXISTS items_creation_trigger BEFORE INSERT ON items FOR EACH ROW",
				"BEGIN",
				"  IF (new.created_on is null)",
				"  THEN",
				"    SET new.created_on = UNIX_TIMESTAMP();",
				"  END IF;",
				"END;",
			}, "\n"),
		},
	}
)

// buildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a maria DB database
func buildMigrationFunc(db *sql.DB) func() {
	return func() {
		driver := darwin.NewGenericDriver(db, darwin.MySQLDialect{})
		if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
			panic(err)
		}
	}
}

// Migrate migrates the database. It does so by invoking the migrateOnce function via sync.Once, so it should be
// safe (as in idempotent, though not necessarily recommended) to call this function multiple times.
func (m *MariaDB) Migrate(ctx context.Context) error {
	m.logger.Info("migrating db")
	if !m.IsReady(ctx) {
		return errors.New("db is not ready yet")
	}

	m.migrateOnce.Do(buildMigrationFunc(m.db))

	return nil
}
