package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/GuiaBolso/darwin"
	"github.com/Masterminds/squirrel"
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
			Script:      buildCreationTriggerScript(queriers.AuditLogEntriesTableName),
		},
		{
			Version:     0.05,
			Description: "create account subscription plans table and default plan",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS account_subscription_plans (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
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
			Script:      buildCreationTriggerScript(queriers.AccountSubscriptionPlansTableName),
		},
		{
			Version:     0.07,
			Description: "create users table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS users (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `username` VARCHAR(128) NOT NULL,",
				"    `avatar_src` VARCHAR(4096) NOT NULL DEFAULT '',",
				"    `hashed_password` VARCHAR(100) NOT NULL,",
				"    `salt` BINARY(16) NOT NULL,",
				"    `requires_password_change` BOOLEAN NOT NULL DEFAULT false,",
				"    `password_last_changed_on` INTEGER UNSIGNED,",
				"    `two_factor_secret` VARCHAR(256) NOT NULL,",
				"    `two_factor_secret_verified_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `is_site_admin` BOOLEAN NOT NULL DEFAULT false,",
				"    `site_admin_permissions` BIGINT NOT NULL DEFAULT 0,",
				"    `reputation` VARCHAR(64) NOT NULL DEFAULT 'unverified',",
				"    `reputation_explanation` VARCHAR(1024) NOT NULL DEFAULT '',",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    PRIMARY KEY (`id`),",
				"    UNIQUE (`username`, `archived_on`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.08,
			Description: "create users table creation trigger",
			Script:      buildCreationTriggerScript(queriers.UsersTableName),
		},
		{
			Version:     0.09,
			Description: "create accounts table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS accounts (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
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
			Script:      buildCreationTriggerScript(queriers.AccountsTableName),
		},
		{
			Version:     0.11,
			Description: "create accounts membership table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS accounts_membership (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `belongs_to_account` BIGINT UNSIGNED NOT NULL,",
				"    `belongs_to_user` BIGINT UNSIGNED NOT NULL,",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to_user`) REFERENCES account_subscription_plans(`id`) ON DELETE CASCADE,",
				"    FOREIGN KEY (`belongs_to_account`) REFERENCES account_subscription_plans(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.12,
			Description: "create accounts membership creation trigger",
			Script:      buildCreationTriggerScript("accounts_membership"),
		},
		{
			Version:     0.13,
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
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_user` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY(`belongs_to_user`) REFERENCES users(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.14,
			Description: "create oauth2_clients table creation trigger",
			Script:      buildCreationTriggerScript(queriers.OAuth2ClientsTableName),
		},
		{
			Version:     0.15,
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
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_user` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to_user`) REFERENCES users(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.16,
			Description: "create webhooks table creation trigger",
			Script:      buildCreationTriggerScript(queriers.WebhooksTableName),
		},
		{
			Version:     0.17,
			Description: "create items table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS items (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `name` LONGTEXT NOT NULL,",
				"    `details` LONGTEXT NOT NULL DEFAULT '',",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `belongs_to_user` BIGINT UNSIGNED NOT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY (`belongs_to_user`) REFERENCES users(`id`) ON DELETE CASCADE",
				");",
			}, "\n"),
		},
		{
			Version:     0.18,
			Description: "create items table creation trigger",
			Script:      buildCreationTriggerScript(queriers.ItemsTableName),
		},
	}
)

// buildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a maria DB database.
func buildMigrationFunc(db *sql.DB) func() {
	return func() {
		driver := darwin.NewGenericDriver(db, darwin.MySQLDialect{})
		if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
			panic(fmt.Errorf("error migrating database: %w", err))
		}
	}
}

// Migrate migrates the database. It does so by invoking the migrateOnce function via sync.Once, so it should be
// safe (as in idempotent, though not necessarily recommended) to call this function multiple times.
func (q *MariaDB) Migrate(ctx context.Context, testUserConfig *types.TestUserCreationConfig) error {
	q.logger.Info("migrating db")

	if !q.IsReady(ctx) {
		return database.ErrDBUnready
	}

	q.migrateOnce.Do(buildMigrationFunc(q.db))

	if testUserConfig != nil {
		query, args, err := q.sqlBuilder.
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
		q.logQueryBuildingError(err)

		res, userCreateErr := q.db.ExecContext(ctx, query, args...)
		if userCreateErr != nil {
			q.logger.Error(userCreateErr, "creating test user")
			return fmt.Errorf("creating test user: %w", userCreateErr)
		}

		id, idRetErr := res.LastInsertId()
		if idRetErr != nil {
			q.logger.Error(idRetErr, "fetching insert ID")
			return fmt.Errorf("fetching insert ID: %w", idRetErr)
		}

		if _, accountCreationErr := q.CreateAccount(ctx, &types.AccountCreationInput{
			Name:          testUserConfig.Username,
			BelongsToUser: uint64(id),
		}); accountCreationErr != nil {
			q.logger.Error(accountCreationErr, "creating test user")
			return fmt.Errorf("creating test user: %w", accountCreationErr)
		}

		q.logger.WithValue(keys.UsernameKey, testUserConfig.Username).Debug("created test user and account")
	}

	return nil
}
