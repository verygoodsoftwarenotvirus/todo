package requests

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_BuildSwitchActiveAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/users/account/select"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		actual, err := helper.builder.BuildSwitchActiveAccountRequest(helper.ctx, exampleAccountID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildSwitchActiveAccountRequest(helper.ctx, "")
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleAccountID := fakes.BuildFakeID()

		actual, err := helper.builder.BuildSwitchActiveAccountRequest(helper.ctx, exampleAccountID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%s"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAccountID)

		actual, err := helper.builder.BuildGetAccountRequest(helper.ctx, exampleAccountID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetAccountRequest(helper.ctx, "")
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleAccountID := fakes.BuildFakeID()

		actual, err := helper.builder.BuildGetAccountRequest(helper.ctx, exampleAccountID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetAccountsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		filter := (*types.QueryFilter)(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := helper.builder.BuildGetAccountsRequest(helper.ctx, filter)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		filter := (*types.QueryFilter)(nil)

		actual, err := helper.builder.BuildGetAccountsRequest(helper.ctx, filter)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildCreateAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		actual, err := helper.builder.BuildCreateAccountRequest(helper.ctx, exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCreateAccountRequest(helper.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCreateAccountRequest(helper.ctx, &types.AccountCreationInput{})
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		actual, err := helper.builder.BuildCreateAccountRequest(helper.ctx, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildUpdateAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%s"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleAccount.ID)

		actual, err := helper.builder.BuildUpdateAccountRequest(helper.ctx, exampleAccount)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildUpdateAccountRequest(helper.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := helper.builder.BuildUpdateAccountRequest(helper.ctx, exampleAccount)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildArchiveAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%s"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleAccountID)

		actual, err := helper.builder.BuildArchiveAccountRequest(helper.ctx, exampleAccountID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildArchiveAccountRequest(helper.ctx, "")
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleAccountID := fakes.BuildFakeID()

		actual, err := helper.builder.BuildArchiveAccountRequest(helper.ctx, exampleAccountID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildAddUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%s/member"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleInput := fakes.BuildFakeAddUserToAccountInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat, exampleInput.AccountID)

		actual, err := helper.builder.BuildAddUserRequest(helper.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildAddUserRequest(helper.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildAddUserRequest(helper.ctx, &types.AddUserToAccountInput{})
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		exampleInput := fakes.BuildFakeAddUserToAccountInput()

		actual, err := helper.builder.BuildAddUserRequest(helper.ctx, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildMarkAsDefaultRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%s/default"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		spec := newRequestSpec(true, http.MethodPost, "", expectedPathFormat, exampleAccountID)

		actual, err := helper.builder.BuildMarkAsDefaultRequest(helper.ctx, exampleAccountID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildMarkAsDefaultRequest(helper.ctx, "")
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleAccountID := fakes.BuildFakeID()

		actual, err := helper.builder.BuildMarkAsDefaultRequest(helper.ctx, exampleAccountID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildRemoveUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%s/members/%s"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		reason := t.Name()
		expectedReason := url.QueryEscape(reason)
		spec := newRequestSpec(false, http.MethodDelete, fmt.Sprintf("reason=%s", expectedReason), expectedPathFormat, exampleAccountID, helper.exampleUser.ID)

		actual, err := helper.builder.BuildRemoveUserRequest(helper.ctx, exampleAccountID, helper.exampleUser.ID, reason)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		reason := t.Name()

		actual, err := helper.builder.BuildRemoveUserRequest(helper.ctx, "", helper.exampleUser.ID, reason)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		reason := t.Name()

		actual, err := helper.builder.BuildRemoveUserRequest(helper.ctx, exampleAccountID, "", reason)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleAccountID := fakes.BuildFakeID()

		reason := t.Name()

		actual, err := helper.builder.BuildRemoveUserRequest(helper.ctx, exampleAccountID, helper.exampleUser.ID, reason)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildModifyMemberPermissionsRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%s/members/%s/permissions"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		spec := newRequestSpec(false, http.MethodPatch, "", expectedPathFormat, exampleAccountID, helper.exampleUser.ID)
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, exampleAccountID, helper.exampleUser.ID, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, "", helper.exampleUser.ID, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, exampleAccountID, "", exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, exampleAccountID, helper.exampleUser.ID, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, exampleAccountID, helper.exampleUser.ID, &types.ModifyUserPermissionsInput{})
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleAccountID := fakes.BuildFakeID()

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, exampleAccountID, helper.exampleUser.ID, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildTransferAccountOwnershipRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%s/transfer"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat, exampleAccountID)
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		actual, err := helper.builder.BuildTransferAccountOwnershipRequest(helper.ctx, exampleAccountID, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		actual, err := helper.builder.BuildTransferAccountOwnershipRequest(helper.ctx, "", exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		actual, err := helper.builder.BuildTransferAccountOwnershipRequest(helper.ctx, exampleAccountID, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccountID := fakes.BuildFakeID()

		actual, err := helper.builder.BuildTransferAccountOwnershipRequest(helper.ctx, exampleAccountID, &types.AccountOwnershipTransferInput{})
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid request builder", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()
		exampleAccountID := fakes.BuildFakeID()

		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		actual, err := helper.builder.BuildTransferAccountOwnershipRequest(helper.ctx, exampleAccountID, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
