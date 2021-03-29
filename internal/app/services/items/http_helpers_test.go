package items

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

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

func newTestHelper(t *testing.T) *itemsServiceHTTPRoutesTestHelper {
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

	reqCtx, err := types.RequestContextFromUser(
		h.exampleUser,
		h.exampleAccount.ID,
		map[uint64]permissions.ServiceUserPermissions{h.exampleAccount.ID: testutil.BuildMaxUserPerms()},
	)
	require.NoError(t, err)

	h.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	h.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}

	req := testutil.BuildTestRequest(t)

	h.req = req.WithContext(context.WithValue(req.Context(), types.RequestContextKey, reqCtx))
	h.req = h.req.WithContext(context.WithValue(h.req.Context(), createMiddlewareCtxKey, h.exampleCreationInput))
	h.req = h.req.WithContext(context.WithValue(h.req.Context(), updateMiddlewareCtxKey, h.exampleUpdateInput))

	h.res = httptest.NewRecorder()

	return h
}
