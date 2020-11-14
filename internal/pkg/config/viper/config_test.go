package viper

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func Test_randString(t *testing.T) {
	t.Parallel()

	actual := config.RandString()
	assert.NotEmpty(t, actual)
	assert.Len(t, actual, 52)
}

func TestBuildConfig(t *testing.T) {
	t.Parallel()

	actual := BuildViperConfig()
	assert.NotNil(t, actual)
}

func TestParseConfigFile(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		tf, err := ioutil.TempFile(os.TempDir(), "*.toml")
		require.NoError(t, err)
		expected := "thisisatest"

		_, err = tf.Write([]byte(fmt.Sprintf(`
[server]
http_port = 1234
debug = false

[database]
provider = "postgres"
debug = true
connection_details = "%s"
`, expected)))
		require.NoError(t, err)

		expectedConfig := &config.ServerConfig{
			Server: config.ServerSettings{
				HTTPPort: 1234,
				Debug:    false,
			},
			Database: config.DatabaseSettings{
				Provider:          "postgres",
				Debug:             true,
				ConnectionDetails: database.ConnectionDetails(expected),
			},
		}

		cfg, err := ParseConfigFile(noop.NewLogger(), tf.Name())
		assert.NoError(t, err)

		assert.Equal(t, expectedConfig.Server.HTTPPort, cfg.Server.HTTPPort)
		assert.Equal(t, expectedConfig.Server.Debug, cfg.Server.Debug)
		assert.Equal(t, expectedConfig.Database.Provider, cfg.Database.Provider)
		assert.Equal(t, expectedConfig.Database.Debug, cfg.Database.Debug)
		assert.Equal(t, expectedConfig.Database.ConnectionDetails, cfg.Database.ConnectionDetails)

		assert.NoError(t, os.Remove(tf.Name()))
	})

	T.Run("unparseable garbage", func(t *testing.T) {
		t.Parallel()
		tf, err := ioutil.TempFile(os.TempDir(), "*.toml")
		require.NoError(t, err)

		_, err = tf.Write([]byte(fmt.Sprintf(`
[server]
http_port = "fart"
debug = ":banana:"
`)))
		require.NoError(t, err)

		cfg, err := ParseConfigFile(noop.NewLogger(), tf.Name())
		assert.Error(t, err)
		assert.Nil(t, cfg)

		assert.NoError(t, os.Remove(tf.Name()))
	})

	T.Run("with invalid run mode", func(t *testing.T) {
		t.Parallel()
		tf, err := ioutil.TempFile(os.TempDir(), "*.toml")
		require.NoError(t, err)

		_, err = tf.Write([]byte(fmt.Sprintf(`
[meta]
run_mode = "party time"
`)))
		require.NoError(t, err)

		cfg, err := ParseConfigFile(noop.NewLogger(), tf.Name())
		assert.Error(t, err)
		assert.Nil(t, cfg)

		assert.NoError(t, os.Remove(tf.Name()))
	})

	T.Run("with nonexistent file", func(t *testing.T) {
		t.Parallel()
		cfg, err := ParseConfigFile(noop.NewLogger(), "/this/doesn't/even/exist/lol")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}
