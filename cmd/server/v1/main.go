package main

import (
	"log"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
)

func main() {
	logger := zerolog.ProvideLogger()

	configFilepath := os.Getenv("CONFIGURATION_FILEPATH")
	if configFilepath == "" {
		panic("no configuration file provided")
	}

	cfg, err := config.ParseConfigFile(configFilepath)
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
