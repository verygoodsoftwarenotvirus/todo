package main

import (
	"log"
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
)

const (
	secure = false

	sqliteSchemaDir         = "database/v1/sqlite/schema"
	sqliteConnectionDetails = "example.db"

	postgresSchemaDir         = "database/v1/postgres/schema"
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

func init() {
	debug = strings.ToLower(os.Getenv("DOCKER")) == "true"
	if debug {
		log.Println("running in a docker environment")
		certToUse, keyToUse = certFile, keyFile
	} else {
		certToUse, keyToUse = localCertFile, localKeyFile
	}
	log.Printf("debug: %v\n", debug)
	log.Printf("using this cert: %q\n", certToUse)
	log.Printf("using this key: %q\n", keyToUse)
}

func main() {
	server, err := BuildServer(
		database.ConnectionDetails(sqliteConnectionDetails),
		sqliteSchemaDir,
		server.CertPair{
			CertFile: certToUse,
			KeyFile:  keyToUse,
		},
		users.CookieName("todo"),
		[]byte(cookieSecret),
		debug,
	)

	if err != nil {
		log.Fatal(err)
	} else {
		server.Serve()
	}
}
