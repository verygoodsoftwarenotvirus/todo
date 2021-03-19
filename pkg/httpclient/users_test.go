package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsers(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(usersTestSuite))
}

type usersBaseSuite struct {
	suite.Suite

	ctx             context.Context
	exampleUser     *types.User
	exampleInput    *types.NewUserCreationInput
	exampleUserList *types.UserList
}

var _ suite.SetupTestSuite = (*usersBaseSuite)(nil)

func (s *usersBaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleUser = fakes.BuildFakeUser()
	// the hashed authentication is never transmitted over the wire.
	s.exampleUser.HashedPassword = ""
	// the two factor secret is transmitted over the wire only on creation.
	s.exampleUser.TwoFactorSecret = ""
	// the two factor secret validation is never transmitted over the wire.
	s.exampleUser.TwoFactorSecretVerifiedOn = nil

	s.exampleInput = fakes.BuildFakeUserCreationInputFromUser(s.exampleUser)
	s.exampleUserList = fakes.BuildFakeUserList()

	for i := 0; i < len(s.exampleUserList.Users); i++ {
		// the hashed authentication is never transmitted over the wire.
		s.exampleUserList.Users[i].HashedPassword = ""
		// the two factor secret is transmitted over the wire only on creation.
		s.exampleUserList.Users[i].TwoFactorSecret = ""
		// the two factor secret validation is never transmitted over the wire.
		s.exampleUserList.Users[i].TwoFactorSecretVerifiedOn = nil
	}
}

type usersTestSuite struct {
	suite.Suite

	usersBaseSuite
}

func (s *usersTestSuite) TestV1Client_GetUser() {
	const expectedPathFormat = "/api/v1/users/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleUser.ID)

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleUser)
		actual, err := c.GetUser(s.ctx, s.exampleUser.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleUser, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetUser(s.ctx, s.exampleUser.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *usersTestSuite) TestV1Client_GetUsers() {
	const expectedPath = "/api/v1/users"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleUserList)
		actual, err := c.GetUsers(s.ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleUserList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetUsers(s.ctx, nil)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *usersTestSuite) TestV1Client_SearchForUsersByUsername() {
	const expectedPath = "/api/v1/users/search"
	exampleUsername := s.exampleUser.Username

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, fmt.Sprintf("q=%s", exampleUsername), expectedPath)

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleUserList.Users)
		actual, err := c.SearchForUsersByUsername(s.ctx, exampleUsername)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleUserList.Users, actual)
	})

	s.Run("with invalid client URL", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.SearchForUsersByUsername(s.ctx, exampleUsername)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *usersTestSuite) TestV1Client_CreateUser() {
	const expectedPath = "/users"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		expected := fakes.BuildUserCreationResponseFromUser(s.exampleUser)
		c := buildTestClientWithRequestBodyValidation(t, spec, &types.NewUserCreationInput{}, s.exampleInput, expected)

		actual, err := c.CreateUser(s.ctx, s.exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateUser(s.ctx, s.exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *usersTestSuite) TestV1Client_ArchiveUser() {
	const expectedPathFormat = "/api/v1/users/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleUser.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		err := c.ArchiveUser(s.ctx, s.exampleUser.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).ArchiveUser(s.ctx, s.exampleUser.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *usersTestSuite) TestV1Client_GetAuditLogForUser() {
	const (
		expectedPath   = "/api/v1/users/%d/audit"
		expectedMethod = http.MethodGet
	)

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleUser.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		c, _ := buildTestClientWithJSONResponse(t, spec, exampleAuditLogEntryList)
		actual, err := c.GetAuditLogForUser(s.ctx, s.exampleUser.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForUser(s.ctx, s.exampleUser.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleUser.ID)

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAuditLogForUser(s.ctx, s.exampleUser.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
