package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	signatureHeaderKey    = "Signature"
	validClientSecretSize = 128
)

func TestClient_fetchAuthTokenForAPIClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		anticipatedResult := "v2.local.QAxIpVe-ECVNI1z4xQbm_qQYomyT3h8FtV8bxkz8pBJWkT8f7HtlOpbroPDEZUKop_vaglyp76CzYy375cHmKCW8e1CCkV0Lflu4GTDyXMqQdpZMM1E6OaoQW27gaRSvWBrR3IgbFIa0AkuUFw.UGFyYWdvbiBJbml0aWF0aXZlIEVudGVycHJpc2Vz"

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				response := &types.PASETOResponse{Token: anticipatedResult}

				assert.NotEmpty(t, req.Header.Get(signatureHeaderKey))

				require.NoError(t, json.NewEncoder(res).Encode(response))
			},
		))

		c := buildTestClient(t, ts)

		exampleClientID := "example_client_id"
		exampleSecret := make([]byte, validClientSecretSize)
		ctx := context.Background()

		token, err := c.fetchAuthTokenForAPIClient(ctx, c.unauthenticatedClient, exampleClientID, exampleSecret)

		assert.NoError(t, err)
		assert.Equal(t, anticipatedResult, token)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()

		c := buildTestClientWithInvalidURL(t)

		exampleClientID := "example_client_id"
		exampleSecret := make([]byte, validClientSecretSize)
		ctx := context.Background()

		token, err := c.fetchAuthTokenForAPIClient(ctx, c.unauthenticatedClient, exampleClientID, exampleSecret)

		assert.Error(t, err)
		assert.Empty(t, token)
	})

	T.Run("with error executing request", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assert.NotEmpty(t, req.Header.Get(signatureHeaderKey))

				time.Sleep(time.Minute)
			},
		))

		c := buildTestClient(t, ts)
		require.NoError(t, c.SetOptions(UsingTimeout(time.Nanosecond)))

		exampleClientID := "example_client_id"
		exampleSecret := make([]byte, validClientSecretSize)
		ctx := context.Background()

		token, err := c.fetchAuthTokenForAPIClient(ctx, c.unauthenticatedClient, exampleClientID, exampleSecret)

		assert.Error(t, err)
		assert.Empty(t, token)
	})

	T.Run("with invalid status code", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assert.NotEmpty(t, req.Header.Get(signatureHeaderKey))

				res.WriteHeader(http.StatusUnauthorized)
			},
		))

		c := buildTestClient(t, ts)

		exampleClientID := "example_client_id"
		exampleSecret := make([]byte, validClientSecretSize)
		ctx := context.Background()

		token, err := c.fetchAuthTokenForAPIClient(ctx, c.unauthenticatedClient, exampleClientID, exampleSecret)

		assert.Error(t, err)
		assert.Empty(t, token)
	})

	T.Run("with invalid response from server", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assert.NotEmpty(t, req.Header.Get(signatureHeaderKey))

				_, err := res.Write([]byte("BLAH"))
				assert.NoError(t, err)
			},
		))

		c := buildTestClient(t, ts)

		exampleClientID := "example_client_id"
		exampleSecret := make([]byte, validClientSecretSize)
		ctx := context.Background()

		token, err := c.fetchAuthTokenForAPIClient(ctx, c.unauthenticatedClient, exampleClientID, exampleSecret)

		assert.Error(t, err)
		assert.Empty(t, token)
	})
}
