package main

import (
	"context"
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
		panic("no configuration file provided")
	}

	cfg, err := config.ParseConfigFile(configFilepath)
	if err != nil || cfg == nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Meta.StartupDeadline)
	ctx, span := trace.StartSpan(ctx, "initialization")

	db, err := cfg.ProvideDatabase(ctx, logger)
	if err != nil {
		log.Fatal(err)
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
