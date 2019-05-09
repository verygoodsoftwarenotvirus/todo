package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/http_client/v1"

	vegeta "github.com/tsenart/vegeta/lib"
)

type (
	targetProvider struct {
		todoClient *client.V1Client
	}
)

var (
	todoClient *client.V1Client
)

func reqToTarget(req *http.Request) (*vegeta.Target, error) {
	var body []byte
	if req.Body != nil {
		bb := new(bytes.Buffer)
		_, _ = bb.ReadFrom(req.Body)
		body = bb.Bytes()
	}

	ts := todoClient.TokenSource()
	token, err := ts.Token()
	if err != nil {
		return nil, err
	}
	token.SetAuthHeader(req)

	return &vegeta.Target{
		Method: req.Method,
		URL:    req.URL.String(),
		Body:   body,
		Header: req.Header,
	}, nil
}

func main() {
	var times uint64 = 50
	duration := 10 * time.Second

	provider := buildJobPoolFromClient(todoClient)

	attacker := vegeta.NewAttacker()
	targeter := vegeta.NewJSONTargeter(provider, nil, nil)

	metrics := &vegeta.Metrics{}
	for res := range attacker.Attack(targeter, uint64(times/10), duration, "Todo Server Testing") {
		metrics.Add(res)
	}
	metrics.Close()

	if err := vegeta.NewTextReporter(metrics).Report(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
