package apiclients

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

type apiClientsServiceHTTPRoutesTestHelper struct {
	suite.Suite

	ctx              context.Context
	req              *http.Request
	res              *httptest.ResponseRecorder
	service          *service
	exampleUser      *types.User
	exampleAccount   *types.Account
	exampleAPIClient *types.APIClient
	exampleInput     *types.APIClientCreationInput
}

var _ suite.SetupTestSuite = (*apiClientsServiceHTTPRoutesTestHelper)(nil)

func (helper *apiClientsServiceHTTPRoutesTestHelper) SetupTest() {
	t := helper.T()

	helper.ctx = context.Background()
	helper.service = buildTestService(t)
	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleAPIClient = fakes.BuildFakeAPIClient()
	helper.exampleAPIClient.BelongsToUser = helper.exampleUser.ID
	helper.exampleInput = fakes.BuildFakeAPIClientCreationInputFromClient(helper.exampleAPIClient)

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

	req := testutil.BuildTestRequest(t)

	helper.req = req.WithContext(context.WithValue(req.Context(), types.RequestContextKey, reqCtx))
	helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), creationMiddlewareCtxKey, helper.exampleInput))

	helper.res = httptest.NewRecorder()
}

var _ suite.WithStats = (*apiClientsServiceHTTPRoutesTestHelper)(nil)

func (helper *apiClientsServiceHTTPRoutesTestHelper) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 11

	testutil.AssertAppropriateNumberOfTestsRan(helper.T(), totalExpectedTestCount, stats)
}
