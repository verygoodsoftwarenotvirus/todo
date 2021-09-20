package workers

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// PostWriteMessage represents an event that asks a worker to write data to the datastore.
type PostWriteMessage struct {
	MessageType          string                           `json:"messageType"`
	Item                 *types.ItemDatabaseCreationInput `json:"item"`
	AttributableToUserID string                           `json:"userID"`
}

// PostWritesWorker writes data from the pending writes topic to the database.
type PostWritesWorker struct {
	logger  logging.Logger
	tracer  tracing.Tracer
	encoder encoding.ClientEncoder
}

// ProvidePostWritesWorker provides a PostWritesWorker.
func ProvidePostWritesWorker(logger logging.Logger) *PostWritesWorker {
	name := "post_writes"

	return &PostWritesWorker{
		logger:  logging.EnsureLogger(logger).WithName(name),
		tracer:  tracing.NewTracer(name),
		encoder: encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
	}
}

// HandleMessage handles a pending write.
func (w *PostWritesWorker) HandleMessage(message []byte) error {
	_, span := w.tracer.StartSpan(context.Background())
	defer span.End()

	w.logger.WithValue("message", message).Info("message received")

	return nil
}
