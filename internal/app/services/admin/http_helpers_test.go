package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/require"
)

type adminServiceHTTPRoutesTestHelper struct {
	ctx            context.Context
	service        *service
	exampleUser    *types.User
	exampleAccount *types.Account
	exampleInput   *types.UserReputationUpdateInput

	req *http.Request
	res *httptest.ResponseRecorder
}

func (helper *adminServiceHTTPRoutesTestHelper) neuterAdminUser() {
	helper.exampleUser.ServiceAdminPermission = 0
	helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
		return types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, map[uint64]*types.UserAccountMembershipInfo{
			helper.exampleAccount.ID: {
				AccountName: helper.exampleAccount.Name,
				Permissions: testutil.BuildMaxUserPerms(),
			},
		})
	}
}

func buildTestHelper(t *testing.T) *adminServiceHTTPRoutesTestHelper {
	t.Helper()

	helper := &adminServiceHTTPRoutesTestHelper{}

	helper.service = buildTestService(t)

	var err error
	helper.ctx, err = helper.service.sessionManager.Load(context.Background(), "")
	require.NoError(t, err)

	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleUser.ServiceAdminPermission = testutil.BuildMaxServiceAdminPerms()
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleInput = fakes.BuildFakeAccountStatusUpdateInput()

	helper.res = httptest.NewRecorder()
	helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://blah.com", nil)
	require.NoError(t, err)
	require.NotNil(t, helper.req)

	helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), accountStatusUpdateMiddlewareCtxKey, helper.exampleInput))

	sessionCtxData, err := types.SessionContextDataFromUser(
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
	helper.service.sessionContextDataFetcher = func(_ *http.Request) (*types.SessionContextData, error) {
		return sessionCtxData, nil
	}
	helper.service.userIDFetcher = func(req *http.Request) uint64 {
		return helper.exampleUser.ID
	}

	return helper
}
