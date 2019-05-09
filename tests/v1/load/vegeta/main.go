package main

import (
	"bytes"
	"encoding/json"
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
	bb := new(bytes.Buffer)
	_, _ = bb.ReadFrom(req.Body)

	ts := todoClient.TokenSource()
	token, err := ts.Token()
	if err != nil {
		return nil, err
	}
	token.SetAuthHeader(req)

	return &vegeta.Target{
		Method: req.Method,
		URL:    req.URL.String(),
		Body:   bb.Bytes(),
		Header: req.Header,
	}, nil
}

func (p *targetProvider) Read(retVal []byte) (int, error) {
	b, _ := json.Marshal(&vegeta.Target{
		Method: "GET",
		URL:    "https://www.google.com",
		Body:   nil,
		Header: nil,
	})

	return bytes.NewReader(append(b, []byte("\n")[0])).Read(retVal)
}

func main() {
	var times uint64 = 50
	duration := 10 * time.Second

	attacker := vegeta.NewAttacker()
	targeter := vegeta.NewJSONTargeter(new(targetProvider), nil, nil)

	metrics := &vegeta.Metrics{}
	for res := range attacker.Attack(targeter, uint64(times/10), duration, "Big Bang!") {

		metrics.Add(res)
	}
	metrics.Close()

	if err := vegeta.NewTextReporter(metrics).Report(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
