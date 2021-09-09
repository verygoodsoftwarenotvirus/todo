package main

import (
	"context"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	dbconfig "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/workers"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/events"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"github.com/nsqio/go-nsq"
)

const (
	devPostgresDBConnDetails = "postgres://dbuser:hunter2@database:5432/todo?sslmode=disable"
)

func main() {
	const (
		addr = "nsqlookupd:4161"
	)

	ctx := context.Background()

	logger := logging.ProvideLogger(logging.Config{
		Provider: logging.ProviderZerolog,
	})

	cfg := &config.InstanceConfig{
		Database: dbconfig.Config{
			Debug:                     true,
			RunMigrations:             false,
			MaxPingAttempts:           50,
			Provider:                  "postgres",
			ConnectionDetails:         devPostgresDBConnDetails,
			MetricsCollectionInterval: time.Second,
		},
	}

	db, err := dbconfig.ProvideDatabaseConnection(logger, &cfg.Database)
	if err != nil {
		logger.Fatal(err)
	}

	dataManager, err := config.ProvideDatabaseClient(ctx, logger, db, cfg)

	pendingWorker := workers.ProvidePendingWriter(logger, dataManager)

	// configure a new Consumer
	pendingWritesConsumer, err := events.NewTopicConsumer(addr, "pending_writes", pendingWorker.HandlePendingWrite)

	if err != nil {
		logger.Fatal(err)
	}
	defer pendingWritesConsumer.Stop()

	// configure a new Consumer
	writesConsumer, err := events.NewTopicConsumer(addr, "writes", func(message *nsq.Message) error {
		logger.WithName("writes_consumer").WithValue("message_body", string(message.Body)).Debug("Got a write message")
		return nil
	})
	if err != nil {
		logger.Fatal(err)
	}
	defer writesConsumer.Stop()

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
}
