package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_randString(t *testing.T) {
	t.Parallel()
	// obligatory

	actual := randString()
	assert.NotEmpty(t, actual)
	assert.Len(t, actual, 52)
}

func Test_buildConfig(t *testing.T) {
	t.Parallel()
	// obligatory

	actual := buildConfig()
	assert.NotNil(t, actual)
}

func TestParseConfigFile(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		tf, err := ioutil.TempFile(os.TempDir(), "*.toml")
		require.NoError(t, err)
		expected := "thisisatest"

		_, err = tf.Write([]byte(fmt.Sprintf(`
[server]
http_port = 1234
debug = false

[database]
type = "postgres"
debug = true
connection_details = "%s"
`, expected)))
		require.NoError(t, err)

		expectedConfig := &ServerConfig{
			Server: ServerSettings{
				HTTPPort: 1234,
				Debug:    false,
			},
			Database: DatabaseSettings{
				Type:              "postgres",
				Debug:             true,
				ConnectionDetails: database.ConnectionDetails(expected),
			},
		}

		cfg, err := ParseConfigFile(tf.Name())
		assert.NoError(t, err)

		assert.Equal(t, expectedConfig.Server.HTTPPort, cfg.Server.HTTPPort)
		assert.Equal(t, expectedConfig.Server.Debug, cfg.Server.Debug)
		assert.Equal(t, expectedConfig.Database.Type, cfg.Database.Type)
		assert.Equal(t, expectedConfig.Database.Debug, cfg.Database.Debug)
		assert.Equal(t, expectedConfig.Database.ConnectionDetails, cfg.Database.ConnectionDetails)
	})

	T.Run("with nonexistent file", func(t *testing.T) {
		cfg, err := ParseConfigFile("/this/doesn't/even/exist/lol")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

}
