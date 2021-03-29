package accountsubscriptionplans

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

type accountSubscriptionPlansServiceHTTPRoutesTestSuite struct {
	suite.Suite

	ctx                            context.Context
	req                            *http.Request
	res                            *httptest.ResponseRecorder
	service                        *service
	exampleUser                    *types.User
	exampleAccount                 *types.Account
	exampleAccountSubscriptionPlan *types.AccountSubscriptionPlan
	exampleInput                   *types.AccountSubscriptionPlanCreationInput
}

var _ suite.SetupTestSuite = (*accountSubscriptionPlansServiceHTTPRoutesTestSuite)(nil)

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) SetupTest() {
	t := s.T()

	s.ctx = context.Background()
	s.service = buildTestService()
	s.exampleUser = fakes.BuildFakeUser()
	s.exampleAccount = fakes.BuildFakeAccountForUser(s.exampleUser)
	s.exampleAccount.BelongsToUser = s.exampleUser.ID
	s.exampleAccountSubscriptionPlan = fakes.BuildFakeAccountSubscriptionPlan()
	s.exampleInput = fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(s.exampleAccountSubscriptionPlan)

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
		s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
	})
	require.NoError(t, err)

	s.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	s.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}
	s.service.accountSubscriptionPlanIDFetcher = func(req *http.Request) uint64 {
		return s.exampleAccountSubscriptionPlan.ID
	}

	s.res = httptest.NewRecorder()
	s.req, err = http.NewRequestWithContext(
		s.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, s.req)
}

var _ suite.WithStats = (*accountSubscriptionPlansServiceHTTPRoutesTestSuite)(nil)

func (s *accountSubscriptionPlansServiceHTTPRoutesTestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 17

	testutil.AssertAppropriateNumberOfTestsRan(s.T(), totalExpectedTestCount, stats)
}
