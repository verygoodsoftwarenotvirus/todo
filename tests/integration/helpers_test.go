package integration

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitForAsynchronousStuffBecauseProperWebhookNotificationsHaveNotBeenImplementedYet() {
	time.Sleep(5 * time.Second)
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func requireNotNilAndNoProblems(t *testing.T, i interface{}, err error) {
	t.Helper()

	require.NoError(t, err)
	require.NotNil(t, i)
}

func createUserAndClientForTest(ctx context.Context, t *testing.T) (user *types.User, cookie *http.Cookie, cookieClient, pasetoClient *httpclient.Client) {
	t.Helper()

	user, err := testutils.CreateServiceUser(ctx, urlToUse, fakes.BuildFakeUser().Username)
	require.NoError(t, err)

	t.Logf("created user %q: %q", user.ID, user.Username)

	cookie, err = testutils.GetLoginCookie(ctx, urlToUse, user)
	require.NoError(t, err)

	cookieClient, err = initializeCookiePoweredClient(cookie)
	require.NoError(t, err)

	apiClient, err := cookieClient.CreateAPIClient(ctx, cookie, &types.APIClientCreationInput{
		Name: t.Name(),
		UserLoginInput: types.UserLoginInput{
			Username:  user.Username,
			Password:  user.HashedPassword,
			TOTPToken: generateTOTPTokenForUser(t, user),
		},
	})
	require.NoError(t, err)

	secretKey, err := base64.RawURLEncoding.DecodeString(apiClient.ClientSecret)
	require.NoError(t, err)

	pasetoClient, err = initializePASETOPoweredClient(apiClient.ClientID, secretKey)
	require.NoError(t, err)

	return user, cookie, cookieClient, pasetoClient
}

func initializeCookiePoweredClient(cookie *http.Cookie) (*httpclient.Client, error) {
	if parsedURLToUse == nil {
		panic("url not set!")
	}

	logger := logging.ProvideLogger(logging.Config{Provider: logging.ProviderZerolog})

	c, err := httpclient.NewClient(parsedURLToUse,
		httpclient.UsingLogger(logger),
		httpclient.UsingCookie(cookie),
	)
	if err != nil {
		return nil, err
	}

	if debug {
		if setOptionErr := c.SetOptions(httpclient.UsingDebug(true)); setOptionErr != nil {
			return nil, setOptionErr
		}
	}

	return c, nil
}

func initializePASETOPoweredClient(clientID string, secretKey []byte) (*httpclient.Client, error) {
	c, err := httpclient.NewClient(parsedURLToUse,
		httpclient.UsingLogger(logging.NewNoopLogger()),
		httpclient.UsingPASETO(clientID, secretKey),
	)
	if err != nil {
		return nil, err
	}

	if debug {
		if setOptionErr := c.SetOptions(httpclient.UsingDebug(true)); setOptionErr != nil {
			return nil, setOptionErr
		}
	}

	return c, nil
}

func buildSimpleClient(t *testing.T) *httpclient.Client {
	t.Helper()

	c, err := httpclient.NewClient(parsedURLToUse)
	require.NoError(t, err)

	return c
}

func generateTOTPTokenForUser(t *testing.T, u *types.User) string {
	t.Helper()

	code, err := totp.GenerateCode(u.TwoFactorSecret, time.Now().UTC())
	require.NotEmpty(t, code)
	require.NoError(t, err)

	return code
}

func buildAdminCookieAndPASETOClients(ctx context.Context, t *testing.T) (cookieClient, pasetoClient *httpclient.Client) {
	t.Helper()

	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	u := testutils.DetermineServiceURL()
	urlToUse = u.String()
	logger := logging.ProvideLogger(logging.Config{Provider: logging.ProviderZerolog})

	logger.WithValue(keys.URLKey, urlToUse).Info("checking server")
	testutils.EnsureServerIsUp(ctx, urlToUse)

	adminCookie, err := testutils.GetLoginCookie(ctx, urlToUse, premadeAdminUser)
	require.NoError(t, err)

	cClient, err := initializeCookiePoweredClient(adminCookie)
	require.NoError(t, err)

	code, err := totp.GenerateCode(premadeAdminUser.TwoFactorSecret, time.Now().UTC())
	require.NoError(t, err)

	apiClient, err := cClient.CreateAPIClient(ctx, adminCookie, &types.APIClientCreationInput{
		Name: fmt.Sprintf("admin_paseto_client_%d", time.Now().UnixNano()),
		UserLoginInput: types.UserLoginInput{
			Username:  premadeAdminUser.Username,
			Password:  premadeAdminUser.HashedPassword,
			TOTPToken: code,
		},
	})
	require.NoError(t, err)

	secretKey, err := base64.RawURLEncoding.DecodeString(apiClient.ClientSecret)
	require.NoError(t, err)

	PASETOClient, err := initializePASETOPoweredClient(apiClient.ClientID, secretKey)
	require.NoError(t, err)

	return cClient, PASETOClient
}

func validateAuditLogEntries(t *testing.T, expectedEntries, actualEntries []*types.AuditLogEntry, relevantID, key string) {
	t.Helper()

	expectedEventTypes := []string{}
	actualEventTypes := []string{}

	for _, e := range expectedEntries {
		expectedEventTypes = append(expectedEventTypes, e.EventType)
	}

	for _, e := range actualEntries {
		actualEventTypes = append(actualEventTypes, e.EventType)

		if relevantID != "" && key != "" {
			if assert.Contains(t, e.Context, key) {
				assert.EqualValues(t, relevantID, e.Context[key])
			}
		}
	}

	assert.Equal(t, len(expectedEntries), len(actualEntries), "expected %q, got %q", strings.Join(expectedEventTypes, ","), strings.Join(actualEventTypes, ","))

	assert.Subset(t, expectedEventTypes, actualEventTypes)
}
