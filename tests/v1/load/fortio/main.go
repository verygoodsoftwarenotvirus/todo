package main

import (
	"fmt"
	"time"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/periodic"
)

const (
	debug = true

	localTestInstanceURL   = "http://localhost"
	defaultTestInstanceURL = "http://todo-server"
)

func main() {
	// productpage should still return 200s when ratings is rate-limited.
	res, err := fhttp.RunHTTPTest(&fhttp.HTTPRunnerOptions{
		RunnerOptions: periodic.RunnerOptions{
			QPS:        10,
			Duration:   1 * time.Minute,
			NumThreads: 8,
		},
		HTTPOptions: fhttp.HTTPOptions{
			URL: fmt.Sprintf("%s/_meta_/health", localTestInstanceURL),
		},
	})

	if res == nil || err != nil {
		panic("bah")
	}
}
