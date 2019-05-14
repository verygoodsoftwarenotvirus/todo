package load

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/http_client/v1"

	"github.com/emicklei/hazana"
)

// TodoServiceAttacker implements hazana's Attacker interface
type TodoServiceAttacker struct {
	todoClient *client.V1Client
}

// Setup implement's hazana's Attacker interface
func (a *TodoServiceAttacker) Setup(c hazana.Config) error {
	return nil
}

// Do implement's hazana's Attacker interface
func (a *TodoServiceAttacker) Do(ctx context.Context) hazana.DoResult {
	// Do performs one request and is executed in a separate goroutine.
	// The context is used to cancel the request on timeout.
	act := RandomAction(a.todoClient)
	req, err := act.Action()
	if err != nil || req == nil {
		if err == ErrUnavailableYet {
			return hazana.DoResult{
				RequestLabel: act.Name,
				Error:        nil,
				StatusCode:   200,
			}
		}
		log.Printf("something has gone awry: %v\n", err)
		return hazana.DoResult{Error: err}
	}

	var (
		sc     int
		bo, bi int64
	)
	if req.Body != nil {
		bi, _ := ioutil.ReadAll(req.Body)
		rdr := ioutil.NopCloser(bytes.NewBuffer(bi))
		req.Body = rdr
	}

	res, err := a.todoClient.AuthenticatedClient().Do(req)
	if res != nil {
		sc = res.StatusCode
		bo = res.ContentLength
	}

	dr := hazana.DoResult{
		RequestLabel: act.Name,
		Error:        err,
		StatusCode:   sc,
		BytesIn:      bi,
		BytesOut:     bo,
	}
	return dr
}

// Teardown implement's hazana's Attacker interface
func (a *TodoServiceAttacker) Teardown() error {
	// Teardown can be used to close the connection to the service.
	return nil
}

// Clone implement's hazana's Attacker interface
func (a *TodoServiceAttacker) Clone() hazana.Attack {
	// Clone should return a fresh new Attack
	// Make sure the new Attack has values for shared struct fields initialized at Setup.
	return a
}

func main() {
	todoClient := initializeClient(clientID, clientSecret)

	attacker := &TodoServiceAttacker{todoClient: todoClient}
	cfg := hazana.Config{
		RPS:           50,
		AttackTimeSec: 1<<32 - 1, // run basically forever
		RampupTimeSec: 5,
		// RampupStrategy: "", string            `json:"rampupStrategy"`
		MaxAttackers: 50,
		// OutputFilename: "", string            `json:"outputFilename,omitempty"`
		Verbose: true,
		// Metadata: "",       map[string]string `json:"metadata,omitempty"`
		DoTimeoutSec: 10,
	}

	r := hazana.Run(attacker, cfg)

	// inspect the report and compute whether the test has failed
	// e.g by looking at the success percentage and mean response time of each metric.
	r.Failed = false

	hazana.PrintReport(r)
}
