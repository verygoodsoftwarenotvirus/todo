package main

import (
	"context"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/workers"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/secrets"

	"github.com/go-redis/redis/v8"
)

func initializeLocalSecretManager(ctx context.Context, envVarKey string) secrets.SecretManager {
	logger := logging.NewNoopLogger()

	cfg := &secrets.Config{
		Provider: secrets.ProviderLocal,
		Key:      os.Getenv(envVarKey),
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

const (
	configFilepathEnvVar = "CONFIGURATION_FILEPATH"
	configStoreEnvVarKey = "TODO_WORKERS_LOCAL_CONFIG_STORE_KEY"
)

func main() {
	const (
		addr = "worker_queue:6379"
	)

	ctx := context.Background()

	logger := logging.ProvideLogger(logging.Config{
		Provider: logging.ProviderZerolog,
	})

	// find and validate our configuration filepath.
	configFilepath := os.Getenv(configFilepathEnvVar)
	if configFilepath == "" {
		log.Fatal("no config provided")
	}

	configBytes, err := os.ReadFile(configFilepath)
	if err != nil {
		logger.Fatal(err)
	}

	sm := initializeLocalSecretManager(ctx, configStoreEnvVarKey)

	var cfg *config.InstanceConfig
	if err = sm.Decrypt(ctx, string(configBytes), &cfg); err != nil || cfg == nil {
		logger.Fatal(err)
	}

	cfg.Database.RunMigrations = false

	dataManager, err := config.ProvideDatabaseClient(ctx, logger, cfg)
	if err != nil {
		logger.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pubsub := rdb.Subscribe(ctx, "pending_writes")

	// Wait for confirmation that subscription is created before publishing anything.
	if _, err = pubsub.Receive(ctx); err != nil {
		logger.Fatal(err)
	}

	pww := workers.ProvidePendingWritesWorker(logger, dataManager)

	// Consume messages.
	for msg := range pubsub.Channel() {
		if err = pww.HandleMessage([]byte(msg.Payload)); err != nil {

		}
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
}
