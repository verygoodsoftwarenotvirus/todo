package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func attachCookieToRequestForTest(t *testing.T, s *service, req *http.Request, user *types.User) (context.Context, *http.Request) {
	t.Helper()

	exampleAccount := fakes.BuildFakeAccount()

	ctx, sessionErr := s.sessionManager.Load(req.Context(), "")
	require.NoError(t, sessionErr)
	require.NoError(t, s.sessionManager.RenewToken(ctx))

	s.sessionManager.Put(ctx, userIDContextKey, user.ID)
	s.sessionManager.Put(ctx, accountIDContextKey, exampleAccount.ID)

	token, _, err := s.sessionManager.Commit(ctx)
	assert.NotEmpty(t, token)
	assert.NoError(t, err)

	c, err := s.buildCookie(token, time.Now().Add(s.config.Cookies.Lifetime))
	require.NoError(t, err)
	req.AddCookie(c)

	return ctx, req.WithContext(ctx)
}

type authServiceHTTPRoutesTestSuite struct {
	suite.Suite

	ctx               context.Context
	req               *http.Request
	res               *httptest.ResponseRecorder
	service           *service
	exampleUser       *types.User
	exampleAccount    *types.Account
	exampleAPIClient  *types.APIClient
	examplePerms      map[uint64]permissions.ServiceUserPermissions
	exampleLoginInput *types.UserLoginInput
}

func (s *authServiceHTTPRoutesTestSuite) setContextFetcher() {
	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, s.examplePerms)
	require.NoError(s.T(), err)

	s.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}
}

var _ suite.SetupTestSuite = (*authServiceHTTPRoutesTestSuite)(nil)

func (s *authServiceHTTPRoutesTestSuite) SetupTest() {
	t := s.T()

	s.ctx = context.Background()
	s.service = buildTestService(t)
	s.exampleUser = fakes.BuildFakeUser()
	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleAccount.BelongsToUser = s.exampleUser.ID
	s.exampleAPIClient = fakes.BuildFakeAPIClient()
	s.exampleAPIClient.BelongsToUser = s.exampleUser.ID
	s.exampleLoginInput = fakes.BuildFakeUserLoginInputFromUser(s.exampleUser)

	s.examplePerms = map[uint64]permissions.ServiceUserPermissions{
		s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
	}

	s.setContextFetcher()

	s.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

	var err error

	s.res = httptest.NewRecorder()
	s.req, err = http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, s.req)
	require.NoError(t, err)
}

var _ suite.WithStats = (*authServiceHTTPRoutesTestSuite)(nil)

func (s *authServiceHTTPRoutesTestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 55

	testutil.AssertAppropriateNumberOfTestsRan(s.T(), totalExpectedTestCount, stats)
}
