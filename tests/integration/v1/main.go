package integration

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1"

	"github.com/bxcodec/faker"
)

const (
	debug                        = false
	nonexistentID                = 999999999
	localTestInstanceURL         = "https://localhost"
	defaultTestInstanceURL       = "https://demo-server"
	defaultTestInstanceAuthToken = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
)

var (
	todoClient *client.V1Client
	fake       = faker.GetLorem()
)

func initializeClient() {
	cfg := &client.Config{
		Client: &http.Client{
			Transport: http.DefaultTransport,
			//Timeout:   5 * time.Second,
		},
		Debug:     debug,
		Address:   defaultTestInstanceURL,
		AuthToken: defaultTestInstanceAuthToken,
	}
	cfg.Client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
		// WARNING: Never this ordinarily, this is an application which will only ever run in a local context
		InsecureSkipVerify: true,
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	todoClient = c
}

func ensureServerIsUp() {
	maxAttempts := 25
	isDown := true
	numberOfAttempts := 0

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
	initializeClient()
	ensureServerIsUp()
}
