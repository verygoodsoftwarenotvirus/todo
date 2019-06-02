package postgres

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgres_buildGetUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		expectedUserID := uint64(123)
		expectedArgCount := 1
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, created_on, updated_on, archived_on FROM users WHERE id = $1"

		actualQuery, args := p.buildGetUserQuery(expectedUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUserID, args[0].(uint64))
	})
}

func TestPostgres_buildGetUsersQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		expectedArgCount := 0
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, created_on, updated_on, archived_on FROM users WHERE archived_on IS NULL LIMIT 20"

		actualQuery, args := p.buildGetUsersQuery(models.DefaultQueryFilter)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestPostgres_buildGetUserByUsernameQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)

		expectedUsername := "username"
		expectedArgCount := 1
		expectedQuery := "SELECT id, username, hashed_password, password_last_changed_on, two_factor_secret, created_on, updated_on, archived_on FROM users WHERE username = $1"

		actualQuery, args := p.buildGetUserByUsernameQuery(expectedUsername)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, expectedUsername, args[0].(string))
	})
}

func TestPostgres_buildCreateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleUser := &models.UserInput{
			Username:        "username",
			Password:        "hashed password",
			TwoFactorSecret: "two factor secret",
		}

		expectedArgCount := 3
		expectedQuery := "INSERT INTO users (username,hashed_password,two_factor_secret) VALUES ($1,$2,$3) RETURNING id, created_on"

		actualQuery, args := p.buildCreateUserQuery(exampleUser)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestPostgres_buildUpdateUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleUser := &models.User{
			ID:              321,
			Username:        "username",
			HashedPassword:  "hashed password",
			TwoFactorSecret: "two factor secret",
		}

		expectedArgCount := 4
		expectedQuery := "UPDATE users SET username = $1, hashed_password = $2, two_factor_secret = $3, updated_on = extract(epoch FROM NOW()) WHERE id = $4 RETURNING updated_on"

		actualQuery, args := p.buildUpdateUserQuery(exampleUser)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
	})
}

func TestPostgres_buildArchiveUserQuery(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		p, _ := buildTestService(t)
		exampleUserID := uint64(321)

		expectedArgCount := 1
		expectedQuery := "UPDATE users SET updated_on = extract(epoch FROM NOW()), archived_on = extract(epoch FROM NOW()) WHERE id = $1 RETURNING archived_on"

		actualQuery, args := p.buildArchiveUserQuery(exampleUserID)
		assert.Equal(t, expectedQuery, actualQuery)
		assert.Len(t, args, expectedArgCount)
		assert.Equal(t, exampleUserID, args[0].(uint64))
	})
}
