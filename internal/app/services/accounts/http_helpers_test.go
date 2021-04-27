package accounts

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

type accountsServiceHTTPRoutesTestHelper struct {
	ctx            context.Context
	req            *http.Request
	res            *httptest.ResponseRecorder
	service        *service
	exampleUser    *types.User
	exampleAccount *types.Account
}

func buildTestHelper(t *testing.T) *accountsServiceHTTPRoutesTestHelper {
	t.Helper()

	helper := &accountsServiceHTTPRoutesTestHelper{}

	helper.ctx = context.Background()
	helper.service = buildTestService()
	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID

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
	helper.service.accountIDFetcher = func(req *http.Request) uint64 {
		return helper.exampleAccount.ID
	}
	helper.service.userIDFetcher = func(req *http.Request) uint64 {
		return helper.exampleUser.ID
	}

	helper.res = httptest.NewRecorder()
	helper.req, err = http.NewRequestWithContext(
		helper.ctx,
		http.MethodGet,
		"https://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)
	require.NotNil(t, helper.req)
	require.NoError(t, err)

	return helper
}
