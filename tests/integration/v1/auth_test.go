package integration

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
)

const (
	expectedUsername   = "username"
	expectedPassword   = "password"
	expectedTOTPSecret = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
)

func loginUser(t *testing.T, username string, password string) *http.Cookie {
	loginURL := fmt.Sprintf("%s://%s/users/login", todoClient.URL.Scheme, todoClient.URL.Hostname())

	code, err := totp.GenerateCode(strings.ToUpper(expectedTOTPSecret), time.Now())
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
	} else {
		t.Logf("wrong number of cookies found: %d", len(cookies))
		t.FailNow()
	}
	return nil
}

func TestAuth(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Auth", func() {
		g.It("should reject an unauthenticated request", func() {
			res, err := todoClient.Client.Post(todoClient.BuildURL(nil, "fart"), "application/json", nil)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		})

		g.Describe("credentials", func() {
			g.It("should accept a valid cookie", func() {
				cookie := loginUser(t, expectedUsername, expectedPassword)
				assert.NotNil(t, cookie)

				req, err := http.NewRequest(http.MethodPost, todoClient.BuildURL(nil, "fart"), nil)
				req.AddCookie(cookie)
				assert.NoError(t, err)
				res, err := todoClient.Client.Do(req)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusTeapot, res.StatusCode)
			})

			g.It("should reject an invalid cookie", func() {
				req, err := http.NewRequest(http.MethodPost, todoClient.BuildURL(nil, "fart"), nil)
				assert.NoError(t, err)
				res, err := todoClient.Client.Do(req)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
			})

			g.It("should accept a valid auth key")
			g.It("should reject an invalid auth key")
		})
	})
}
