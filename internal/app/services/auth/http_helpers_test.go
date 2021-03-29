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

type authServiceHTTPRoutesTestHelper struct {
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

func (helper *authServiceHTTPRoutesTestHelper) setContextFetcher() {
	reqCtx, err := types.RequestContextFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
	require.NoError(helper.T(), err)

	helper.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}
}

var _ suite.SetupTestSuite = (*authServiceHTTPRoutesTestHelper)(nil)

func (helper *authServiceHTTPRoutesTestHelper) SetupTest() {
	t := helper.T()

	helper.ctx = context.Background()
	helper.service = buildTestService(t)
	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleAPIClient = fakes.BuildFakeAPIClient()
	helper.exampleAPIClient.BelongsToUser = helper.exampleUser.ID
	helper.exampleLoginInput = fakes.BuildFakeUserLoginInputFromUser(helper.exampleUser)

	helper.examplePerms = map[uint64]permissions.ServiceUserPermissions{
		helper.exampleAccount.ID: testutil.BuildMaxUserPerms(),
	}

	helper.setContextFetcher()

	helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

	var err error

	helper.res = httptest.NewRecorder()
	helper.req, err = http.NewRequestWithContext(
		helper.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, helper.req)
	require.NoError(t, err)
}

var _ suite.WithStats = (*authServiceHTTPRoutesTestHelper)(nil)

func (helper *authServiceHTTPRoutesTestHelper) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 55

	testutil.AssertAppropriateNumberOfTestsRan(helper.T(), totalExpectedTestCount, stats)
}
