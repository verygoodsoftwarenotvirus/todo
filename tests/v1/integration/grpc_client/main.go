package grpcclient

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/grpc_client/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/proto/v1"

	"github.com/icrowley/fake"
)

const (
	targetGRPCPort = "8888"
)

var (
	localTestInstanceURL   = fmt.Sprintf("localhost:%s", targetGRPCPort)
	defaultTestInstanceURL = fmt.Sprintf("todo-server:%s", targetGRPCPort)

	livenessCheckURI,
	testUserTwoFactorSecret,
	urlToUse string
	todoClient todoproto.TodoClient
)

func init() {
	fake.Seed(time.Now().UnixNano())

	var err error
	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		urlToUse = defaultTestInstanceURL
	} else {
		urlToUse = localTestInstanceURL
	}
	livenessCheckURI = strings.ReplaceAll(urlToUse, fmt.Sprintf(":%s", targetGRPCPort), "")

	ensureServerIsUp()

	todoClient, err = grpcclient.NewAuthorizedClient(urlToUse, "clientID", "clientSecret")

	ctx := context.Background()
	user, err := todoClient.CreateUser(ctx, &todoproto.CreateUserRequest{
		Username: fake.UserName(),
		Password: fake.Password(64, 64, true, true, true),
	})
	if err != nil {
		log.Fatal(err)
	}
	testUserTwoFactorSecret = user.GetTwoFactorSecret()

	todoClient.CreateOAuth2Client(ctx, nil)
}

// mostly duplicated code from the http_client

func ensureServerIsUp() {
	var (
		isDown           = true
		maxAttempts      = 25
		numberOfAttempts = 0
	)

	for isDown {
		if !isUp(livenessCheckURI) {
			log.Printf("waiting before pinging %q again\n", livenessCheckURI)
			time.Sleep(500 * time.Millisecond)
			numberOfAttempts++
			if numberOfAttempts >= maxAttempts {
				log.Fatal("Maximum number of attempts made, something's gone awry")
			}
		} else {
			isDown = false
		}
	}
}

func isUp(address string) bool {
	uri := fmt.Sprintf("http://%s/_meta_/health", address)
	req, _ := http.NewRequest(http.MethodGet, uri, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}

	return res.StatusCode == http.StatusOK
}
