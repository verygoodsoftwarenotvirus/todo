package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config/viper"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password/bcrypt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/logging/zerolog"
)

const (
	useNoOpLoggerEnvVar  = "USE_NOOP_LOGGER"
	configFilepathEnvVar = "CONFIGURATION_FILEPATH"
)

func main() {
	var (
		ctx            = context.Background()
		logger         = zerolog.NewLogger()
		configFilepath string
	)

	if x, err := strconv.ParseBool(os.Getenv(useNoOpLoggerEnvVar)); x && err == nil {
		logger = logging.NewNonOperationalLogger()
	}

	// find and validate our configuration filepath.
	if configFilepath = os.Getenv(configFilepathEnvVar); configFilepath == "" {
		logger.Fatal(errors.New("no configuration file provided"))
	}

	// parse our config file.
	cfg, err := viper.ParseConfigFile(ctx, logger, configFilepath)
	if err != nil || cfg == nil {
		logger.WithValue("config_filepath", configFilepath).Fatal(fmt.Errorf("parsing configuration file: %w", err))
	}

	if initializeTracerErr := cfg.Observability.Tracing.Initialize(logger); initializeTracerErr != nil {
		logger.Error(initializeTracerErr, "initializing tracer")
	}

	// only allow initialization to take so long.
	ctx, cancel := context.WithTimeout(ctx, cfg.Server.StartupDeadline)
	ctx, initSpan := tracing.StartSpan(ctx)

	logger.Debug("connecting to database")

	// connect to database
	ctx, databaseConnectionSpan := tracing.StartSpan(ctx)

	rawDB, err := cfg.Database.ProvideDatabaseConnection(logger)
	if err != nil {
		logger.Fatal(fmt.Errorf("connecting to database: %w", err))
	}

	databaseConnectionSpan.End()
	logger.Debug("setting up database client")

	// setup DB client
	ctx, databaseClientSetupSpan := tracing.StartSpan(ctx)
	authenticator := bcrypt.ProvideAuthenticator(bcrypt.ProvideHashCost(), logger)

	dbClient, err := cfg.ProvideDatabaseClient(ctx, logger, rawDB)
	if err != nil {
		logger.Fatal(fmt.Errorf("initializing database client: %w", err))
	}

	databaseClientSetupSpan.End()

	// build our server struct.
	server, err := BuildServer(ctx, cfg, logger, dbClient, rawDB, authenticator)
	if err != nil {
		logger.Fatal(fmt.Errorf("initializing HTTP server: %w", err))
	}

	initSpan.End()
	cancel()

	// I slept and dreamt that life was joy.
	//   I awoke and saw that life was service.
	//   	I acted and behold, service deployed.
	server.Serve()
}
