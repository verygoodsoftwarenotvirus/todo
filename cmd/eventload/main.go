// This example declares a durable Exchange, and publishes a single message to
// that Exchange with a given routing key.
//
package main

import (
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"log"
	"time"

	"github.com/streadway/amqp"
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

	channel, err := setupWriteChannel(logger, connection)
	if err != nil {
		logger.Fatal(err)
	}

	stopChan := make(chan bool)

	go func() {
		for range time.Tick(time.Second) {
			log.Printf("declared Exchange, publishing %dB body (%q)", len(body), body)
			if err = channel.Publish(
				exchange,   // publish to an exchange
				routingKey, // routing to 0 or more queues
				false,      // mandatory
				false,      // immediate
				amqp.Publishing{
					Headers:         amqp.Table{},
					ContentType:     "application/json",
					ContentEncoding: "",
					Body:            []byte(body),
					DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
					Priority:        0,              // 0-9
					// a bunch of application/implementation-specific fields
				},
			); err != nil {
				logger.Fatal(fmt.Errorf("Exchange Publish: %s", err))
			}
		}
	}()

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

func setupWriteChannel(logger logging.Logger, connection *amqp.Connection) (*amqp.Channel, error) {
	// This function dials, connects, declares, publishes, and tears down,
	// all in one go. In a real service, you probably want to maintain a
	// long-lived connection as state, and publish against that.

	log.Printf("got Connection, getting Channel")
	channel, err := connection.Channel()
	if err != nil {
		logger.Fatal(fmt.Errorf("Channel: %s", err))
	}

	log.Printf("got Channel, declaring %q Exchange (%q)", exchangeType, exchange)
	if err = channel.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		logger.Fatal(fmt.Errorf("Exchange Declare: %s", err))
	}

	q, err := channel.QueueDeclare(
		"things.and.stuff", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		logger.Fatal(fmt.Errorf("queue Declare: %s", err))
	}

	if err = channel.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil,
	); err != nil {
		logger.Fatal(fmt.Errorf("queue bind: %s", err))
	}

	// Reliable publisher confirms require confirm.select support from the  connection.
	if reliable {
		log.Printf("enabling publishing confirms.")
		if err = channel.Confirm(false); err != nil {
			logger.Fatal(fmt.Errorf("Channel could not be put into confirm mode: %s", err))
		}

		confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

		defer confirmOne(confirms)
	}

	return channel, nil
}
