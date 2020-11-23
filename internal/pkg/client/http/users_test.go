package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildGetUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/users/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleUser := fakes.BuildFakeUser()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleUser.ID)

		actual, err := c.BuildGetUserRequest(ctx, exampleUser.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_GetUser(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/users/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleUser.ID)

		// the hashed password is never transmitted over the wire.
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

	T.Run("with invalid client URL", func(t *testing.T) {
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

func TestV1Client_BuildGetUsersRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetUsersRequest(ctx, nil)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
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
		// the hashed password is never transmitted over the wire.
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

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetUsers(ctx, nil)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildSearchForUsersByUsernameRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users/search"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUsername := fakes.BuildFakeUser().Username
		spec := newRequestSpec(false, http.MethodGet, fmt.Sprintf("q=%s", exampleUsername), expectedPath)

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildSearchForUsersByUsernameRequest(ctx, exampleUsername)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
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
		// the hashed password is never transmitted over the wire.
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

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		exampleUsername := fakes.BuildFakeUser().Username

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildCreateUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(exampleUser)
		c := buildTestClient(t, ts)

		actual, err := c.BuildCreateUserRequest(ctx, exampleInput)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
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

					var x *types.UserCreationInput
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

	T.Run("with invalid client URL", func(t *testing.T) {
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

func TestV1Client_BuildArchiveUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/users/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleUser := fakes.BuildFakeUser()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleUser.ID)

		actual, err := c.BuildArchiveUserRequest(ctx, exampleUser.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
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

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		err := buildTestClientWithInvalidURL(t).ArchiveUser(ctx, exampleUser.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildLoginRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/login"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		actual, err := c.BuildLoginRequest(ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		req, err := c.BuildLoginRequest(ctx, nil)
		assert.Nil(t, req)
		assert.Error(t, err)
	})
}

func TestV1Client_Login(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/login"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					http.SetCookie(res, &http.Cookie{Name: exampleUser.Username})
				},
			),
		)
		c := buildTestClient(t, ts)

		cookie, err := c.Login(ctx, exampleInput)
		require.NotNil(t, cookie)
		assert.NoError(t, err)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		cookie, err := c.Login(ctx, nil)
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		c := buildTestClientWithInvalidURL(t)

		cookie, err := c.Login(ctx, exampleInput)
		assert.Nil(t, cookie)
		assert.Error(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)
					time.Sleep(10 * time.Hour)
				},
			),
		)
		c := buildTestClient(t, ts)
		c.plainClient.Timeout = 500 * time.Microsecond

		cookie, err := c.Login(ctx, exampleInput)
		require.Nil(t, cookie)
		assert.Error(t, err)
	})

	T.Run("with missing cookie", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)
				},
			),
		)
		c := buildTestClient(t, ts)

		cookie, err := c.Login(ctx, exampleInput)
		require.Nil(t, cookie)
		assert.Error(t, err)
	})
}

func TestV1Client_BuildVerifyTOTPSecretRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/totp_secret/verify"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		actual, err := c.BuildVerifyTOTPSecretRequest(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_VerifyTOTPSecret(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/totp_secret/verify"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusAccepted)
				},
			),
		)
		c := buildTestClient(t, ts)

		err := c.VerifyTOTPSecret(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.NoError(t, err)
	})

	T.Run("with bad request response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusBadRequest)
				},
			),
		)
		c := buildTestClient(t, ts)

		err := c.VerifyTOTPSecret(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTOTPToken, err)
	})

	T.Run("with otherwise invalid status code response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusInternalServerError)
				},
			),
		)
		c := buildTestClient(t, ts)

		err := c.VerifyTOTPSecret(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		c := buildTestClientWithInvalidURL(t)

		err := c.VerifyTOTPSecret(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretValidationInputForUser(exampleUser)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					time.Sleep(10 * time.Minute)

					res.WriteHeader(http.StatusAccepted)
				},
			),
		)
		c := buildTestClient(t, ts)
		c.plainClient.Timeout = time.Millisecond

		err := c.VerifyTOTPSecret(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})
}
