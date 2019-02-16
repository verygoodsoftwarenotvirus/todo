package main

import (
	"log"
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
)

const (
	secure = false

	sqliteConnectionDetails   = "example.db"
	postgresConnectionDetails = "postgres://todo:hunter2@database:5432/todo?sslmode=disable"

	certFile = "certs/cert.pem"
	keyFile  = "certs/key.pem"

	localCertFile = "dev_files/certs/server/cert.pem"
	localKeyFile  = "dev_files/certs/server/key.pem"

	cookieSecret = "HEREISA32CHARSECRETWHICHISMADEUP"
)

var (
	certToUse, keyToUse string
	debug               bool
)

func main() {
	debug = strings.ToLower(os.Getenv("DOCKER")) == "true"
	if debug {
		certToUse, keyToUse = certFile, keyFile
	} else {
		certToUse, keyToUse = localCertFile, localKeyFile
	}

	server, err := BuildServer(
		database.ConnectionDetails(postgresConnectionDetails),
		server.CertPair{
			CertFile: certToUse,
			KeyFile:  keyToUse,
		},
		users.CookieName("todo"),
		metrics.Namespace("todo-server"),
		[]byte(cookieSecret),
		debug,
	)

	log.Printf(`
	debug: %v
	using this cert: %q
	using this key: %q
	`, debug, certToUse, keyToUse)

	if err != nil {
		log.Fatal(err)
	} else {
		server.Serve()
	}
}
