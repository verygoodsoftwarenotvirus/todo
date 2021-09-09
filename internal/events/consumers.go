package events

import (
	"log"
	"time"

	"github.com/nsqio/go-nsq"
)

const (
	channel = "todo-service"
)

type obligatoryNSQHandler struct {
	f func(message *nsq.Message) error
}

// HandleMessage implements the message handler interface.
func (h *obligatoryNSQHandler) HandleMessage(message *nsq.Message) error {
	return h.f(message)
}

// NewTopicConsumer builds a new NSQ consumer for a given address and topic.
func NewTopicConsumer(addr, topic string, handler func(*nsq.Message) error) (*nsq.Consumer, error) {
	const (
		maxConnectionAttempts  = 50
		connectionWaitInterval = time.Second
	)

	// configure a new Consumer
	config := nsq.NewConfig()
	config.LookupdPollInterval = connectionWaitInterval
	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return nil, err
	}

	// register our message handler with the consumer
	consumer.AddHandler(&obligatoryNSQHandler{f: handler})

	// connect to NSQ and start receiving messages
	attempts := 0
	for {
		log.Println("attempting to connect")

		connectionErr := consumer.ConnectToNSQLookupd(addr)
		if connectionErr != nil {
			log.Println("error encountered, incrementing attempts and sleeping")

			attempts++
			if attempts == maxConnectionAttempts {
				return nil, connectionErr
			}
			time.Sleep(connectionWaitInterval)
		} else {
			log.Println("connection established")
			break
		}
	}

	return consumer, nil
}
