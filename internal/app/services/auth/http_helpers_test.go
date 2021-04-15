package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func attachCookieToRequestForTest(t *testing.T, s *service, req *http.Request, user *types.User) (context.Context, *http.Request) {
	t.Helper()

	exampleAccount := fakes.BuildFakeAccount()

	ctx, sessionErr := s.sessionManager.Load(req.Context(), "")
	require.NoError(t, sessionErr)
	require.NoError(t, s.sessionManager.RenewToken(ctx))

	s.sessionManager.Put(ctx, userIDContextKey, user.ID)
	s.sessionManager.Put(ctx, accountIDContextKey, exampleAccount.ID)

	token, _, err := s.sessionManager.Commit(ctx)
	assert.NotEmpty(t, token)
	assert.NoError(t, err)

	c, err := s.buildCookie(token, time.Now().Add(s.config.Cookies.Lifetime))
	require.NoError(t, err)
	req.AddCookie(c)

	return ctx, req.WithContext(ctx)
}

type authServiceHTTPRoutesTestHelper struct {
	ctx               context.Context
	req               *http.Request
	res               *httptest.ResponseRecorder
	sessionCtxData    *types.SessionContextData
	service           *service
	exampleUser       *types.User
	exampleAccount    *types.Account
	exampleAPIClient  *types.APIClient
	examplePerms      map[uint64]*types.UserAccountMembershipInfo
	exampleLoginInput *types.UserLoginInput
}

func (helper *authServiceHTTPRoutesTestHelper) setContextFetcher(t *testing.T) {
	t.Helper()

	sessionCtxData, err := types.SessionContextDataFromUser(helper.exampleUser, helper.exampleAccount.ID, helper.examplePerms)
	require.NoError(t, err)

	helper.sessionCtxData = sessionCtxData
	helper.service.sessionContextDataFetcher = func(_ *http.Request) (*types.SessionContextData, error) {
		return sessionCtxData, nil
	}
}

func buildTestHelper(t *testing.T) *authServiceHTTPRoutesTestHelper {
	t.Helper()

	helper := &authServiceHTTPRoutesTestHelper{}

	helper.ctx = context.Background()
	helper.service = buildTestService(t)
	helper.exampleUser = fakes.BuildFakeUser()
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleAPIClient = fakes.BuildFakeAPIClient()
	helper.exampleAPIClient.BelongsToUser = helper.exampleUser.ID
	helper.exampleLoginInput = fakes.BuildFakeUserLoginInputFromUser(helper.exampleUser)

	helper.examplePerms = map[uint64]*types.UserAccountMembershipInfo{
		helper.exampleAccount.ID: {
			AccountName: helper.exampleAccount.Name,
			Permissions: testutil.BuildMaxUserPerms(),
		},
	}

	helper.setContextFetcher(t)

	helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

	var err error

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
