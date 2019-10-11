package main

import (
	"context"
	"errors"
	"log"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/zerolog"
	"go.opencensus.io/trace"
)

func main() {
	// initialize our logger of choice
	logger := zerolog.NewZeroLogger()

	// find and validate our configuration filepath
	configFilepath := os.Getenv("CONFIGURATION_FILEPATH")
	if configFilepath == "" {
		logger.Fatal(errors.New("no configuration file provided"))
	}

	// parse our config file
	cfg, err := config.ParseConfigFile(configFilepath)
	if err != nil || cfg == nil {
		logger.Fatal(err)
	}

	// only allow initialization to take so long
	tctx, cancel := context.WithTimeout(context.Background(), cfg.Meta.StartupDeadline)
	ctx, span := trace.StartSpan(tctx, "initialization")

	// connect to our database
	db, err := cfg.ProvideDatabase(ctx, logger)
	if err != nil {
		logger.Fatal(err)
	}

	// build our server struct
	server, err := BuildServer(ctx, cfg, logger, db)
	span.End()
	cancel()

	if err != nil {
		log.Fatal(err)
	}

	// I slept and dreamt that life was joy.
	//   I awoke and saw that life was service.
	//   	I acted and behold, service deployed.
	server.Serve()
}
