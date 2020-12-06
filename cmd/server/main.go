package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config/viper"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/bcrypt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
)

const (
	useNoOpLoggerEnvVar  = "USE_NOOP_LOGGER"
	configFilepathEnvVar = "CONFIGURATION_FILEPATH"
)

func main() {
	var (
		logger         logging.Logger = zerolog.NewLogger()
		configFilepath string
	)

	if x, err := strconv.ParseBool(os.Getenv(useNoOpLoggerEnvVar)); x {
		_ = err
		logger = noop.NewLogger()
	}

	// find and validate our configuration filepath.
	if configFilepath = os.Getenv(configFilepathEnvVar); configFilepath == "" {
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
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Meta.StartupDeadline)
	ctx, initSpan := tracing.StartSpan(ctx)

	logger.Debug("connecting to database")

	// connect to database
	ctx, databaseConnectionSpan := tracing.StartSpan(ctx)

	rawDB, err := cfg.ProvideDatabaseConnection(logger)
	if err != nil {
		logger.Fatal(fmt.Errorf("error connecting to database: %w", err))
	}

	databaseConnectionSpan.End()
	logger.Debug("setting up database client")

	// setup DB client
	ctx, databaseClientSetupSpan := tracing.StartSpan(ctx)
	authenticator := bcrypt.ProvideAuthenticator(bcrypt.ProvideHashCost(), logger)

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
