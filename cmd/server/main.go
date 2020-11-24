package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config/viper"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/bcrypt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
)

func main() {
	ctx := context.Background()

	// initialize our logger of choice.
	logger := zerolog.NewLogger()

	// find and validate our configuration filepath.
	var configFilepath string
	if configFilepath = os.Getenv("CONFIGURATION_FILEPATH"); configFilepath == "" {
		logger.Fatal(errors.New("no configuration file provided"))
	}

	// parse our config file.
	cfg, err := viper.ParseConfigFile(logger, configFilepath)
	if err != nil || cfg == nil {
		logger.WithValue("config_filepath", configFilepath).Fatal(fmt.Errorf("error parsing configuration file: %w", err))
	}

	if initializeTracerErr := cfg.InitializeTracer(logger); errors.Is(initializeTracerErr, config.ErrInvalidTracingProvider) {
		logger.Fatal(fmt.Errorf("error providing tracer: %w", initializeTracerErr))
	}

	// only allow initialization to take so long.
	ctx, cancel := context.WithTimeout(ctx, cfg.Meta.StartupDeadline)
	ctx, initSpan := tracing.StartSpan(ctx)
	ctx, databaseConnectionSpan := tracing.StartSpan(ctx)

	logger.Debug("connecting to database")

	rawDB, err := cfg.ProvideDatabaseConnection(logger)
	if err != nil {
		logger.Fatal(fmt.Errorf("error connecting to database: %w", err))
	}

	databaseConnectionSpan.End()

	authenticator := bcrypt.ProvideAuthenticator(bcrypt.ProvideHashCost(), logger)
	ctx, databaseClientSetupSpan := tracing.StartSpan(ctx)

	logger.Debug("setting up database client")

	dbClient, err := cfg.ProvideDatabaseClient(ctx, logger, rawDB, authenticator)
	if err != nil {
		logger.Fatal(fmt.Errorf("error initializing database client: %w", err))
	}

	databaseClientSetupSpan.End()

	// build our server struct.
	server, err := BuildServer(ctx, cfg, logger, dbClient, rawDB, authenticator)
	if err != nil {
		logger.Fatal(fmt.Errorf("error initializing HTTP server: %w", err))
	}

	initSpan.End()
	cancel()

	// I slept and dreamt that life was joy.
	//   I awoke and saw that life was service.
	//   	I acted and behold, service deployed.
	server.Serve()
}
