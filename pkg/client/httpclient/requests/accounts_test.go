package requests

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_BuildSwitchActiveAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/users/account/select"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		actual, err := helper.builder.BuildSwitchActiveAccountRequest(helper.ctx, exampleAccount.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildSwitchActiveAccountRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAccount.ID)

		actual, err := helper.builder.BuildGetAccountRequest(helper.ctx, exampleAccount.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetAccountRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
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
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCreateAccountRequest(helper.ctx, &types.AccountCreationInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildUpdateAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

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
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildArchiveAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleAccount.ID)

		actual, err := helper.builder.BuildArchiveAccountRequest(helper.ctx, exampleAccount.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildArchiveAccountRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildAddUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d/member"

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
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildAddUserRequest(helper.ctx, &types.AddUserToAccountInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildMarkAsDefaultRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d/default"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(true, http.MethodPost, "", expectedPathFormat, exampleAccount.ID)

		actual, err := helper.builder.BuildMarkAsDefaultRequest(helper.ctx, exampleAccount.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildMarkAsDefaultRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildRemoveUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d/members/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		reason := t.Name()
		expectedReason := url.QueryEscape(reason)
		spec := newRequestSpec(false, http.MethodDelete, fmt.Sprintf("reason=%s", expectedReason), expectedPathFormat, exampleAccount.ID, helper.exampleUser.ID)

		actual, err := helper.builder.BuildRemoveUserRequest(helper.ctx, exampleAccount.ID, helper.exampleUser.ID, reason)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		reason := t.Name()

		actual, err := helper.builder.BuildRemoveUserRequest(helper.ctx, 0, helper.exampleUser.ID, reason)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		reason := t.Name()

		actual, err := helper.builder.BuildRemoveUserRequest(helper.ctx, exampleAccount.ID, 0, reason)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildModifyMemberPermissionsRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d/members/%d/permissions"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(false, http.MethodPatch, "", expectedPathFormat, exampleAccount.ID, helper.exampleUser.ID)
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, exampleAccount.ID, helper.exampleUser.ID, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, 0, helper.exampleUser.ID, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, exampleAccount.ID, 0, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, exampleAccount.ID, helper.exampleUser.ID, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := helper.builder.BuildModifyMemberPermissionsRequest(helper.ctx, exampleAccount.ID, helper.exampleUser.ID, &types.ModifyUserPermissionsInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildTransferAccountOwnershipRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d/transfer"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat, exampleAccount.ID)
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		actual, err := helper.builder.BuildTransferAccountOwnershipRequest(helper.ctx, exampleAccount.ID, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		actual, err := helper.builder.BuildTransferAccountOwnershipRequest(helper.ctx, 0, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := helper.builder.BuildTransferAccountOwnershipRequest(helper.ctx, exampleAccount.ID, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := helper.builder.BuildTransferAccountOwnershipRequest(helper.ctx, exampleAccount.ID, &types.AccountOwnershipTransferInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAuditLogForAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts/%d/audit"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := helper.builder.BuildGetAuditLogForAccountRequest(helper.ctx, exampleAccount.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleAccount.ID)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetAuditLogForAccountRequest(helper.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
