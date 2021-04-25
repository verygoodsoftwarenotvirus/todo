package http

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestUsers(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(usersTestSuite))
}

type usersBaseSuite struct {
	suite.Suite

	ctx             context.Context
	exampleUser     *types.User
	exampleUserList *types.UserList
}

var _ suite.SetupTestSuite = (*usersBaseSuite)(nil)

func (s *usersBaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleUser = fakes.BuildFakeUser()
	// the hashed passwords is never transmitted over the wire.
	s.exampleUser.HashedPassword = ""
	// the two factor secret is transmitted over the wire only on creation.
	s.exampleUser.TwoFactorSecret = ""
	// the two factor secret validation is never transmitted over the wire.
	s.exampleUser.TwoFactorSecretVerifiedOn = nil

	s.exampleUserList = fakes.BuildFakeUserList()
	for i := 0; i < len(s.exampleUserList.Users); i++ {
		// the hashed passwords is never transmitted over the wire.
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

func (s *usersTestSuite) TestClient_GetUser() {
	const expectedPathFormat = "/api/v1/users/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleUser.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleUser)

		actual, err := c.GetUser(s.ctx, s.exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleUser, actual)
	})

	s.Run("with invalid user ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.GetUser(s.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.GetUser(s.ctx, s.exampleUser.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.GetUser(s.ctx, s.exampleUser.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *usersTestSuite) TestClient_GetUsers() {
	const expectedPath = "/api/v1/users"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleUserList)

		actual, err := c.GetUsers(s.ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleUserList, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.GetUsers(s.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.GetUsers(s.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *usersTestSuite) TestClient_SearchForUsersByUsername() {
	const expectedPath = "/api/v1/users/search"
	exampleUsername := s.exampleUser.Username

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, fmt.Sprintf("q=%s", exampleUsername), expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleUserList.Users)

		actual, err := c.SearchForUsersByUsername(s.ctx, exampleUsername)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleUserList.Users, actual)
	})

	s.Run("with empty query", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.SearchForUsersByUsername(s.ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.SearchForUsersByUsername(s.ctx, exampleUsername)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.SearchForUsersByUsername(s.ctx, exampleUsername)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *usersTestSuite) TestClient_CreateUser() {
	const expectedPath = "/users"

	s.Run("standard", func() {
		t := s.T()

		expected := fakes.BuildUserCreationResponseFromUser(s.exampleUser)
		exampleInput := fakes.BuildFakeUserCreationInputFromUser(s.exampleUser)
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c := buildTestClientWithRequestBodyValidation(t, spec, &types.UserCreationInput{}, exampleInput, expected)

		actual, err := c.CreateUser(s.ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.CreateUser(s.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(s.exampleUser)
		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateUser(s.ctx, exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeUserCreationInputFromUser(s.exampleUser)
		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.CreateUser(s.ctx, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func (s *usersTestSuite) TestClient_ArchiveUser() {
	const expectedPathFormat = "/api/v1/users/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleUser.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		err := c.ArchiveUser(s.ctx, s.exampleUser.ID)
		assert.NoError(t, err)
	})

	s.Run("with invalid user ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		err := c.ArchiveUser(s.ctx, 0)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		err := c.ArchiveUser(s.ctx, s.exampleUser.ID)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		err := c.ArchiveUser(s.ctx, s.exampleUser.ID)
		assert.Error(t, err)
	})
}

func (s *usersTestSuite) TestClient_GetAuditLogForUser() {
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
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid user ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.GetAuditLogForUser(s.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.GetAuditLogForUser(s.ctx, s.exampleUser.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleUser.ID)
		c := buildTestClientWithInvalidResponse(t, spec)

		actual, err := c.GetAuditLogForUser(s.ctx, s.exampleUser.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *usersTestSuite) TestClient_UploadNewAvatar() {
	const expectedPath = "/api/v1/users/avatar/upload"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)
		exampleAvatar := []byte(t.Name())
		exampleExtension := png

		err := c.UploadNewAvatar(s.ctx, exampleAvatar, exampleExtension)
		assert.NoError(t, err)
	})

	s.Run("with invalid avatar", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleExtension := png

		err := c.UploadNewAvatar(s.ctx, nil, exampleExtension)
		assert.Error(t, err)
	})

	s.Run("with invalid extension", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleAvatar := []byte(t.Name())

		err := c.UploadNewAvatar(s.ctx, exampleAvatar, "")
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		exampleAvatar := []byte(t.Name())
		exampleExtension := png

		err := c.UploadNewAvatar(s.ctx, exampleAvatar, exampleExtension)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)
		exampleAvatar := []byte(t.Name())
		exampleExtension := png

		err := c.UploadNewAvatar(s.ctx, exampleAvatar, exampleExtension)
		assert.Error(t, err)
	})
}
