package workers

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

// PreArchiveMessage represents an event that asks a worker to archive data to the datastore.
type PreArchiveMessage struct {
	MessageType       string `json:"messageType"`
	RelevantID        string `json:"relevantID"`
	AccountID         string `json:"accountID"`
	AttributeToUserID string `json:"userID"`
}

// PreArchivesWorker archives data from the pending archives topic to the database.
type PreArchivesWorker struct {
	logger                logging.Logger
	tracer                tracing.Tracer
	encoder               encoding.ClientEncoder
	postArchivesPublisher publishers.Publisher
	dataManager           database.DataManager
}

// ProvidePreArchivesWorker provides a PreArchivesWorker.
func ProvidePreArchivesWorker(logger logging.Logger, dataManager database.DataManager, postArchivesPublisher publishers.Publisher) *PreArchivesWorker {
	const name = "pre_archives"

	return &PreArchivesWorker{
		logger:                logging.EnsureLogger(logger).WithName(name).WithValue("topic", name),
		tracer:                tracing.NewTracer(name),
		encoder:               encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON),
		postArchivesPublisher: postArchivesPublisher,
		dataManager:           dataManager,
	}
}

const (
	itemMessageType = "item"
)

// HandleMessage handles a pending archive.
func (w *PreArchivesWorker) HandleMessage(message []byte) error {
	ctx, span := w.tracer.StartSpan(context.Background())
	defer span.End()

	var msg *PreArchiveMessage

	if err := w.encoder.Unmarshal(ctx, message, &msg); err != nil {
		return observability.PrepareError(err, w.logger, span, "unmarshalling message")
	}

	tracing.AttachUserIDToSpan(span, msg.AttributeToUserID)

	w.logger.WithValue("message_type", msg.MessageType).Debug("message read")

	switch msg.MessageType {
	case itemMessageType:
		if err := w.dataManager.ArchiveItem(ctx, msg.RelevantID, msg.AccountID); err != nil {
			return observability.PrepareError(err, w.logger, span, "creating item")
		}

		if w.postArchivesPublisher != nil {
			if err := w.postArchivesPublisher.Publish(ctx, msg.RelevantID); err != nil {
				w.logger.Error(err, "publishing to post-archives topic")
			}
		}
	}

	return nil
}
