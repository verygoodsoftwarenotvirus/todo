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
	dbCfg := database.Config{
		Debug:            true,
		ConnectionString: dbFile,
		SchemaDir:        schemaDir,
	}

	cfg := server.ServerConfig{
		DebugMode:    true,
		CookieSecret: []byte(cookieSecret),
		CertFile:     certToUse,
		KeyFile:      keyToUse,
		DBBuilder:    sqlite.NewSqlite,
	}

	if server, err := server.NewDebug(cfg, dbCfg); err != nil {
		panic(err)
	} else {
		server.Serve()
	}
}
