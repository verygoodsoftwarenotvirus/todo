package main

import (
	"context"
	"errors"
	"log"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/zerolog"

	"go.opencensus.io/trace"
)

func main() {
	logger := zerolog.NewZeroLogger()

	configFilepath := os.Getenv("CONFIGURATION_FILEPATH")
	if configFilepath == "" {
		logger.Fatal(errors.New("no configuration file provided"))
	}

	cfg, err := config.ParseConfigFile(configFilepath)
	if err != nil || cfg == nil {
		logger.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Meta.StartupDeadline)
	ctx, span := trace.StartSpan(ctx, "initialization")

	db, err := cfg.ProvideDatabase(ctx, logger)
	if err != nil {
		logger.Fatal(err)
	}

	server, err := BuildServer(ctx, cfg, logger, db)
	cancel()
	span.End()

	if err != nil {
		log.Fatal(err)
	} else {
		server.Serve()
	}
}
