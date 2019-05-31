package users

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_randString(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		actual, err := randString()
		assert.NotEmpty(t, actual)
		assert.NoError(t, err)
	})
}
