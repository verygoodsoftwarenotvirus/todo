package webhooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/require"
)

type webhooksServiceHTTPRoutesTestHelper struct {
	ctx                  context.Context
	req                  *http.Request
	res                  *httptest.ResponseRecorder
	service              *service
	exampleUser          *types.User
	exampleAccount       *types.Account
	exampleWebhook       *types.Webhook
	exampleCreationInput *types.WebhookCreationInput
	exampleUpdateInput   *types.WebhookUpdateInput
}

func newTestHelper(t *testing.T) *webhooksServiceHTTPRoutesTestHelper {
	t.Helper()

	h := &webhooksServiceHTTPRoutesTestHelper{}

	h.ctx = context.Background()
	h.service = buildTestService()
	h.exampleUser = fakes.BuildFakeUser()
	h.exampleAccount = fakes.BuildFakeAccount()
	h.exampleAccount.BelongsToUser = h.exampleUser.ID
	h.exampleWebhook = fakes.BuildFakeWebhook()
	h.exampleWebhook.BelongsToAccount = h.exampleAccount.ID
	h.exampleCreationInput = fakes.BuildFakeWebhookCreationInputFromWebhook(h.exampleWebhook)
	h.exampleUpdateInput = fakes.BuildFakeWebhookUpdateInputFromWebhook(h.exampleWebhook)

	h.service.webhookIDFetcher = func(*http.Request) uint64 {
		return h.exampleWebhook.ID
	}

	sessionCtxData, err := types.SessionContextDataFromUser(
		h.exampleUser,
		h.exampleAccount.ID,
		map[uint64]*types.UserAccountMembershipInfo{
			h.exampleAccount.ID: {
				AccountName: h.exampleAccount.Name,
				Permissions: testutil.BuildMaxUserPerms(),
			},
		},
	)
	require.NoError(t, err)

	h.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	h.service.sessionContextDataFetcher = func(_ *http.Request) (*types.SessionContextData, error) {
		return sessionCtxData, nil
	}

	req := testutil.BuildTestRequest(t)

	h.req = req.WithContext(context.WithValue(req.Context(), types.SessionContextDataKey, sessionCtxData))
	h.req = h.req.WithContext(context.WithValue(h.req.Context(), createMiddlewareCtxKey, h.exampleCreationInput))
	h.req = h.req.WithContext(context.WithValue(h.req.Context(), updateMiddlewareCtxKey, h.exampleUpdateInput))

	h.res = httptest.NewRecorder()

	return h
}
