package integration

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
)

func loginUser(t *testing.T, username, password, totpSecret string) *http.Cookie {
	loginURL := fmt.Sprintf("%s://%s:%s/users/login", todoClient.URL.Scheme, todoClient.URL.Hostname(), todoClient.URL.Port())

	code, err := totp.GenerateCode(strings.ToUpper(totpSecret), time.Now())
	assert.NoError(t, err)

	body := strings.NewReader(fmt.Sprintf(`
		{
			"username": %q,
			"password": %q,
			"totp_token": %q
		}
	`, username, password, code))

	req, _ := http.NewRequest(http.MethodPost, loginURL, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode, "login should be successful")

	cookies := resp.Cookies()
	if len(cookies) == 1 {
		return resp.Cookies()[0]
	}
	t.Logf("wrong number of cookies found: %d", len(cookies))
	t.FailNow()

	return nil
}

func TestAuth(test *testing.T) {
	test.Parallel()

	test.Run("should reject an unauthenticated request", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, todoClient.BuildURL(nil, "items"), nil)
		assert.NoError(t, err)

		res, err := (*http.Client)(&http.Client{Timeout: 10 * time.Second}).Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	test.Run("should accept a login cookie if a token is missing", func(t *testing.T) {
		t.SkipNow()
		// create user
		_, _, cookie := buildDummyUser(test)
		assert.NotNil(test, cookie)

		req, err := http.NewRequest(http.MethodGet, todoClient.BuildURL(nil, "items"), nil)
		assert.NoError(t, err)
		req.AddCookie(cookie)

		res, err := (*http.Client)(&http.Client{Timeout: 10 * time.Second}).Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	// test.Run("should only allow users to see their own content", func(t *testing.T) {
	// 	tctx := context.Background()

	// 	// create user
	// 	x, y, cookie := buildDummyUser(context.Background(), test)
	// 	assert.NotNil(test, cookie)

	// 	input := buildDummyOAuth2ClientInput(test, x.Username, y.Password, x.TwoFactorSecret)
	// 	premade, err := todoClient.CreateOAuth2Client(context.Background(), input, cookie)
	// 	checkValueAndError(test, premade, err)

	// 	c, err := client.NewClient(
	// 		premade.ClientID,
	// 		premade.ClientSecret,
	// 		todoClient.URL,
	// 		noop.ProvideNoopLogger(),
	// 		buildHTTPClient(),
	// 		tracing.ProvideNoopTracer(),
	// 		true,
	// 	)
	// 	checkValueAndError(test, c, err)

	// 	// Create item for user A
	// 	a, err := todoClient.CreateItem(
	// 		tctx,
	// 		&models.ItemInput{
	// 			Name:    "name A",
	// 			Details: "details A",
	// 		})
	// 	checkValueAndError(t, a, err)

	// 	// Create item for user B
	// 	b, err := c.CreateItem(
	// 		tctx,
	// 		&models.ItemInput{
	// 			Name:    "name B",
	// 			Details: "details B",
	// 		})
	// 	checkValueAndError(t, b, err)

	// 	i, err := c.GetItem(tctx, a.ID)
	// 	assert.Nil(t, i)
	// 	assert.Error(t, err, "should experience error trying to fetch item they're not authorized for")

	// 	// Clean up
	// 	assert.NoError(t, todoClient.DeleteItem(tctx, a.ID))
	// 	assert.NoError(t, todoClient.DeleteItem(tctx, b.ID))

	// })

}
