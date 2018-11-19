package main

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server"
)

const (
	secure    = false
	dbFile    = "example.db"
	schemaDir = "database/sqlite/schema"
	certFile  = "certs/cert.pem"
	keyFile   = "certs/key.pem"
)

func main() {
	dbCfg := database.Config{
		Debug:            true,
		ConnectionString: dbFile,
	}

	cfg := server.ServerConfig{
		DebugMode: !secure,
		CertFile:  certFile,
		KeyFile:   keyFile,
		SchemaDir: schemaDir,

		DBConfig:  dbCfg,
		DBBuilder: sqlite.NewSqlite,
	}

	server, err := server.NewDebug(cfg)
	if err != nil {
		panic(err)
	}

	server.Serve()
}
