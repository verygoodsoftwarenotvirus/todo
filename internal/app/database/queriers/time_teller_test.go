package queriers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StandardTimeTeller_Now(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		assert.NotZero(t, (&StandardTimeTeller{}).Now())
	})
}
