package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/secrets"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/workers"

	"github.com/go-redis/redis/v8"
)

const (
	preWritesTopicName  = "pre_writes"
	postWritesTopicName = "post_writes"
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

	logger.Info("starting workers...")

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

	pcfg := &publishers.Config{
		Provider:     "redis",
		QueueAddress: addr,
	}

	publisherProvider, err := publishers.ProvidePublisherProvider(logger, pcfg)
	if err != nil {
		logger.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err = setupPostWritesWorker(ctx, logger, redisClient); err != nil {
		logger.Fatal(err)
	}

	if err = setupPreWritesWorker(ctx, logger, dataManager, redisClient, publisherProvider); err != nil {
		logger.Fatal(err)
	}

	logger.Info("working...")

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
}

func setupPreWritesWorker(ctx context.Context, logger logging.Logger, dataManager database.DataManager, redisClient *redis.Client, publisherProvider publishers.PublisherProvider) error {
	preWritesSubscription := redisClient.Subscribe(ctx, preWritesTopicName)

	postWritesPublisher, err := publisherProvider.ProviderPublisher(postWritesTopicName)
	if err != nil {
		return err
	}

	pww := workers.ProvidePreWritesWorker(logger, dataManager, postWritesPublisher)

	// Consume messages.
	go func() {
		for msg := range preWritesSubscription.Channel() {
			if err = pww.HandleMessage([]byte(msg.Payload)); err != nil {
				logger.Error(err, "handling pre-write message")
			}
		}
	}()

	return nil
}

func setupPostWritesWorker(ctx context.Context, logger logging.Logger, redisClient *redis.Client) error {
	preWritesSubscription := redisClient.Subscribe(ctx, "post_writes")

	// Wait for confirmation that subscription is created before publishing anything.
	if _, err := preWritesSubscription.Receive(ctx); err != nil {
		return err
	}

	pww := workers.ProvidePostWritesWorker(logger)

	// Consume messages.
	go func() {
		for msg := range preWritesSubscription.Channel() {
			if err := pww.HandleMessage([]byte(msg.Payload)); err != nil {
				logger.Error(err, "handling post-write message")
			}
		}
	}()

	return nil
}
