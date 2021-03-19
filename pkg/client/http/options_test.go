package http

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

func TestV1Client_SetOption(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		expectedURL, err := url.Parse("https://notarealplace.lol")
		require.NoError(t, err)

		c, _ := buildSimpleTestClient(t)
		assert.NotEqual(t, expectedURL, c.URL(), "expected and actual URLs match somehow")

		exampleOption := func(client *Client) error {
			client.url = expectedURL
			return nil
		}

		require.NoError(t, c.SetOptions(exampleOption))

		assert.Equal(t, expectedURL, c.URL(), "expected and actual URLs do not match")
	})
}

func TestWithURL(T *testing.T) {
	T.Parallel()

	T.Run("normal use", func(t *testing.T) {
		t.Parallel()

		expectedURL, err := url.Parse("https://todo.verygoodsoftwarenotvirus.ru")
		require.NoError(t, err)

		c, err := NewClient(expectedURL)

		assert.NotNil(t, c)
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, c.URL(), "expected and actual URLs do not match")
	})
}

func TestWithLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expectedURL, err := url.Parse("https://todo.verygoodsoftwarenotvirus.ru")
		require.NoError(t, err)

		c, err := NewClient(expectedURL, UsingLogger(logging.NewNonOperationalLogger()))

		assert.NotNil(t, c)
		assert.NoError(t, err)
	})
}
