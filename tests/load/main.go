package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	ctx := context.Background()
	rate := vegeta.Rate{Freq: 100, Per: time.Second}
	runTime := 10 * time.Minute

	if runTimeEnvVar := os.Getenv("LOADTEST_RUN_TIME"); runTimeEnvVar != "" {
		rt, err := time.ParseDuration(runTimeEnvVar)
		if err != nil {
			log.Fatal(err)
		}

		runTime = rt
	}

	var (
		metricsHat         sync.Mutex
		metrics            vegeta.Metrics
		attackersWaitGroup sync.WaitGroup
	)

	for _, pair := range buildAttackTargetPairs(ctx) {
		p := pair

		attackersWaitGroup.Add(1)

		go func() {
			for res := range p.attacker.Attack(p.helper.Targeter(), rate, runTime, fmt.Sprintf("%s load test", p.name)) {
				metricsHat.Lock()
				p.helper.HandleResult(res)
				metrics.Add(res)
				metricsHat.Unlock()
			}

			attackersWaitGroup.Done()
		}()
	}

	attackersWaitGroup.Wait()
	metrics.Close()

	log.Printf("99th percentile: %s\n", metrics.Latencies.P99)
}
