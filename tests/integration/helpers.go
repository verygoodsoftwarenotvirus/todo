package integration

import (
	"context"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"
)

func checkValueAndError(t *testing.T, i interface{}, err error) {
	t.Helper()

	require.NoError(t, err)
	require.NotNil(t, i)
}

func createUserAndClientForTest(ctx context.Context, t *testing.T) (*types.User, *httpclient.Client) {
	t.Helper()

	user, err := testutil.CreateServiceUser(ctx, urlToUse, "", debug)
	require.NoError(t, err)

	oa2Client, err := testutil.CreateObligatoryOAuth2Client(ctx, urlToUse, user)
	require.NoError(t, err)

	return user, initializeClient(oa2Client)
}

func generateTOTPTokenForUser(t *testing.T, u *types.User) string {
	t.Helper()

	code, err := totp.GenerateCode(u.TwoFactorSecret, time.Now().UTC())
	require.NotEmpty(t, code)
	require.NoError(t, err)

	return code
}

/*
func runTestForClientAndCookie(ctx context.Context, t *testing.T, testFunc func(*testing.T, *httpclient.Client)) {
	t.Helper()

	user, testClient := createUserAndClientForTest(ctx, t)
	cookie, err := testClient.Login(ctx, &types.UserLoginInput{
		Username:  user.Username,
		Password:  user.HashedPassword,
		TOTPToken: generateTOTPTokenForUser(t, user),
	})
	require.NoError(t, err)

	testFunc(t, testClient)
	testClient.SetOption(httpclient.WithCookieCredentials(cookie))
	testFunc(t, testClient)
}
*/
