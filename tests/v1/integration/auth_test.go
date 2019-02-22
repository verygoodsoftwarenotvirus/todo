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

	pc := todoClient.PlainClient()

	test.Run("should reject an unauthenticated request", func(t *testing.T) {
		res, err := pc.Post(todoClient.BuildURL(nil, "fart"), "application/json", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

}
