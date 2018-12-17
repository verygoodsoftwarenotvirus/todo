// https://github.com/golang/oauth2/blob/d668ce993890a79bda886613ee587a69dd5da7a6/example_test.go
package integration

import (
	// "context"
	// "fmt"
	// "log"
	// "net/http"
	"testing"
	// "time"

	// "golang.org/x/oauth2"
	"github.com/stretchr/testify/assert"
)

const (
	defaultSecret = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
)

func TestSomething(test *testing.T) {
	// test.Parallel()

	test.Run("buh", func(t *testing.T) {
		// Create user
		x, y, c := buildDummyUser(test)
		assert.NotNil(test, c)

		input := buildDummyOauth2ClientInput(t, x.Username, y.Password, x.TwoFactorSecret)
		premade, err := todoClient.CreateOauth2Client(input)
		checkValueAndError(t, premade, err)

		// Fetch oauth2Client
		actual, err := todoClient.GetOauth2Client(premade.ClientID)
		assert.NoError(t, err)

		// Assert oauth2Client equality
		checkOauth2ClientEquality(t, input, actual)

		// Clean up
		assert.NoError(t, todoClient.DeleteUser(actual.ID))
		assert.NoError(t, todoClient.DeleteOauth2Client(actual.ID))
	})

	// test.Run("fuck", func(t *testing.T) {
	// 	ctx := context.Background()

	// 	bp := fmt.Sprintf("%s://%s", todoClient.URL.Scheme, todoClient.URL.Hostname())
	// 	turl, aurl := fmt.Sprintf("%s/oauth2/token", bp), fmt.Sprintf("%s/oauth2/authorize", bp)

	// 	conf := &oauth2.Config{
	// 		ClientID:     defaultSecret,
	// 		ClientSecret: defaultSecret,
	// 		Scopes:       []string{"*"},
	// 		Endpoint:     oauth2.Endpoint{TokenURL: turl, AuthURL: aurl},
	// 	}

	// 	// Use the custom HTTP client when requesting a token.
	// 	httpClient := &http.Client{Timeout: 2 * time.Second}
	// 	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	// 	tok, err := conf.Exchange(ctx, defaultSecret)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	client := conf.Client(ctx, tok)
	// 	_ = client
	// })
}
