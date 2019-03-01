package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/periodic"
)

const (
	debug = true

	localTestInstanceURL   = "http://localhost"
	defaultTestInstanceURL = "http://todo-server"
)

var urlToUse = defaultTestInstanceURL

func init() {
	ensureServerIsUp()
}

func ensureServerIsUp() {
	var (
		isDown           = true
		maxAttempts      = 25
		numberOfAttempts = 0
	)

	for isDown {
		if !isUp() {
			log.Println("waiting half a second before pinging again")
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

func isUp() bool {
	uri := fmt.Sprintf("%s/_meta_/health", urlToUse)

	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	res, err := (*http.Client)(&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return false
	}

	return res.StatusCode == http.StatusOK
}

func main() {
	// productpage should still return 200s when ratings is rate-limited.
	_, err := fhttp.RunHTTPTest(&fhttp.HTTPRunnerOptions{
		RunnerOptions: periodic.RunnerOptions{
			QPS:        10,
			Duration:   1 * time.Minute,
			NumThreads: 8,
		},
		HTTPOptions: fhttp.HTTPOptions{
			URL: fmt.Sprintf("%s/_meta_/health", urlToUse),
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
