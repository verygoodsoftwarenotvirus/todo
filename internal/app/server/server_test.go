package server

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"

	"github.com/stretchr/testify/assert"
)

func TestProvideServer(T *testing.T) {
	T.Parallel()

	T.Run("with nil config provided", func(t *testing.T) {
		t.Parallel()

		x, err := ProvideServer(nil, nil)

		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("with nil server provided", func(t *testing.T) {
		t.Parallel()

		cfg := &config.ServerConfig{}
		x, err := ProvideServer(cfg, nil)

		assert.Nil(t, x)
		assert.Error(t, err)
	})
}
