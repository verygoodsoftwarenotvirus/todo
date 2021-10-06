package httpclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// SubscribeToDataChangeNotifications subscribes to a websocket to receive DataChangeMessages.
func (c *Client) SubscribeToDataChangeNotifications(ctx context.Context, stopChan <-chan bool) (chan *types.DataChangeMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if stopChan == nil {
		stopChan = make(chan bool)
	}

	logger := c.logger
	uri := c.requestBuilder.BuildSubscribeToDataChangesWebsocketURL(ctx)

	header, err := c.authHeaderBuilder.BuildRequestHeaders(ctx)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "preparing websocket request headers")
	}

	conn, _, err := c.websocketDialer.DialContext(ctx, uri, header)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "dialing websocket")
	}

	dataChangeMessages := make(chan *types.DataChangeMessage)
	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				_, message, msgReadErr := conn.ReadMessage()
				if msgReadErr != nil {
					observability.AcknowledgeError(msgReadErr, logger, span, "receiving data change message")
					return
				}

				var msg *types.DataChangeMessage
				if unmarshalErr := c.encoder.Unmarshal(ctx, message, &msg); unmarshalErr != nil {
					observability.AcknowledgeError(unmarshalErr, logger, span, "decoding data change message")
					return
				}

				if msg != nil {
					dataChangeMessages <- msg
				}
			}
		}
	}()

	return dataChangeMessages, nil
}
