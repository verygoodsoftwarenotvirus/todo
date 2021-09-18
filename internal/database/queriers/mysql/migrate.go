package mysql

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/GuiaBolso/darwin"
	"github.com/segmentio/ksuid"
)

const (
	driverName = "instrumented-mysql"

	// defaultTestUserTwoFactorSecret is the default TwoFactorSecret we give to test users when we initialize them.
	// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
	defaultTestUserTwoFactorSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
)

const testUserExistenceQuery = `
	SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.service_roles, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = ? AND users.two_factor_secret_verified_on IS NOT NULL
`

const testUserCreationQuery = `
	INSERT INTO users (id,username,hashed_password,two_factor_secret,avatar_src,reputation,reputation_explanation,service_roles,two_factor_secret_verified_on,created_on) VALUES (?,?,?,?,?,?,?,?,UNIX_TIMESTAMP(),UNIX_TIMESTAMP())
`

// Migrate is a simple wrapper around the core querier Migrate call.
func (q *SQLQuerier) Migrate(ctx context.Context, maxAttempts uint8, testUserConfig *types.TestUserCreationConfig) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	q.logger.Info("migrating db")

	if !q.IsReady(ctx, maxAttempts) {
		return database.ErrDatabaseNotReady
	}

	q.migrateOnce.Do(q.migrationFunc)

	if testUserConfig != nil {
		q.logger.Debug("creating test user")

		testUserExistenceArgs := []interface{}{testUserConfig.Username}

		userRow := q.getOneRow(ctx, q.db, "user", testUserExistenceQuery, testUserExistenceArgs)
		_, _, _, err := q.scanUser(ctx, userRow, false)
		if err != nil {
			if testUserConfig.ID == "" {
				testUserConfig.ID = ksuid.New().String()
			}

			testUserCreationArgs := []interface{}{
				testUserConfig.ID,
				testUserConfig.Username,
				testUserConfig.HashedPassword,
				defaultTestUserTwoFactorSecret,
				"",
				types.GoodStandingAccountStatus,
				"",
				authorization.ServiceAdminRole.String(),
			}

			// these structs will be fleshed out by createUser
			user := &types.User{
				ID:       testUserConfig.ID,
				Username: testUserConfig.Username,
			}
			account := &types.Account{
				ID: ksuid.New().String(),
			}

			if err = q.createUser(ctx, user, account, testUserCreationQuery, testUserCreationArgs); err != nil {
				return observability.PrepareError(err, q.logger, span, "creating test user")
			}
			q.logger.WithValue(keys.UsernameKey, testUserConfig.Username).Debug("created test user and account")
		}
	}

	return nil
}

var (
	migrations = []darwin.Migration{
		{
			Version:     0.01,
			Description: "create sessions table for session manager",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS sessions (",
				"`token` CHAR(43) PRIMARY KEY,",
				"`data` BLOB NOT NULL,",
				"`expiry` TIMESTAMP(6) NOT NULL,",
				"`created_on` BIGINT UNSIGNED",
				");",
			}, "\n"),
		},
		{
			Version:     0.02,
			Description: "create sessions table for session manager",
			Script:      "CREATE INDEX sessions_expiry_idx ON sessions (expiry);",
		},
		{
			Version:     0.03,
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
			Version:     0.04,
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
			Version:     0.05,
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
			Version:     0.06,
			Description: "create API clients table",
			Script: strings.Join([]string{
				"CREATE TABLE IF NOT EXISTS api_clients (",
				"    `id` CHAR(27) NOT NULL,",
				"    `name` VARCHAR(128),",
				"    `client_id` VARCHAR(64) NOT NULL,",
				"    `secret_key` BINARY(128) NOT NULL,",
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
			Version:     0.07,
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
			Version:     0.08,
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
// migrate a postgres database.
func (q *SQLQuerier) migrationFunc() {
	driver := darwin.NewGenericDriver(q.db, darwin.MySQLDialect{})
	if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
		panic(fmt.Errorf("migrating database: %w", err))
	}
}
