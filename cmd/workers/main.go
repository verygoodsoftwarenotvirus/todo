package main

import (
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/events"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"github.com/nsqio/go-nsq"
)

func main() {
	const (
		addr = "nsqlookupd:4161"
	)

	logger := logging.ProvideLogger(logging.Config{
		Provider: logging.ProviderZerolog,
	})

	// configure a new Consumer
	pendingWritesConsumer, err := events.NewTopicConsumer(addr, "pending_writes", func(message *nsq.Message) error {
		logger.WithName("pending_writes_consumer").WithValue("message_body", string(message.Body)).Debug("Got a write message")
		return nil
	})

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
