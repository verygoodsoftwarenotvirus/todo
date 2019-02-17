package main

import (
	"log"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
)

const (
	postgresConnectionDetails = "postgres://todo:hunter2@database:5432/todo?sslmode=disable"

	cookieSecret = "HEREISA32CHARSECRETWHICHISMADEUP"
)

var (
	debug bool
)

func main() {
	server, err := BuildServer(
		database.ConnectionDetails(postgresConnectionDetails),
		users.CookieName("todo"),
		metrics.Namespace("todo-server"),
		[]byte(cookieSecret),
		debug,
	)

	if err != nil {
		log.Fatal(err)
	} else {
		server.Serve()
	}
}
