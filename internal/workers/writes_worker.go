package workers

import (
	"context"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/events"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/nsqio/go-nsq"
)

// WriteMessage represents an event that asks a worker to write data to the datastore.
type WriteMessage struct {
	MessageType          string                           `json:"messageType"`
	Item                 *types.ItemDatabaseCreationInput `json:"item"`
	AttributableToUserID string                           `json:"userID"`
}

// AfterWriteWorker writes data from the pending writes topic to the database.
type AfterWriteWorker struct {
	logger         logging.Logger
	tracer         tracing.Tracer
	errorsProducer events.Producer
	dataManager    database.DataManager
}

// ProvideAfterWriteWorker provides a AfterWriteWorker.
func ProvideAfterWriteWorker(logger logging.Logger, errorsProducer events.Producer) *AfterWriteWorker {
	name := "after_writes"

	return &AfterWriteWorker{
		logger:         logging.EnsureLogger(logger).WithName(name),
		tracer:         tracing.NewTracer(name),
		errorsProducer: errorsProducer,
	}
}

// HandleMessage handles a pending write.
func (w *AfterWriteWorker) HandleMessage(message *nsq.Message) error {
	_, span := w.tracer.StartSpan(context.Background())
	defer span.End()

	w.logger.WithValue("message", message.Body).Debug("message read")

	message.Finish()

	return nil
}
