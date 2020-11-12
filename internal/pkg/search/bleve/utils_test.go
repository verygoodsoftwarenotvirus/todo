package bleve

import (
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestEnsureQueryIsRestrictedToUser(T *testing.T) {
	T.Parallel()

	T.Run("leaves good queries alone", func(t *testing.T) {
		t.Parallel()
		exampleUserID := fakes.BuildFakeUser().ID

		exampleQuery := fmt.Sprintf("things +belongsToUser:%d", exampleUserID)
		expectation := fmt.Sprintf("things +belongsToUser:%d", exampleUserID)

		actual := ensureQueryIsRestrictedToUser(exampleQuery, exampleUserID)
		assert.Equal(t, expectation, actual, "expected %q to equal %q", expectation, actual)
	})

	T.Run("basic replacement", func(t *testing.T) {
		t.Parallel()
		exampleUserID := fakes.BuildFakeUser().ID

		exampleQuery := "things"
		expectation := fmt.Sprintf("things +belongsToUser:%d", exampleUserID)

		actual := ensureQueryIsRestrictedToUser(exampleQuery, exampleUserID)
		assert.Equal(t, expectation, actual, "expected %q to equal %q", expectation, actual)
	})

	T.Run("with invalid user restriction", func(t *testing.T) {
		t.Parallel()
		exampleUserID := fakes.BuildFakeUser().ID

		exampleQuery := fmt.Sprintf("stuff belongsToUser:%d", exampleUserID)
		expectation := fmt.Sprintf("stuff +belongsToUser:%d", exampleUserID)

		actual := ensureQueryIsRestrictedToUser(exampleQuery, exampleUserID)
		assert.Equal(t, expectation, actual, "expected %q to equal %q", expectation, actual)
	})
}
