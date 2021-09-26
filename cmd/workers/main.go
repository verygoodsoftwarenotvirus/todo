package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/workers"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	msgconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/consumers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/secrets"
)

const (
	preWritesTopicName   = "pre_writes"
	dataChangesTopicName = "data_changes"
	preUpdatesTopicName  = "pre_updates"
	preArchivesTopicName = "pre_archives"
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

	cfg.Observability.Tracing.Jaeger.ServiceName = "workers"

	flushFunc, initializeTracerErr := cfg.Observability.Tracing.Initialize(logger)
	if initializeTracerErr != nil {
		logger.Error(initializeTracerErr, "initializing tracer")
	}

	// if tracing is disabled, this will be nil
	if flushFunc != nil {
		defer flushFunc()
	}

	cfg.Database.RunMigrations = false

	dataManager, err := config.ProvideDatabaseClient(ctx, logger, cfg)
	if err != nil {
		logger.Fatal(err)
	}

	pcfg := &msgconfig.Config{
		Provider: msgconfig.ProviderRedis,
		RedisConfig: msgconfig.RedisConfig{
			QueueAddress: addr,
		},
	}

	consumerProvider := consumers.ProvideRedisConsumerProvider(logger, addr)

	publisherProvider, err := msgconfig.ProvidePublisherProvider(logger, pcfg)
	if err != nil {
		logger.Fatal(err)
	}

	// post-writes worker

	postWritesWorker := workers.ProvideDataChangesWorker(logger)
	postWritesConsumer, err := consumerProvider.ProviderConsumer(ctx, dataChangesTopicName, postWritesWorker.HandleMessage)
	if err != nil {
		logger.Fatal(err)
	}

	go postWritesConsumer.Consume(nil, nil)

	// pre-writes worker

	postWritesPublisher, err := publisherProvider.ProviderPublisher(dataChangesTopicName)
	if err != nil {
		logger.Fatal(err)
	}
	preWritesWorker := workers.ProvidePreWritesWorker(logger, dataManager, postWritesPublisher)

	preWritesConsumer, err := consumerProvider.ProviderConsumer(ctx, preWritesTopicName, preWritesWorker.HandleMessage)
	if err != nil {
		logger.Fatal(err)
	}

	go preWritesConsumer.Consume(nil, nil)
	// pre-updates worker

	postUpdatesPublisher, err := publisherProvider.ProviderPublisher(dataChangesTopicName)
	if err != nil {
		logger.Fatal(err)
	}

	preUpdatesWorker := workers.ProvidePreUpdatesWorker(logger, dataManager, postUpdatesPublisher)

	preUpdatesConsumer, err := consumerProvider.ProviderConsumer(ctx, preUpdatesTopicName, preUpdatesWorker.HandleMessage)
	if err != nil {
		logger.Fatal(err)
	}

	go preUpdatesConsumer.Consume(nil, nil)

	// pre-archives worker

	postArchivesPublisher, err := publisherProvider.ProviderPublisher(dataChangesTopicName)
	if err != nil {
		logger.Fatal(err)
	}

	preArchivesWorker := workers.ProvidePreArchivesWorker(logger, dataManager, postArchivesPublisher)

	preArchivesConsumer, err := consumerProvider.ProviderConsumer(ctx, preArchivesTopicName, preArchivesWorker.HandleMessage)
	if err != nil {
		logger.Fatal(err)
	}

	go preArchivesConsumer.Consume(nil, nil)

	logger.Info("working...")

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
}
