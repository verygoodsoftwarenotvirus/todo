package mariadb

import (
	"database/sql"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"

	"github.com/GuiaBolso/darwin"
)

func buildCreationTriggerScript(tableName string) string {
	return strings.Join([]string{
		fmt.Sprintf("CREATE TRIGGER IF NOT EXISTS %s_creation_trigger BEFORE INSERT ON %s FOR EACH ROW", tableName, tableName),
		"BEGIN",
		"  IF (new.created_on is null)",
		"  THEN",
		"    SET new.created_on = UNIX_TIMESTAMP(NOW());",
		"  END IF;",
		"END;",
	}, "\n")
}

var (
	migrations = []darwin.Migration{
		{
			Version:     0.00,
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
			Version:     0.02,
			Description: "create sessions table creation trigger",
			Script:      buildCreationTriggerScript("sessions"),
		},
		{
			Version:     0.03,
			Description: "create audit log table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS audit_log (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `external_id` VARCHAR(36) NOT NULL,",
				"    `event_type` VARCHAR(256) NOT NULL,",
				"    `context` JSON NOT NULL,",
				"    `created_on` BIGINT UNSIGNED,",
				"    PRIMARY KEY (`id`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.04,
			Description: "create audit_log table creation trigger",
			Script:      buildCreationTriggerScript(querybuilding.AuditLogEntriesTableName),
		},
		{
			Version:     0.05,
			Description: "create account subscription plans table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS account_subscription_plans (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `external_id` VARCHAR(36) NOT NULL,",
				"    `name` VARCHAR(128) NOT NULL,",
				"    `description` VARCHAR(128) NOT NULL DEFAULT '',",
				"    `price` INT UNSIGNED NOT NULL,",
				"    `period` VARCHAR(128) NOT NULL DEFAULT '0m0s',",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    PRIMARY KEY (`id`),",
				"    UNIQUE (`name`, `archived_on`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.06,
			Description: "create account subscription plans table creation trigger",
			Script:      buildCreationTriggerScript(querybuilding.AccountSubscriptionPlansTableName),
		},
		{
			Version:     0.07,
			Description: "create users table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS users (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `external_id` VARCHAR(36) NOT NULL,",
				"    `username` VARCHAR(128) NOT NULL,",
				"    `avatar_src` LONGTEXT NOT NULL DEFAULT '',",
				"    `hashed_password` VARCHAR(100) NOT NULL,",
				"    `requires_password_change` BOOLEAN NOT NULL DEFAULT false,",
				"    `password_last_changed_on` INTEGER UNSIGNED,",
				"    `two_factor_secret` VARCHAR(256) NOT NULL,",
				"    `two_factor_secret_verified_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `service_roles` LONGTEXT NOT NULL DEFAULT 'service_user',",
				"    `reputation` VARCHAR(64) NOT NULL DEFAULT 'unverified',",
				"    `reputation_explanation` VARCHAR(1024) NOT NULL DEFAULT '',",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    PRIMARY KEY (`id`),",
				"    UNIQUE (`username`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.08,
			Description: "create users table creation trigger",
			Script:      buildCreationTriggerScript(querybuilding.UsersTableName),
		},
		{
			Version:     0.09,
			Description: "create accounts table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS accounts (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `external_id` VARCHAR(36) NOT NULL,",
				"    `name` LONGTEXT NOT NULL,",
				"    `plan_id` BIGINT UNSIGNED,",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_user` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`plan_id`) REFERENCES account_subscription_plans(`id`) ON DELETE RESTRICT,",
				"    FOREIGN KEY (`belongs_to_user`) REFERENCES users(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.10,
			Description: "create accounts table creation trigger",
			Script:      buildCreationTriggerScript(querybuilding.AccountsTableName),
		},
		{
			Version:     0.11,
			Description: "create account user memberships table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS account_user_memberships (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `belongs_to_account` BIGINT UNSIGNED NOT NULL,",
				"    `belongs_to_user` BIGINT UNSIGNED NOT NULL,",
				"    `default_account` BOOLEAN NOT NULL DEFAULT false,",
				"    `account_roles` LONGTEXT NOT NULL DEFAULT 'account_user',",
				"    `created_on` BIGINT UNSIGNED,",
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
			Version:     0.12,
			Description: "create accounts membership creation trigger",
			Script:      buildCreationTriggerScript("account_user_memberships"),
		},
		{
			Version:     0.13,
			Description: "create API clients table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS api_clients (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `external_id` VARCHAR(36) NOT NULL,",
				"    `name` VARCHAR(128) DEFAULT '',",
				"    `client_id` VARCHAR(64) NOT NULL,",
				"    `secret_key` BINARY(128) NOT NULL,",
				"    `account_roles` LONGTEXT NOT NULL DEFAULT 'account_member',",
				"    `for_admin` BOOLEAN NOT NULL DEFAULT false,",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_user` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    UNIQUE (`name`, `belongs_to_user`),",
				"    FOREIGN KEY (`belongs_to_user`) REFERENCES users(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.14,
			Description: "create api_clients table creation trigger",
			Script:      buildCreationTriggerScript(querybuilding.APIClientsTableName),
		},
		{
			Version:     0.15,
			Description: "create webhooks table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS webhooks (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `external_id` VARCHAR(36) NOT NULL,",
				"    `name` VARCHAR(128) NOT NULL,",
				"    `content_type` VARCHAR(64) NOT NULL,",
				"    `url` LONGTEXT NOT NULL,",
				"    `method` VARCHAR(8) NOT NULL,",
				"    `events` VARCHAR(256) NOT NULL,",
				"    `data_types` VARCHAR(256) NOT NULL,",
				"    `topics` VARCHAR(256) NOT NULL,",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_account` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to_account`) REFERENCES accounts(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.16,
			Description: "create webhooks table creation trigger",
			Script:      buildCreationTriggerScript(querybuilding.WebhooksTableName),
		},
		{
			Version:     0.17,
			Description: "create items table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS items (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `external_id` VARCHAR(36) NOT NULL,",
				"    `name` LONGTEXT NOT NULL,",
				"    `details` LONGTEXT NOT NULL DEFAULT '',",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_account` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to_account`) REFERENCES accounts(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.18,
			Description: "create items table creation trigger",
			Script:      buildCreationTriggerScript(querybuilding.ItemsTableName),
		},
	}
)

// BuildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a maria DB database.
func (b *MariaDB) BuildMigrationFunc(db *sql.DB) func() {
	return func() {
		driver := darwin.NewGenericDriver(db, darwin.MySQLDialect{})
		if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
			panic(fmt.Errorf("migrating database: %w", err))
		}
	}
}
