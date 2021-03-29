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

type adminServiceHTTPRoutesTestSuite struct {
	suite.Suite

	ctx            context.Context
	service        *service
	exampleUser    *types.User
	exampleAccount *types.Account
	exampleInput   *types.UserReputationUpdateInput

	req *http.Request
	res *httptest.ResponseRecorder
}

func (s *adminServiceHTTPRoutesTestSuite) neuterAdminUser() {
	s.exampleUser.ServiceAdminPermissions = 0
	s.service.requestContextFetcher = func(*http.Request) (*types.RequestContext, error) {
		return types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
			s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
		})
	}
}

var _ suite.SetupTestSuite = (*adminServiceHTTPRoutesTestSuite)(nil)

func (s *adminServiceHTTPRoutesTestSuite) SetupTest() {
	t := s.T()

	s.service = buildTestService(t)

	var err error
	s.ctx, err = s.service.sessionManager.Load(context.Background(), "")
	require.NoError(t, err)

	s.exampleUser = fakes.BuildFakeUser()
	s.exampleUser.ServiceAdminPermissions = testutil.BuildMaxServiceAdminPerms()
	s.exampleAccount = fakes.BuildFakeAccount()
	s.exampleAccount.BelongsToUser = s.exampleUser.ID
	s.exampleInput = fakes.BuildFakeAccountStatusUpdateInput()

	s.res = httptest.NewRecorder()
	s.req, err = http.NewRequestWithContext(s.ctx, http.MethodPost, "https://blah.com", nil)
	require.NoError(t, err)
	require.NotNil(t, s.req)

	s.req = s.req.WithContext(context.WithValue(s.req.Context(), accountStatusUpdateMiddlewareCtxKey, s.exampleInput))

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
		s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
	})
	require.NoError(s.T(), err)

	s.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	s.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}
	s.service.userIDFetcher = func(req *http.Request) uint64 {
		return s.exampleUser.ID
	}
}

var _ suite.WithStats = (*adminServiceHTTPRoutesTestSuite)(nil)

func (s *adminServiceHTTPRoutesTestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 9

	testutil.AssertAppropriateNumberOfTestsRan(s.T(), totalExpectedTestCount, stats)
}
