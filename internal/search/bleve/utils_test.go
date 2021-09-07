package bleve

import (
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

func TestEnsureQueryIsRestrictedToUser(T *testing.T) {
	T.Parallel()

	T.Run("leaves good queries alone", func(t *testing.T) {
		t.Parallel()
		exampleAccountID := fakes.BuildFakeAccount().ID

		exampleQuery := fmt.Sprintf("things +belongsToAccount:%s", exampleAccountID)
		expectation := fmt.Sprintf("things +belongsToAccount:%s", exampleAccountID)

		actual := ensureQueryIsRestrictedToAccount(exampleQuery, exampleAccountID)
		assert.Equal(t, expectation, actual, "expected %q to equal %q", expectation, actual)
	})

	T.Run("basic replacement", func(t *testing.T) {
		t.Parallel()
		exampleAccountID := fakes.BuildFakeAccount().ID

		exampleQuery := "things"
		expectation := fmt.Sprintf("things +belongsToAccount:%s", exampleAccountID)

		actual := ensureQueryIsRestrictedToAccount(exampleQuery, exampleAccountID)
		assert.Equal(t, expectation, actual, "expected %q to equal %q", expectation, actual)
	})

	T.Run("with invalid account restriction", func(t *testing.T) {
		t.Parallel()
		exampleAccountID := ksuid.New().String()

		exampleQuery := fmt.Sprintf("stuff belongsToAccount:%s", exampleAccountID)
		expectation := fmt.Sprintf("stuff +belongsToAccount:%s", exampleAccountID)

		actual := ensureQueryIsRestrictedToAccount(exampleQuery, exampleAccountID)
		assert.Equal(t, expectation, actual, "expected %q to equal %q", expectation, actual)
	})
}
