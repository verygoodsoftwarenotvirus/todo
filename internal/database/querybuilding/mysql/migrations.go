package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/GuiaBolso/darwin"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     0.0,
			Description: "create sessions table for session manager",
			Script: strings.Join([]string{
				"CREATE TABLE sessions (",
				"`token` CHAR(43) PRIMARY KEY,",
				"`data` BLOB NOT NULL,",
				"`expiry` TIMESTAMP(6) NOT NULL,",
				"`created_on` BIGINT UNSIGNED",
				");",
			}, "\n"),
		},
		{
			Version:     0.01,
			Description: "create sessions table for session manager",
			Script:      "CREATE INDEX sessions_expiry_idx ON sessions (expiry);",
		},
		{
			Version:     0.03,
			Description: "create audit log table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS audit_log (",
				"    `id` CHAR(27) NOT NULL,",
				"    `event_type` VARCHAR(256) NOT NULL,",
				"    `context` JSON NOT NULL,",
				"    `created_on` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.05,
			Description: "create users table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS users (",
				"    `id` CHAR(27) NOT NULL,",
				"    `username` VARCHAR(128) NOT NULL,",
				"    `avatar_src` LONGTEXT NOT NULL,",
				"    `hashed_password` VARCHAR(100) NOT NULL,",
				"    `requires_password_change` BOOLEAN NOT NULL DEFAULT false,",
				"    `password_last_changed_on` INTEGER UNSIGNED,",
				"    `two_factor_secret` VARCHAR(256) NOT NULL,",
				"    `two_factor_secret_verified_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `service_roles` LONGTEXT NOT NULL,",
				"    `reputation` VARCHAR(64) NOT NULL,",
				"    `reputation_explanation` VARCHAR(1024) NOT NULL,",
				"    `created_on` BIGINT UNSIGNED NOT NULL,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    PRIMARY KEY (`id`),",
				"    UNIQUE (`username`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.07,
			Description: "create accounts table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS accounts (",
				"    `id` CHAR(27) NOT NULL,",
				"    `name` LONGTEXT NOT NULL,",
				"    `billing_status` TEXT NOT NULL,",
				"    `contact_email` TEXT NOT NULL,",
				"    `contact_phone` TEXT NOT NULL,",
				"    `payment_processor_customer_id` TEXT NOT NULL,",
				"    `subscription_plan_id` VARCHAR(128),",
				"    `created_on` BIGINT UNSIGNED NOT NULL,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_user` CHAR(27) NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to_user`) REFERENCES users(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.09,
			Description: "create account user memberships table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS account_user_memberships (",
				"    `id` CHAR(27) NOT NULL,",
				"    `belongs_to_account` CHAR(27) NOT NULL,",
				"    `belongs_to_user` CHAR(27) NOT NULL,",
				"    `default_account` BOOLEAN NOT NULL DEFAULT false,",
				"    `account_roles` LONGTEXT NOT NULL,",
				"    `created_on` BIGINT UNSIGNED NOT NULL,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to_user`) REFERENCES users(`id`) ON DELETE CASCADE,",
				"    FOREIGN KEY (`belongs_to_account`) REFERENCES accounts(`id`) ON DELETE CASCADE,",
				"    UNIQUE (`belongs_to_account`, `belongs_to_user`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.11,
			Description: "create API clients table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS api_clients (",
				"    `id` CHAR(27) NOT NULL,",
				"    `name` VARCHAR(128),",
				"    `client_id` VARCHAR(64) NOT NULL,",
				"    `secret_key` BINARY(128) NOT NULL,",
				"    `account_roles` LONGTEXT NOT NULL,",
				"    `for_admin` BOOLEAN NOT NULL DEFAULT false,",
				"    `created_on` BIGINT UNSIGNED NOT NULL,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_user` CHAR(27) NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    UNIQUE (`name`, `belongs_to_user`),",
				"    FOREIGN KEY (`belongs_to_user`) REFERENCES users(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.13,
			Description: "create webhooks table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS webhooks (",
				"    `id` CHAR(27) NOT NULL,",
				"    `name` VARCHAR(128) NOT NULL,",
				"    `content_type` VARCHAR(64) NOT NULL,",
				"    `url` LONGTEXT NOT NULL,",
				"    `method` VARCHAR(8) NOT NULL,",
				"    `events` VARCHAR(256) NOT NULL,",
				"    `data_types` VARCHAR(256) NOT NULL,",
				"    `topics` VARCHAR(256) NOT NULL,",
				"    `created_on` BIGINT UNSIGNED NOT NULL,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_account` CHAR(27) NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to_account`) REFERENCES accounts(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.15,
			Description: "create items table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS items (",
				"    `id` CHAR(27) NOT NULL,",
				"    `name` LONGTEXT NOT NULL,",
				"    `details` LONGTEXT NOT NULL,",
				"    `created_on` BIGINT UNSIGNED NOT NULL,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_account` CHAR(27) NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to_account`) REFERENCES accounts(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
	}
)

// BuildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a maria DB database.
func (b *MySQL) BuildMigrationFunc(db *sql.DB) func() {
	return func() {
		driver := darwin.NewGenericDriver(db, darwin.MySQLDialect{})
		if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
			panic(fmt.Errorf("migrating database: %w", err))
		}
	}
}
