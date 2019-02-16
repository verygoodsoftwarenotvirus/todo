package loadtest

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/go"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/icrowley/fake"
	"github.com/pkg/errors"
	"github.com/tsenart/vegeta/lib"
)

const (
	debug = true

	nonexistentID          = 999999999
	localTestInstanceURL   = "https://localhost"
	defaultTestInstanceURL = "https://todo-server"

	// dockerPostgresAddress = "postgres://todo:hunter2@database:5432/todo?sslmode=disable"
	// localPostgresAddress  = "postgres://todo:hunter2@localhost:2345/todo?sslmode=disable"

	defaultSecret                   = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultTestInstanceClientID     = "HEREISACLIENTIDWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultTestInstanceClientSecret = defaultSecret
)

var (
	oneDay             = 24 * time.Hour
	activeAddress      string
	activeClientID     string
	activeClientSecret string
)

func init() {
	activeAddress = defaultTestInstanceURL
	activeClientID = defaultTestInstanceClientID
	activeClientSecret = defaultTestInstanceClientSecret

	if strings.ToLower(os.Getenv("DOCKER")) == "True" {
		activeAddress = localTestInstanceURL
	}
}

func ensureServerIsUp(todoClient *client.V1Client, logger logging.Logger) {
	var (
		isDown           = true
		maxAttempts      = 25
		numberOfAttempts = 0
		napTime          = 500 * time.Millisecond
	)

	for isDown {
		if !todoClient.IsUp() {
			logger.WithValue("waiting", napTime.String()).Info("waiting before pinging again")
			time.Sleep(napTime)

			numberOfAttempts++
			napTime += napTime

			if numberOfAttempts >= maxAttempts {
				logger.Fatal(errors.New("Maximum number of attempts made, something's gone awry"))
			}
		} else {
			isDown = false
		}
	}
}

// NewLoadTester provides a new load tester
func NewLoadTester() (*LoadTester, error) {
	httpc := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}

	// WARNING: Never do this ordinarily, this is an application which will only ever run in a local context
	// NOTE: I recognize the irony that this is copy/pasted in a couple of places, but bare with me
	httpc.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	logger := zerolog.ProvideLogger(zerolog.ProvideZerologger())
	u, _ := url.Parse(activeAddress)

	if debug {
		logger.SetLevel(logging.DebugLevel)
	}

	c, err := client.NewClient(
		activeClientID,
		activeClientSecret,
		u,
		logger,
		httpc,
		nil, // tracer,
		debug,
	)

	if err != nil {
		return nil, err
	}

	ensureServerIsUp(c, logger)

	lt := &LoadTester{
		client: c,
		logger: logger,
		attacker: vegeta.NewAttacker(
			vegeta.Client(c.AuthenticatedClient()),
		),
	}

	return lt, nil
}

// LoadTester is a load tester
type LoadTester struct {
	logger   logging.Logger
	client   *client.V1Client
	attacker *vegeta.Attacker
}

// Run runs the load test
func (lt *LoadTester) Run() {
	results := lt.attacker.Attack(
		(&ItemTester{}).ItemCreationAndDeletionTargeter(lt.client),
		vegeta.Rate{
			Freq: 50,
			Per:  time.Second,
		},
		oneDay,
		"load testing",
	)

	for {
		select {
		case <-results:
			// case r := <-results:
			// lt.logger.WithValues(map[string]interface{}{
			// 	"Attack":    r.Attack,
			// 	"Seq":       r.Seq,
			// 	"Code":      r.Code,
			// 	"Timestamp": r.Timestamp,
			// 	"Latency":   r.Latency,
			// 	"BytesOut":  r.BytesOut,
			// 	"BytesIn":   r.BytesIn,
			// 	"Error":     r.Error,
			// 	"Body":      r.Body,
			// }).Info("result acquired!")
		}
	}
}

// ItemTester tests item routes
type ItemTester struct {
	lock sync.Mutex
}

// ItemCreationAndDeletionTargeter is something we can provide to a Vegeta attacker to generate requests in the right order
func (it *ItemTester) ItemCreationAndDeletionTargeter(client *client.V1Client) vegeta.Targeter {

	targets := []vegeta.Target{
		{
			Method: http.MethodPost,
			URL:    client.BuildURL(nil, "items"),
			Body:   structToBytes(randomItemInput()),
			Header: nil,
		},
	}

	i := int64(-1)
	return func(tgt *vegeta.Target) error {
		*tgt = targets[atomic.AddInt64(&i, 1)%int64(len(targets))]
		return nil
	}
}

func structToBytes(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func randomItemInput() *models.ItemInput {
	ii := &models.ItemInput{
		Name:    fake.Word(),
		Details: fake.Sentence(),
	}
	return ii
}
