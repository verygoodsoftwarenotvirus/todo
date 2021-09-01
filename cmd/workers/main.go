package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsqio/go-nsq"
)

// MyHandler handles NSQ messages from the channel being subscribed to
type MyHandler struct {
}

func (h *MyHandler) HandleMessage(message *nsq.Message) error {
	log.Printf("Got a message: %s", string(message.Body))

	return nil
}

func main() {
	const (
		addr    = "nsqlookupd:4161"
		topic   = "todo-events"
		channel = "things.and.stuff"
	)

	// configure a new Consumer
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		log.Fatal(err)
	}

	// register our message handler with the consumer
	consumer.AddHandler(&MyHandler{})

	// connect to NSQ and start receiving messages
	//err = consumer.ConnectToNSQD("nsqd:4150")
	if err = consumer.ConnectToNSQLookupd(addr); err != nil {
		log.Fatal(err)
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	// disconnect
	consumer.Stop()
}
