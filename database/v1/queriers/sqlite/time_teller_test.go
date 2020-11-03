package sqlite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_stdLibTimeTeller_Now(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		tt := &stdLibTimeTeller{}

		assert.NotZero(t, tt.Now())
	})
}
