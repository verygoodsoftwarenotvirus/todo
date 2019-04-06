package grpcclient

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/proto/v1"

	"google.golang.org/grpc"
)

const (
	targetGRPCPort = "41214"
)

var (
	localTestInstanceURL   = fmt.Sprintf("localhost:%s", targetGRPCPort)
	defaultTestInstanceURL = fmt.Sprintf("todo-server:%s", targetGRPCPort)

	livenessCheckURI,
	urlToUse string
	grpcConn   *grpc.ClientConn
	todoClient todoproto.TodoClient
)

func init() {
	var err error

	if strings.ToLower(os.Getenv("DOCKER")) == "true" {
		urlToUse = defaultTestInstanceURL
	} else {
		urlToUse = localTestInstanceURL
	}
	livenessCheckURI = strings.ReplaceAll(urlToUse, fmt.Sprintf(":%s", targetGRPCPort), "")

	ensureServerIsUp()

	grpcConn, err = grpc.Dial(urlToUse, grpc.WithInsecure())
	if err != nil || grpcConn == nil {
		log.Fatalf("Failed to start gRPC connection: %v", err)
	}

	todoClient = todoproto.NewTodoClient(grpcConn)
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
