package httpclient

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAccounts(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(accountsTestSuite))
}

type accountsTestSuite struct {
	suite.Suite

	ctx                context.Context
	exampleAccount     *types.Account
	exampleUser        *types.User
	exampleAccountList *types.AccountList
}

var _ suite.SetupTestSuite = (*accountsTestSuite)(nil)

func (s *accountsTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleUser = fakes.BuildFakeUser()
	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleAccount.BelongsToUser = s.exampleUser.ID
	s.exampleAccountList = fakes.BuildFakeAccountList()
}

func (s *accountsTestSuite) TestClient_SwitchActiveAccount() {
	const expectedPath = "/users/account/select"

	s.Run("standard", func() {
		t := s.T()

		s.exampleAccount.BelongsToUser = 0

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusAccepted)
		c.authMethod = cookieAuthMethod

		assert.NoError(t, c.SwitchActiveAccount(s.ctx, s.exampleAccount.ID))
	})

	s.Run("with invalid account ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		c.authMethod = cookieAuthMethod

		assert.Error(t, c.SwitchActiveAccount(s.ctx, 0))
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		c.authMethod = cookieAuthMethod

		assert.Error(t, c.SwitchActiveAccount(s.ctx, s.exampleAccount.ID))
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)
		c.authMethod = cookieAuthMethod

		assert.Error(t, c.SwitchActiveAccount(s.ctx, s.exampleAccount.ID))
	})
}

func (s *accountsTestSuite) TestClient_GetAccount() {
	const expectedPathFormat = "/api/v1/accounts/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccount.ID)

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAccount)

		actual, err := c.GetAccount(s.ctx, s.exampleAccount.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleAccount, actual)
	})

	s.Run("with invalid account ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.GetAccount(s.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccount(s.ctx, s.exampleAccount.ID)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAccount.ID)

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAccount(s.ctx, s.exampleAccount.ID)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *accountsTestSuite) TestClient_GetAccounts() {
	const expectedPath = "/api/v1/accounts"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
	filter := (*types.QueryFilter)(nil)

	s.Run("standard", func() {
		t := s.T()

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAccountList)
		actual, err := c.GetAccounts(s.ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleAccountList, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAccounts(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAccounts(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *accountsTestSuite) TestClient_CreateAccount() {
	const expectedPath = "/api/v1/accounts"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		s.exampleAccount.BelongsToUser = 0
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(s.exampleAccount)

		c := buildTestClientWithRequestBodyValidation(t, spec, exampleInput, exampleInput, s.exampleAccount)
		actual, err := c.CreateAccount(s.ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleAccount, actual)
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.CreateAccount(s.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with invalid input", func() {
		t := s.T()

		exampleInput := &types.AccountCreationInput{}
		c, _ := buildSimpleTestClient(t)

		actual, err := c.CreateAccount(s.ctx, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(s.exampleAccount)
		c := buildTestClientWithInvalidURL(t)

		actual, err := c.CreateAccount(s.ctx, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		s.exampleAccount.BelongsToUser = 0
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(s.exampleAccount)

		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.CreateAccount(s.ctx, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *accountsTestSuite) TestClient_UpdateAccount() {
	const expectedPathFormat = "/api/v1/accounts/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleAccount.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAccount)

		assert.NoError(t, c.UpdateAccount(s.ctx, s.exampleAccount), "no error should be returned")
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		assert.Error(t, c.UpdateAccount(s.ctx, nil), "error should be returned")
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		err := c.UpdateAccount(s.ctx, s.exampleAccount)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		assert.Error(t, c.UpdateAccount(s.ctx, s.exampleAccount), "error should be returned")
	})
}

func (s *accountsTestSuite) TestClient_ArchiveAccount() {
	const expectedPathFormat = "/api/v1/accounts/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAccount.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		assert.NoError(t, c.ArchiveAccount(s.ctx, s.exampleAccount.ID), "no error should be returned")
	})

	s.Run("with invalid account ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		assert.Error(t, c.ArchiveAccount(s.ctx, 0), "no error should be returned")
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		assert.Error(t, c.ArchiveAccount(s.ctx, s.exampleAccount.ID), "error should be returned")
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		assert.Error(t, c.ArchiveAccount(s.ctx, s.exampleAccount.ID), "no error should be returned")
	})
}

func (s *accountsTestSuite) TestClient_AddUserToAccount() {
	const expectedPathFormat = "/api/v1/accounts/%d/member"

	s.Run("standard", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAddUserToAccountInput()
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat, exampleInput.AccountID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		assert.NoError(t, c.AddUserToAccount(s.ctx, exampleInput))
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		assert.Error(t, c.AddUserToAccount(s.ctx, nil))
	})

	s.Run("with invalid input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		assert.Error(t, c.AddUserToAccount(s.ctx, &types.AddUserToAccountInput{}))
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		exampleInput := fakes.BuildFakeAddUserToAccountInput()

		assert.Error(t, c.AddUserToAccount(s.ctx, exampleInput))
	})

	s.Run("with error executing request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAddUserToAccountInput()
		c, _ := buildTestClientThatWaitsTooLong(t)

		assert.Error(t, c.AddUserToAccount(s.ctx, exampleInput))
	})
}

func (s *accountsTestSuite) TestClient_MarkAsDefault() {
	const expectedPathFormat = "/api/v1/accounts/%d/default"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodPost, "", expectedPathFormat, s.exampleAccount.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		assert.NoError(t, c.MarkAsDefault(s.ctx, s.exampleAccount.ID))
	})

	s.Run("with invalid account ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		assert.Error(t, c.MarkAsDefault(s.ctx, 0))
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		assert.Error(t, c.MarkAsDefault(s.ctx, s.exampleAccount.ID))
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		assert.Error(t, c.MarkAsDefault(s.ctx, s.exampleAccount.ID))
	})
}

func (s *accountsTestSuite) TestClient_RemoveUserFromAccount() {
	const expectedPathFormat = "/api/v1/accounts/%d/members/%d"

	s.Run("standard", func() {
		t := s.T()

		query := url.Values{keys.ReasonKey: []string{t.Name()}}.Encode()
		spec := newRequestSpec(true, http.MethodDelete, query, expectedPathFormat, s.exampleAccount.ID, s.exampleUser.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		assert.NoError(t, c.RemoveUserFromAccount(s.ctx, s.exampleAccount.ID, s.exampleUser.ID, t.Name()))
	})

	s.Run("with invalid account ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		assert.Error(t, c.RemoveUserFromAccount(s.ctx, 0, s.exampleUser.ID, t.Name()))
	})

	s.Run("with invalid user ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		assert.Error(t, c.RemoveUserFromAccount(s.ctx, s.exampleAccount.ID, 0, t.Name()))
	})

	s.Run("with invalid reason", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		assert.Error(t, c.RemoveUserFromAccount(s.ctx, s.exampleAccount.ID, s.exampleUser.ID, ""))
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		assert.Error(t, c.RemoveUserFromAccount(s.ctx, s.exampleAccount.ID, s.exampleUser.ID, t.Name()))
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		assert.Error(t, c.RemoveUserFromAccount(s.ctx, s.exampleAccount.ID, s.exampleUser.ID, t.Name()))
	})
}

func (s *accountsTestSuite) TestClient_ModifyMemberPermissions() {
	const expectedPathFormat = "/api/v1/accounts/%d/members/%d/permissions"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPatch, "", expectedPathFormat, s.exampleAccount.ID, s.exampleUser.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		assert.NoError(t, c.ModifyMemberPermissions(s.ctx, s.exampleAccount.ID, s.exampleUser.ID, exampleInput))
	})

	s.Run("with invalid account ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		assert.Error(t, c.ModifyMemberPermissions(s.ctx, 0, s.exampleUser.ID, exampleInput))
	})

	s.Run("with invalid user ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		assert.Error(t, c.ModifyMemberPermissions(s.ctx, s.exampleAccount.ID, 0, exampleInput))
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		assert.Error(t, c.ModifyMemberPermissions(s.ctx, s.exampleAccount.ID, s.exampleUser.ID, nil))
	})

	s.Run("with invalid input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleInput := &types.ModifyUserPermissionsInput{}

		assert.Error(t, c.ModifyMemberPermissions(s.ctx, s.exampleAccount.ID, s.exampleUser.ID, exampleInput))
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		assert.Error(t, c.ModifyMemberPermissions(s.ctx, s.exampleAccount.ID, s.exampleUser.ID, exampleInput))
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)
		exampleInput := fakes.BuildFakeUserPermissionModificationInput()

		assert.Error(t, c.ModifyMemberPermissions(s.ctx, s.exampleAccount.ID, s.exampleUser.ID, exampleInput))
	})
}

func (s *accountsTestSuite) TestClient_TransferAccountOwnership() {
	const expectedPathFormat = "/api/v1/accounts/%d/transfer"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat, s.exampleAccount.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		assert.NoError(t, c.TransferAccountOwnership(s.ctx, s.exampleAccount.ID, exampleInput))
	})

	s.Run("with invalid account ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		assert.Error(t, c.TransferAccountOwnership(s.ctx, 0, exampleInput))
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		assert.Error(t, c.TransferAccountOwnership(s.ctx, s.exampleAccount.ID, nil))
	})

	s.Run("with invalid input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleInput := &types.TransferAccountOwnershipInput{}

		assert.Error(t, c.TransferAccountOwnership(s.ctx, s.exampleAccount.ID, exampleInput))
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		assert.Error(t, c.TransferAccountOwnership(s.ctx, s.exampleAccount.ID, exampleInput))
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)
		exampleInput := fakes.BuildFakeTransferAccountOwnershipInput()

		assert.Error(t, c.TransferAccountOwnership(s.ctx, s.exampleAccount.ID, exampleInput))
	})
}

func (s *accountsTestSuite) TestClient_GetAuditLogForAccount() {
	const (
		expectedPath   = "/api/v1/accounts/%d/audit"
		expectedMethod = http.MethodGet
	)

	s.Run("standard", func() {
		t := s.T()

		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAccount.ID)

		c, _ := buildTestClientWithJSONResponse(t, spec, exampleAuditLogEntryList)

		actual, err := c.GetAuditLogForAccount(s.ctx, s.exampleAccount.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid account ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.GetAuditLogForAccount(s.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.GetAuditLogForAccount(s.ctx, s.exampleAccount.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAccount.ID)
		c := buildTestClientWithInvalidResponse(t, spec)

		actual, err := c.GetAuditLogForAccount(s.ctx, s.exampleAccount.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
