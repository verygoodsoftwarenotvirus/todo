package integration

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
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

func createUserAndClientForTest(ctx context.Context, t *testing.T) (user *types.User, cookie *http.Cookie, cookieClient, pasetoClient *httpclient.Client) {
	t.Helper()

	var err error

	user, err = testutil.CreateServiceUser(ctx, urlToUse, "", debug)
	require.NoError(t, err)

	t.Logf("created user: %q", user.Username)

	cookie, err = testutil.GetLoginCookie(ctx, urlToUse, user)
	require.NoError(t, err)

	cookieClient = initializeCookiePoweredClient(cookie)

	delegatedClient, err := cookieClient.CreateDelegatedClient(ctx, cookie, &types.DelegatedClientCreationInput{
		Name: t.Name(),
		UserLoginInput: types.UserLoginInput{
			Username:  user.Username,
			Password:  user.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, user),
		},
	})
	require.NoError(t, err)

	secretKey, err := base64.RawURLEncoding.DecodeString(delegatedClient.ClientSecret)
	require.NoError(t, err)

	return user, cookie, cookieClient, initializePASETOPoweredClient(delegatedClient.ClientID, secretKey)
}

func generateTOTPTokenForUser(t *testing.T, u *types.User) string {
	t.Helper()

	code, err := totp.GenerateCode(u.TwoFactorSecret, time.Now().UTC())
	require.NotEmpty(t, code)
	require.NoError(t, err)

	return code
}

func runTestForAllAuthMethodsAsAdmin(t *testing.T, testName string, testFunc func(ctx context.Context, client *httpclient.Client) func(*testing.T)) {
	t.Helper()

	ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
	defer span.End()

	adminClientLock.Lock()
	defer adminClientLock.Unlock()

	t.Run(fmt.Sprintf("%s with cookie", testName), testFunc(ctx, adminCookieClient))
	t.Run(fmt.Sprintf("%s with paseto", testName), testFunc(ctx, adminPASETOClient))
}

func runTestForAllAuthMethods(t *testing.T, testName string, testFunc func(ctx context.Context, user *types.User, cookie *http.Cookie, client *httpclient.Client) func(*testing.T)) {
	t.Helper()

	ctx, span := tracing.StartCustomSpan(context.Background(), t.Name())
	defer span.End()

	user, cookie, cookieClient, pasetoClient := createUserAndClientForTest(ctx, t)

	t.Run(fmt.Sprintf("%s with cookie", testName), testFunc(ctx, user, cookie, cookieClient))
	t.Run(fmt.Sprintf("%s with paseto token", testName), testFunc(ctx, user, cookie, pasetoClient))
}

func validateAuditLogEntries(t *testing.T, expectedEntries, actualEntries []*types.AuditLogEntry, relevantID uint64, key string) {
	t.Helper()

	expectedEventTypes := []string{}
	actualEventTypes := []string{}

	for _, e := range expectedEntries {
		expectedEventTypes = append(expectedEventTypes, e.EventType)
	}

	for _, e := range actualEntries {
		actualEventTypes = append(actualEventTypes, e.EventType)

		if relevantID != 0 && key != "" {
			if assert.Contains(t, e.Context, key) {
				assert.EqualValues(t, relevantID, e.Context[key])
			}
		}
	}

	assert.Equal(t, len(expectedEntries), len(actualEntries), "expected %q, got %q", strings.Join(expectedEventTypes, ","), strings.Join(actualEventTypes, ","))

	assert.Subset(t, expectedEventTypes, actualEventTypes)
}
