package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildGetOAuth2ClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/oauth2/clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleOAuth2Client.ID)

		actual, err := c.BuildGetOAuth2ClientRequest(ctx, exampleOAuth2Client.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetOAuth2Client(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/oauth2/clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleOAuth2Client.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleOAuth2Client))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetOAuth2Client(ctx, exampleOAuth2Client.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleOAuth2Client, actual)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetOAuth2Client(ctx, exampleOAuth2Client.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildGetOAuth2ClientsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/oauth2/clients"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := c.BuildGetOAuth2ClientsRequest(ctx, nil)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetOAuth2Clients(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/oauth2/clients"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2ClientList := fakes.BuildFakeOAuth2ClientList()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleOAuth2ClientList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetOAuth2Clients(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleOAuth2ClientList, actual)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetOAuth2Clients(ctx, nil)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildCreateOAuth2ClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/oauth2/client"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		req, err := c.BuildCreateOAuth2ClientRequest(ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, req, spec)
	})
}

func TestV1Client_CreateOAuth2Client(T *testing.T) {
	T.Parallel()

	const expectedPath = "/oauth2/client"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleOAuth2Client))
				},
			),
		)
		c := buildTestClient(t, ts)

		actual, err := c.CreateOAuth2Client(ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleOAuth2Client, actual)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.CreateOAuth2Client(ctx, &http.Cookie{}, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response from server", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		exampleInput := fakes.BuildFakeOAuth2ClientCreationInputFromClient(exampleOAuth2Client)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					_, err := res.Write([]byte("BLAH"))
					assert.NoError(t, err)
				},
			),
		)
		c := buildTestClient(t, ts)

		actual, err := c.CreateOAuth2Client(ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("without cookie", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		_, err := c.CreateOAuth2Client(ctx, nil, nil)
		assert.Error(t, err)
	})
}

func TestV1Client_BuildArchiveOAuth2ClientRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/oauth2/clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleOAuth2Client.ID)
		c := buildTestClient(t, ts)

		actual, err := c.BuildArchiveOAuth2ClientRequest(ctx, exampleOAuth2Client.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_ArchiveOAuth2Client(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/oauth2/clients/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleOAuth2Client.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		err := buildTestClientWithInvalidURL(t).ArchiveOAuth2Client(ctx, exampleOAuth2Client.ID)
		assert.Error(t, err, "error should be returned")
	})
}
