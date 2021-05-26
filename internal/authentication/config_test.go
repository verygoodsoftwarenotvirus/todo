package authentication

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c := Config{
			Provider: argon2Provider,
		}

		assert.NoError(t, c.ValidateWithContext(context.Background()))
	})
}
