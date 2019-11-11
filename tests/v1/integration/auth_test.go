package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"
	randmodel "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil/rand/model"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

func loginUser(t *testing.T, username, password, totpSecret string) *http.Cookie {
	loginURL := fmt.Sprintf("%s://%s:%s/users/login", todoClient.URL.Scheme, todoClient.URL.Hostname(), todoClient.URL.Port())

	code, err := totp.GenerateCode(strings.ToUpper(totpSecret), time.Now().UTC())
	assert.NoError(t, err)

	bodyStr := fmt.Sprintf(`
	{
		"username": %q,
		"password": %q,
		"totp_token": %q
	}
`, username, password, code)

	body := strings.NewReader(bodyStr)
	req, _ := http.NewRequest(http.MethodPost, loginURL, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "login should be successful")

	cookies := resp.Cookies()
	if len(cookies) == 1 {
		return cookies[0]
	}
	t.Logf("wrong number of cookies found: %d", len(cookies))
	t.FailNow()

	return nil
}

func TestAuth(test *testing.T) {
	test.Parallel()

	test.Run("should be able to login", func(t *testing.T) {
		tctx := context.Background()

		// create a user
		ui := randmodel.RandomUserInput()
		req, err := todoClient.BuildCreateUserRequest(tctx, ui)
		checkValueAndError(t, req, err)

		res, err := todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)

		// load user response
		ucr := &models.UserCreationResponse{}
		require.NoError(t, json.NewDecoder(res.Body).Decode(ucr))

		// create login request
		token, err := totp.GenerateCode(ucr.TwoFactorSecret, time.Now().UTC())
		checkValueAndError(t, token, err)
		r := &models.UserLoginInput{
			Username:  ucr.Username,
			Password:  ui.Password,
			TOTPToken: token,
		}
		out, err := json.Marshal(r)
		require.NoError(t, err)
		body := bytes.NewReader(out)

		u, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u.Path = "/users/login"

		req, err = http.NewRequest(http.MethodPost, u.String(), body)
		checkValueAndError(t, req, err)

		// execute login request
		res, err = todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusNoContent, res.StatusCode)

		cookies := res.Cookies()
		assert.Len(t, cookies, 1)
	})

	test.Run("should be able to logout", func(t *testing.T) {
		tctx := context.Background()

		ui := randmodel.RandomUserInput()
		req, err := todoClient.BuildCreateUserRequest(tctx, ui)
		checkValueAndError(t, req, err)

		res, err := todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)

		ucr := &models.UserCreationResponse{}
		require.NoError(t, json.NewDecoder(res.Body).Decode(ucr))

		token, err := totp.GenerateCode(ucr.TwoFactorSecret, time.Now().UTC())
		checkValueAndError(t, token, err)
		r := &models.UserLoginInput{
			Username:  ucr.Username,
			Password:  ui.Password,
			TOTPToken: token,
		}
		out, err := json.Marshal(r)
		require.NoError(t, err)
		body := bytes.NewReader(out)

		u, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u.Path = "/users/login"

		req, err = http.NewRequest(http.MethodPost, u.String(), body)
		checkValueAndError(t, req, err)

		// execute login request
		res, err = todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusNoContent, res.StatusCode)

		// extract cookie
		cookies := res.Cookies()
		require.Len(t, cookies, 1)
		loginCookie := cookies[0]

		// build logout request
		u2, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u2.Path = "/users/logout"

		req, err = http.NewRequest(http.MethodPost, u2.String(), nil)
		checkValueAndError(t, req, err)
		req.AddCookie(loginCookie)

		// execute logout request
		res, err = todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	test.Run("login request without body fails", func(t *testing.T) {
		u, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u.Path = "/users/login"

		req, err := http.NewRequest(http.MethodPost, u.String(), nil)
		checkValueAndError(t, req, err)

		// execute login request
		res, err := todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	test.Run("should not be able to log in with the wrong password", func(t *testing.T) {
		tctx := context.Background()

		// create a user
		ui := randmodel.RandomUserInput()
		req, err := todoClient.BuildCreateUserRequest(tctx, ui)
		checkValueAndError(t, req, err)

		res, err := todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)

		// load user response
		ucr := &models.UserCreationResponse{}
		require.NoError(t, json.NewDecoder(res.Body).Decode(ucr))

		// create login request
		var badPassword string
		for _, v := range ui.Password {
			badPassword = string(v) + badPassword
		}

		// create login request
		token, err := totp.GenerateCode(ucr.TwoFactorSecret, time.Now().UTC())
		checkValueAndError(t, token, err)
		r := &models.UserLoginInput{
			Username:  ucr.Username,
			Password:  badPassword,
			TOTPToken: token,
		}
		out, err := json.Marshal(r)
		require.NoError(t, err)
		body := bytes.NewReader(out)

		u, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u.Path = "/users/login"

		req, err = http.NewRequest(http.MethodPost, u.String(), body)
		checkValueAndError(t, req, err)

		// execute login request
		res, err = todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	test.Run("should not be able to login as someone that doesn't exist", func(t *testing.T) {
		ui := randmodel.RandomUserInput()

		s, err := randString()
		require.NoError(t, err)

		token, err := totp.GenerateCode(s, time.Now().UTC())
		checkValueAndError(t, token, err)
		r := &models.UserLoginInput{
			Username:  ui.Username,
			Password:  ui.Password,
			TOTPToken: token,
		}
		out, err := json.Marshal(r)
		require.NoError(t, err)
		body := bytes.NewReader(out)

		u, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u.Path = "/users/login"

		req, err := http.NewRequest(http.MethodPost, u.String(), body)
		checkValueAndError(t, req, err)

		res, err := todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

		cookies := res.Cookies()
		assert.Len(t, cookies, 0)
	})

	test.Run("should reject an unauthenticated request", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, todoClient.BuildURL(nil, "webhooks"), nil)
		assert.NoError(t, err)

		res, err := todoClient.PlainClient().Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	test.Run("should be able to change password", func(t *testing.T) {
		// create user
		user, ui, cookie := buildDummyUser(test)
		require.NotNil(test, cookie)

		// create login request
		var backwardsPass string
		for _, v := range ui.Password {
			backwardsPass = string(v) + backwardsPass
		}

		// create password update request
		token, err := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
		checkValueAndError(t, token, err)
		r := &models.PasswordUpdateInput{
			CurrentPassword: ui.Password,
			TOTPToken:       token,
			NewPassword:     backwardsPass,
		}
		out, err := json.Marshal(r)
		require.NoError(t, err)
		body := bytes.NewReader(out)

		u, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u.Path = "/users/password/new"

		req, err := http.NewRequest(http.MethodPut, u.String(), body)
		checkValueAndError(t, req, err)
		req.AddCookie(cookie)

		// execute password update request
		res, err := todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusAccepted, res.StatusCode)

		// logout

		u2, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u2.Path = "/users/logout"

		req, err = http.NewRequest(http.MethodPost, u2.String(), nil)
		checkValueAndError(t, req, err)
		req.AddCookie(cookie)

		res, err = todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		// create login request
		newToken, err := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
		checkValueAndError(t, newToken, err)
		l, err := json.Marshal(&models.UserLoginInput{
			Username:  user.Username,
			Password:  backwardsPass,
			TOTPToken: newToken,
		})
		require.NoError(t, err)
		body = bytes.NewReader(l)

		u3, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u3.Path = "/users/login"

		req, err = http.NewRequest(http.MethodPost, u3.String(), body)
		checkValueAndError(t, req, err)

		// execute login request
		res, err = todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusNoContent, res.StatusCode)

		cookies := res.Cookies()
		require.Len(t, cookies, 1)
		assert.NotEqual(t, cookie, cookies[0])
	})

	test.Run("should be able to change 2FA Token", func(t *testing.T) {
		// create user
		user, ui, cookie := buildDummyUser(test)
		require.NotNil(test, cookie)

		// create TOTP secret update request
		token, err := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
		checkValueAndError(t, token, err)
		ir := &models.TOTPSecretRefreshInput{
			CurrentPassword: ui.Password,
			TOTPToken:       token,
		}
		out, err := json.Marshal(ir)
		require.NoError(t, err)
		body := bytes.NewReader(out)

		u, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u.Path = "/users/totp_secret/new"

		req, err := http.NewRequest(http.MethodPost, u.String(), body)
		checkValueAndError(t, req, err)
		req.AddCookie(cookie)

		// execute TOTP secret update request
		res, err := todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusAccepted, res.StatusCode)

		// load user response
		r := &models.TOTPSecretRefreshResponse{}
		require.NoError(t, json.NewDecoder(res.Body).Decode(r))
		require.NotEqual(t, user.TwoFactorSecret, r.TwoFactorSecret)

		// logout

		u2, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u2.Path = "/users/logout"

		req, err = http.NewRequest(http.MethodPost, u2.String(), nil)
		checkValueAndError(t, req, err)
		req.AddCookie(cookie)

		res, err = todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		// create login request
		newToken, err := totp.GenerateCode(r.TwoFactorSecret, time.Now().UTC())
		checkValueAndError(t, newToken, err)
		l, err := json.Marshal(&models.UserLoginInput{
			Username:  user.Username,
			Password:  ui.Password,
			TOTPToken: newToken,
		})
		require.NoError(t, err)
		body = bytes.NewReader(l)

		u3, err := url.Parse(todoClient.BuildURL(nil))
		require.NoError(t, err)
		u3.Path = "/users/login"

		req, err = http.NewRequest(http.MethodPost, u3.String(), body)
		checkValueAndError(t, req, err)

		// execute login request
		res, err = todoClient.PlainClient().Do(req)
		checkValueAndError(t, res, err)
		assert.Equal(t, http.StatusNoContent, res.StatusCode)

		cookies := res.Cookies()
		require.Len(t, cookies, 1)
		assert.NotEqual(t, cookie, cookies[0])
	})

	test.Run("should accept a login cookie if a token is missing", func(t *testing.T) {
		// create user
		_, _, cookie := buildDummyUser(test)
		assert.NotNil(test, cookie)

		req, err := http.NewRequest(http.MethodGet, todoClient.BuildURL(nil, "webhooks"), nil)
		assert.NoError(t, err)
		req.AddCookie(cookie)

		res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	test.Run("should only allow users to see their own content", func(t *testing.T) {
		tctx := context.Background()

		// create user and oauth2 client A
		userA, err := testutil.CreateObligatoryUser(urlToUse, debug)
		require.NoError(t, err)

		ca, err := testutil.CreateObligatoryClient(urlToUse, userA)
		require.NoError(t, err)

		clientA, err := client.NewClient(
			tctx,
			ca.ClientID,
			ca.ClientSecret,
			todoClient.URL,
			noop.ProvideNoopLogger(),
			buildHTTPClient(),
			ca.Scopes,
			true,
		)
		checkValueAndError(test, clientA, err)

		// create user and oauth2 client B
		userB, err := testutil.CreateObligatoryUser(urlToUse, debug)
		require.NoError(t, err)

		cb, err := testutil.CreateObligatoryClient(urlToUse, userB)
		require.NoError(t, err)

		clientB, err := client.NewClient(
			tctx,
			cb.ClientID,
			cb.ClientSecret,
			todoClient.URL,
			noop.ProvideNoopLogger(),
			buildHTTPClient(),
			cb.Scopes,
			true,
		)
		checkValueAndError(test, clientA, err)

		// create webhook for user A
		webhookA, err := clientA.CreateWebhook(tctx, &models.WebhookCreationInput{
			Method: http.MethodPatch,
			Name:   "A",
		})
		checkValueAndError(t, webhookA, err)

		// create webhook for user B
		webhookB, err := clientB.CreateWebhook(tctx, &models.WebhookCreationInput{
			Method: http.MethodPatch,
			Name:   "B",
		})
		checkValueAndError(t, webhookB, err)

		i, err := clientB.GetWebhook(tctx, webhookA.ID)
		assert.Nil(t, i)
		assert.Error(t, err, "should experience error trying to fetch entry they're not authorized for")

		// Clean up
		assert.NoError(t, todoClient.ArchiveWebhook(tctx, webhookA.ID))
		assert.NoError(t, todoClient.ArchiveWebhook(tctx, webhookB.ID))
	})

	test.Run("should only allow clients with a given scope to see that scope's content", func(t *testing.T) {
		tctx := context.Background()

		// create user
		x, y, cookie := buildDummyUser(test)
		assert.NotNil(test, cookie)

		input := buildDummyOAuth2ClientInput(test, x.Username, y.Password, x.TwoFactorSecret)
		input.Scopes = []string{"absolutelynevergonnaexistascopelikethis"}
		premade, err := todoClient.CreateOAuth2Client(tctx, cookie, input)
		checkValueAndError(test, premade, err)

		c, err := client.NewClient(
			context.Background(),
			premade.ClientID,
			premade.ClientSecret,
			todoClient.URL,
			noop.ProvideNoopLogger(),
			buildHTTPClient(),
			premade.Scopes,
			true,
		)
		checkValueAndError(test, c, err)

		i, err := c.GetOAuth2Clients(tctx, nil)
		assert.Nil(t, i)
		assert.Error(t, err, "should experience error trying to fetch entry they're not authorized for")
	})
}
