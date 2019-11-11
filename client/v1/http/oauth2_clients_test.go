package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildGetOAuth2ClientRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedMethod := http.MethodGet

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		expectedID := uint64(1)

		actual, err := c.BuildGetOAuth2ClientRequest(ctx, expectedID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", expectedID)))
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.OAuth2Client{
			ID:           1,
			ClientID:     "example",
			ClientSecret: "blah",
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.True(t, strings.HasSuffix(req.URL.String(), strconv.Itoa(int(expected.ID))))
					assert.Equal(t, fmt.Sprintf("/api/v1/oauth2/clients/%d", expected.ID), req.URL.Path, "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(expected))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetOAuth2Client(ctx, expected.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildGetOAuth2ClientsRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedMethod := http.MethodGet

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		actual, err := c.BuildGetOAuth2ClientsRequest(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_GetOAuth2Clients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.OAuth2ClientList{
			Clients: []models.OAuth2Client{
				{
					ID:           1,
					ClientID:     "example",
					ClientSecret: "blah",
				},
			},
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/api/v1/oauth2/clients", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(expected))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetOAuth2Clients(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildCreateOAuth2ClientRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}
		req, err := c.BuildCreateOAuth2ClientRequest(ctx, &http.Cookie{}, exampleInput)

		require.NotNil(t, req)
		assert.NoError(t, err)
		assert.Equal(t, http.MethodPost, req.Method)
	})
}

func TestV1Client_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}

		exampleOutput := &models.OAuth2Client{
			ClientID:     "EXAMPLECLIENTID",
			ClientSecret: "EXAMPLECLIENTSECRET",
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, "/oauth2/client", req.URL.Path, "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)
					require.NoError(t, json.NewEncoder(res).Encode(exampleOutput))
				},
			),
		)
		c := buildTestClient(t, ts)

		oac, err := c.CreateOAuth2Client(ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)
		assert.NotNil(t, oac)
	})

	T.Run("with invalid body", func(t *testing.T) {
		ctx := context.Background()
		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/oauth2/client", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)
					_, err := res.Write([]byte("BLAH"))
					assert.NoError(t, err)
				},
			),
		)
		c := buildTestClient(t, ts)

		oac, err := c.CreateOAuth2Client(ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, oac)
	})

	T.Run("with timeout", func(t *testing.T) {
		ctx := context.Background()
		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/oauth2/client", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)
					time.Sleep(10 * time.Hour)
				},
			),
		)
		c := buildTestClient(t, ts)
		c.plainClient.Timeout = 500 * time.Millisecond

		oac, err := c.CreateOAuth2Client(ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, oac)
	})

	T.Run("with 404", func(t *testing.T) {
		ctx := context.Background()
		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/oauth2/client", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)
					res.WriteHeader(http.StatusNotFound)
				},
			),
		)
		c := buildTestClient(t, ts)

		oac, err := c.CreateOAuth2Client(ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Equal(t, err, ErrNotFound)
		assert.Nil(t, oac)
	})

	T.Run("with no cookie", func(t *testing.T) {
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		_, err := c.CreateOAuth2Client(ctx, nil, nil)
		assert.Error(t, err)
	})
}

func TestV1Client_BuildArchiveOAuth2ClientRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedMethod := http.MethodDelete
		ts := httptest.NewTLSServer(nil)

		expectedID := uint64(1)
		c := buildTestClient(t, ts)
		actual, err := c.BuildArchiveOAuth2ClientRequest(ctx, expectedID)

		require.NotNil(t, actual)
		require.NotNil(t, actual.URL)
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", expectedID)))
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_ArchiveOAuth2Client(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := uint64(1)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/oauth2/clients/%d", expected), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodDelete)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveOAuth2Client(ctx, expected)
		assert.NoError(t, err, "no error should be returned")
	})
}
