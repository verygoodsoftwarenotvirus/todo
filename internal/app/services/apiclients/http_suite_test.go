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

type apiClientsServiceHTTPRoutesTestSuite struct {
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

var _ suite.SetupTestSuite = (*apiClientsServiceHTTPRoutesTestSuite)(nil)

func (s *apiClientsServiceHTTPRoutesTestSuite) SetupTest() {
	t := s.T()

	s.ctx = context.Background()
	s.service = buildTestService(t)
	s.exampleUser = fakes.BuildFakeUser()
	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleAccount.BelongsToUser = s.exampleUser.ID
	s.exampleAPIClient = fakes.BuildFakeAPIClient()
	s.exampleAPIClient.BelongsToUser = s.exampleUser.ID
	s.exampleInput = fakes.BuildFakeAPIClientCreationInputFromClient(s.exampleAPIClient)

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
		s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
	})
	require.NoError(s.T(), err)

	s.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	s.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}

	req := testutil.BuildTestRequest(t)

	s.req = req.WithContext(context.WithValue(req.Context(), types.RequestContextKey, reqCtx))
	s.req = s.req.WithContext(context.WithValue(s.req.Context(), creationMiddlewareCtxKey, s.exampleInput))

	s.res = httptest.NewRecorder()
}

var _ suite.WithStats = (*apiClientsServiceHTTPRoutesTestSuite)(nil)

func (s *apiClientsServiceHTTPRoutesTestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 11

	testutil.AssertAppropriateNumberOfTestsRan(s.T(), totalExpectedTestCount, stats)
}
