package bleve

import (
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestEnsureQueryIsRestrictedToUser(T *testing.T) {
	T.Parallel()

	T.Run("leaves good queries alone", func(t *testing.T) {
		t.Parallel()
		exampleUserID := fakes.BuildFakeUser().ID

		exampleQuery := fmt.Sprintf("things +belongsToAccount:%d", exampleUserID)
		expectation := fmt.Sprintf("things +belongsToAccount:%d", exampleUserID)

		actual := ensureQueryIsRestrictedToUser(exampleQuery, exampleUserID)
		assert.Equal(t, expectation, actual, "expected %q to equal %q", expectation, actual)
	})

	T.Run("basic replacement", func(t *testing.T) {
		t.Parallel()
		exampleUserID := fakes.BuildFakeUser().ID

		exampleQuery := "things"
		expectation := fmt.Sprintf("things +belongsToAccount:%d", exampleUserID)

		actual := ensureQueryIsRestrictedToUser(exampleQuery, exampleUserID)
		assert.Equal(t, expectation, actual, "expected %q to equal %q", expectation, actual)
	})

	T.Run("with invalid user restriction", func(t *testing.T) {
		t.Parallel()
		exampleUserID := fakes.BuildFakeUser().ID

		exampleQuery := fmt.Sprintf("stuff belongsToAccount:%d", exampleUserID)
		expectation := fmt.Sprintf("stuff +belongsToAccount:%d", exampleUserID)

		actual := ensureQueryIsRestrictedToUser(exampleQuery, exampleUserID)
		assert.Equal(t, expectation, actual, "expected %q to equal %q", expectation, actual)
	})
}