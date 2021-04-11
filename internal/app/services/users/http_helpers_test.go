package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
)

type usersServiceHTTPRoutesTestHelper struct {
	ctx            context.Context
	req            *http.Request
	res            *httptest.ResponseRecorder
	service        *service
	exampleUser    *types.User
	exampleAccount *types.Account
}

func newTestHelper(t *testing.T) *usersServiceHTTPRoutesTestHelper {
	t.Helper()

	h := &usersServiceHTTPRoutesTestHelper{}

	h.ctx = context.Background()
	h.service = buildTestService(t)
	h.exampleUser = fakes.BuildFakeUser()
	h.exampleAccount = fakes.BuildFakeAccount()
	h.exampleAccount.BelongsToUser = h.exampleUser.ID

	h.service.userIDFetcher = func(*http.Request) uint64 {
		return h.exampleUser.ID
	}

	reqCtx, err := types.RequestContextFromUser(
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
	h.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}

	req := testutil.BuildTestRequest(t)
	h.req = req.WithContext(context.WithValue(req.Context(), types.RequestContextKey, reqCtx))
	h.res = httptest.NewRecorder()

	return h
}
