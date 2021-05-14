package requests

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_BuildUserStatusRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/auth/status"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPath)

		actual, err := h.builder.BuildUserStatusRequest(h.ctx)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildLoginRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/login"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(h.exampleUser)

		actual, err := h.builder.BuildLoginRequest(h.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		req, err := h.builder.BuildLoginRequest(h.ctx, nil)
		assert.Nil(t, req)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildLogoutRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/logout"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		spec := newRequestSpec(true, http.MethodPost, "", expectedPath)

		actual, err := h.builder.BuildLogoutRequest(h.ctx)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildChangePasswordRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/password/new"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakePasswordUpdateInput()
		spec := newRequestSpec(false, http.MethodPut, "", expectedPath)

		actual, err := h.builder.BuildChangePasswordRequest(h.ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildChangePasswordRequest(h.ctx, &http.Cookie{}, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error building request", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		h.builder = buildTestRequestBuilderWithInvalidURL()
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		actual, err := h.builder.BuildChangePasswordRequest(h.ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildCycleTwoFactorSecretRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/totp_secret/new"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := h.builder.BuildCycleTwoFactorSecretRequest(h.ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil cookie", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		actual, err := h.builder.BuildCycleTwoFactorSecretRequest(h.ctx, nil, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildCycleTwoFactorSecretRequest(h.ctx, &http.Cookie{}, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildCycleTwoFactorSecretRequest(h.ctx, &http.Cookie{}, &types.TOTPSecretRefreshInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error building request", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		h.builder = buildTestRequestBuilderWithInvalidURL()
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		actual, err := h.builder.BuildCycleTwoFactorSecretRequest(h.ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildVerifyTOTPSecretRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/totp_secret/verify"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(h.exampleUser)

		actual, err := h.builder.BuildVerifyTOTPSecretRequest(h.ctx, h.exampleUser.ID, exampleInput.TOTPToken)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(h.exampleUser)

		actual, err := h.builder.BuildVerifyTOTPSecretRequest(h.ctx, 0, exampleInput.TOTPToken)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid token", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildVerifyTOTPSecretRequest(h.ctx, h.exampleUser.ID, " nope lol ")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
