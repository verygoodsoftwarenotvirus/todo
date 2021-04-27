package httpclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newPASETORoundTripper(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, _ := buildSimpleTestClient(t)
		exampleClientID := "example_client_id"
		exampleSecret := make([]byte, validClientSecretSize)

		assert.NotNil(t, newPASETORoundTripper(c, exampleClientID, exampleSecret))
	})
}
