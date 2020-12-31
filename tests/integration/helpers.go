package integration

import (
	"context"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"
)

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
