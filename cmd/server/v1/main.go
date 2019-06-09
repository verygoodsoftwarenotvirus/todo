package main

import (
	"context"
	"log"
	"os"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ctx, span := trace.StartSpan(ctx, "initialization")
	defer span.End()

	db, err := cfg.ProvideDatabase(ctx, logger)
	if err != nil {
		log.Fatal(err)
	}

	server, err := BuildServer(ctx, cfg, logger, db)
	if err != nil {
		log.Fatal(err)
	} else {
		server.Serve()
	}
}
