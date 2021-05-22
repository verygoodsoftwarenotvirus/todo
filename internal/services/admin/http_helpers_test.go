package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
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
	helper.exampleUser.ServiceRoles = []string{authorization.ServiceUserRole.String()}
	helper.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
		return &types.SessionContextData{
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
		}, nil
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
	helper.exampleUser.ServiceRoles = []string{authorization.ServiceAdminRole.String()}
	helper.exampleAccount = fakes.BuildFakeAccount()
	helper.exampleAccount.BelongsToUser = helper.exampleUser.ID
	helper.exampleInput = fakes.BuildFakeUserReputationUpdateInput()

	helper.res = httptest.NewRecorder()
	helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://blah.com", nil)
	require.NoError(t, err)
	require.NotNil(t, helper.req)

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
	helper.service.userIDFetcher = func(req *http.Request) uint64 {
		return helper.exampleUser.ID
	}

	return helper
}
