package main

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestRunMain(t *testing.T) {
	d, err := time.ParseDuration(os.Getenv("RUNTIME_DURATION"))
	if err != nil {
		log.Fatal(err)
	}

	go main()

	time.Sleep(d)
}
