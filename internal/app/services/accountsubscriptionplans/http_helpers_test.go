package accountsubscriptionplans

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/require"
)

type accountSubscriptionPlansServiceHTTPRoutesTestHelper struct {
	ctx                            context.Context
	req                            *http.Request
	res                            *httptest.ResponseRecorder
	service                        *service
	exampleUser                    *types.User
	exampleAccount                 *types.Account
	exampleAccountSubscriptionPlan *types.AccountSubscriptionPlan
	exampleInput                   *types.AccountSubscriptionPlanCreationInput
}

func buildTestHelper(t *testing.T) *accountSubscriptionPlansServiceHTTPRoutesTestHelper {
	t.Helper()

	helper := &accountSubscriptionPlansServiceHTTPRoutesTestHelper{}

	helper.ctx = context.Background()
	helper.service = buildTestService()
	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleAccount = fakes.BuildFakeAccountForUser(helper.exampleUser)
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleAccountSubscriptionPlan = fakes.BuildFakeAccountSubscriptionPlan()
	helper.exampleInput = fakes.BuildFakeAccountSubscriptionPlanCreationInputFromAccountSubscriptionPlan(helper.exampleAccountSubscriptionPlan)

	reqCtx, err := types.RequestContextFromUser(
		helper.exampleUser,
		helper.exampleAccount.ID,
		map[uint64]*types.UserAccountMembershipInfo{
			helper.exampleAccount.ID: {
				AccountName: helper.exampleAccount.Name,
				Permissions: testutil.BuildMaxUserPerms(),
			},
		},
	)
	require.NoError(t, err)

	helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	helper.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}
	helper.service.accountSubscriptionPlanIDFetcher = func(req *http.Request) uint64 {
		return helper.exampleAccountSubscriptionPlan.ID
	}

	helper.res = httptest.NewRecorder()
	helper.req, err = http.NewRequestWithContext(
		helper.ctx,
		http.MethodGet,
		"http://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, helper.req)

	return helper
}
