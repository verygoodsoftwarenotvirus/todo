package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/GuiaBolso/darwin"
	"github.com/segmentio/ksuid"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	// defaultTestUserTwoFactorSecret is the default TwoFactorSecret we give to test users when we initialize them.
	// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
	defaultTestUserTwoFactorSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	testUserExistenceQuery = `
		SELECT users.id, users.username, users.avatar_src, users.hashed_password, users.requires_password_change, users.password_last_changed_on, users.two_factor_secret, users.two_factor_secret_verified_on, users.service_roles, users.reputation, users.reputation_explanation, users.created_on, users.last_updated_on, users.archived_on FROM users WHERE users.archived_on IS NULL AND users.username = $1 AND users.two_factor_secret_verified_on IS NOT NULL
	`

	testUserCreationQuery = `
		INSERT INTO users (id,username,hashed_password,two_factor_secret,reputation,service_roles,two_factor_secret_verified_on) VALUES ($1,$2,$3,$4,$5,$6,extract(epoch FROM NOW()))
	`
)

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

var (
	//go:embed migrations/00001_initial.sql
	initMigration string

	//go:embed migrations/00002_items.sql
	itemsMigration string

	migrations = []darwin.Migration{
		{
			Version:     0.01,
			Description: "basic infrastructural tables",
			Script:      initMigration,
		},
		{
			Version:     0.02,
			Description: "create items table",
			Script:      itemsMigration,
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
