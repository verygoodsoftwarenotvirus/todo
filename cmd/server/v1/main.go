package main

import (
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
)

func main() {
	dbCfg := database.Config{
		Debug:            !secure,
		ConnectionString: dbFile,
		SchemaDir:        schemaDir,
	}

	cfg := server.ServerConfig{
		DebugMode: !secure,
		CertFile:  certFile,
		KeyFile:   keyFile,
		DBBuilder: sqlite.NewSqlite,
	}

	if server, err := server.NewDebug(cfg, dbCfg); err != nil {
		panic(err)
	} else {
		server.Serve()
	}
}
