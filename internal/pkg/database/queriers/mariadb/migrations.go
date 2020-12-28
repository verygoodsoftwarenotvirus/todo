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
		"    SET new.created_on = UNIX_TIMESTAMP();",
		"  END IF;",
		"END;",
	}, "\n")
}

var (
	migrations = []darwin.Migration{
		{
			Version:     0.00,
			Description: "create plans table and default plan",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS plans (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `name` VARCHAR(128) NOT NULL,",
				"    `price` INT UNSIGNED NOT NULL,",
				"    `period` INT UNSIGNED NOT NULL,",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    PRIMARY KEY (`id`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.01,
			Description: "create plans table creation trigger",
			Script:      buildCreationTriggerScript("plans"),
		},
		{
			Version:     0.02,
			Description: "create default plan",
			Script:      "INSERT INTO plans (name,price,period) VALUES ('free', 0, 0);",
		},
		{
			Version:     0.03,
			Description: "create users table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS users (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `username` VARCHAR(128) NOT NULL,",
				"    `hashed_password` VARCHAR(100) NOT NULL,",
				"    `salt` BINARY(16) NOT NULL,",
				"    `requires_password_change` BOOLEAN NOT NULL DEFAULT false,",
				"    `password_last_changed_on` INTEGER UNSIGNED,",
				"    `two_factor_secret` VARCHAR(256) NOT NULL,",
				"    `two_factor_secret_verified_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `is_admin` BOOLEAN NOT NULL DEFAULT false,",
				"    `admin_permissions` BIGINT NOT NULL DEFAULT 0,",
				"    `account_status` VARCHAR(32) NOT NULL DEFAULT 'created',",
				"    `status_explanation` VARCHAR(32) NOT NULL DEFAULT '',",
				"    `plan_id` BIGINT UNSIGNED NOT NULL DEFAULT 1,",
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    PRIMARY KEY (`id`),",
				"    FOREIGN KEY(`plan_id`) REFERENCES plans(`id`),",
				"    UNIQUE (`username`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.04,
			Description: "create users table creation trigger",
			Script:      buildCreationTriggerScript("users"),
		},
		{
			Version:     0.05,
			Description: "create sessions table for session manager",
			Script: strings.Join([]string{
				"CREATE TABLE sessions (",
				"`token` CHAR(43) PRIMARY KEY,",
				"`data` BLOB NOT NULL,",
				"`expiry` TIMESTAMP(6) NOT NULL",
				");",
			}, "\n"),
		},
		{
			Version:     0.06,
			Description: "create sessions table for session manager",
			Script:      "CREATE INDEX sessions_expiry_idx ON sessions (expiry);",
		},
		{
			Version:     0.07,
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
				"    FOREIGN KEY(`belongs_to_user`) REFERENCES users(`id`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.08,
			Description: "create oauth2_clients table creation trigger",
			Script:      buildCreationTriggerScript("oauth2_clients"),
		},
		{
			Version:     0.09,
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
				"    FOREIGN KEY (`belongs_to_user`) REFERENCES users(`id`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.10,
			Description: "create webhooks table creation trigger",
			Script:      buildCreationTriggerScript("webhooks"),
		},
		{
			Version:     0.11,
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
			Version:     0.12,
			Description: "create audit_log table creation trigger",
			Script:      buildCreationTriggerScript("audit_log"),
		},
		{
			Version:     0.13,
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
				"    FOREIGN KEY (`belongs_to_user`) REFERENCES users(`id`)",
				");",
			}, "\n"),
		},
		{
			Version:     0.14,
			Description: "create items table creation trigger",
			Script:      buildCreationTriggerScript("items"),
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
				queriers.UsersTableAccountStatusColumn,
				queriers.UsersTableAdminPermissionsColumn,
				queriers.UsersTableTwoFactorVerifiedOnColumn,
			).
			Values(
				testUserConfig.Username,
				testUserConfig.HashedPassword,
				[]byte("aaaaaaaaaaaaaaaa"),
				// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
				"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				testUserConfig.IsAdmin,
				types.GoodStandingAccountStatus,
				math.MaxUint32,
				squirrel.Expr(currentUnixTimeQuery),
			).
			ToSql()
		q.logQueryBuildingError(err)

		if _, dbErr := q.db.ExecContext(ctx, query, args...); dbErr != nil {
			return dbErr
		}

		q.logger.WithValue(keys.UsernameKey, testUserConfig.Username).Debug("created user")
	}

	return nil
}
