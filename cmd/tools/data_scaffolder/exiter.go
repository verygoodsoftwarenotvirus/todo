package main

import (
	"log"
	"os"
)

type Quitter interface {
	Quit(code int)
	ComplainAndQuit(...interface{})
	ComplainAndQuitf(string, ...interface{})
}

// fatal quitter

var _ Quitter = fatalQuitter{}

type fatalQuitter struct{}

func (e fatalQuitter) Quit(code int) {
	os.Exit(code)
}

func (e fatalQuitter) ComplainAndQuit(v ...interface{}) {
	log.Fatal(v...)
}

func (e fatalQuitter) ComplainAndQuitf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
