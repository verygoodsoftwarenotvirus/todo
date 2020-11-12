package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config/viper"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
)

func main() {
	// initialize our logger of choice.
	logger := zerolog.NewLogger()

	// find and validate our configuration filepath.
	configFilepath := os.Getenv("CONFIGURATION_FILEPATH")
	if configFilepath == "" {
		logger.Fatal(errors.New("no configuration file provided"))
	}

	// parse our config file.
	cfg, err := viper.ParseConfigFile(logger, configFilepath)
	if err != nil || cfg == nil {
		logger.WithValue("config_filepath", configFilepath).Fatal(fmt.Errorf("error parsing configuration file: %w", err))
	}

	// only allow initialization to take so long.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Meta.StartupDeadline)
	ctx, span := tracing.StartSpan(ctx, "initialization")

	logger.Debug("connecting to database")

	rawDB, err := cfg.ProvideDatabaseConnection(logger)
	if err != nil {
		logger.Fatal(fmt.Errorf("error connecting to database: %w", err))
	}

	logger.Debug("setting up database client")
	authenticator := auth.ProvideBcryptAuthenticator(auth.ProvideBcryptHashCost(), logger)

	dbClient, err := cfg.ProvideDatabaseClient(ctx, logger, rawDB, authenticator)
	if err != nil {
		logger.Fatal(fmt.Errorf("error initializing database client: %w", err))
	}

	// build our server struct.
	logger.Debug("building server")
	server, err := BuildServer(ctx, cfg, logger, dbClient, rawDB, authenticator)

	span.End()
	cancel()

	if err != nil {
		logger.Fatal(fmt.Errorf("error initializing HTTP server: %w", err))
	}

	// I slept and dreamt that life was joy.
	//   I awoke and saw that life was service.
	//   	I acted and behold, service deployed.
	server.Serve()
}
