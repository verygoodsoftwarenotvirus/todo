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

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		actual, err := h.builder.BuildSwitchActiveAccountRequest(h.ctx, exampleAccount.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildSwitchActiveAccountRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleAccount.ID)

		actual, err := h.builder.BuildGetAccountRequest(h.ctx, exampleAccount.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetAccountRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAccountsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		filter := (*types.QueryFilter)(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := h.builder.BuildGetAccountsRequest(h.ctx, filter)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildCreateAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		actual, err := h.builder.BuildCreateAccountRequest(h.ctx, exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildCreateAccountRequest(h.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildCreateAccountRequest(h.ctx, &types.AccountCreationInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildUpdateAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleAccount.ID)

		actual, err := h.builder.BuildUpdateAccountRequest(h.ctx, exampleAccount)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildUpdateAccountRequest(h.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildArchiveAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleAccount.ID)

		actual, err := h.builder.BuildArchiveAccountRequest(h.ctx, exampleAccount.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildArchiveAccountRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildAddUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d/member"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		exampleInput := fakes.BuildFakeAddUserToAccountInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat, exampleInput.AccountID)

		actual, err := h.builder.BuildAddUserRequest(h.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildAddUserRequest(h.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildAddUserRequest(h.ctx, &types.AddUserToAccountInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildMarkAsDefaultRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d/default"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(true, http.MethodPost, "", expectedPathFormat, exampleAccount.ID)

		actual, err := h.builder.BuildMarkAsDefaultRequest(h.ctx, exampleAccount.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildMarkAsDefaultRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildRemoveUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d/members/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		reason := t.Name()
		expectedReason := url.QueryEscape(reason)
		spec := newRequestSpec(false, http.MethodDelete, fmt.Sprintf("reason=%s", expectedReason), expectedPathFormat, exampleAccount.ID, h.exampleUser.ID)

		actual, err := h.builder.BuildRemoveUserRequest(h.ctx, exampleAccount.ID, h.exampleUser.ID, reason)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		reason := t.Name()

		actual, err := h.builder.BuildRemoveUserRequest(h.ctx, 0, h.exampleUser.ID, reason)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		reason := t.Name()

		actual, err := h.builder.BuildRemoveUserRequest(h.ctx, exampleAccount.ID, 0, reason)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildModifyMemberPermissionsRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d/members/%d/permissions"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(false, http.MethodPatch, "", expectedPathFormat, exampleAccount.ID, h.exampleUser.ID)
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		actual, err := h.builder.BuildModifyMemberPermissionsRequest(h.ctx, exampleAccount.ID, h.exampleUser.ID, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		actual, err := h.builder.BuildModifyMemberPermissionsRequest(h.ctx, 0, h.exampleUser.ID, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		actual, err := h.builder.BuildModifyMemberPermissionsRequest(h.ctx, exampleAccount.ID, 0, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := h.builder.BuildModifyMemberPermissionsRequest(h.ctx, exampleAccount.ID, h.exampleUser.ID, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := h.builder.BuildModifyMemberPermissionsRequest(h.ctx, exampleAccount.ID, h.exampleUser.ID, &types.ModifyUserPermissionsInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildTransferAccountOwnershipRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/accounts/%d/transfer"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat, exampleAccount.ID)
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		actual, err := h.builder.BuildTransferAccountOwnershipRequest(h.ctx, exampleAccount.ID, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		actual, err := h.builder.BuildTransferAccountOwnershipRequest(h.ctx, 0, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := h.builder.BuildTransferAccountOwnershipRequest(h.ctx, exampleAccount.ID, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := h.builder.BuildTransferAccountOwnershipRequest(h.ctx, exampleAccount.ID, &types.TransferAccountOwnershipInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAuditLogForAccountRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/accounts/%d/audit"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleAccount := fakes.BuildFakeAccount()

		actual, err := h.builder.BuildGetAuditLogForAccountRequest(h.ctx, exampleAccount.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleAccount.ID)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetAuditLogForAccountRequest(h.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
