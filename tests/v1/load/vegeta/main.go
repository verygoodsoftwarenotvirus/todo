package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/go"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/icrowley/fake"
	"github.com/tsenart/vegeta/lib"
)

const (
	debug = true

	localTestInstanceURL   = "http://localhost"
	defaultTestInstanceURL = "http://todo-server"
)

var (
	clientID,
	clientSecret string
	oneDay   = 24 * time.Hour * 1000
	urlToUse = defaultTestInstanceURL
)

func main() {
	lt, err := NewLoadTester()

	if err != nil {
		log.Fatal(err)
	}

	lt.Run()
}

/////////////////////////////////////////

// NewLoadTester provides a new load tester
func NewLoadTester() (*LoadTester, error) {
	httpc := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}

	logger := zerolog.ProvideLogger(zerolog.ProvideZerologger())
	u, _ := url.Parse(urlToUse)

	if debug {
		logger.SetLevel(logging.DebugLevel)
	}

	c, err := client.NewClient(
		"",
		"",
		u,
		logger,
		httpc,
		nil, // tracer,
		debug,
	)

	if err != nil {
		return nil, err
	}

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
		(&ItemTester{}).ItemCreationTargeter(lt.client),
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
		default:

		}
	}
}

// ItemTester tests item routes
type ItemTester struct {
	lock sync.Mutex
}

// ItemCreationTargeter is something
func (it *ItemTester) ItemCreationTargeter(client *client.V1Client) vegeta.Targeter {
	targets := []vegeta.Target{
		{
			Method: http.MethodPost,
			URL:    client.BuildURL(nil, "items"),
			Body:   structToBytes(randomItemInput()),
			// Header: nil,
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

// ItemRetrievalTargeter is something
func (it *ItemTester) ItemRetrievalTargeter(client *client.V1Client) vegeta.Targeter {
	targets := []vegeta.Target{
		{
			Method: http.MethodGet,
			URL:    client.BuildURL(nil, "items"),
			Body:   structToBytes(randomItemInput()),
			// Header: nil,
		},
	}

	i := int64(-1)
	return func(tgt *vegeta.Target) error {
		*tgt = targets[atomic.AddInt64(&i, 1)%int64(len(targets))]
		return nil
	}
}
