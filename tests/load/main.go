package main

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"

	"github.com/emicklei/hazana"
)

// ServiceAttacker implements hazana's Attacker interface.
type ServiceAttacker struct {
	todoClient *httpclient.V1Client
}

// Setup implements hazana's Attacker interface.
func (a *ServiceAttacker) Setup(_ hazana.Config) error {
	return nil
}

// Do implements hazana's Attacker interface.
func (a *ServiceAttacker) Do(_ context.Context) hazana.DoResult {
	// Do performs one request and is executed in a separate goroutine.
	// The context is used to cancel the request on timeout.
	act := RandomAction(a.todoClient)

	req, err := act.Action()
	if err != nil || req == nil {
		if errors.Is(err, ErrUnavailableYet) {
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
		sc int
		bo int64
		bi []byte
	)

	if req.Body != nil {
		bi, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return hazana.DoResult{Error: err}
		}

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
		BytesIn:      int64(len(bi)),
		BytesOut:     bo,
	}

	return dr
}

// Teardown implements hazana's Attacker interface.
func (a *ServiceAttacker) Teardown() error {
	return nil
}

// Clone implements hazana's Attacker interface.
func (a *ServiceAttacker) Clone() hazana.Attack {
	return a
}

func main() {
	runTime := 10 * time.Minute
	todoClient := initializeClient(oa2Client)

	if rt := os.Getenv("LOADTEST_RUN_TIME"); rt != "" {
		_rt, err := time.ParseDuration(rt)
		if err != nil {
			panic(err)
		}

		runTime = _rt
	}

	attacker := &ServiceAttacker{todoClient: todoClient}
	cfg := hazana.Config{
		RPS:           50,
		AttackTimeSec: int(runTime.Seconds()),
		RampupTimeSec: 5,
		MaxAttackers:  50,
		Verbose:       true,
		DoTimeoutSec:  10,
	}

	r := hazana.Run(attacker, cfg)

	// inspect the report and compute whether the test has failed
	// e.g by looking at the success percentage and mean response time of each metric.
	r.Failed = false

	hazana.PrintReport(r)
}
