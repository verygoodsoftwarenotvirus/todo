package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/build/server"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/secrets"

	chimiddleware "github.com/go-chi/chi/middleware"
	flag "github.com/spf13/pflag"
)

const (
	useNoOpLoggerEnvVar  = "USE_NOOP_LOGGER"
	configFilepathEnvVar = "CONFIGURATION_FILEPATH"
)

var (
	configFilepath string
	errNoConfig    = errors.New("no configuration file provided")
)

func init() {
	flag.StringVarP(&configFilepath, "config", "c", "", "the config filepath")
}

func initializeLocalSecretManager(ctx context.Context) secrets.SecretManager {
	logger := logging.NewNoopLogger()

	cfg := &secrets.Config{
		Provider: secrets.ProviderLocal,
		Key:      os.Getenv("TODO_SERVICE_LOCAL_CONFIG_STORE_KEY"),
	}

	k, err := secrets.ProvideSecretKeeper(ctx, cfg)
	if err != nil {
		panic(err)
	}

	sm, err := secrets.ProvideSecretManager(logger, k)
	if err != nil {
		panic(err)
	}

	return sm
}

func main() {
	flag.Parse()

	var (
		ctx    = context.Background()
		logger = logging.ProvideLogger(logging.Config{Provider: logging.ProviderZerolog})
	)

	logger.SetLevel(logging.DebugLevel)

	logger.SetRequestIDFunc(func(req *http.Request) string {
		return chimiddleware.GetReqID(req.Context())
	})

	if x, err := strconv.ParseBool(os.Getenv(useNoOpLoggerEnvVar)); x && err == nil {
		logger = logging.NewNoopLogger()
	}

	// find and validate our configuration filepath.
	if configFilepath == "" {
		if configFilepath = os.Getenv(configFilepathEnvVar); configFilepath == "" {
			logger.Fatal(errNoConfig)
		}
	}

	configBytes, err := os.ReadFile(configFilepath)
	if err != nil {
		logger.Fatal(err)
	}

	sm := initializeLocalSecretManager(ctx)

	var cfg *config.InstanceConfig
	if err = sm.Decrypt(ctx, string(configBytes), &cfg); err != nil || cfg == nil {
		logger.Fatal(err)
	}

	flushFunc, initializeTracerErr := cfg.Observability.Tracing.Initialize(logger)
	if initializeTracerErr != nil {
		logger.Error(initializeTracerErr, "initializing tracer")
	}

	// if tracing is disabled, this will be nil
	if flushFunc != nil {
		defer flushFunc()
	}

	// only allow initialization to take so long.
	ctx, cancel := context.WithTimeout(ctx, cfg.Server.StartupDeadline)
	ctx, initSpan := tracing.StartSpan(ctx)

	// build our server struct.
	srv, err := server.Build(ctx, cfg, logger)
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
