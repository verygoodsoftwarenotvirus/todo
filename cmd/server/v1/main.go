package main

import (
	"log"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
)

const (
	debug = false

	sqliteConnectionDetails   = database.ConnectionDetails("example.db")
	postgresConnectionDetails = database.ConnectionDetails("postgres://todo:hunter2@database:5432/todo?sslmode=disable")
)

var cookieSecret = []byte("HEREISA32CHARSECRETWHICHISMADEUP")

func main() {
	server, err := BuildServer(
		postgresConnectionDetails,
		users.CookieName("todocookie"),
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
