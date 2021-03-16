package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_GetUser(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/users/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleUser.ID)

		// the hashed authentication is never transmitted over the wire.
		exampleUser.HashedPassword = ""
		// the two factor secret is transmitted over the wire only on creation.
		exampleUser.TwoFactorSecret = ""
		// the two factor secret validation is never transmitted over the wire.
		exampleUser.TwoFactorSecretVerifiedOn = nil

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleUser))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetUser(ctx, exampleUser.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleUser, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleUser.Salt = nil
		exampleUser.HashedPassword = ""

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetUser(ctx, exampleUser.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_GetUsers(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUserList := fakes.BuildFakeUserList()
		// the hashed authentication is never transmitted over the wire.
		exampleUserList.Users[0].HashedPassword = ""
		exampleUserList.Users[1].HashedPassword = ""
		exampleUserList.Users[2].HashedPassword = ""
		// the two factor secret is transmitted over the wire only on creation.
		exampleUserList.Users[0].TwoFactorSecret = ""
		exampleUserList.Users[1].TwoFactorSecret = ""
		exampleUserList.Users[2].TwoFactorSecret = ""
		// the two factor secret validation is never transmitted over the wire.
		exampleUserList.Users[0].TwoFactorSecretVerifiedOn = nil
		exampleUserList.Users[1].TwoFactorSecretVerifiedOn = nil
		exampleUserList.Users[2].TwoFactorSecretVerifiedOn = nil

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleUserList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetUsers(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleUserList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetUsers(ctx, nil)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_SearchForUsersByUsername(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users/search"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUsername := fakes.BuildFakeUser().Username
		spec := newRequestSpec(true, http.MethodGet, fmt.Sprintf("q=%s", exampleUsername), expectedPath)

		exampleUserList := fakes.BuildFakeUserList()
		// the hashed authentication is never transmitted over the wire.
		exampleUserList.Users[0].HashedPassword = ""
		exampleUserList.Users[1].HashedPassword = ""
		exampleUserList.Users[2].HashedPassword = ""
		// the two factor secret is transmitted over the wire only on creation.
		exampleUserList.Users[0].TwoFactorSecret = ""
		exampleUserList.Users[1].TwoFactorSecret = ""
		exampleUserList.Users[2].TwoFactorSecret = ""
		// the two factor secret validation is never transmitted over the wire.
		exampleUserList.Users[0].TwoFactorSecretVerifiedOn = nil
		exampleUserList.Users[1].TwoFactorSecretVerifiedOn = nil
		exampleUserList.Users[2].TwoFactorSecretVerifiedOn = nil
		exampleUsers := exampleUserList.Users

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleUsers))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleUsers, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		exampleUsername := fakes.BuildFakeUser().Username

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_CreateUser(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)
		expected := fakes.BuildUserCreationResponseFromUser(exampleUser)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					var x *types.NewUserCreationInput
					require.NoError(t, json.NewDecoder(req.Body).Decode(&x))
					assert.Equal(t, exampleInput, x)

					require.NoError(t, json.NewEncoder(res).Encode(expected))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.CreateUser(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateUser(ctx, exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_ArchiveUser(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/users/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleUser.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveUser(ctx, exampleUser.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		err := buildTestClientWithInvalidURL(t).ArchiveUser(ctx, exampleUser.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_GetAuditLogForUser(T *testing.T) {
	T.Parallel()

	const (
		expectedPath   = "/api/v1/users/%d/audit"
		expectedMethod = http.MethodGet
	)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleUser := fakes.BuildFakeUser()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, exampleUser.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAuditLogEntryList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForUser(ctx, exampleUser.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForUser(ctx, exampleUser.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, exampleUser.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForUser(ctx, exampleUser.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
