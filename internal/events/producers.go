package events

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/nsqio/go-nsq"
)

type TopicProducer struct {
	topic    string
	logger   logging.Logger
	tracer   tracing.Tracer
	producer *nsq.Producer
}

func (w *TopicProducer) Publish(message string) error {
	w.logger.Debug("publishing message")
	return w.producer.Publish(w.topic, []byte(message))
}

func (w *TopicProducer) Stop() {
	w.producer.Stop()
}

func NewTopicProducer(logger logging.Logger, addr, topic string) (*TopicProducer, error) {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(addr, config)
	if err != nil {
		return nil, err
	}

	tw := &TopicProducer{
		topic:    topic,
		logger:   logging.EnsureLogger(logger).WithValue("topic", topic),
		tracer:   tracing.NewTracer(fmt.Sprintf("%s_writer", topic)),
		producer: producer,
	}

	return tw, nil
}
