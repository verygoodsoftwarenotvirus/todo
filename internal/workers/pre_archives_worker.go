package workers

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/messagequeue/publishers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

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

// HandleMessage handles a pending archive.
func (w *PreArchivesWorker) HandleMessage(ctx context.Context, message []byte) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	var msg *types.PreArchiveMessage

	if err := w.encoder.Unmarshal(ctx, message, &msg); err != nil {
		return observability.PrepareError(err, w.logger, span, "unmarshalling message")
	}

	tracing.AttachUserIDToSpan(span, msg.AttributableToUserID)

	w.logger.WithValue("data_type", msg.DataType).Debug("message read")

	switch msg.DataType {
	case types.ItemDataType:
		if err := w.dataManager.ArchiveItem(ctx, msg.RelevantID, msg.AttributableToAccountID); err != nil {
			return observability.PrepareError(err, w.logger, span, "creating item")
		}

		if w.postArchivesPublisher != nil {
			dcm := &types.DataChangeMessage{
				DataType:                msg.DataType,
				AttributableToUserID:    msg.AttributableToUserID,
				AttributableToAccountID: msg.AttributableToAccountID,
			}

			if err := w.postArchivesPublisher.Publish(ctx, dcm); err != nil {
				w.logger.Error(err, "publishing to post-archives topic")
			}
		}
	case types.WebhookDataType:
		if err := w.dataManager.ArchiveWebhook(ctx, msg.RelevantID, msg.AttributableToAccountID); err != nil {
			return observability.PrepareError(err, w.logger, span, "creating item")
		}

		if w.postArchivesPublisher != nil {
			dcm := &types.DataChangeMessage{
				DataType:                msg.DataType,
				AttributableToUserID:    msg.AttributableToUserID,
				AttributableToAccountID: msg.AttributableToAccountID,
			}

			if err := w.postArchivesPublisher.Publish(ctx, dcm); err != nil {
				w.logger.Error(err, "publishing to post-archives topic")
			}
		}
	case types.UserMembershipDataType:
		break
	}

	return nil
}
