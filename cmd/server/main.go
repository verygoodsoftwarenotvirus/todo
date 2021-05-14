package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/build/server"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/viper"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"

	chimiddleware "github.com/go-chi/chi/middleware"
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

	logger.SetLevel(logging.DebugLevel)

	logger.SetRequestIDFunc(func(req *http.Request) string {
		return chimiddleware.GetReqID(req.Context())
	})

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

	flushFunc, initializeTracerErr := cfg.Observability.Tracing.Initialize(logger)
	if initializeTracerErr != nil {
		logger.Error(initializeTracerErr, "initializing tracer")
	}

	// it's possible that tracing is disabled
	if flushFunc != nil {
		defer flushFunc()
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
	authenticator := passwords.ProvideArgon2Authenticator(logger)

	dbClient, err := cfg.ProvideDatabaseClient(ctx, logger, rawDB)
	if err != nil {
		logger.Fatal(fmt.Errorf("initializing database client: %w", err))
	}

	databaseClientSetupSpan.End()

	// build our server struct.
	srv, err := server.Build(ctx, cfg, logger, dbClient, rawDB, authenticator)
	if err != nil {
		logger.Fatal(fmt.Errorf("initializing HTTP server: %w", err))
	}

	initSpan.End()
	cancel()

	// I slept and dreamt that life was joy.
	//   I awoke and saw that life was service.
	//   	I acted and behold, service deployed.
	srv.Serve()
}
