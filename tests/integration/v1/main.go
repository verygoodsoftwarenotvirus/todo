package integration

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"
)

const (
	debug                  = false
	nonexistentID          = "999999999"
	localTestInstanceURL   = "https://localhost"
	defaultTestInstanceURL = "https://demo-server"

	defaultSecret                   = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultTestInstanceClientID     = "HEREISACLIENTIDWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultTestInstanceClientSecret = defaultSecret
)

var (
	urlToUse   string
	todoClient *client.V1Client
)

func checkValueAndError(t *testing.T, i interface{}, err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, i)
}

func initializeClient() {
	cfg := &client.Config{
		Client: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   5 * time.Second,
		},
		Debug:   debug,
		Address: urlToUse,

		ClientID:     defaultTestInstanceClientID,
		ClientSecret: defaultTestInstanceClientSecret,
		RedirectURI:  defaultTestInstanceURL,
	}

	// WARNING: Never do this ordinarily, this is an application which will only ever run in a local context
	cfg.Client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	c, err := client.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	todoClient = c
}

func ensureServerIsUp() {
	var (
		isDown           = true
		maxAttempts      = 25
		numberOfAttempts = 0
	)

	for isDown {
		if !todoClient.IsUp() {
			log.Printf("waiting half a second before pinging again")
			time.Sleep(500 * time.Millisecond)
			numberOfAttempts++
			if numberOfAttempts >= maxAttempts {
				log.Fatalf("Maximum number of attempts made, something's gone awry")
			}
		} else {
			isDown = false
		}
	}

}

func init() {
	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		urlToUse = defaultTestInstanceURL
	} else {
		urlToUse = localTestInstanceURL
	}

	initializeClient()
	ensureServerIsUp()
	//testOAuth()
}
