package integration

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"
)

type testClientWrapper struct {
	main  *httpclient.Client
	admin *httpclient.Client
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func checkValueAndError(t *testing.T, i interface{}, err error) {
	t.Helper()

	require.NoError(t, err)
	require.NotNil(t, i)
}

func createUserAndClientForTest(ctx context.Context, t *testing.T) (user *types.User, cookie *http.Cookie, cookieClient, pasetoClient *httpclient.Client) {
	t.Helper()

	var err error

	user, err = utils.CreateServiceUser(ctx, urlToUse, "")
	require.NoError(t, err)

	t.Logf("created user: %q", user.Username)

	cookie, err = utils.GetLoginCookie(ctx, urlToUse, user)
	require.NoError(t, err)

	cookieClient, err = initializeCookiePoweredClient(cookie)
	require.NoError(t, err)

	apiClient, err := cookieClient.CreateAPIClient(ctx, cookie, &types.APICientCreationInput{
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

func buildHTTPClient() *http.Client {
	return &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   timeout,
	}
}

func initializeCookiePoweredClient(cookie *http.Cookie) (*httpclient.Client, error) {
	c, err := httpclient.NewClient(
		httpclient.UsingURI(urlToUse),
		httpclient.UsingLogger(logging.NewNonOperationalLogger()),
		httpclient.UsingHTTPClient(buildHTTPClient()),
		httpclient.UsingCookie(cookie),
	)
	if err != nil {
		return nil, err
	}

	if debug {
		if setOptionErr := c.SetOptions(httpclient.WithDebug()); setOptionErr != nil {
			return nil, setOptionErr
		}
	}

	return c, nil
}
func initializePASETOPoweredClient(clientID string, secretKey []byte) (*httpclient.Client, error) {
	c, err := httpclient.NewClient(
		httpclient.UsingURI(urlToUse),
		httpclient.UsingLogger(logging.NewNonOperationalLogger()),
		httpclient.UsingHTTPClient(buildHTTPClient()),
		httpclient.UsingPASETO(clientID, secretKey),
	)
	if err != nil {
		return nil, err
	}

	if debug {
		if setOptionErr := c.SetOptions(httpclient.WithDebug()); setOptionErr != nil {
			return nil, setOptionErr
		}
	}

	return c, nil
}

func buildSimpleClient(t *testing.T) *httpclient.Client {
	t.Helper()

	c, err := httpclient.NewClient(httpclient.UsingURI(urlToUse))
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

func buildAdminCookieAndPASETOClients(ctx context.Context) (cookieClient, pasetoClient *httpclient.Client) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	urlToUse = testutil.DetermineServiceURL()
	logger := zerolog.NewLogger()

	logger.WithValue(keys.URLKey, urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(ctx, urlToUse)

	adminCookie, err := utils.GetLoginCookie(ctx, urlToUse, premadeAdminUser)
	if err != nil {
		logger.Fatal(err)
	}

	cClient, err := initializeCookiePoweredClient(adminCookie)
	if err != nil {
		logger.Fatal(err)
	}

	code, err := totp.GenerateCode(premadeAdminUser.TwoFactorSecret, time.Now().UTC())
	if err != nil {
		logger.Fatal(err)
	}

	apiClient, err := cClient.CreateAPIClient(ctx, adminCookie, &types.APICientCreationInput{
		Name: "admin_paseto_client",
		UserLoginInput: types.UserLoginInput{
			Username:  premadeAdminUser.Username,
			Password:  premadeAdminUser.HashedPassword,
			TOTPToken: code,
		},
	})
	if err != nil {
		logger.Fatal(err)
	}

	secretKey, err := base64.RawURLEncoding.DecodeString(apiClient.ClientSecret)
	if err != nil {
		logger.Fatal(err)
	}

	PASETOClient, err := initializePASETOPoweredClient(apiClient.ClientID, secretKey)
	if err != nil {
		logger.Fatal(err)
	}

	return cClient, PASETOClient
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
