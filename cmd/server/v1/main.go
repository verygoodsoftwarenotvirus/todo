package main

import (
	"log"
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
)

const (
	secure        = false
	dbFile        = "example.db"
	schemaDir     = "database/v1/sqlite/schema"
	certFile      = "certs/cert.pem"
	keyFile       = "certs/key.pem"
	localCertFile = "dev_files/certs/server/cert.pem"
	localKeyFile  = "dev_files/certs/server/key.pem"
	cookieSecret  = "HEREISA32CHARSECRETWHICHISMADEUP"
)

func main() {
	debug := strings.ToLower(os.Getenv("DEBUG")) == "true"
	log.Printf("debug: %v\n", debug)
	dbCfg := database.Config{
		Debug:            debug,
		ConnectionString: dbFile,
		SchemaDir:        schemaDir,
	}

	cfg := server.ServerConfig{
		DebugMode:    debug,
		CookieSecret: []byte(cookieSecret),
		CertFile:     localCertFile,
		KeyFile:      localKeyFile,
		DBBuilder:    sqlite.NewSqlite,
	}

	if server, err := server.NewDebug(cfg, dbCfg); err != nil {
		panic(err)
	} else {
		server.Serve()
	}
}
