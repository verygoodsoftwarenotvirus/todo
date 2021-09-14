package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/querybuilding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/GuiaBolso/darwin"
	"github.com/segmentio/ksuid"
)

const testUserExistenceQuery = `
	SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.service_roles, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = $1 AND users.two_factor_secret_verified_on IS NOT NULL
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

			testUserCreationQuery := "INSERT INTO users (id,username,hashed_password,two_factor_secret,reputation,service_roles,two_factor_secret_verified_on) VALUES ($1,$2,$3,$4,$5,$6,extract(epoch FROM NOW()))"
			testUserCreationArgs := []interface{}{
				testUserConfig.ID,
				testUserConfig.Username,
				testUserConfig.HashedPassword,
				querybuilding.DefaultTestUserTwoFactorSecret,
				types.GoodStandingAccountStatus,
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

////////////////////////////////////

var (
	//go:embed migrations/00001_initial.sql
	initSqlScript string

	migrations = []darwin.Migration{
		{
			Version:     0.0,
			Description: "create basic structures",
			Script:      initSqlScript,
		},
		{
			Version:     0.01,
			Description: "create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				id CHAR(27) NOT NULL PRIMARY KEY,
				name TEXT NOT NULL,
				details TEXT NOT NULL DEFAULT '',
				created_on BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				last_updated_on BIGINT DEFAULT NULL,
				archived_on BIGINT DEFAULT NULL,
				belongs_to_account CHAR(27) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE
			);`,
		},
	}
)

// BuildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a postgres database.
func (q *SQLQuerier) migrationFunc() {
	driver := darwin.NewGenericDriver(q.db, darwin.PostgresDialect{})
	if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
		panic(fmt.Errorf("migrating database: %w", err))
	}
}
