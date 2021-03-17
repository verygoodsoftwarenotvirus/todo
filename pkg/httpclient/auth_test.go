package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_Login(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/login"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				http.SetCookie(res, &http.Cookie{Name: exampleUser.Username})
			},
		))
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

	T.Run("with invalid client url", func(t *testing.T) {
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

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)
				time.Sleep(10 * time.Hour)
			},
		))
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

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)
			},
		))
		c := buildTestClient(t, ts)

		cookie, err := c.Login(ctx, exampleInput)
		require.Nil(t, cookie)
		assert.Error(t, err)
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
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				res.WriteHeader(http.StatusAccepted)
			},
		))
		c := buildTestClient(t, ts)

		err := c.VerifyTOTPSecret(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.NoError(t, err)
	})

	T.Run("with bad request response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				res.WriteHeader(http.StatusBadRequest)
			},
		))
		c := buildTestClient(t, ts)

		err := c.VerifyTOTPSecret(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidTOTPToken, err)
	})

	T.Run("with otherwise invalid status code response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				res.WriteHeader(http.StatusInternalServerError)
			},
		))
		c := buildTestClient(t, ts)

		err := c.VerifyTOTPSecret(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)

		c := buildTestClientWithInvalidURL(t)

		err := c.VerifyTOTPSecret(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(exampleUser)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				time.Sleep(10 * time.Minute)

				res.WriteHeader(http.StatusAccepted)
			},
		))
		c := buildTestClient(t, ts)
		c.plainClient.Timeout = time.Millisecond

		err := c.VerifyTOTPSecret(ctx, exampleUser.ID, exampleInput.TOTPToken)
		assert.Error(t, err)
	})
}
