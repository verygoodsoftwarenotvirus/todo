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

func TestAuth(test *testing.T) {
	// test.Parallel()

	pc := todoClient.PlainClient()

	test.Run("should reject an unauthenticated request", func(t *testing.T) {
		res, err := pc.Post(todoClient.BuildURL(nil, "fart"), "application/json", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	test.Run("should accept a valid token", func(t *testing.T) {
		cookie := loginUser(t, expectedUsername, expectedPassword)
		assert.NotNil(t, cookie)

		req, err := http.NewRequest(http.MethodPost, todoClient.BuildURL(nil, "fart"), nil)
		req.AddCookie(cookie)
		assert.NoError(t, err)

		ac := todoClient.AuthenticatedClient()
		res, err := ac.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusTeapot, res.StatusCode)
	})

	test.Run("should reject an invalid cookie", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, todoClient.BuildURL(nil, "fart"), nil)
		assert.NoError(t, err)
		res, err := pc.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	// 	test.Run("Creating", func(T *testing.T) {
	// 		T.Run("should be creatable", func(t *testing.T) {
	// 		})
	// 	})
	// 	test.Run("Reading", func(T *testing.T) {
	// 		T.Run("it should return an error when trying to read something that doesn't exist", func(t *testing.T) {
	// 		})
	// 		T.Run("it should be readable", func(t *testing.T) {
	// 		})
	// 	})
	// 	test.Run("Updating", func(T *testing.T) {
	// 		T.Run("it should be updatable", func(t *testing.T) {
	// 		})
	// 	})
	// 	test.Run("Deleting", func(T *testing.T) {
	// 		T.Run("should be able to be deleted", func(t *testing.T) {
	// 		})
	// 	})
	// 	test.Run("Listing", func(T *testing.T) {
	// 		T.Run("should be able to be read in a list", func(t *testing.T) {
	// 		})
	// 	})
	// 	test.Run("Counting", func(T *testing.T) {
	// 		T.Run("it should be able to be counted", func(t *testing.T) {
	// 			t.Skip()
	// 		})
	// 	})
}
