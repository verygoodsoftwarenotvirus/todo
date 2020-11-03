package users

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvideUserDataServer(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, ProvideUserDataServer(buildTestService(t)))
	})
}
