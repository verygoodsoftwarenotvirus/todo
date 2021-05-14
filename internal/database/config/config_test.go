package config

import (
	"context"
	"database/sql"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/auth"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	invalidProvider = "blah"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Provider:          PostgresProvider,
			ConnectionDetails: "example_connection_string",
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestConfig_ProvideDatabaseConnection(T *testing.T) {
	T.Parallel()

	T.Run("standard for postgres", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		cfg := &Config{
			Provider:          PostgresProvider,
			ConnectionDetails: "example_connection_string",
		}

		db, err := cfg.ProvideDatabaseConnection(logger)
		assert.NotNil(t, db)
		assert.NoError(t, err)
	})

	T.Run("standard for sqlite", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		cfg := &Config{
			Provider:          SqliteProvider,
			ConnectionDetails: "example_connection_string",
		}

		db, err := cfg.ProvideDatabaseConnection(logger)
		assert.NotNil(t, db)
		assert.NoError(t, err)
	})

	T.Run("standard for mariadb", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		cfg := &Config{
			Provider:          MariaDBProvider,
			ConnectionDetails: "dbuser:hunter2@tcp(database:3306)/todo",
		}

		db, err := cfg.ProvideDatabaseConnection(logger)
		assert.NotNil(t, db)
		assert.NoError(t, err)
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		cfg := &Config{
			Provider:          invalidProvider,
			ConnectionDetails: "example_connection_string",
		}

		db, err := cfg.ProvideDatabaseConnection(logger)
		assert.Nil(t, db)
		assert.Error(t, err)
	})
}

func TestConfig_ProvideDatabasePlaceholderFormat(T *testing.T) {
	T.Parallel()

	T.Run("standard for postgres", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          PostgresProvider,
			ConnectionDetails: "example_connection_string",
		}

		pf, err := cfg.ProvideDatabasePlaceholderFormat()
		assert.NotNil(t, pf)
		assert.NoError(t, err)
	})

	T.Run("standard for sqlite", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          SqliteProvider,
			ConnectionDetails: "example_connection_string",
		}

		pf, err := cfg.ProvideDatabasePlaceholderFormat()
		assert.NotNil(t, pf)
		assert.NoError(t, err)
	})

	T.Run("standard for mariadb", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          MariaDBProvider,
			ConnectionDetails: "example_connection_string",
		}

		pf, err := cfg.ProvideDatabasePlaceholderFormat()
		assert.NotNil(t, pf)
		assert.NoError(t, err)
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          invalidProvider,
			ConnectionDetails: "example_connection_string",
		}

		pf, err := cfg.ProvideDatabasePlaceholderFormat()
		assert.Nil(t, pf)
		assert.Error(t, err)
	})
}

func TestConfig_ProvideJSONPluckQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard for postgres", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          PostgresProvider,
			ConnectionDetails: "example_connection_string",
		}

		assert.NotEmpty(t, cfg.ProvideJSONPluckQuery())
	})

	T.Run("standard for sqlite", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          SqliteProvider,
			ConnectionDetails: "example_connection_string",
		}

		assert.NotEmpty(t, cfg.ProvideJSONPluckQuery())
	})

	T.Run("standard for mariadb", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          MariaDBProvider,
			ConnectionDetails: "example_connection_string",
		}

		assert.NotEmpty(t, cfg.ProvideJSONPluckQuery())
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          invalidProvider,
			ConnectionDetails: "example_connection_string",
		}

		assert.Empty(t, cfg.ProvideJSONPluckQuery())
	})
}

func TestConfig_ProvideCurrentUnixTimestampQuery(T *testing.T) {
	T.Parallel()

	T.Run("standard for postgres", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          PostgresProvider,
			ConnectionDetails: "example_connection_string",
		}

		assert.NotEmpty(t, cfg.ProvideCurrentUnixTimestampQuery())
	})

	T.Run("standard for sqlite", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          SqliteProvider,
			ConnectionDetails: "example_connection_string",
		}

		assert.NotEmpty(t, cfg.ProvideCurrentUnixTimestampQuery())
	})

	T.Run("standard for mariadb", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          MariaDBProvider,
			ConnectionDetails: "example_connection_string",
		}

		assert.NotEmpty(t, cfg.ProvideCurrentUnixTimestampQuery())
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:          invalidProvider,
			ConnectionDetails: "example_connection_string",
		}

		assert.Empty(t, cfg.ProvideCurrentUnixTimestampQuery())
	})
}

func TestProvideSessionManager(T *testing.T) {
	T.Parallel()

	T.Run("with nil database", func(t *testing.T) {
		t.Parallel()

		cookieConfig := auth.CookieConfig{}
		cfg := Config{
			Provider:          PostgresProvider,
			ConnectionDetails: "example_connection_string",
		}

		sessionManager, err := ProvideSessionManager(cookieConfig, cfg, nil)
		assert.Nil(t, sessionManager)
		assert.Error(t, err)
	})

	T.Run("standard for postgres", func(t *testing.T) {
		t.Parallel()

		cookieConfig := auth.CookieConfig{}
		cfg := Config{
			Provider:          PostgresProvider,
			ConnectionDetails: "example_connection_string",
		}

		sessionManager, err := ProvideSessionManager(cookieConfig, cfg, &sql.DB{})
		assert.NotNil(t, sessionManager)
		assert.NoError(t, err)
	})

	T.Run("standard for sqlite", func(t *testing.T) {
		t.Parallel()

		cookieConfig := auth.CookieConfig{}
		cfg := Config{
			Provider:          SqliteProvider,
			ConnectionDetails: "example_connection_string",
		}

		sessionManager, err := ProvideSessionManager(cookieConfig, cfg, &sql.DB{})
		assert.NotNil(t, sessionManager)
		assert.NoError(t, err)
	})

	T.Run("standard for mariadb", func(t *testing.T) {
		t.Parallel()

		cookieConfig := auth.CookieConfig{}
		cfg := Config{
			Provider:          MariaDBProvider,
			ConnectionDetails: "example_connection_string",
		}

		db, mockDB, err := sqlmock.New()
		require.NoError(t, err)
		require.NotNil(t, mockDB)

		mockDB.ExpectQuery("SELECT VERSION()").
			WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("1.2.3"))

		sessionManager, err := ProvideSessionManager(cookieConfig, cfg, db)
		assert.NotNil(t, sessionManager)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet())
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		cookieConfig := auth.CookieConfig{}
		cfg := Config{
			Provider:          invalidProvider,
			ConnectionDetails: "example_connection_string",
		}

		sessionManager, err := ProvideSessionManager(cookieConfig, cfg, &sql.DB{})
		assert.Nil(t, sessionManager)
		assert.Error(t, err)
	})
}
