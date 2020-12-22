package httpclient

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestV1Client_SetOption(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		expectedURL, err := url.Parse("https://notarealplace.lol")
		require.NoError(t, err)

		c := buildTestClient(t, nil)
		assert.NotEqual(t, expectedURL, c.URL, "expected and actual URLs match somehow")

		exampleOption := func(client *V1Client) {
			client.URL = expectedURL
		}

		c.SetOption(exampleOption)

		assert.Equal(t, expectedURL, c.URL, "expected and actual URLs do not match")
	})
}

func TestWithURL(T *testing.T) {
	T.Parallel()

	T.Run("normal use", func(t *testing.T) {
		t.Parallel()

		expectedURL, err := url.Parse("https://todo.verygoodsoftwarenotvirus.ru")
		require.NoError(t, err)

		c := NewClient(
			WithURL(expectedURL),
		)

		assert.NotNil(t, c)
		assert.Equal(t, expectedURL, c.URL, "expected and actual URLs do not match")
	})
}

func TestWithLogger(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		c := NewClient(
			WithLogger(noop.NewLogger()),
		)

		assert.NotNil(t, c)
	})
}
