package mariadb

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

const (
	defaultLimit = uint8(20)
)

func buildTestService(t *testing.T) (*MariaDB, sqlmock.Sqlmock) {
	_, mock, err := sqlmock.New()
	require.NoError(t, err)

	return ProvideMariaDB(logging.NewNonOperationalLogger()), mock
}

func assertArgCountMatchesQuery(t *testing.T, query string, args []interface{}) {
	t.Helper()

	queryArgCount := len(regexp.MustCompile(`\?+`).FindAllString(query, -1))

	if len(args) > 0 {
		assert.Equal(t, queryArgCount, len(args))
	} else {
		assert.Zero(t, queryArgCount)
	}
}

func TestProvideMariaDB(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		buildTestService(t)
	})
}

func TestMariaDB_logQueryBuildingError(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		q, _ := buildTestService(t)
		q.logQueryBuildingError(errors.New("blah"))
	})
}

func TestProvideMariaDBConnection(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		_, err := ProvideMariaDBConnection(logging.NewNonOperationalLogger(), "")
		assert.NoError(t, err)
	})
}
