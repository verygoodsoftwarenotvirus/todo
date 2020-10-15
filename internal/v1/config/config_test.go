package config

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandString(t *testing.T) {
	t.Parallel()

	actual := RandString()
	assert.NotEmpty(t, actual)
	assert.Len(t, actual, 52)
}

func TestServerConfig_EncodeToFile(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		cfg := &ServerConfig{}

		f, err := ioutil.TempFile("", "")
		require.NoError(t, err)

		assert.NoError(t, cfg.EncodeToFile(f.Name(), json.Marshal))
	})
}
