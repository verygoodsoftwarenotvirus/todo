package requests

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

type usersTestSuite struct {
	suite.Suite

	ctx             context.Context
	builder         *Builder
	exampleUser     *types.User
	exampleInput    *types.NewUserCreationInput
	exampleUserList *types.UserList
}

var _ suite.SetupTestSuite = (*usersTestSuite)(nil)

func (s *usersTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.builder = buildTestRequestBuilder()

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

func (s *usersTestSuite) TestBuilder_BuildGetUserRequest() {
	const expectedPathFormat = "/api/v1/users/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleUser.ID)

		actual, err := s.builder.BuildGetUserRequest(s.ctx, s.exampleUser.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *usersTestSuite) TestBuilder_BuildGetUsersRequest() {
	const expectedPath = "/api/v1/users"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := s.builder.BuildGetUsersRequest(s.ctx, nil)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *usersTestSuite) TestBuilder_BuildSearchForUsersByUsernameRequest() {
	const expectedPath = "/api/v1/users/search"
	exampleUsername := fakes.BuildFakeUser().Username

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodGet, fmt.Sprintf("q=%s", exampleUsername), expectedPath)

		actual, err := s.builder.BuildSearchForUsersByUsernameRequest(s.ctx, exampleUsername)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *usersTestSuite) TestBuilder_BuildCreateUserRequest() {
	const expectedPath = "/users"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := s.builder.BuildCreateUserRequest(s.ctx, s.exampleInput)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *usersTestSuite) TestBuilder_BuildArchiveUserRequest() {
	const expectedPathFormat = "/api/v1/users/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleUser.ID)

		actual, err := s.builder.BuildArchiveUserRequest(s.ctx, s.exampleUser.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *usersTestSuite) TestBuilder_BuildGetAuditLogForUserRequest() {
	const expectedPath = "/api/v1/users/%d/audit"

	s.Run("happy path", func() {
		t := s.T()

		actual, err := s.builder.BuildGetAuditLogForUserRequest(s.ctx, s.exampleUser.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, s.exampleUser.ID)
		assertRequestQuality(t, actual, spec)
	})
}
