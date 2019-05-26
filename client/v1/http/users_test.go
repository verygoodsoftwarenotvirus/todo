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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

func TestV1Client_BuildGetUserRequest(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expectedMethod := http.MethodGet
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		expectedID := uint64(1)

		actual, err := c.BuildGetUserRequest(ctx, expectedID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", expectedID)))
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expected := &models.User{
			ID: 1,
		}

		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.True(t,
						strings.HasSuffix(
							req.URL.String(),
							strconv.Itoa(int(expected.ID)),
						),
					)
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/users/%d", expected.ID), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(expected))
				},
			),
		)

		c := buildTestClient(t, ts)

		actual, err := c.GetUser(ctx, expected.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildGetUsersRequest(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expectedMethod := http.MethodGet
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetUsersRequest(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expected := &models.UserList{
			Users: []models.User{
				{
					ID: 1,
				},
			},
		}

		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/users", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(expected))
				},
			),
		)

		c := buildTestClient(t, ts)

		actual, err := c.GetUsers(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildCreateUserRequest(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expectedMethod := http.MethodPost
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		exampleInput := &models.UserInput{
			//
		}
		c := buildTestClient(t, ts)
		actual, err := c.BuildCreateUserRequest(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expected := &models.UserCreationResponse{
			ID: 1,
		}

		exampleInput := &models.UserInput{
			//
		}

		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/users", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)

					var x *models.UserInput
					require.NoError(t, json.NewDecoder(req.Body).Decode(&x))
					assert.Equal(t, exampleInput, x)

					require.NoError(t, json.NewEncoder(res).Encode(expected))
					res.WriteHeader(http.StatusOK)
				},
			),
		)

		c := buildTestClient(t, ts)

		actual, err := c.CreateUser(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildDeleteUserRequest(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expectedMethod := http.MethodDelete
		expectedID := uint64(1)
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildDeleteUserRequest(ctx, expectedID)

		require.NotNil(t, actual)
		require.NotNil(t, actual.URL)
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", expectedID)))
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_DeleteUser(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expected := uint64(1)
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/users/%d", expected), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodDelete)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).DeleteUser(ctx, expected)

		assert.NoError(t, err, "no error should be returned")
	})
}

func TestV1Client_BuildLoginRequest(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		req, err := c.BuildLoginRequest("username", "password", "123456")
		require.NotNil(t, req)
		assert.Equal(t, req.Method, http.MethodPost)
		assert.NoError(t, err)
	})
}

func TestV1Client_Login(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		ctx := context.Background()
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/users/login", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)

					http.SetCookie(res, &http.Cookie{Name: "hi"})
					res.WriteHeader(http.StatusOK)
				},
			),
		)
		c := buildTestClient(t, ts)

		cookie, err := c.Login(ctx, "username", "password", "123456")
		require.NotNil(t, cookie)
		assert.NoError(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		ctx := context.Background()
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/users/login", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)

					time.Sleep(10 * time.Hour)
					res.WriteHeader(http.StatusOK)
				},
			),
		)
		c := buildTestClient(t, ts)

		c.plainClient.Timeout = 500 * time.Microsecond
		cookie, err := c.Login(ctx, "username", "password", "123456")
		require.Nil(t, cookie)
		assert.Error(t, err)
	})

	T.Run("with missing cookie", func(t *testing.T) {
		ctx := context.Background()
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/users/login", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)

					res.WriteHeader(http.StatusOK)
				},
			),
		)
		c := buildTestClient(t, ts)

		cookie, err := c.Login(ctx, "username", "password", "123456")
		require.Nil(t, cookie)
		assert.Error(t, err)
	})
}
