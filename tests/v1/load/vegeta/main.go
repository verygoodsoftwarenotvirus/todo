package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

type targetProvider struct {
	bytesRead uint64
}

func (p *targetProvider) Read(retVal []byte) (int, error) {
	t := &vegeta.Target{
		Method: "GET",
		URL:    "https://www.google.com",
		Body:   nil,
		Header: nil,
	}

	b, err := json.Marshal(t)
	if err != nil {
		return -1, err
	}
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
