package main

import (
	"log"

	"gitlab.com/verygoodsoftwarenotvirus/todo/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
)

func main() {
	logger := zerolog.ProvideLogger()

	cfg, err := config.ParseConfigFile("production.toml")
	if err != nil || cfg == nil {
		log.Fatal(err)
	}

	db, err := cfg.ProvideDatabase(logger)
	if err != nil || cfg == nil {
		log.Fatal(err)
	}

	server, err := BuildServer(cfg, logger, db)

	if err != nil {
		log.Fatal(err)
	} else {
		server.Serve()
	}
}
