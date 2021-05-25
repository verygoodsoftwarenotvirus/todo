package apiclients

import (
	"testing"

	"github.com/stretchr/testify/assert"

	authservice "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
)

func TestProvideConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ProvideConfig(&authservice.Config{}))
	})
}
