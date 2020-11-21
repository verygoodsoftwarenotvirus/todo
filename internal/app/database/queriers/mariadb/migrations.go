package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/auth"

	"github.com/GuiaBolso/darwin"
	"github.com/Masterminds/squirrel"
)

var currentMigration float64 = 0

func incrementMigrationVersion() float64 {
	currentMigration++
	return currentMigration
}

var (
	migrations = []darwin.Migration{
		{
			Version:     incrementMigrationVersion(),
			Description: "create users table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS users (",
				"    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,",
				"    `username` VARCHAR(150) NOT NULL,",
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
				"    `created_on` BIGINT UNSIGNED,",
				"    `last_updated_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    `archived_on` BIGINT UNSIGNED DEFAULT NULL,",
				"    PRIMARY KEY (`id`),",
				"    UNIQUE (`username`)",
				");",
			}, "\n"),
		},
		{
			Version:     incrementMigrationVersion(),
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
			Version:     incrementMigrationVersion(),
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
			Version:     incrementMigrationVersion(),
			Description: "create sessions table for session manager",
			Script:      "CREATE INDEX sessions_expiry_idx ON sessions (expiry);",
		},
		{
			Version:     incrementMigrationVersion(),
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
			Version:     incrementMigrationVersion(),
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
			Version:     incrementMigrationVersion(),
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
			Version:     incrementMigrationVersion(),
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
			Version:     incrementMigrationVersion(),
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
			Version:     incrementMigrationVersion(),
			Description: "create audit_log table creation trigger",
			Script: strings.Join([]string{
				"CREATE TRIGGER IF NOT EXISTS audit_log_creation_trigger BEFORE INSERT ON audit_log FOR EACH ROW",
				"BEGIN",
				"  IF (new.created_on is null)",
				"  THEN",
				"    SET new.created_on = UNIX_TIMESTAMP();",
				"  END IF;",
				"END;",
			}, "\n"),
		},
		{
			Version:     incrementMigrationVersion(),
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
			Version:     incrementMigrationVersion(),
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
func (m *MariaDB) Migrate(ctx context.Context, authenticator auth.Authenticator, testUserConfig *database.UserCreationConfig) error {
	m.logger.Info("migrating db")

	if !m.IsReady(ctx) {
		return database.ErrDBUnready
	}

	m.migrateOnce.Do(buildMigrationFunc(m.db))

	if testUserConfig != nil {
		hp, err := authenticator.HashPassword(ctx, testUserConfig.Password)
		if err != nil {
			return err
		}

		query, args, err := m.sqlBuilder.
			Insert(queriers.UsersTableName).
			Columns(
				queriers.UsersTableUsernameColumn,
				queriers.UsersTableHashedPasswordColumn,
				queriers.UsersTableSaltColumn,
				queriers.UsersTableTwoFactorColumn,
				queriers.UsersTableIsAdminColumn,
				queriers.UsersTableAdminPermissionsColumn,
				queriers.UsersTableTwoFactorVerifiedOnColumn,
			).
			Values(
				testUserConfig.Username,
				hp,
				[]byte("aaaaaaaaaaaaaaaa"),
				// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
				"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				testUserConfig.IsAdmin,
				math.MaxUint32,
				squirrel.Expr(currentUnixTimeQuery),
			).
			ToSql()
		m.logQueryBuildingError(err)

		if _, dbErr := m.db.ExecContext(ctx, query, args...); dbErr != nil {
			return dbErr
		}

		m.logger.WithValue("username", testUserConfig.Username).Debug("created user")
	}

	return nil
}
