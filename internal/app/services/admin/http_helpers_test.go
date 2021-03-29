package admin

import (
	"context"
	"net/http"
	"net/http/httptest"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type adminServiceHTTPRoutesTestHelper struct {
	suite.Suite

	ctx            context.Context
	service        *service
	exampleUser    *types.User
	exampleAccount *types.Account
	exampleInput   *types.UserReputationUpdateInput

	req *http.Request
	res *httptest.ResponseRecorder
}

func (helper *adminServiceHTTPRoutesTestHelper) neuterAdminUser() {
	helper.exampleUser.ServiceAdminPermissions = 0
	helper.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
		return types.RequestContextFromUser(helper.exampleUser, helper.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
			helper.exampleAccount.ID: testutil.BuildMaxUserPerms(),
		})
	}
}

var _ suite.SetupTestSuite = (*adminServiceHTTPRoutesTestHelper)(nil)

func (helper *adminServiceHTTPRoutesTestHelper) SetupTest() {
	t := helper.T()

	helper.service = buildTestService(t)

	var err error
	helper.ctx, err = helper.service.sessionManager.Load(context.Background(), "")
	require.NoError(t, err)

	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleUser.ServiceAdminPermissions = testutil.BuildMaxServiceAdminPerms()
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleInput = fakes.BuildFakeAccountStatusUpdateInput()

	helper.res = httptest.NewRecorder()
	helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://blah.com", nil)
	require.NoError(t, err)
	require.NotNil(t, helper.req)

	helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), accountStatusUpdateMiddlewareCtxKey, helper.exampleInput))

	reqCtx, err := types.RequestContextFromUser(
		helper.exampleUser,
		helper.exampleAccount.ID,
		map[uint64]permissions.ServiceUserPermissions{helper.exampleAccount.ID: testutil.BuildMaxUserPerms()},
	)
	require.NoError(helper.T(), err)

	helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	helper.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}
	helper.service.userIDFetcher = func(req *http.Request) uint64 {
		return helper.exampleUser.ID
	}
}

var _ suite.WithStats = (*adminServiceHTTPRoutesTestHelper)(nil)

func (helper *adminServiceHTTPRoutesTestHelper) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 9

	testutil.AssertAppropriateNumberOfTestsRan(helper.T(), totalExpectedTestCount, stats)
}
