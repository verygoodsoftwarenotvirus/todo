package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/events"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/secrets"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/workers"

	"github.com/nsqio/go-nsq"
)

const (
	configFilepathEnvVar = "CONFIGURATION_FILEPATH"
	configStoreEnvVarKey = "TODO_WORKERS_LOCAL_CONFIG_STORE_KEY"
)

type nsqLogger struct {
	logger logging.Logger
}

func (l *nsqLogger) Output(calldepth int, s string) error {
	if !strings.Contains(s, "TOPIC_NOT_FOUND") &&
		!strings.Contains(s, "querying nsqlookupd") &&
		!strings.Contains(s, "retrying with next nsqlookupd") {
		l.logger.WithValue("calldepth", calldepth).Info(s)
	}

	return nil
}

func initializeLocalSecretManager(ctx context.Context) secrets.SecretManager {
	logger := logging.NewNoopLogger()

	cfg := &secrets.Config{
		Provider: secrets.ProviderLocal,
		Key:      os.Getenv(configStoreEnvVarKey),
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
	const (
		addr = "nsqlookupd:4161"
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

	sm := initializeLocalSecretManager(ctx)

	var cfg *config.InstanceConfig
	if err = sm.Decrypt(ctx, string(configBytes), &cfg); err != nil || cfg == nil {
		logger.Fatal(err)
	}

	cfg.Database.RunMigrations = false

	dataManager, err := config.ProvideDatabaseClient(ctx, logger, cfg)
	if err != nil {
		logger.Fatal(err)
	}

	pendingWorker := workers.ProvidePendingWriter(logger, dataManager)

	// configure a new Consumer
	pendingWritesConsumer, err := events.NewTopicConsumer(addr, "pending_writes", pendingWorker.HandlePendingWrite)
	if err != nil {
		logger.Fatal(err)
	}
	defer pendingWritesConsumer.Stop()

	pendingWritesConsumer.SetLogger(&nsqLogger{logger: logger}, nsq.LogLevelInfo)
	pendingWritesConsumer.SetLoggerLevel(nsq.LogLevelInfo)

	// configure a new Consumer
	writesConsumer, err := events.NewTopicConsumer(addr, "writes", func(message *nsq.Message) error {
		logger.WithName("writes_consumer").WithValue("message_body", string(message.Body)).Debug("Got a write message")
		return nil
	})
	if err != nil {
		logger.Fatal(err)
	}
	defer writesConsumer.Stop()

	writesConsumer.SetLogger(&nsqLogger{logger: logger}, nsq.LogLevelInfo)
	writesConsumer.SetLoggerLevel(nsq.LogLevelInfo)

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
}
