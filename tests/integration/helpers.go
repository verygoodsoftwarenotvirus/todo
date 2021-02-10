package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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

func runTestForClientAndCookie(ctx context.Context, t *testing.T, testName string, testFunc func(*httpclient.Client) func(*testing.T)) {
	t.Helper()

	_, testClient := createUserAndClientForTest(ctx, t)

	//	, cookie, err := testClient.Login(ctx, &types.UserLoginInput{
	//	, 	Username:  user.Username,
	//	, 	Password:  user.HashedPassword,
	//	, 	TOTPToken: generateTOTPTokenForUser(t, user),
	//	, })
	//	, require.NoError(t, err)
	//  ,
	//	, t.Run(testName, testFunc(testClient))
	//	, testClient.SetOption(httpclient.WithCookieCredentials(cookie))
	//	, t.Run(fmt.Sprintf("%s with cookie", testName), testFunc(testClient))

	t.Run(testName, testFunc(testClient))
}

func validateAuditLogEntries(t *testing.T, expectedEntries, actualEntries []*types.AuditLogEntry, relevantID uint64, key string) {
	t.Helper()

	require.Len(t, actualEntries, len(expectedEntries))

	expectedEventTypes := []string{}
	for _, e := range expectedEntries {
		expectedEventTypes = append(expectedEventTypes, e.EventType)
	}

	actualEventTypes := []string{}
	for _, e := range actualEntries {
		actualEventTypes = append(actualEventTypes, e.EventType)

		if relevantID != 0 && key != "" {
			require.Contains(t, e.Context, key)
			assert.EqualValues(t, relevantID, e.Context[key])
		}
	}

	assert.Subset(t, expectedEventTypes, actualEventTypes)
}
