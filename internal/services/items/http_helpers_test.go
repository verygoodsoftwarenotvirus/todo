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

	"github.com/stretchr/testify/require"
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

	h := &itemsServiceHTTPRoutesTestHelper{}

	h.ctx = context.Background()
	h.service = buildTestService()
	h.exampleUser = fakes.BuildFakeUser()
	h.exampleAccount = fakes.BuildFakeAccount()
	h.exampleAccount.BelongsToUser = h.exampleUser.ID
	h.exampleItem = fakes.BuildFakeItem()
	h.exampleItem.BelongsToAccount = h.exampleAccount.ID
	h.exampleCreationInput = fakes.BuildFakeItemCreationInputFromItem(h.exampleItem)
	h.exampleUpdateInput = fakes.BuildFakeItemUpdateInputFromItem(h.exampleItem)

	h.service.itemIDFetcher = func(*http.Request) uint64 {
		return h.exampleItem.ID
	}

	sessionCtxData, err := types.SessionContextDataFromUser(h.exampleUser, h.exampleAccount.ID, map[uint64]authorization.AccountRolePermissionsChecker{
		h.exampleAccount.ID: authorization.NewAccountRolePermissionChecker(authorization.AccountMemberRole.String()),
	})
	require.NoError(t, err)

	h.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	h.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
		return sessionCtxData, nil
	}

	req := testutil.BuildTestRequest(t)

	h.req = req.WithContext(context.WithValue(req.Context(), types.SessionContextDataKey, sessionCtxData))

	h.res = httptest.NewRecorder()

	return h
}
