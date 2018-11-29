package client_test

import (
	"net/url"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleURL = "https://todo.verygoodsoftwarenotvirus.ru"
)

func TestBuildURL(T *testing.T) {
	T.Parallel()

	T.Run("various urls", func(t *testing.T) {
		t.Parallel()

		cfg := &client.Config{Address: exampleURL}
		c, err := client.NewClient(cfg)
		require.NoError(t, err)

		testCases := []struct {
			expectation string
			inputParts  []string
			inputQuery  url.Values
		}{
			{
				expectation: "https://todo.verygoodsoftwarenotvirus.ru/api/v1/things",
				inputParts:  []string{"things"},
			},
			{
				expectation: "https://todo.verygoodsoftwarenotvirus.ru/api/v1/stuff?key=value",
				inputQuery: map[string][]string{
					"key": {"value"},
				},
				inputParts: []string{"stuff"},
			},
			{
				expectation: "https://todo.verygoodsoftwarenotvirus.ru/api/v1/things/and/stuff?key=value1&key=value2&yek=eulav",
				inputQuery: map[string][]string{
					"key": {"value1", "value2"},
					"yek": {"eulav"},
				},
				inputParts: []string{"things", "and", "stuff"},
			},
		}

		for _, tc := range testCases {
			actual := c.BuildURL(tc.inputQuery, tc.inputParts...)
			assert.Equal(t, tc.expectation, actual)
		}
	})
}
