package main

import (
	"github.com/nsqio/go-nsq"
	"log"
	"time"
)

func main() {
	var (
		addr    = "events:4150"
		topic   = "todo-events"
		message = `{"things": "stuff"}`
	)

	// configure a new Producer
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(addr, config)
	if err != nil {
		log.Fatal(err)
	}

	for range time.Tick(time.Second) {
		log.Println("publishing to queue")
		// publish a message to the producer
		if err = producer.Publish(topic, []byte(message)); err != nil {
			log.Fatal(err)
		}
	}

	// disconnect
	producer.Stop()
}
