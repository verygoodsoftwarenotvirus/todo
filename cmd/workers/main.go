// This example declares a durable Exchange, and publishes a single message to
// that Exchange with a given routing key.
//
package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"log"
)

const (
	amqpURI      = "amqp://guest:guest@events:5672/" // flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")"
	exchange     = "todo-service"                    // flag.String("exchange", "test-exchange", "Durable AMQP exchange name")"
	exchangeType = "topic"                           // flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")"
	routingKey   = "testing"                         // flag.String("key", "test-key", "AMQP routing key")"
	body         = `{"things":"stuff"}`              // flag.String("body", "foobar", "Body of message")"
	reliable     = false                             // flag.Bool("reliable", true, "Wait for the publisher confirmation before exiting")
)

func main() {
	logger := logging.ProvideLogger(logging.Config{Provider: logging.ProviderZerolog})

	log.Printf("dialing %q", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		logger.Fatal(fmt.Errorf("Dial: %s", err))
	}

	stopChan := make(chan bool)

	go setupReadChannel(logger, connection)

	<-stopChan
}

// One would typically keep a channel of publishings, a sequence number, and a
// set of unacknowledged sequence numbers and loop until the publishing channel
// is closed.
func confirmOne(confirms <-chan amqp.Confirmation) {
	log.Printf("waiting for confirmation of one publishing")

	if confirmed := <-confirms; confirmed.Ack {
		log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
	} else {
		log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func setupReadChannel(logger logging.Logger, connection *amqp.Connection) error {
	ch, err := connection.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"things.and.stuff", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name,   // queue
		"worker", // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			if err = d.Ack(true); err != nil {
				logger.Error(err, "acknowledging msg")
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	return nil
}
