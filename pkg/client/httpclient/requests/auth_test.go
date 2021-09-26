package requests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func TestBuilder_BuildUserStatusRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/auth/status"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPath)

		actual, err := helper.builder.BuildUserStatusRequest(helper.ctx)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		actual, err := helper.builder.BuildUserStatusRequest(helper.ctx)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildLoginRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/login"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		exampleInput := fakes.BuildFakeUserLoginInputFromUser(helper.exampleUser)

		actual, err := helper.builder.BuildLoginRequest(helper.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		req, err := helper.builder.BuildLoginRequest(helper.ctx, nil)
		assert.Nil(t, req)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		exampleInput := fakes.BuildFakeUserLoginInputFromUser(helper.exampleUser)

		actual, err := helper.builder.BuildLoginRequest(helper.ctx, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildLogoutRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/logout"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		spec := newRequestSpec(true, http.MethodPost, "", expectedPath)

		actual, err := helper.builder.BuildLogoutRequest(helper.ctx)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		actual, err := helper.builder.BuildLogoutRequest(helper.ctx)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildChangePasswordRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/password/new"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleInput := fakes.BuildFakePasswordUpdateInput()
		spec := newRequestSpec(false, http.MethodPut, "", expectedPath)

		actual, err := helper.builder.BuildChangePasswordRequest(helper.ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildChangePasswordRequest(helper.ctx, &http.Cookie{}, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleInput := fakes.BuildFakePasswordUpdateInput()

		actual, err := helper.builder.BuildChangePasswordRequest(helper.ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildCycleTwoFactorSecretRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/totp_secret/new"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := helper.builder.BuildCycleTwoFactorSecretRequest(helper.ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil cookie", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		actual, err := helper.builder.BuildCycleTwoFactorSecretRequest(helper.ctx, nil, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCycleTwoFactorSecretRequest(helper.ctx, &http.Cookie{}, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCycleTwoFactorSecretRequest(helper.ctx, &http.Cookie{}, &types.TOTPSecretRefreshInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleInput := fakes.BuildFakeTOTPSecretRefreshInput()

		actual, err := helper.builder.BuildCycleTwoFactorSecretRequest(helper.ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildVerifyTOTPSecretRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users/totp_secret/verify"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)

		actual, err := helper.builder.BuildVerifyTOTPSecretRequest(helper.ctx, helper.exampleUser.ID, exampleInput.TOTPToken)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)

		actual, err := helper.builder.BuildVerifyTOTPSecretRequest(helper.ctx, "", exampleInput.TOTPToken)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid token", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildVerifyTOTPSecretRequest(helper.ctx, helper.exampleUser.ID, " nope lol ")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		exampleInput := fakes.BuildFakeTOTPSecretVerificationInputForUser(helper.exampleUser)

		actual, err := helper.builder.BuildVerifyTOTPSecretRequest(helper.ctx, helper.exampleUser.ID, exampleInput.TOTPToken)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
