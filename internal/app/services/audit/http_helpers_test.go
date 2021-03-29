package audit

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

type auditServiceHTTPRoutesTestHelper struct {
	suite.Suite

	ctx                  context.Context
	req                  *http.Request
	res                  *httptest.ResponseRecorder
	service              *service
	exampleUser          *types.User
	exampleAccount       *types.Account
	exampleAuditLogEntry *types.AuditLogEntry
}

var _ suite.SetupTestSuite = (*auditServiceHTTPRoutesTestHelper)(nil)

func (helper *auditServiceHTTPRoutesTestHelper) SetupTest() {
	t := helper.T()

	helper.ctx = context.Background()
	helper.service = buildTestService()
	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleAuditLogEntry = fakes.BuildFakeAuditLogEntry()

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
	helper.service.auditLogEntryIDFetcher = func(req *http.Request) uint64 {
		return helper.exampleAuditLogEntry.ID
	}

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

var _ suite.WithStats = (*auditServiceHTTPRoutesTestHelper)(nil)

func (helper *auditServiceHTTPRoutesTestHelper) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 6

	testutil.AssertAppropriateNumberOfTestsRan(helper.T(), totalExpectedTestCount, stats)
}
