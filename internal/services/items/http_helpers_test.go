package items

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
)

type itemsServiceHTTPRoutesTestHelper struct {
	ctx                  context.Context
	req                  *http.Request
	res                  *httptest.ResponseRecorder
	service              *service
	exampleUser          *types.User
	exampleAccount       *types.Account
	exampleItem          *types.Item
	exampleCreationInput *types.ItemCreationInput
	exampleUpdateInput   *types.ItemUpdateInput
}

func buildTestHelper(t *testing.T) *itemsServiceHTTPRoutesTestHelper {
	t.Helper()

	helper := &itemsServiceHTTPRoutesTestHelper{}

	helper.ctx = context.Background()
	helper.service = buildTestService()
	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleItem = fakes.BuildFakeItem()
	helper.exampleItem.BelongsToAccount = helper.exampleAccount.ID
	helper.exampleCreationInput = fakes.BuildFakeItemCreationInputFromItem(helper.exampleItem)
	helper.exampleUpdateInput = fakes.BuildFakeItemUpdateInputFromItem(helper.exampleItem)

	helper.service.itemIDFetcher = func(*http.Request) uint64 {
		return helper.exampleItem.ID
	}

	sessionCtxData := &types.SessionContextData{
		Requester: types.RequesterInfo{
			UserID:                helper.exampleUser.ID,
			Reputation:            helper.exampleUser.ServiceAccountStatus,
			ReputationExplanation: helper.exampleUser.ReputationExplanation,
			ServicePermissions:    authorization.NewServiceRolePermissionChecker(helper.exampleUser.ServiceRoles...),
		},
		ActiveAccountID: helper.exampleAccount.ID,
		AccountPermissions: map[uint64]authorization.AccountRolePermissionsChecker{
			helper.exampleAccount.ID: authorization.NewAccountRolePermissionChecker(authorization.AccountMemberRole.String()),
		},
	}

	helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
		return sessionCtxData, nil
	}

	req := testutil.BuildTestRequest(t)

	helper.req = req.WithContext(context.WithValue(req.Context(), types.SessionContextDataKey, sessionCtxData))

	helper.res = httptest.NewRecorder()

	return helper
}
