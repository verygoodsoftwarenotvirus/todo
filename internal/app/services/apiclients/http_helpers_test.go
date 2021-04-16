package apiclients

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

type apiClientsServiceHTTPRoutesTestHelper struct {
	ctx              context.Context
	req              *http.Request
	res              *httptest.ResponseRecorder
	service          *service
	exampleUser      *types.User
	exampleAccount   *types.Account
	exampleAPIClient *types.APIClient
	exampleInput     *types.APIClientCreationInput
}

func buildTestHelper(t *testing.T) *apiClientsServiceHTTPRoutesTestHelper {
	t.Helper()

	helper := &apiClientsServiceHTTPRoutesTestHelper{}

	helper.ctx = context.Background()
	helper.service = buildTestService(t)
	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleAPIClient = fakes.BuildFakeAPIClient()
	helper.exampleAPIClient.BelongsToUser = helper.exampleUser.ID
	helper.exampleInput = fakes.BuildFakeAPIClientCreationInputFromClient(helper.exampleAPIClient)

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
	helper.service.urlClientIDExtractor = func(*http.Request) uint64 {
		return helper.exampleAPIClient.ID
	}

	req := testutil.BuildTestRequest(t)

	helper.req = req.WithContext(context.WithValue(req.Context(), types.SessionContextDataKey, sessionCtxData))
	helper.req = helper.req.WithContext(context.WithValue(helper.req.Context(), creationMiddlewareCtxKey, helper.exampleInput))

	helper.res = httptest.NewRecorder()

	return helper
}
