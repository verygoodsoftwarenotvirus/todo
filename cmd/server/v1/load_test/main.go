package main

import (
	"log"

	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/load"
)

func main() {
	if lt, err := loadtest.NewLoadTester(); err == nil {
		lt.Run()
	} else {
		log.Fatal(err)
	}
}
