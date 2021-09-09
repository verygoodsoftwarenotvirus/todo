package workers

import (
	"bytes"
	"context"
	"encoding/gob"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/nsqio/go-nsq"
)

const (
	name = "pending_writer"
)

// PendingWriteMessage represents an event that asks a worker to write data to the datastore.
type PendingWriteMessage struct {
	MessageType          string                   `json:"messageType"`
	Item                 *types.ItemCreationInput `json:"item"`
	AttributableToUserID string                   `json:"userID"`
}

type PendingWriter struct {
	logger      logging.Logger
	tracer      tracing.Tracer
	dataManager database.DataManager
}

func ProvidePendingWriter(logger logging.Logger, dataManager database.DataManager) *PendingWriter {
	return &PendingWriter{
		logger:      logging.EnsureLogger(logger).WithName(name),
		tracer:      tracing.NewTracer(name),
		dataManager: dataManager,
	}
}

func (w *PendingWriter) HandlePendingWrite(message *nsq.Message) error {
	ctx := context.Background()

	var msg *PendingWriteMessage

	if err := gob.NewDecoder(bytes.NewReader(message.Body)).Decode(&msg); err != nil {
		message.Touch()
		return err
	}

	//if err := json.Unmarshal(message.Body, &msg); err != nil {
	//	message.Touch()
	//	return err
	//}

	w.logger.WithValue("message_type", msg.MessageType).WithValue("item", msg.Item).Debug("message read")

	_, err := w.dataManager.CreateItem(ctx, msg.Item, msg.AttributableToUserID)
	if err != nil {
		message.Touch()
		return err
	}

	message.Finish()

	return nil
}
