package main

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestRunMain(_ *testing.T) {
	// This test is built specifically to capture the coverage that the integration
	// tests exhibit. We run the main function (i.e. a production server)
	// on an independent goroutine and sleep for long enough that the integration
	// tests can run, then we quit.
	d, err := time.ParseDuration(os.Getenv("RUNTIME_DURATION"))
	if err != nil {
		log.Fatal(err)
	}

	go main()

	time.Sleep(d)
}
